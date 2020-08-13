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

    let NAMESPACE = 'qor.logo',
        SELECTOR = '.qor-logo[data-src]',
        EVENT_LOAD = 'load.'+NAMESPACE,
        EVENT_ERROR = 'error.'+NAMESPACE,
        EVENT_ENABLE = 'enable.' + NAMESPACE,
        EVENT_DISABLE = 'disable.' + NAMESPACE;

    function QorLogo(element, options) {
        this.$el = $(element);
        this.data = this.$el.data();
        this.$img = $("<img />");
        this.fallback = false;
        this.init();
    }

    QorLogo.prototype = {
        constructor: QorLogo,

        init: function () {
            this.build();
        },

        build: function () {
            this.$img
                .attr('src', this.data.src)
                .on(EVENT_LOAD, this.onload.bind(this))
                .on(EVENT_ERROR, this.onerror.bind(this));

            if (this.data.alt) {
                this.$img.attr('alt', this.data.alt)
            }
            if (this.data.title) {
                this.$img.attr('title', this.data.title)
            }

            this.$el.show().append(this.$img);
        },

        onerror: function(e) {
            if (this.fallback) {
                this.destroy();
                return
            }
            if (this.data.fallback) {
                this.fallback = true;
                this.$img.attr('src', this.data.fallback);
            } else {
                this.destroy();
            }
        },

        onload: function(e) {
            this.unbind();
        },

        unbind: function () {
            this.$img
                .off(EVENT_LOAD, this.onload)
                .off(EVENT_ERROR, this.onerror);
        },

        destroy: function () {
            this.unbind();
            this.$el.removeData(NAMESPACE);
            this.$el.hide();
            this.$img.remove();
        }
    };

    QorLogo.DEFAULTS = {};

    QorLogo.plugin = function (options) {
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
                data = new QorLogo(this, options);
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
                QorLogo.plugin.call($(SELECTOR, e.target), 'destroy');
            })
            .on(EVENT_ENABLE, function (e) {
                QorLogo.plugin.call($(SELECTOR, e.target), options);
            })
            .triggerHandler(EVENT_ENABLE);
    });

    return QorLogo;
});