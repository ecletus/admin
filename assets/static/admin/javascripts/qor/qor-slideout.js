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

    let $document = $(document),
        FormData = window.FormData,
        _ = window._,
        NAMESPACE = 'qor.slideout',
        EVENT_KEYUP = 'keyup.' + NAMESPACE,
        EVENT_CLICK = 'click.' + NAMESPACE,
        EVENT_SUBMIT = 'submit.' + NAMESPACE,
        EVENT_SHOW = 'show.' + NAMESPACE,
        EVENT_SLIDEOUT_SUBMIT_COMPLEMENT = 'slideoutSubmitComplete.' + NAMESPACE,
        EVENT_SLIDEOUT_CLOSED = 'slideoutClosed.' + NAMESPACE,
        EVENT_SLIDEOUT_LOADED = 'slideoutLoaded.' + NAMESPACE,
        EVENT_SLIDEOUT_BEFORESEND = 'slideoutBeforeSend.' + NAMESPACE,
        EVENT_SHOWN = 'shown.' + NAMESPACE,
        EVENT_HIDE = 'hide.' + NAMESPACE,
        EVENT_HIDDEN = 'hidden.' + NAMESPACE,
        EVENT_TRANSITIONEND = 'transitionend',
        CLASS_OPEN = 'qor-slideout-open',
        CLASS_MINI = 'qor-slideout-mini',
        CLASS_IS_SHOWN = 'is-shown',
        CLASS_IS_SLIDED = 'is-slided',
        CLASS_IS_SELECTED = 'is-selected',
        CLASS_MAIN_CONTENT = '.mdl-layout__content.qor-page',
        CLASS_HEADER_LOCALE = '.qor-actions__locale',
        CLASS_BODY_LOADING = '.qor-body__loading',
        CLASS_ACTION_BUTTON = '.qor-action-button',
        CLASS_FOOTER_COPYRIGTH = '.qor-page__footer_copy-rigth';

    function replaceHtml(el, html) {
        let oldEl = typeof el === 'string' ? document.getElementById(el) : el,
            newEl = oldEl.cloneNode(false);
        newEl.innerHTML = html;
        oldEl.parentNode.replaceChild(newEl, oldEl);
        return newEl;
    }

    function pushArrary($ele, isScript) {
        let array = [],
            prop = 'href';

        isScript && (prop = 'src');
        $ele.each(function () {
            array.push($(this).attr(prop));
        });
        return _.uniq(array);
    }

    function execSlideoutEvents(url, response) {
        // exec qorSliderAfterShow after script loaded
        var qorSliderAfterShow = $.fn.qorSliderAfterShow;
        for (var name in qorSliderAfterShow) {
            if (qorSliderAfterShow.hasOwnProperty(name) && !qorSliderAfterShow[name]['isLoaded']) {
                qorSliderAfterShow[name]['isLoaded'] = true;
                qorSliderAfterShow[name].call(this, url, response);
            }
        }
    }

    function loadScripts(srcs, data, callback) {
        let scriptsLoaded = 0;

        for (let i = 0, len = srcs.length; i < len; i++) {
            let script = document.createElement('script');

            script.onload = function () {
                scriptsLoaded++;

                if (scriptsLoaded === srcs.length) {
                    if ($.isFunction(callback)) {
                        callback();
                    }
                }

                if (data && data.url && data.response) {
                    execSlideoutEvents(data.url, data.response);
                }
            };

            script.src = srcs[i];
            document.body.appendChild(script);
        }
    }

    function loadStyles(srcs) {
        let ss = document.createElement('link'),
            src = srcs.shift();

        ss.type = 'text/css';
        ss.rel = 'stylesheet';
        ss.onload = function () {
            if (srcs.length) {
                loadStyles(srcs);
            }
        };
        ss.href = src;
        document.getElementsByTagName('head')[0].appendChild(ss);
    }

    function compareScripts($scripts) {
        let $currentPageScripts = $('script'),
            slideoutScripts = pushArrary($scripts, true),
            currentPageScripts = pushArrary($currentPageScripts, true),
            scriptDiff = _.difference(slideoutScripts, currentPageScripts);
        return scriptDiff;
    }

    function compareLinks($links) {
        let $currentStyles = $('link'),
            slideoutStyles = pushArrary($links),
            currentStyles = pushArrary($currentStyles),
            styleDiff = _.difference(slideoutStyles, currentStyles);

        return styleDiff;
    }

    function QorSlideout(element, options) {
        this.$element = $(element);
        this.options = $.extend({}, QorSlideout.DEFAULTS, $.isPlainObject(options) && options);
        this.options.language = this.options.language || this.$element.attr("lang") || $('html').attr('lang');
        this.slided = false;
        this.disabled = false;
        this.slideoutType = false;
        this.init();
    }

    QorSlideout.prototype = {
        constructor: QorSlideout,

        init: function () {
            this.build();
            this.bind();
        },

        build: function () {
            var $slideout;

            this.$slideout = $slideout = $(QorSlideout.TEMPLATE).appendTo('body');
            this.$slideoutTemplate = $slideout.html();
        },

        unbuild: function () {
            this.$slideout.remove();
        },

        bind: function () {
            this.$slideout
                .on(EVENT_SUBMIT, 'form', this.submit.bind(this))
                .on(EVENT_CLICK, '.qor-slideout__fullscreen', this.toggleSlideoutMode.bind(this))
                .on(EVENT_CLICK, '[data-dismiss="slideout"]', this.closeSlideout.bind(this));

            $document.on(EVENT_KEYUP, $.proxy(this.keyup, this));
        },

        unbind: function () {
            this.$slideout.off(EVENT_SUBMIT, this.submit).off(EVENT_CLICK);

            $document.off(EVENT_KEYUP, this.keyup);
        },

        keyup: function (e) {
            if (e.which === 27) {
                if ($('.qor-bottomsheets').is(':visible') || $('.qor-modal').is(':visible') || $('#redactor-modal-box').length || $('#dialog').is(':visible')) {
                    return;
                }

                this.hide();
                this.removeSelectedClass();
            }
        },

        loadExtraResource: function (data) {
            let styleDiff = compareLinks(data.$links),
                scriptDiff = compareScripts(data.$scripts);

            if (styleDiff.length) {
                loadStyles(styleDiff);
            }

            if (scriptDiff.length) {
                loadScripts(scriptDiff, data);
            }
        },

        removeSelectedClass: function () {
            this.$element.find('[data-url]').removeClass(CLASS_IS_SELECTED);
        },

        addLoading: function () {
            $(CLASS_BODY_LOADING).remove();
            var $loading = $(QorSlideout.TEMPLATE_LOADING);
            $loading.appendTo($('body')).trigger('enable');
        },

        toggleSlideoutMode: function () {
            this.$slideout
                .toggleClass('qor-slideout__fullscreen')
                .find('.qor-slideout__fullscreen i')
                .toggle();
        },

        submit: function (e) {
            var $slideout = this.$slideout;
            var $body = this.$body;
            var form = e.target;
            var $form = $(form);
            var _this = this;
            var $submit = $form.find(':submit');

            $slideout.trigger(EVENT_SLIDEOUT_BEFORESEND);

            if (FormData) {
                e.preventDefault();
                let action = $form.prop('action'),
                    continueEditing = /[?|&]continue_editing=true/.test(action),
                    headers = {
                        'X-Layout': 'lite'
                    };

                if (continueEditing) {
                    action = action.replace(/([?|&]continue_editing)=true/, '$1_url=true')
                } else {
                    headers["X-Redirection-Disabled"] = "true"
                }
                $.ajax(action, {
                    method: $form.prop('method'),
                    data: new FormData(form),
                    dataType: 'html',
                    processData: false,
                    contentType: false,
                    headers: headers,
                    beforeSend: function () {
                        $submit.prop('disabled', true);
                        $.fn.qorSlideoutBeforeHide = null;
                    },
                    success: function (html, statusText, jqXHR) {
                        $form.parent().find('.qor-error').remove();
                        $slideout.trigger(EVENT_SLIDEOUT_SUBMIT_COMPLEMENT);
                        let xLocation = jqXHR.getResponseHeader('X-Location');

                        if (xLocation) {
                            if (e.originalEvent && e.originalEvent.explicitOriginalTarget) {
                                let data = $(e.originalEvent.explicitOriginalTarget).data();
                                if (data.submitSuccessTarget === "window") {
                                    window.location.href = xLocation;
                                    return;
                                }
                            }
                            _this.load(xLocation);
                            return
                        }

                        var returnUrl = $form.data('returnUrl'),
                            refreshUrl = $form.data('refreshUrl');

                        if (refreshUrl) {
                            window.location.href = refreshUrl;
                            return;
                        }

                        if (returnUrl === 'refresh') {
                            _this.refresh();
                            return;
                        }

                        if (returnUrl && returnUrl !== 'refresh') {
                            _this.load(returnUrl);
                        } else {
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
                            _this.refresh();
                        }
                    },
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

                            $slideout.scrollTop(0);
                        } else {
                            QOR.ajaxError.apply(this, arguments)
                        }
                    },
                    complete: function () {
                        $submit.prop('disabled', false);
                    }
                });
            }
        },

        showOnLoad: function(url, response) {
            this.show();

            // callback for after slider loaded HTML
            // this callback is deprecated, use slideoutLoaded.qor.slideout event.
            var qorSliderAfterShow = $.fn.qorSliderAfterShow;
            if (qorSliderAfterShow) {
                for (var name in qorSliderAfterShow) {
                    if (qorSliderAfterShow.hasOwnProperty(name) && $.isFunction(qorSliderAfterShow[name])) {
                        qorSliderAfterShow[name]['isLoaded'] = true;
                        qorSliderAfterShow[name].call(this, url, response);
                    }
                }
            }

            // will trigger slideoutLoaded.qor.slideout event after slideout loaded
            this.$slideout.trigger(EVENT_SLIDEOUT_LOADED, [url, response]);
        },

        load: function (url, data) {
            var options = this.options;
            var method;
            var dataType;
            var load;
            var $slideout = this.$slideout;
            var $title;

            if (!url) {
                return;
            }

            data = $.isPlainObject(data) ? data : {};

            if (data.image) {
                // reset slideout header and body
                $slideout.html(this.$slideoutTemplate);
                let $title = $slideout.find('.qor-slideout__title');
                if (data.title) {
                    $title.html(data.title)
                } else {
                    $title.remove();
                }
                this.$body = $slideout.find('.qor-slideout__body')
                let response = '<div class="center-text"><img src="'+url+'"></div>';
                this.$body.html(response);

                $(CLASS_BODY_LOADING).remove();
                $slideout
                    .one(EVENT_SHOWN, function () {
                        $(this).trigger('enable');
                    })
                    .one(EVENT_HIDDEN, function () {
                        $(this).trigger('disable');
                    });
                this.showOnLoad(url, response, false);
                return;
            }

            method = data.method ? data.method : 'GET';
            dataType = data.datatype ? data.datatype : 'html';

            load = $.proxy(function () {
                $.ajax(url, {
                    method: method,
                    dataType: dataType,
                    cache: true,
                    ifModified: true,
                    headers: {
                        'X-Layout': 'lite'
                    },
                    success: $.proxy(function (response) {
                        let $response, $content, $qorFormContainer, $form, $scripts, $links, bodyClass;

                        $(CLASS_BODY_LOADING).remove();

                        if (method === 'GET') {
                            $response = $(response);
                            $content = $response.find(CLASS_MAIN_CONTENT);
                            if (!$content.length) {
                                return;
                            }

                            $content.find(CLASS_ACTION_BUTTON).attr('data-disable-success-redirection', 'true');
                            $content.find(CLASS_FOOTER_COPYRIGTH).remove();
                            $qorFormContainer = $content.find('.qor-form-container');
                            this.slideoutType = $qorFormContainer.length && $qorFormContainer.data().slideoutType;

                            $form = $qorFormContainer.find('form');
                            if ($form.length === 1) {
                                $form.data(QOR.SUBMITER, this.submit.bind(this));
                            }

                            let bodyHtml = response.match(/<\s*body.*>[\s\S]*<\s*\/body\s*>/gi);
                            if (bodyHtml) {
                                bodyHtml = bodyHtml
                                    .join('')
                                    .replace(/<\s*body/gi, '<div')
                                    .replace(/<\s*\/body/gi, '</div');
                                bodyClass = $(bodyHtml).prop('class');
                                $('body').addClass(bodyClass);

                                let data = {
                                    $scripts: $response.filter('script'),
                                    $links: $response.filter('link'),
                                    url: url,
                                    response: response
                                };

                                this.loadExtraResource(data);
                            }

                            $content
                                .find('.qor-button--cancel')
                                .attr('data-dismiss', 'slideout')
                                .removeAttr('href');

                            $scripts = compareScripts($content.find('script[src]'));
                            $links = compareLinks($content.find('link[href]'));

                            if ($scripts.length) {
                                let data = {
                                    url: url,
                                    response: response
                                };

                                loadScripts($scripts, data, function () {
                                });
                            }

                            if ($links.length) {
                                loadStyles($links);
                            }

                            $content.find('script[src],link[href]').remove();

                            // reset slideout header and body
                            $slideout.html(this.$slideoutTemplate);
                            $title = $slideout.find('.qor-slideout__title');
                            this.$body = $slideout.find('.qor-slideout__body');

                            $title.html($response.find(options.title).html());
                            replaceHtml($slideout.find('.qor-slideout__body')[0], $content.html());
                            this.$body.find(CLASS_HEADER_LOCALE).remove();

                            $slideout
                                .one(EVENT_SHOWN, function () {
                                    $(this).trigger('enable');
                                })
                                .one(EVENT_HIDDEN, function () {
                                    $(this).trigger('disable');
                                });

                            $slideout.find('.qor-slideout__opennew').attr('href', url);
                            this.showOnLoad(url, response);
                        } else {
                            if (data.returnUrl) {
                                this.load(data.returnUrl);
                            } else {
                                this.refresh();
                            }
                        }
                    }, this),

                    error: $.proxy(function () {
                        $(CLASS_BODY_LOADING).remove();
                        if ($('.qor-error span').length > 0) {
                            let errors = $('.qor-error span')
                                .map(function () {
                                    return $(this).text();
                                })
                                .get()
                                .join(', ');
                            QOR.alert(errors)
                        } else {
                            QOR.ajaxError.apply(this, arguments)
                        }
                    }, this)
                });
            }, this);

            if (this.slided) {
                this.hide(true);
                this.$slideout.one(EVENT_HIDDEN, load);
            } else {
                load();
            }
        },

        open: function (options) {
            this.addLoading();
            return this.load(options.url, options.data || options);
        },

        reload: function (url) {
            this.hide(true);
            this.load(url);
        },

        show: function () {
            var $slideout = this.$slideout;
            var showEvent;

            if (this.slided) {
                return;
            }

            showEvent = $.Event(EVENT_SHOW);
            $slideout.trigger(showEvent);

            if (showEvent.isDefaultPrevented()) {
                return;
            }

            $slideout.removeClass(CLASS_MINI);
            this.slideoutType == 'mini' && $slideout.addClass(CLASS_MINI);

            $slideout.addClass(CLASS_IS_SHOWN).get(0).offsetWidth;
            $slideout
                .one(EVENT_TRANSITIONEND, $.proxy(this.shown, this))
                .addClass(CLASS_IS_SLIDED)
                .scrollTop(0);
        },

        shown: function () {
            this.slided = true;
            // Disable to scroll body element
            $('body').addClass(CLASS_OPEN);
            this.$slideout
                .trigger('beforeEnable.qor.slideout')
                .trigger(EVENT_SHOWN)
                .trigger('afterEnable.qor.slideout');
        },

        closeSlideout: function () {
            this.hide();
        },

        hide: function (isReload) {
            let _this = this,
                message = QOR.messages.slideout;

            if ($.fn.qorSlideoutBeforeHide) {
                window.QOR.qorConfirm(message, function (confirm) {
                    if (confirm) {
                        _this.hideSlideout(isReload);
                    }
                });
            } else {
                this.hideSlideout(isReload);
            }

            this.removeSelectedClass();
        },

        hideSlideout: function (isReload) {
            var $slideout = this.$slideout;
            var hideEvent;
            var $datePicker = $('.qor-datepicker').not('.hidden');

            // remove onbeforeunload event
            window.onbeforeunload = null;
            $.fn.qorSlideoutBeforeHide = null;

            if ($datePicker.length) {
                $datePicker.addClass('hidden');
            }

            if (!this.slided) {
                return;
            }

            hideEvent = $.Event(EVENT_HIDE);
            $slideout.trigger(hideEvent);

            if (hideEvent.isDefaultPrevented()) {
                return;
            }

            $slideout.one(EVENT_TRANSITIONEND, $.proxy(this.hidden, this)).removeClass(`${CLASS_IS_SLIDED} qor-slideout__fullscreen`);
            !isReload && $slideout.trigger(EVENT_SLIDEOUT_CLOSED);
        },

        hidden: function () {
            this.slided = false;

            // Enable to scroll body element
            $('body').removeClass(CLASS_OPEN);

            this.$slideout.removeClass(CLASS_IS_SHOWN).trigger(EVENT_HIDDEN);
        },

        refresh: function () {
            this.hide();

            setTimeout(function () {
                window.location.reload();
            }, 350);
        },

        destroy: function () {
            this.unbind();
            this.unbuild();
            this.$element.removeData(NAMESPACE);
        }
    };

    $.extend({}, QOR.messages, {
        slideout:{
            confirm: 'You have unsaved changes on this slideout. If you close this slideout, you will lose all ' +
                'unsaved changes. Are you sure you want to close the slideout?'
        }
    }, true);

    QorSlideout.DEFAULTS = {
        title: '.qor-form-title, .mdl-layout-title',
        content: false
    };

    QorSlideout.TEMPLATE = `<div class="qor-slideout">
            <div class="qor-slideout__header">
                <div class="qor-slideout__header-link">
                    <a href="#" target="_blank" class="mdl-button mdl-button--icon mdl-js-button mdl-js-repple-effect qor-slideout__opennew"><i class="material-icons">open_in_new</i></a>
                    <a href="#" class="mdl-button mdl-button--icon mdl-js-button mdl-js-repple-effect qor-slideout__fullscreen">
                        <i class="material-icons">fullscreen</i>
                        <i class="material-icons" style="display: none;">fullscreen_exit</i>
                    </a>
                </div>
                <button type="button" class="mdl-button mdl-button--icon mdl-js-button mdl-js-repple-effect qor-slideout__close" data-dismiss="slideout">
                    <span class="material-icons">close</span>
                </button>
                <h3 class="qor-slideout__title"></h3>
            </div>
            <div class="qor-slideout__body"></div>
        </div>`;

    QorSlideout.TEMPLATE_LOADING = `<div class="qor-body__loading">
            <div><div class="mdl-spinner mdl-js-spinner is-active qor-layout__bottomsheet-spinner"></div></div>
        </div>`;

    QorSlideout.plugin = function (options) {
        return this.each(function () {
            var $this = $(this);
            var data = $this.data(NAMESPACE);
            var fn;

            if (!data) {
                if (/destroy/.test(options)) {
                    return;
                }

                $this.data(NAMESPACE, (data = new QorSlideout(this, options)));
            }

            if (typeof options === 'string' && $.isFunction((fn = data[options]))) {
                fn.apply(data);
            }
        });
    };

    $.fn.qorSlideout = QorSlideout.plugin;

    $document.data('qor.slideout', function (cb) {
        return cb.call(QorSlideout);
    });

    return QorSlideout;
});
