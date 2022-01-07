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

    const NAMESPACE = 'qor.dialog',
        EVENT_ENABLE = 'enable.' + NAMESPACE,
        EVENT_DISABLE = 'disable.' + NAMESPACE,
        EVENT_CLICK = 'click.' + NAMESPACE,
        SELECTOR_DIALOG = 'dialog.mdl-dialog',
        SELECTOR_DIALOG_SHOW_BTN = '[data-dialog]';

    function Dialog(el) {
        this.dialog = el;
        this.$el = $(el);
        this.init();
    }

    Dialog.prototype = {
        init: function () {
            if (!this.dialog.showModal) {
                dialogPolyfill.registerDialog(this.dialog);
            }
            this.$closers = this.$el.find('> .mdl-dialog__actions .close');
            this.bind();
        },

        bind: function () {
            this.$closers.on(EVENT_CLICK, this.hide.bind(this))
        },

        unbind: function () {
            if (this.$closers) this.$closers.off(EVENT_CLICK);
        },

        hide: function () {
            this.dialog.close()
        },

        show: function () {
            this.dialog.showModal()
        },

        destroy: function () {
            this.unbind();
            this.$el = this.$closers = this.dialog = null;
        }
    }

    Dialog.plugin = function (option) {
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
                $this.data(NAMESPACE, (data = new Dialog(this, options)));
            } else if (/destroy/.test(option)) {
                $this.removeData(NAMESPACE)
            }

            if (typeof option === 'string' && $.isFunction((fn = data[option]))) {
                fn.apply(data);
            }
        });
    };


    function OpenDialog(el, options) {
        this.$el = $(el);
        this.init();
    }

    OpenDialog.prototype = {
        init: function (options) {
            let dialog = options && options.target || this.$el.data().dialog,
                $dialog = dialog && $(dialog) || null;
            if (!$dialog) {
                return
            }
            this.$dialog = $dialog;
            this.bind();
        },

        bind: function () {
            this.$el.on(EVENT_CLICK, this.show.bind(this))
        },

        unbind: function () {
            this.$el.off(EVENT_CLICK);
        },

        show: function () {
            this.$dialog.data(NAMESPACE).show()
        },

        destroy: function () {
            this.unbind();
            this.$el = this.$dialog = null;
        }
    }

    OpenDialog.plugin = function (option) {
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
                $this.data(NAMESPACE, (data = new OpenDialog(this, options)));
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
                Dialog.plugin.call($(SELECTOR_DIALOG, e.target));
                OpenDialog.plugin.call($(SELECTOR_DIALOG_SHOW_BTN, e.target));
            })
            .on(EVENT_DISABLE, function (e) {
                Dialog.plugin.call($(SELECTOR_DIALOG, e.target), 'destroy');
                OpenDialog.plugin.call($(SELECTOR_DIALOG_SHOW_BTN, e.target), 'destroy');
            })
            .triggerHandler(EVENT_ENABLE)
    });
});
