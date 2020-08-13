(function(factory) {
    if (typeof define === 'function' && define.amd) {
        // AMD. Register as anonymous module.
        define(['jquery'], factory);
    } else if (typeof exports === 'object') {
        // Node / CommonJS
        factory(require('jquery'));
    } else {
        // Browser globals.
        factory(jQuery);
    }
})(function($) {
    'use strict';

    let location = window.location,
        NAMESPACE = 'qor.filter',
        EVENT_FILTER_CHANGE = 'filterChanged.' + NAMESPACE,
        EVENT_ENABLE = 'enable.' + NAMESPACE,
        EVENT_DISABLE = 'disable.' + NAMESPACE,
        EVENT_CLICK = 'click.' + NAMESPACE,
        EVENT_CHANGE = 'change.' + NAMESPACE,
        CLASS_IS_ACTIVE = 'is-active',
        CLASS_BOTTOMSHEETS = '.qor-bottomsheets';

    let re = /([^&=]+)(=([^&]*))?/g;
    let decodeRE = /\+/g;  // Regex for replacing addition symbol with a space

    function decode(str) {
        return decodeURIComponent( str.replace(decodeRE, " ") )
    }


    function decodeSearch(search) {
        let data = [];

        if (search && search.indexOf('?') > -1) {
            search = search.replace(/\+/g, ' ').split('?')[1];

            if (search && search.indexOf('#') > -1) {
                search = search.split('#')[0];
            }

            if (search) {
                // search = search.toLowerCase();
                data = $.map(search.split('&'), function(n) {
                    let param = [];
                    let value;

                    n = n.split('=');
                    if (/page/.test(n[0])) {
                        return;
                    }
                    value = n[1];
                    param.push(n[0]);

                    if (value) {
                        value = $.trim(decodeURIComponent(value));

                        if (value) {
                            param.push(value);
                        }
                    }

                    return param.join('=');
                });
            }
        }

        return data;
    }

    function parseParams(data) {
        let query = decodeURI(data === undefined ? location.search : data),
            params = {}, e, search;
        if (query && query[0] === '?') {
            query = query.substring(1)
        }
        while ( e = re.exec(query) ) {
            let k = decode( e[1] ), v = decode( e[3] );
            if (k.substring(k.length - 2) === '[]') {
                (params[k] || (params[k] = [])).push(v);
            }
            else params[k] = v;
        }
        return {
            params: params,
            isArray: function(key) {
                return key.substring(key.length - 2) === '[]'
            },
            remove: function (key) {
                delete (this.params[key])
            },
            set: function(key, value) {
                if (key.substring(key.length - 2) === '[]') {
                    this.params[key] = this.params[key] || []
                    this.params[key].push(value);
                }
                else this.params[key] = value;
            },
            removeAny: function (key, values) {
                if (!this.params.hasOwnProperty(key)) return;
                this.params[key] = this.params[key].filter(val => values.indexOf(val) < 0);
                if (!this.params[key].length)
                    this.remove(key)
            },
            removeItem: function (key, item) {
                if (!this.params.hasOwnProperty(key)) return;
                this.params[key] = this.params[key].filter(val => val !== item);
                if (!this.params[key].length)
                    this.remove(key)
            },
            encode: function () {
                const search = $.param(this.params);
                return search.length ? '?' + search : '';
            }
        }
    }

    function QorFilter(element, options) {
        this.$element = $(element);
        this.options = $.extend({}, QorFilter.DEFAULTS, $.isPlainObject(options) && options);
        this.init();
    }

    QorFilter.prototype = {
        constructor: QorFilter,

        init: function() {
            // this.parse();
            this.bind();
        },

        bind: function() {
            var options = this.options;

            this.$element
                .on(EVENT_CLICK, options.label, $.proxy(this.toggle, this))
                .on(EVENT_CHANGE, options.group, $.proxy(this.toggle, this));
        },

        unbind: function() {
            this.$element.off(EVENT_CLICK, this.toggle).off(EVENT_CHANGE, this.toggle);
        },

        toggle: function(e) {
            let $target = $(e.currentTarget),
                params = parseParams(),
                paramName,
                value,
                search;

            if ($target.is('select')) {
                paramName = $target.attr('name');
                value = $target.val();
                if (params.isArray(paramName)) {
                    let values = [];
                    $target.children().each((_, el) => values[values.length] = $(el).prop('value'));
                    params.removeAny(paramName, values);
                    if (value) {
                        params.set(paramName, value)
                    }
                } else {
                    if (value) params.set(paramName, value)
                    else params.remove(paramName)
                }
                search = params.encode()
            } else if ($target.is('a')) {
                e.preventDefault();
                let uri = $target.attr('href'),
                    pos = uri.indexOf('?');
                if (pos >= 0) {
                    search = uri.substring(0, pos)
                } else {
                    search = "?"
                }
            } else if ($target.is('input')) {
                paramName = $target.attr('name');
                value = $target.val();
                if (value)
                    params.set(paramName, value);
                else
                    params.remove(paramName);
                search = params.encode()
            }
            this.applySearch(search, paramName)
        },

        applySearch: function(search, paramName) {
            if (this.$element.closest(CLASS_BOTTOMSHEETS).length) {
                $(CLASS_BOTTOMSHEETS).trigger(EVENT_FILTER_CHANGE, [search, paramName]);
            } else {
                location.search = search;
            }
        },

        destroy: function() {
            this.unbind();
            this.$element.removeData(NAMESPACE);
        }
    };

    QorFilter.DEFAULTS = {
        label: false,
        group: false
    };

    QorFilter.plugin = function(options) {
        return this.each(function() {
            var $this = $(this);
            var data = $this.data(NAMESPACE);
            var fn;

            if (!data) {
                if (/destroy/.test(options)) {
                    return;
                }

                $this.data(NAMESPACE, (data = new QorFilter(this, options)));
            }

            if (typeof options === 'string' && $.isFunction((fn = data[options]))) {
                fn.apply(data);
            }
        });
    };

    $(function() {
        var selector = '[data-toggle="qor.filter"]';
        var options = {
            label: 'a',
            group: 'select,input'
        };

        $(document)
            .on(EVENT_DISABLE, function(e) {
                QorFilter.plugin.call($(selector, e.target), 'destroy');
            })
            .on(EVENT_ENABLE, function(e) {
                QorFilter.plugin.call($(selector, e.target), options);
            })
            .triggerHandler(EVENT_ENABLE);
    });

    return QorFilter;
});
