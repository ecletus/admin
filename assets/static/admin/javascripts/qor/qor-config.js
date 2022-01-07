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
    // init for slideout after show event
    $.fn.qorSliderAfterShow = $.fn.qorSliderAfterShow || {};

    // change Mustache tags from {{}} to [[]]
    window.Mustache && (window.Mustache.tags = ['[[', ']]']);

    // clear close alert after ajax complete
    $(document).ajaxComplete(function (event, xhr, settings) {
        if (settings.type === "POST" || settings.type === "PUT") {
            if ($.fn.qorSlideoutBeforeHide) {
                $.fn.qorSlideoutBeforeHide = null;
                window.onbeforeunload = null;
            }
        }
    });

    // select2 ajax common options
    $.fn.select2 = $.fn.select2 || function () {};
    $.fn.select2.ajaxCommonOptions = function (select2Data) {
        let remoteDataPrimaryKey = select2Data.remoteDataPrimaryKey,
            remoteDataDisplayKey = select2Data.remoteDataDisplayKey,
            remoteDataIconKey = select2Data.remoteDataIconKey,
            remoteDataCache = !(select2Data.remoteDataCache === 'false');

        return {
            dataType: 'json',
            cache: remoteDataCache,
            delay: 250,
            data: function (params) {
                return {
                    keyword: params.term || '', // search term
                    page: params.page,
                    per_page: 20
                };
            },
            processResults: function (data, params) {
                // parse the results into the format expected by Select2
                // since we are using custom formatting functions we do not need to
                // alter the remote JSON data, except to indicate that infinite
                // scrolling can be used
                params.page = params.page || 1;

                var processedData = $.map(data, function (obj) {
                    obj.id = obj[remoteDataPrimaryKey] || obj.primaryKey || obj.Id || obj.ID;
                    if (!obj.text) {
                        if (remoteDataDisplayKey) {
                            obj.text = obj[remoteDataDisplayKey];
                        } else if (!(obj.text = obj.text = obj.value || obj.Label || obj.Value)) {
                            let parts = [];
                            for (let key in obj) {
                                if (key.toLowerCase() !== "id" && obj[key]) {
                                    parts[parts.length] = obj[key]
                                }
                            }
                            obj.text = parts.join(" ");
                        }
                    }
                    if (remoteDataIconKey) {
                        obj.icon = obj[remoteDataIconKey];
                        if (obj.icon && /\.svg/.test(obj.icon)) {
                            obj.iconSVG = true;
                        }
                    }
                    return obj;
                });

                return {
                    results: processedData,
                    pagination: {
                        more: processedData.length >= 20
                    }
                };
            }
        };

    };

    // select2 ajax common options
    // format ajax template data
    $.fn.select2.ajaxFormatResult = function (data, tmpl) {
        var result = "";
        if (tmpl.length > 0) {
            result = window.Mustache.render(tmpl.html().replace(/{{(.*?)}}/g, '[[&$1]]'), data);
        } else {
            result = data.text || data.html || data.Name || data.Title || data.Code || data[Object.keys(data)[0]];
        }

        // if is HTML
        if (/<(.*)(\/>|<\/.+>)/.test(result)) {
            return $(result);
        }
        return result;
    };
});