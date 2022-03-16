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

    const NAMESPACE = 'qor.chooser',
        EVENT_ENABLE = 'enable.' + NAMESPACE,
        EVENT_DISABLE = 'disable.' + NAMESPACE;

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
                option = $.extend({
                    minimumResultsForSearch: 8,
                    dropdownParent: $this.parent()
                }, select2Data.select2 || {}),
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
                let url = select2Data.ajaxUrl || select2Data.originalAjaxUrl,
                    xurl = QOR.Xurl(url, $this),
                    primaryKey = select2Data.remoteDataPrimaryKey;

                delete select2Data["ajaxUrl"];
                $this.removeAttr('data-ajax-url');
                $this.attr('data-original-ajax-url', url);

                option.ajax.url = function (params) {
                    xurl.query.keyword = [params.term];
                    xurl.query.page = params.page;
                    xurl.query.per_page = 20;
                    let result = xurl.build();
                    if (!result.notFound.length && !result.empties.length) {
                        return result.url
                    }
                    return 'https://unsolvedy.dependency.localhost/' + result.notFound.concat(result.empties).toString()
                };

                let $field = $this.parents('.qor-field:eq(0)');
                this.$templateResult = $field.find('[name="select2-result-template"]');

                let renderTemplate = function ($tmpl, data) {
                    if (data.text && (data.loading || data.selected || data.id === "")) {
                        return data.text
                    }
                    if ($tmpl.length) {
                        if ($tmpl.data("raw")) {
                            let f = $tmpl.data("func");
                            if (!f) {
                                f = new Function("data", $tmpl.html());
                                $tmpl.data('func', f);
                            }
                            return f(data)
                        } else {
                            let tmpl = $tmpl.html().replace(/^\s+|\s+$/g, '').replace(/\[\[ *&amp;/g, '[[&'),
                                res = Mustache.render(tmpl, data);
                            return res;
                        }
                    }
                    return $.fn.select2.ajaxFormatResult(data, $tmpl);
                }.bind(this);

                option.templateResult = function(data) {
                    data.QorChooserOptions = dataOptions;
                    let text = renderTemplate(this.$templateResult, data).replace(/^\s+|\s+$/g, '');
                    return text;
                }.bind(this);

                this.$templateSelection = $field.find('[name="select2-selection-template"]');

                option.templateSelection = function(data) {
                    if (data.loading) return data.text;
                    data.QorChooserOptions = dataOptions;
                    if (data.element) {
                        if (primaryKey && data.hasOwnProperty(primaryKey)) {
                            $(data.element).attr('data-value', data[primaryKey]);
                        } else {
                            $(data.element).attr('data-value', data.id);
                        }
                    }
                    let text = renderTemplate(this.$templateSelection, data).replace(/^\s+|\s+$/g, '');
                    if (text === '') {
                        return data.text
                    }
                    return text
                }.bind(this);
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
            if (!this.$element) {
                this.destroy()
                return;
            }

            let $container, select2 = this.$element.data().select2;
            if (select2 && select2.$container) {
                $container = select2.$container;
                $container.width($container.parent().width());
            }

        },

        destroy: function() {
            if (this.$element) {
                const $el = this.$element;
                this.$element = null;
                $el.removeData(NAMESPACE);
                try {
                    $el.select2('destroy');
                } catch (e) {
                }
            }
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
