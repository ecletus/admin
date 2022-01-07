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

    function QorMark(key, $el) {
        if (key[0] === '#') {
            key = key.substr(1)
        }
        let pos = key.lastIndexOf(':');

        if (pos > 0) {
            let pk = key.substr(pos + 1),
                $target;

            key = key.substr(0, pos);

            $target = $el.find(`[data-mark="QorResource.${key}"][data-primary-key="${pk}"]:first`);
            $target.addClass('qor-marked').focus();
        }
    }

    const NAMESPACE = 'qor.marked',
        EVENT_ENABLE = 'enable.' + NAMESPACE,
        EVENT_DISABLE = 'disable.' + NAMESPACE;

    window.QOR.Mark = QorMark

    $(function () {
        $(document)
            .on(EVENT_ENABLE, function (e) {
                const $target = $(e.target);
                let urlS = $target.attr('data-src'),
                    url;
                if (urlS) {
                    if (urlS[0] === '/') {
                        urlS = location.protocol + "//" + location.host + urlS;
                    }
                    url = new URL(urlS);
                    if (url && url.hash) {
                        QorMark(url.hash, $target)
                    }
                }
                //QorAction.plugin.call($(selector, e.target), options);
            })
        if (location.hash !== '') {
            QorMark(location.hash, $(document))
        }
    });
})