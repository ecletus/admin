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

    let NAMESPACE = 'qor.single_edit',
        EVENT_ENABLE = 'enable.' + NAMESPACE,
        EVENT_DISABLE = 'disable.' + NAMESPACE,
        EVENT_CHANGE = 'change.' + NAMESPACE;

    function QorSingleEdit(element, options) {
        let $el = $(element);
        let value = $el.data('name');
        if (value) {
            this.$el = $el;
            this.$block = $el.find('.qor-field__block:eq(0)');
            this.$toggle = this.$el.find("[name='"+value+".@enabled']")
            this.init();
        } else {
            this.maker = null;
        }
    }

    QorSingleEdit.prototype = {
        constructor: QorSingleEdit,

        init: function() {
            this.bind();
            if (this.$toggle.length) {
                this.toggle()
            }
        },

        bind: function() {
            this.$toggle.on(EVENT_CHANGE, this.toggle.bind(this))
        },

        unbind: function() {
            this.$toggle.off(EVENT_CHANGE);
        },

        destroy: function() {
            this.unbind();
            this.$el.removeData(NAMESPACE);
        },

        toggle: function () {
            if (this.$toggle.is(':checked')) {
                this.$block.show()
            } else {
                this.$block.hide()
            }
        }
    };

    QorSingleEdit.DEFAULTS = {};

    QorSingleEdit.plugin = function(options) {
        return this.each(function() {
            let $this = $(this),
                data = $this.data(NAMESPACE),
                fn;

            if (!data) {
                if (/destroy/.test(options)) {
                    return;
                }
                data = new QorSingleEdit(this, options);
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
        let selector = '.single-edit',
            options = {};

        $(document)
            .on(EVENT_DISABLE, function(e) {
                QorSingleEdit.plugin.call($(selector, e.target), 'destroy');
            })
            .on(EVENT_ENABLE, function(e) {
                QorSingleEdit.plugin.call($(selector, e.target), options);
            })
            .triggerHandler(EVENT_ENABLE);
    });

    return QorSingleEdit;
});