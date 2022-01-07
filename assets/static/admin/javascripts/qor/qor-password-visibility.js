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

    let NAMESPACE = 'qor.password_visibility',
        SELECTOR = '[data-toggle="qor.password_visibility"]',
        EVENT_ENABLE = 'enable.' + NAMESPACE,
        EVENT_DISABLE = 'disable.' + NAMESPACE,
        EVENT_CLICK = 'click.' + NAMESPACE;

    function QorPasswordVisibility(element, options) {
        this.$el = $(element);
        this.init();
    }

    QorPasswordVisibility.prototype = {
        constructor: QorPasswordVisibility,

        init: function () {
            this.flag = false;
            this.$icon = this.$el.find('i');
            this.$target = this.$el.parents('div:eq(0)').children('input[type=password]');
            this.icons = [this.$icon.text(), this.$el.data('toggleIcon')];
            this.bind();
        },

        bind: function () {
            this.$el.bind(EVENT_CLICK, this.toggle.bind(this));
        },

        toggle: function () {
            this.flag = !this.flag;
            this.$icon.html(this.icons[+this.flag]);
            this.$target.attr('type', this.flag?'text':'password');
        },

        destroy: function () {
            this.$el.off(EVENT_CLICK, this.toggle);
            this.$el.removeData(NAMESPACE);
        }
    };

    QorPasswordVisibility.DEFAULTS = {};

    QorPasswordVisibility.plugin = function (options) {
        let args = Array.prototype.slice.call(arguments, 1),
            result;

        return this.each(function () {
            var $this = $(this);
            var data = $this.data(NAMESPACE);
            var fn;

            if (!data) {
                if (/destroy/.test(options)) {
                    return;
                }
                data = new QorPasswordVisibility(this, options);
                if (("$el" in data)) {
                    $this.data(NAMESPACE, data);
                } else {
                    return
                }
            }

            if (typeof options === 'string' && $.isFunction((fn = data[options]))) {
                result = fn.apply(data, args);
            }
        });

        return (typeof result === 'undefined') ? this : result;
    };

    $(function () {
        var options = {};

        $(document)
            .on(EVENT_DISABLE, function (e) {
                QorPasswordVisibility.plugin.call($(SELECTOR, e.target), 'destroy');
            })
            .on(EVENT_ENABLE, function (e) {
                QorPasswordVisibility.plugin.call($(SELECTOR, e.target), options);
            })
            .triggerHandler(EVENT_ENABLE);
    });

    return QorPasswordVisibility;
});