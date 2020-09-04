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

    let NAMESPACE = 'qor.asyncFormSubmiter',
        EVENT_ENABLE = 'enable.' + NAMESPACE,
        EVENT_DISABLE = 'disable.' + NAMESPACE,
        EVENT_SUBMIT = 'submit.' + NAMESPACE;

    function QorAsyncFormSubmiter(element, options) {
        this.$el = $(element);
        this.options = $.extend({}, QorAsyncFormSubmiter.DEFAULTS, $.isPlainObject(options) && options);
        this.init();
    }

    QorAsyncFormSubmiter.prototype = {
        constructor: QorAsyncFormSubmiter,

        init: function() {
            this.bind();
        },

        bind: function() {
            this.$el.off(EVENT_SUBMIT, this.submit.bind(this));
            this.$el.bind(EVENT_SUBMIT, this.submit.bind(this))
        },

        unbind: function() {
            this.$el.off(EVENT_SUBMIT, this.submit.bind(this));
        },

        destroy: function() {
            this.unbind();
            this.$el.removeData(NAMESPACE);
        },

        updateOptions: function(options) {
            this.options = $.extend(this.options, options);
        },

        submit: function (e) {
            let form = e.target,
                $form = $(form),
                $submit = $form.find(':submit'),
                beforeSubmit = this.options.onBeforeSubmit,
                submitSuccess = this.options.onSubmitSuccess,
                openPage = this.options.openPage;

            if (window.FormData) {
                e.preventDefault();
                let action = $form.prop('action'),
                    continueEditing = /[?|&]continue_editing=true/.test(action),
                    cfg;

                if (continueEditing) {
                    action = action.replace(/([?|&]continue_editing)=true/, '$1_url=true')
                }

                cfg = {
                    continueEditing: continueEditing,
                    method: $form.prop('method'),
                    data: new FormData(form),
                    dataType: 'html',
                    processData: false,
                    contentType: false,
                    headers: {
                        'X-Layout': 'lite'
                    },
                    beforeSend: function (jqXHR, cfg) {
                        $submit.prop('disabled', true);
                    },
                    success: function (html, statusText, jqXHR) {
                        $form.parent().find('.qor-error').remove();

                        let xLocation = jqXHR.getResponseHeader('X-Location');

                        if (xLocation) {
                            if ($.isFunction(openPage)) {
                                openPage(xLocation)
                                return;
                            }
                            window.location.href = xLocation;
                            return
                        }

                        let returnUrl = $form.data('returnUrl'),
                            refreshUrl = $form.data('refreshUrl');

                        if (refreshUrl) {
                            window.location.href = refreshUrl;
                            return;
                        }

                        if (returnUrl) {
                            if ($.isFunction(openPage)) {
                                openPage(returnUrl)
                                return;
                            }
                            location.href = returnUrl
                            return;
                        }

                        if ($.isFunction(submitSuccess)) {
                            submitSuccess(html, statusText, jqXHR)
                            return;
                        }

                        var prefix = '/' + location.pathname.split('/')[1];
                        var flashStructs = [];
                        $(html)
                            .find('.qor-alert')
                            .each(function (i, e) {
                                var message = $(e)
                                    .find('.qor-alert-message')
                                    .text()
                                    .trim();
                                var type = $(e).data('type');
                                if (message !== '') {
                                    flashStructs.push({
                                        Type: type,
                                        Message: message,
                                        Keep: true
                                    });
                                }
                            });
                        if (flashStructs.length > 0) {
                            document.cookie = 'qor-flashes=' + btoa(unescape(encodeURIComponent(JSON.stringify(flashStructs)))) + '; path=' + prefix;
                        }
                    }.bind(this),
                    error: function (xhr, textStatus, errorThrown) {
                        $form.parent().find('.qor-error').remove();

                        var $error;

                        if (xhr.status === 422) {
                            $form
                                .find('.qor-field')
                                .removeClass('is-error')
                                .find('.qor-field__error')
                                .remove();

                            $error = $(xhr.responseText).find('.qor-error');
                            $form.before($error);

                            $error.find('> li > label').each(function () {
                                var $label = $(this);
                                var id = $label.attr('for');

                                if (id) {
                                    $form
                                        .find('#' + id)
                                        .closest('.qor-field')
                                        .addClass('is-error')
                                        .append($label.clone().addClass('qor-field__error'));
                                }
                            });

                            $('main').scrollTop($('main').scrollTop()+$form.siblings('.qor-error').offset().top)
                        } else {
                            QOR.ajaxError.apply(this, arguments)
                        }
                    }.bind(this),
                    complete: function () {
                        $submit.prop('disabled', false);
                    }
                };

                if ($.isFunction(beforeSubmit)) {
                    beforeSubmit(this, cfg);
                }

                $.ajax(action, cfg);
            }
        }
    };

    QorAsyncFormSubmiter.DEFAULTS = {};

    QorAsyncFormSubmiter.plugin = function(options) {
        let args = Array.prototype.slice.call(arguments, 1);
        return this.each(function() {
            var $this = $(this);
            var data = $this.data(NAMESPACE);
            var fn;

            if (!data) {
                if (/destroy/.test(options)) {
                    return;
                }

                if (typeof options === 'string') {
                    data = new QorAsyncFormSubmiter(this, {});
                } else {
                    data = new QorAsyncFormSubmiter(this, options);
                }

                $this.data(NAMESPACE, data);
            }

            if (typeof options === 'string' && $.isFunction((fn = data[options]))) {
                fn.apply(data, args);
            }
        });
    };

    $.fn.qorAsyncFormSubmiter = QorAsyncFormSubmiter.plugin;

    $(function() {
        var selector = 'form[data-async="true"]';
        var options = {};

        $(document)
            .on(EVENT_DISABLE, function(e) {
                let $form = $(selector, e.target);
                if (!$form.length) {
                    return
                }
                if ($form.parents('.qor-slideout').length) {
                    return
                }
                QorAsyncFormSubmiter.plugin.call($form, 'destroy');
            })
            .on(EVENT_ENABLE, function(e) {
                let $form = $(selector, e.target);
                if (!$form.length) {
                    return
                }
                if ($form.parents('.qor-slideout').length) {
                    return
                }
                QorAsyncFormSubmiter.plugin.call($form, options);
            })
            .triggerHandler(EVENT_ENABLE);
    });

    return QorAsyncFormSubmiter;
});