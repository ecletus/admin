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
    $.fn.Ready = function (fn) {
        this.ready(fn);
        this.on("qor-ready", function (event, fragment) {
            fragment = fragment || this;
            fn.call(fragment)
        });
        return this;
    };
});