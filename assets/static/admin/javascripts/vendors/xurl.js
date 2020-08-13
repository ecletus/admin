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

        var data = {},
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
    }

    this.build = function () {
        let url = this.url, notFound = {}, empties = {};
        if (this.query) {
            let query = [];

            for (let key in this.query) {
                for (let i in this.query[key]) {
                    if (this.query[key][i] !== undefined) {
                        query[query.length] = encodeURIComponent(key) + "=" + encodeURIComponent(this.query[key][i]);
                    }
                }
            }
            url += "?" + query.join("&");
        }
        if (this.fragment) {
            url += "#" + this.fragment
        }

        if (this.originalUrl.indexOf('{') !== -1) {
            const regex = /\{[^\}]+\}/gm;
            let dget = this.dget.bind(this),
                m;

            while ((m = regex.exec(this.originalUrl)) !== null) {
                // This is necessary to avoid infinite loops with zero-width matches
                if (m.index === regex.lastIndex) {
                    regex.lastIndex++;
                }

                m.forEach((match) => {
                    let name = match.substring(1, match.length - 1),
                        value = dget(name);
                    if (value[1]) {
                        if (!value[0]) {
                            empties[name] = 1
                        }
                        url = url.replace(match, value[0]);
                        url = url.replace(encodeURIComponent(match), encodeURIComponent(value[0]));
                    } else {
                        notFound[name] = 1
                    }
                });
            }
        }
        return {url:url, notFound:Object.keys(notFound), empties:Object.keys(empties)}
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