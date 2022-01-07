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
            this.successCallbacks = [];
        },

        onSuccess: function (cb) {
            this.successCallbacks.push(cb);
        },

        bind: function() {
            this.$el.bind(EVENT_SUBMIT, this.submit.bind(this))
        },

        unbind: function() {
            this.$el.off(EVENT_SUBMIT);
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
                    formData = new FormData(form),
                    cfg;


                if (e.originalEvent && e.originalEvent.constructor === SubmitEvent) {
                    const $submitter = $(e.originalEvent.submitter),
                        name = $submitter.attr('name'),
                        val = $submitter.attr('value');

                    if (name && val) {
                        formData.append(name, val);
                    }
                }

                QOR.prepareFormData(formData)

                if (continueEditing) {
                    action = action.replace(/([?|&]continue_editing)=true/, '$1_url=true')
                }

                cfg = {
                    continueEditing: continueEditing,
                    method: $form.prop('method'),
                    data: formData,
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
                        $form.parent().find('.qor-error').trigger('disable').remove();
                        $form.closest('.qor-page__body').children('#flashes,.qor-error').trigger('disable').remove();

                        if (jqXHR.getResponseHeader('X-Frame-Reload') === 'render-body') {
                            if ($form.closest('.qor-slideout').length) {
                                $form.closest('.qor-page__body').trigger('disable').html(html).trigger('enable');
                            } else {
                                const $error = $(html).find('.qor-error');
                                $form.before($error);

                                $error.find('> li > label').each(function () {
                                    let $label = $(this),
                                        id = $label.attr('for');

                                    if (id) {
                                        $form
                                            .find('#' + id)
                                            .closest('.qor-field')
                                            .addClass('is-error')
                                            .append($label.clone().addClass('qor-field__error'));
                                    }
                                });

                                const $messages = $form.closest('.qor-page__body').find('#flashes,.qor-error').eq(0);
                                if ($messages.length) {
                                    const $scroller = $messages.scrollParent();
                                    $scroller.scrollTop($messages.scrollTop() + $messages.offset().top)
                                }


                                html = $(html).find('.qor-page__body form:eq(0)').html();
                                $form.children().trigger('disable').remove();
                                $form.html(html).children().trigger('enable');
                            }
                            return
                        }

                        // handle file download from form submit
                        let disposition = jqXHR.getResponseHeader('Content-Disposition');
                        if (disposition && disposition.indexOf('attachment') !== -1) {
                            let fileNameRegex = /filename[^;=\n]*=((['"]).*?\2|[^;\n]*)/,
                                matches = fileNameRegex.exec(disposition),
                                contentType = jqXHR.getResponseHeader('Content-Type'),
                                fileName = '';

                            if (matches != null && matches[1]) {
                                fileName = matches[1].replace(/['"]/g, '');
                            }

                            window.QOR.qorAjaxHandleFile(action, contentType, fileName, formData);
                            $submit.prop('disabled', false);

                            return;
                        }

                        let xLocation = jqXHR.getResponseHeader('X-Location'),
                            wLocation = jqXHR.getResponseHeader('X-Location-Window');

                        if (wLocation) {
                            window.location.href = wLocation;
                            return
                        }

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
                            submitSuccess(html, statusText, jqXHR);
                            this.successCallbacks.forEach(cb => (cb(html, statusText, jqXHR)));
                            return;
                        }

                        let prefix = '/' + location.pathname.split('/')[1],
                            flashStructs = [];

                        $(html)
                            .find('.qor-alert')
                            .each(function (i, e) {
                                let message = $(e)
                                    .find('.qor-alert-message')
                                    .text()
                                    .trim(),
                                    type = $(e).data('type');
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

                        this.successCallbacks.forEach(cb => (cb(html, statusText, jqXHR)));
                    }.bind(this),
                    error: function (xhr, textStatus, errorThrown) {
                        $form.parent().find('.qor-error').remove();
                        $form.closest('.qor-page__body').children('#flashes,.qor-error').remove();

                        let $error;

                        if (xhr.status === 422) {
                            $form
                                .find('.qor-field')
                                .removeClass('is-error')
                                .find('.qor-field__error')
                                .remove();

                            $error = $(xhr.responseText).find('.qor-error');
                            $form.before($error);

                            $error.find('> li > label').each(function () {
                                let $label = $(this),
                                    id = $label.attr('for');

                                if (id) {
                                    $form
                                        .find('#' + id)
                                        .closest('.qor-field')
                                        .addClass('is-error')
                                        .append($label.clone().addClass('qor-field__error'));
                                }
                            });

                            const $messages = $form.closest('.qor-page__body').find('#flashes,.qor-error').eq(0);
                            if ($messages.length) {
                                const $scroller = $messages.scrollParent();
                                $scroller.scrollTop($messages.scrollTop() + $messages.offset().top)
                            }
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
            let $this = $(this),
                data = $this.data(NAMESPACE),
                fn = null;

            if (!$this.is('form')) {
                return
            }

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

    function accept($form) {
        return $form.length > 0 && $form.parents('.qor-slideout').length === 0;
    }

    $(function() {
        let selector = 'form[data-async="true"]',
            options = {};

        $(document)
            .on(EVENT_DISABLE, function(e) {
                let $form = $(selector, e.target);
                if (!accept($form)) {
                    return
                }
                QorAsyncFormSubmiter.plugin.call($form, 'destroy');
            })
            .on(EVENT_ENABLE, function(e) {
                let $form = $(selector, e.target);
                if (!accept($form)) {
                    return
                }
                QorAsyncFormSubmiter.plugin.call($form, options);
            })
            .triggerHandler(EVENT_ENABLE);
    });

    return QorAsyncFormSubmiter;
});