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
        NAMESPACE = 'qor.content_copy',
        EVENT_ENABLE = 'enable.' + NAMESPACE,
        EVENT_DISABLE = 'disable.' + NAMESPACE,
        EVENT_CLICK = 'click.' + NAMESPACE;

    function QorContentCopy(element, options) {
        this.$element = $(element);
        this.options = $.extend({}, QorContentCopy.DEFAULTS, $.isPlainObject(options) && options);
        this.init();
    }

    QorContentCopy.prototype = {
        constructor: QorContentCopy,

        init: function() {
            this.bind();
        },

        bind: function() {
            var options = this.options;
            this.$element.attr('href', 'javascript:void(0);')
                .html('<i class="material-icons">content_copy</i>')
                .on(EVENT_CLICK, options.label, $.proxy(this.do, this));
        },

        unbind: function() {
            this.$element.off(EVENT_CLICK, this.do);
        },

        do: function(e) {
            let $this = $(e.currentTarget),
                value = $this.data('value');
            if (!value) {
                value = $this.data('value-b64');
                if (value) {
                    value = atob(value)
                }
            }
            if (!value) {
                let $el = $(this).parent().find("[data-content-copy-value]");
                if ($el.length === 0) {
                    return
                }
                value = $el.text()
            }
            let $temp = $("<input style='position: absolute; top: -200px'>");
            $("body").append($temp);
            $temp.val(value).select();
            document.execCommand("copy");
            $temp.remove();
        },

        destroy: function() {
            this.unbind();
            this.$element.removeData(NAMESPACE);
        }
    };

    QorContentCopy.DEFAULTS = {};

    QorContentCopy.plugin = function(options) {
        return this.each(function() {
            var $this = $(this);
            var data = $this.data(NAMESPACE);
            var fn;

            if (!data) {
                if (/destroy/.test(options)) {
                    return;
                }

                $this.data(NAMESPACE, (data = new QorContentCopy(this, options)));
            }

            if (typeof options === 'string' && $.isFunction((fn = data[options]))) {
                fn.apply(data);
            }
        });
    };

    $(function() {
        var selector = '[data-content-copy]';
        var options = {};

        $(document)
            .on(EVENT_DISABLE, function(e) {
                QorContentCopy.plugin.call($(selector, e.target), 'destroy');
            })
            .on(EVENT_ENABLE, function(e) {
                QorContentCopy.plugin.call($(selector, e.target), options);
            })
            .triggerHandler(EVENT_ENABLE);
    });

    return QorContentCopy;
});