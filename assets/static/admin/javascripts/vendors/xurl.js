function Xurl(url, depGet) {
    /**
     * @function decodeParam
     *
     * Takes a string of name value pairs and returns a Object literal that represents those params.
     *
     * @param {String} params a string like <code>"foo=bar&person[age]=3&items[]=5&items[]=8"</code>
     * @return {Object} A JavaScript Object that represents the params:
     *
     *     {
     *       "foo": ["bar"],
     *       "person[age]": ["3"]
     *       "items[]": ["5", "8"]
     *     }
     */
    this.decodeParam = function (params) {
        if (!params) {
            return {};
        }

        let data = {},
            pairs = params.split('&'),
            current;

        for (let i = 0; i < pairs.length; i++) {
            current = data;
            let pos = pairs[i].indexOf('='),
                key = decodeURIComponent(pairs[i].substring(0, pos)),
                value = decodeURIComponent(pairs[i].substring(pos + 1));

            data[key] = data[key] || [];
            data[key][data[key].length] = value;
        }
        return data;
    };

    this.originalUrl = url;
    this.url = "";
    this.queryString = "";
    this.query = {};
    this.deps = {};
    this.fragment = "";

    this.dset = function (depName, value) {
        this.deps[depName] = value;
    };

    this.dget = function (depName) {
        if ((depName in this.deps))
            return this.deps[depName];
        return depGet(depName);
    };

    this.qset = function (queryName, value) {
        let values = [];
        for (i = 1; i < arguments.length; i++) {
            values[i - 1] = arguments[i]
        }
        this.query[queryName] = values
    };

    this.qget = function (queryName) {
        if ((queryName in this.query) && this.query[queryName].length > 0) {
            return this.query[queryName][0]
        }
        return undefined
    };

    this.qdel = function (queryName) {
        if ((queryName in this.query)) {
            delete this.query[queryName]
        }
    };

    this.qgetAll = function (queryName) {
        if ((queryName in this.query) && this.query[queryName].length > 0) {
            return this.query[queryName]
        }
        return undefined
    };

    this.toString = function () {
        return this.build().url
    };

    this.resolve = function (s, state) {
        const regex = /{[^}]+}/gm;
        let dget = this.dget.bind(this),
            m;

        while ((m = regex.exec(s)) !== null) {
            // This is necessary to avoid infinite loops with zero-width matches
            if (m.index === regex.lastIndex) {
                regex.lastIndex++;
            }

            m.forEach((match) => {
                let name = match.substring(1, match.length - 1),
                    required = name.substr(-1, 1) === "*",
                    value = null,
                    defaultValue = "!UNDEFINED!",
                    parts;

                if (required) {
                    name = name.substring(0, name.length-1)
                }

                parts = name.split("|")
                if (parts.length === 2) {
                    name = parts[0];
                    defaultValue = parts[1];
                }

                value = dget(name)

                if (value && value[1]) {
                    if (!value[0]) {
                        state.empties[name] = 1
                        value[0] = encodeURIComponent(defaultValue);
                    }
                    s = s.replace(match, value[0]);
                    s = s.replace(encodeURIComponent(match), encodeURIComponent(value[0]));
                } else {
                    state.notFound[name] = 1
                    if (!required) {
                        s = undefined
                    }
                }
            });
        }
        return s
    }

    this.build = function () {
        const state = {
            notFound: {},
            empties: {}
        };
        let url = this.resolve(this.url, state);
        if (this.query) {
            let query = [];

            for (let key in this.query) {
                for (let i in this.query[key]) {
                    if (this.query[key][i] !== undefined) {
                        query[query.length] = encodeURIComponent(this.resolve(key, state)) + "=" + encodeURIComponent(this.resolve(this.query[key][i], state));
                    }
                }
            }
            url += "?" + query.join("&");
        }
        if (this.fragment) {
            url += "#" + this.fragment
        }
        return {url:url, notFound:Object.keys(state.notFound), empties:Object.keys(state.empties)}
    };

    this.init = function () {
        if (!depGet) {
            depGet = function (name) {
                return [null, false]
            }
        }
        let pos = url.indexOf('#');
        this.url = url;

        if (pos !== -1) {
            this.fragment = this.url.substring(pos + 1);
            this.url = this.url.substring(0, pos);
        }

        pos = url.indexOf('?');
        if (pos !== -1) {
            this.queryString = this.url.substring(pos + 1);
            this.query = this.decodeParam(this.queryString);
            this.url = this.url.substring(0, pos);
        }
    };
    this.init();
}