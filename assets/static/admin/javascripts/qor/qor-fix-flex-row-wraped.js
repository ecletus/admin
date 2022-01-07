(function (factory) {
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
})(function ($) {

    'use strict';

    const $window = $(window),
        NAMESPACE = 'qor.flex-row-wraped-fixer',
        EVENT_ENABLE = 'enable.' + NAMESPACE,
        EVENT_DISABLE = 'disable.' + NAMESPACE,
        EVENT_RESIZE = 'resize.' + NAMESPACE,
        EVENT_BEFORE_PRINT = 'beforeprint.' + NAMESPACE,
        EVENT_AFTER_PRINT = 'afterprint.' + NAMESPACE,
        SELECTOR = '.qor-flex-row-wrap';

    function Fixer(element, options) {
        this.$el = $(element);
        this.options = $.extend({}, Fixer.DEFAULTS, $.isPlainObject(options) && options);
        this.$children = this.$el.children();
        if (!this.$children.length) {
            return
        }
        this.init();
    }

    Fixer.prototype = {
        constructor: Fixer,

        init: function () {
            this.bind();
            this.fix();
        },

        bind: function () {
            $window.on(EVENT_RESIZE, this.fix.bind(this))
                .on(EVENT_AFTER_PRINT, this.afterPrint.bind(this))
                .on(EVENT_BEFORE_PRINT, this.beforePrint.bind(this));
        },

        unbind: function () {
            $window.off(EVENT_RESIZE, this.fix)
                .off(EVENT_AFTER_PRINT, this.afterPrint)
                .off(EVENT_BEFORE_PRINT, this.beforePrint);
        },

        beforePrint: function () {
            this.fix();
        },

        afterPrint: function () {
            this.fix()
        },

        fix: function () {
            const distance = this.$el.css('--item-distance'),
                $children = this.$children,
                dw = parseInt(this.$el.width());

            if (!distance) return;

            let row = [],
                rows = [row],
                rw = 0;

            $children.each(function () {
                $(this).children().css({marginLeft: 0, marginRight: 0, width: 'auto'});
            })

            $children.each(function (i) {
                let $el = $(this),
                    w = parseInt($el.width());
                if (rw > 0 && (rw + w) > dw) {
                    row = [];
                    rows[rows.length] = row
                    rw = w
                } else {
                    rw += w
                }
                row[row.length] = [i, w];
            });

            rows.forEach((row) => {
                row.forEach((el, i) => {
                    let $el = $($children[el[0]]),
                        rdw = $el.width();

                    if (row.length > 1) {
                        if (i === 0) {
                            $el.children().css({marginRight: distance / 2, width: rdw - distance / 2});
                        } else if (i === row.length - 1) {
                            $el.children().css({marginLeft: distance / 2, width: rdw - distance / 2});
                        } else {
                            $el.children().css({
                                marginLeft: distance / 2,
                                marginRight: distance / 2,
                                width: rdw - distance
                            })
                        }
                    }
                })
            })
        },

        destroy: function () {
            this.unbind();
            this.$el.removeData(NAMESPACE);
        }
    };

    Fixer.DEFAULTS = {
        header: false,
        content: false
    };

    Fixer.plugin = function (options) {
        return this.each(function () {
            var $this = $(this);
            var data = $this.data(NAMESPACE);
            var fn;

            if (!data) {
                $this.data(NAMESPACE, (data = new Fixer(this, options)));
            }

            if (typeof options === 'string' && $.isFunction(fn = data[options])) {
                fn.call(data);
            }
        });
    };

    return;

    $(function () {
        $(document).on(EVENT_ENABLE, function (e) {
            Fixer.plugin.call($(SELECTOR, e.target));
        }).on(EVENT_DISABLE, function (e) {
            Fixer.plugin.call($(SELECTOR, e.target), 'destroy');
        }).triggerHandler(EVENT_ENABLE);
    });

    return Fixer;

});