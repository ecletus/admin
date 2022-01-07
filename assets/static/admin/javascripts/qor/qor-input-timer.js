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

    const NAMESPACE = 'qor.input_timer',
        EVENT_ENABLE = 'enable.' + NAMESPACE;

    $(function() {
        const selector = 'input.mdl-textfield__input[type="datetime-local"],input.mdl-textfield__input[type="date"]';

        $(document)
            .on(EVENT_ENABLE, function(e) {
                $(selector, e.target).each(function () {
                    const field = $(this).closest('.mdl-textfield');
                    field.length === 1 && new MaterialTextfield(field[0])
                })
            })
            .triggerHandler(EVENT_ENABLE);
    });
});