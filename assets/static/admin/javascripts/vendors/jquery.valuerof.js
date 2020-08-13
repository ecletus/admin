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

    // Register as jQuery plugin
    $.fn.valuerOf = function () {
        var $this = $(this[0]);
        let $form = $this.parents('form'),
            names = $this.attr('name').split('.').slice(0, -1);

        return function (name) {
            let l = names.length,
                discovery = name[0] === '*',
                field,
                value;
            if (discovery) {
                let tmpName;
                name = name.substring(1);
                do {
                    tmpName = l > 0 ? names.slice(0, l) + '.' + name : name;
                    field = $form.find(`[name='${tmpName}']:last`);
                    l--;
                } while (l >= 0 && field.length === 0);

                if (field.length === 0) {
                    if (name !== 'ID' || !$form.data('id')) {
                        value = '';
                    } else {
                        value = $form.data('id');
                    }
                } else {
                    value = field.val();
                }
            } else {
                while (name !== "" && name[0] === '.') {
                    l--;
                    name = name.substring(1);
                }
                name = names.slice(0, l).join('.') + '.' + name;
                field = $form.find(`[name='${name}']:last`);
                if (field.length === 0) {
                    return [null, false]
                }
                value = field.val();
            }
            return [value, true]
        }
    };
});