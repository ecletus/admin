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

    var NAMESPACE = 'qor.chooser';
    var EVENT_ENABLE = 'enable.' + NAMESPACE;
    var EVENT_DISABLE = 'disable.' + NAMESPACE;

    function QorChooser(element, options) {
        this.$element = $(element);
        this.options = $.extend({}, QorChooser.DEFAULTS, $.isPlainObject(options) && options);
        this.init();
    }

    QorChooser.prototype = {
        constructor: QorChooser,

        init: function() {
            let $this = this.$element,
                select2Data = $this.data(),
                resetSelect2Width,
                option = {
                    minimumResultsForSearch: 8,
                    dropdownParent: $this.parent()
                },
                dataOptions = {
                    displayKey: select2Data.remoteDataDisplayKey,
                    iconKey: select2Data.remoteDataIconKey,
                    getKey: function (data, key, defaul) {
                        if (key) {
                            let tmp = data, keys = key.split('.');
                            for (let i = 0; (typeof tmp !== 'undefined') && i < keys.length; i++) {
                                tmp = tmp[i]
                            }
                            if (typeof tmp !== 'undefined') {
                                return tmp;
                            }
                        }
                        return defaul
                    }
                };

            if (select2Data.remoteData) {
                option.ajax = $.fn.select2.ajaxCommonOptions(select2Data);
                let xurl = QOR.Xurl(select2Data["ajax-Url"], $this);

                delete select2Data["ajax-Url"];
                option.ajax.url = function (params) {
                    xurl.query.keyword = [params.term];
                    xurl.query.page = params.page;
                    xurl.query.per_page = 20;
                    let url = xurl.toString();
                    return url
                };

                option.templateResult = function(data) {
                    data.QorChooserOptions = dataOptions;
                    let tmpl = $this.parents('.qor-field').find('[name="select2-result-template"]');
                    if (tmpl.length > 0 && tmpl.data("raw")) {
                        var f = tmpl.data("func");
                        if (!f) {
                            f = new Function("data", tmpl.html());
                            tmpl.data('func', f);
                        }
                        return f(data);
                    }
                    return $.fn.select2.ajaxFormatResult(data, tmpl);
                };

                option.templateSelection = function(data) {
                    if (data.loading) return data.text;
                    data.QorChooserOptions = dataOptions;
                    let tmpl = $this.parents('.qor-field').find('[name="select2-selection-template"]');
                    if (tmpl.length > 0 && tmpl.data("raw")) {
                        var f = tmpl.data("func");
                        if (!f) {
                            f = new Function("data", tmpl.html());
                            tmpl.data('func', f);
                        }
                        return f(data)
                    }
                    return $.fn.select2.ajaxFormatResult(data, tmpl);
                };
            }

            $this.on('select2:select', function(evt) {
                $(evt.target).attr('chooser-selected', 'true');
            }).on('select2:unselect', function(evt) {
                $(evt.target).attr('chooser-selected', '');
            });

            $this.select2(option);

            // reset select2 container width
            this.resetSelect2Width();
            resetSelect2Width = window._.debounce(this.resetSelect2Width.bind(this), 300);
            $(window).resize(resetSelect2Width);

            if ($this.val()) {
                $this.attr('chooser-selected', 'true');
            }
        },

        resetSelect2Width: function() {
            var $container, select2 = this.$element.data().select2;
            if (select2 && select2.$container) {
                $container = select2.$container;
                $container.width($container.parent().width());
            }

        },

        destroy: function() {
            this.$element.select2('destroy').removeData(NAMESPACE);
        }
    };

    QorChooser.DEFAULTS = {};

    QorChooser.plugin = function(options) {
        return this.each(function() {
            var $this = $(this);
            var data = $this.data(NAMESPACE);
            var fn;

            if (!data) {

                if (/destroy/.test(options)) {
                    return;
                }

                $this.data(NAMESPACE, (data = new QorChooser(this, options)));
            }

            if (typeof options === 'string' && $.isFunction(fn = data[options])) {
                fn.apply(data);
            }
        });
    };

    $(function() {
        var selector = 'select[data-toggle="qor.chooser"]';

        $(document).
        on(EVENT_DISABLE, function(e) {
            QorChooser.plugin.call($(selector, e.target), 'destroy');
        }).
        on(EVENT_ENABLE, function(e) {
            QorChooser.plugin.call($(selector, e.target));
        }).
        triggerHandler(EVENT_ENABLE);
    });

    return QorChooser;

});
