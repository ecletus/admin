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

    const $window = $(window),
        _ = window._,
        NAMESPACE = 'qor.fixer',
        EVENT_ENABLE = 'enable.' + NAMESPACE,
        EVENT_DISABLE = 'disable.' + NAMESPACE;

    $(function() {
        $(document).
        on(EVENT_ENABLE, function(e) {
            $('.collection-edit-tabled', e.target).each(function (){
               const $el = $(this), $p2 = $el.parent().parent();
               if ($p2.is('.sec-col')) {
                   $p2.css({marginRight: 0, marginLeft: 0})
               }
            });
        }).
        on(EVENT_DISABLE, function(e) {
        }).
        triggerHandler(EVENT_ENABLE);
    });
});