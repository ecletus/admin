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

    const NAMESPACE = 'qor.layout_drawer_fixer',
        SELECTOR = '.mdl-layout__drawer .qor-layout__sidebar',
        EVENT_ENABLE = 'enable.'+NAMESPACE,
        EVENT_DISABLE = 'disable.'+NAMESPACE,
        EVENT_RESIZE = 'resize.' + NAMESPACE,
        $WINDOW = $(window);

    function QorLayoutDrawerFixer(element) {
        this.$el = $(element);
        this.init();
    }

    QorLayoutDrawerFixer.prototype = {
        init: function () {
            this.$header = this.$el.find('.sidebar-header')
            this.$body = this.$el.find('.sidebar-body')
            this.$footer = this.$el.find('.sidebar-footer')

            this.bind();
            this.resize();
        },

        bind: function () {
            $WINDOW.on(EVENT_RESIZE, this.resize.bind(this));
        },

        unbind: function() {
            $WINDOW.off(EVENT_RESIZE, this.resize)
        },

        destroy: function () {
            this.unbind();
            this.$el = this.$header = $this.$body = this.$footer = null;
        },

        resize : function () {
            this.$body.height(this.$el.height()-this.$header.height()-this.$footer.height());
            this.$body.css('top', this.$header.height());
        }
    }


    QorLayoutDrawerFixer.plugin = function (options) {
        let args = Array.prototype.slice.call(arguments, 1),
            result;

        this.each(function () {
            let $this = $(this),
                data = $this.data(NAMESPACE),
                fn;

            if (!data) {
                if (/destroy/.test(options)) {
                    return;
                }

                if (typeof options === "object")
                    options = $.extend({}, options, true)

                data = new QorLayoutDrawerFixer(this, options);
                $this.data(NAMESPACE, data)
            }

            if (typeof options === 'string' && $.isFunction((fn = data[options]))) {
                result = fn.apply(data, args);
            }
        });

        return (typeof result === 'undefined') ? this : result;
    };

    $(function () {
        let options = {};

        $(document)
            .on(EVENT_DISABLE, function (e) {
                QorLayoutDrawerFixer.plugin.call($(SELECTOR, e.target), 'destroy');
            })
            .on(EVENT_ENABLE, function (e) {
                QorLayoutDrawerFixer.plugin.call($(SELECTOR, e.target), options);
            })
            .triggerHandler(EVENT_ENABLE);
    });
});