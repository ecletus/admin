// init for slideout after show event
$.fn.qorSliderAfterShow = $.fn.qorSliderAfterShow || {};
window.QOR = {
    Xurl: function (url, $this) {
        let $form = $this.parents('form'),
            names = $this.attr('name').split('.').slice(0, -1),
            depGet = function (name) {
                let l = names.length,
                    discovery = name[0] === '*',
                    field,
                    value;
                if (discovery) {
                    let tmpName;
                    name = name.substring(1);
                    do {
                        tmpName = l > 0 ? names.slice(0, l) + '.' + name : name;
                        field = $form.find(`[name='${tmpName}']`);
                        l--;
                    } while(l >= 0 && field.length === 0);

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
                    field = $form.find(`[name='${name}']`);
                    if (field.length === 0) {
                        return [null, false]
                    }
                    value = field.val();
                }
                return [value, true]
            };

        return new Xurl(url, depGet);
    },

    submitContinueEditing: function (e) {
        let $form = $(e).parents('form'),
            action = $form.attr('action') || window.location.href;
        if (!/(\?|&)continue_editing=/.test(action)) {
            let param = 'continue_editing=true';
            if (action.indexOf('?') === -1) {
                action += '?' + param
            } else if (action[action.length-1] !== '?') {
                action += '&' + param
            } else {
                action += param
            }
        }
        $form.attr('action', action);
        $form.submit();
        return false;
    }
};

// change Mustache tags from {{}} to [[]]
window.Mustache && (window.Mustache.tags = ['[[', ']]']);

// clear close alert after ajax complete
$(document).ajaxComplete(function(event, xhr, settings) {
    if (settings.type === "POST" || settings.type === "PUT") {
        if ($.fn.qorSlideoutBeforeHide) {
            $.fn.qorSlideoutBeforeHide = null;
            window.onbeforeunload = null;
        }
    }
});

$(function () {
    let $header = $('.qor-page__header');
    if ($header.css('position') === 'fixed') {
        $('.qor-page__body').css({marginTop:$header.height(), paddingTop:0});
    }
});

// select2 ajax common options
$.fn.select2 = $.fn.select2 || function(){};
$.fn.select2.ajaxCommonOptions = function(select2Data) {
    let remoteDataPrimaryKey = select2Data.remoteDataPrimaryKey,
        remoteDataDisplayKey = select2Data.remoteDataDisplayKey,
        remoteDataIconKey = select2Data.remoteDataIconKey,
        remoteDataCache = !(select2Data.remoteDataCache === 'false');

    return {
        dataType: 'json',
        cache: remoteDataCache,
        delay: 250,
        data: function(params) {
            return {
                keyword: params.term || '', // search term
                page: params.page,
                per_page: 20
            };
        },
        processResults: function(data, params) {
            // parse the results into the format expected by Select2
            // since we are using custom formatting functions we do not need to
            // alter the remote JSON data, except to indicate that infinite
            // scrolling can be used
            params.page = params.page || 1;

            var processedData = $.map(data, function(obj) {
                obj.id = obj[remoteDataPrimaryKey] || obj.primaryKey || obj.Id || obj.ID;
                if (remoteDataDisplayKey) {
                    obj.text = obj[remoteDataDisplayKey];
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
$.fn.select2.ajaxFormatResult = function(data, tmpl) {
    var result = "";
    if (tmpl.length > 0) {
        result = window.Mustache.render(tmpl.html().replace(/{{(.*?)}}/g, '[[$1]]'), data);
    } else {
        result = data.text || data.html || data.Name || data.Title || data.Code || data[Object.keys(data)[0]];
    }

    // if is HTML
    if (/<(.*)(\/>|<\/.+>)/.test(result)) {
        return $(result);
    }
    return result;
};
