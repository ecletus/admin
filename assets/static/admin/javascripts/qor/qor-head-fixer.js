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

    var $window = $(window);
    var _ = window._;
    var NAMESPACE = 'qor.head-fixer';
    var EVENT_ENABLE = 'enable.' + NAMESPACE;
    var EVENT_DISABLE = 'disable.' + NAMESPACE;
    var EVENT_RESIZE = 'resize.' + NAMESPACE;
    var EVENT_BEFORE_PRINT = 'beforeprint.' + NAMESPACE;
    var EVENT_AFTER_PRINT = 'afterprint.' + NAMESPACE;
    var CLASS_HEADER = '.qor-page__header';
    var CLASS_BODY = '.qor-page__body';

    function QorHeadFixer(element, options) {
        this.$element = $(element);
        this.options = $.extend({}, QorHeadFixer.DEFAULTS, $.isPlainObject(options) && options);
        this.$clone = null;
        this.init();
    }

    QorHeadFixer.prototype = {
        constructor: QorHeadFixer,

        init: function() {
            this.bind();
            this.fix();
        },

        bind: function() {
            $window.on(EVENT_RESIZE, this.fix.bind(this))
                .on(EVENT_AFTER_PRINT, this.afterPrint.bind(this))
                .on(EVENT_BEFORE_PRINT, this.beforePrint.bind(this));
        },

        unbind: function() {
            $window.off(EVENT_RESIZE, this.fix)
                .off(EVENT_AFTER_PRINT, this.afterPrint)
                .off(EVENT_BEFORE_PRINT, this.beforePrint);
        },

        beforePrint: function() {
            $(CLASS_BODY).each(function () {
                $(this).removeAttr('style');
            })
        },

        afterPrint: function() {
            this.fix()
        },

        fix: function() {
            $(CLASS_BODY).each(function () {
                let $this = $(this),
                    $header = $this.siblings(CLASS_HEADER);
                if ($header.length === 0) return;
                $this.css('paddingTop', 0);
                if ($header.css('position') !== 'fixed') {
                    $this.css('marginTop', 0);
                } else {
                    $this.css('marginTop', $header.height());
                }
                if ($header.children(':visible').length) {
                    $header.removeClass('no-visibile-items', true).show()
                } else {
                    $header.addClass('no-visibile-items', true).hide()
                }
            })
        },

        destroy: function() {
            this.unbind();
            this.$element.removeData(NAMESPACE);
        }
    };

    QorHeadFixer.DEFAULTS = {
        header: false,
        content: false
    };

    QorHeadFixer.plugin = function(options) {
        return this.each(function() {
            var $this = $(this);
            var data = $this.data(NAMESPACE);
            var fn;

            if (!data) {
                $this.data(NAMESPACE, (data = new QorHeadFixer(this, options)));
            }

            if (typeof options === 'string' && $.isFunction(fn = data[options])) {
                fn.call(data);
            }
        });
    };


    $('.qor-page > .qor-page__header').each(function (){
        const $thead = $(this).siblings('.qor-page__body').find('> .qor-table-container > table > thead');
        if (!$thead.length) return;

        const resize_ob = new ResizeObserver(function(entries) {
            // since we are observing only a single element, so we access the first element in entries array
            let rect = entries[0].contentRect;

            // current width & height
            let width = rect.width;
            let height = rect.height;
            $thead.css({top:rect.height})
        });
        resize_ob.observe(this);
    })

// start observing for resize

    return;

    $(function() {
        if (/[?&]prin(t&|t$)/.test(location.search)) {
            return
        }
        var selector = '.qor-js-table';
        var options = {
            header: '.mdl-layout__header',
            subHeader: '.qor-page__header',
            content: '.mdl-layout__content',
            paddingHeight: 2 // Fix sub header height bug
        };

        $(document).
        on(EVENT_DISABLE, function(e) {
            QorHeadFixer.plugin.call($(e.target), 'destroy');
        }).
        on(EVENT_ENABLE, function(e) {
            QorHeadFixer.plugin.call($(e.target), options);
        }).
        triggerHandler(EVENT_ENABLE);
    });

    return QorHeadFixer;

});