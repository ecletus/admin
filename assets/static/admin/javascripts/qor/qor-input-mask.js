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

    let NAMESPACE = 'qor.input_mask',
        EVENT_ENABLE = 'enable.' + NAMESPACE,
        EVENT_DISABLE = 'disable.' + NAMESPACE;

    function QorInputMask(element, options) {
        let $el = $(element);
        let value = $el.data('masker');
        if (value) {
            this.$el = $el;
            this.masker = atob(value);
            this.options = $.extend({}, QorInputMask.DEFAULTS, $.isPlainObject(options) && options);
            this.init();
        } else {
            this.maker = null;
        }
    }

    QorInputMask.prototype = {
        constructor: QorInputMask,

        init: function() {
            this.bind();
        },

        bind: function() {
            (function (masker) {
                eval(masker)
            }).call(this.$el, this.masker);
        },

        unbind: function() {
            this.$el.unmask();
        },

        destroy: function() {
            this.unbind();
            this.$el.removeData(NAMESPACE);
        }
    };

    QorInputMask.DEFAULTS = {};

    QorInputMask.plugin = function(options) {
        return this.each(function() {
            var $this = $(this);
            var data = $this.data(NAMESPACE);
            var fn;

            if (!data) {
                if (/destroy/.test(options)) {
                    return;
                }
                data = new QorInputMask(this, options);
                if (("masker" in data)) {
                    $this.data(NAMESPACE, data);
                } else {
                    return
                }
            }

            if (typeof options === 'string' && $.isFunction((fn = data[options]))) {
                fn.apply(data);
            }
        });
    };

    $(function() {
        var selector = '[data-masker]';
        var options = {};

        $(document)
            .on(EVENT_DISABLE, function(e) {
                QorInputMask.plugin.call($(selector, e.target), 'destroy');
            })
            .on(EVENT_ENABLE, function(e) {
                QorInputMask.plugin.call($(selector, e.target), options);
            })
            .triggerHandler(EVENT_ENABLE);
    });

    return QorInputMask;
});