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

    const NAMESPACE = 'qor.table',
        EVENT_ENABLE = 'enable.' + NAMESPACE,
        EVENT_DISABLE = 'disable.' + NAMESPACE,
        SELECTOR = '.qor-table';


    function QorTable(el, options) {
        this.$el = $(el);
        this.init();
    }

    QorTable.prototype = {
        init: function (options) {
            this.$el.resizableColumns();
        },

        bind: function () {
        },

        unbind: function () {
        },

        destroy: function () {
            this.unbind();
            this.$el = null;
        }
    }

    QorTable.plugin = function (option) {
        return this.each(function () {
            const $this = $(this);
            let data = $this.data(NAMESPACE),
                options,
                fn;

            if (!data) {
                if (/destroy/.test(option)) {
                    return;
                }

                options = $.extend(true, {}, $this.data(), typeof option === 'object' && option);
                $this.data(NAMESPACE, (data = new QorTable(this, options)));
            } else if (/destroy/.test(option)) {
                $this.removeData(NAMESPACE)
            }

            if (typeof option === 'string' && $.isFunction((fn = data[option]))) {
                fn.apply(data);
            }
        });
    };

    $(function () {
        $(document)
            .on(EVENT_ENABLE, function (e) {
                QorTable.plugin.call($(SELECTOR, e.target));
            })
            .on(EVENT_DISABLE, function (e) {
                QorTable.plugin.call($(SELECTOR, e.target), 'destroy');
            })
            .triggerHandler(EVENT_ENABLE)
    });
});
