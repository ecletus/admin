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

    let NAMESPACE = 'qor.input_money',
        SELECTOR = 'input.input-money',
        EVENT_ENABLE = 'enable.' + NAMESPACE,
        EVENT_DISABLE = 'disable.' + NAMESPACE,
        EVENT_CHANGE = 'change.' + NAMESPACE;

    function QorInputMoney(element, options) {
        this.$el = $(element);
        let $p = this.$el.closest('.mdl-js-textfield');
        this.MDL = $p.length ? $p[0].MaterialTextfield : null;
        this.init();
    }

    QorInputMoney.prototype = {
        constructor: QorInputMoney,

        init: function () {
            this.bind();
        },

        bind: function () {
            this.$el.maskMoney();
            this.$el.on(EVENT_CHANGE, this.changed.bind(this))
        },

        destroy: function () {
            this.$el.maskMoney('destroy');
            this.$el.removeData(NAMESPACE);
            this.$el.off(EVENT_CHANGE, this.changed);
            this.MDL = null;
        },

        changed: function (e) {
            if (!this.MDL) {
                let $p = this.$el.closest('.mdl-js-textfield');
                this.MDL = $p.length ? $p[0].MaterialTextfield : null;
            }
            if (this.MDL) {
                this.MDL.checkDirty()
            }
        }
    };

    QorInputMoney.DEFAULTS = {};

    QorInputMoney.plugin = function (options) {
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
                data = new QorInputMoney(this, options);
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
                QorInputMoney.plugin.call($(SELECTOR, e.target), 'destroy');
            })
            .on(EVENT_ENABLE, function (e) {
                QorInputMoney.plugin.call($(SELECTOR, e.target), options);
            })
            .triggerHandler(EVENT_ENABLE);
    });

    return QorInputMoney;
});