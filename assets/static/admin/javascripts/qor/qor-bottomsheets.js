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

    let _ = window._,
        NAMESPACE = 'qor.bottomsheets',
        EVENT_CLICK = 'click.' + NAMESPACE,
        EVENT_SUBMIT = 'submit.' + NAMESPACE,
        EVENT_SUBMITED = 'ajaxSuccessed.' + NAMESPACE,
        EVENT_RELOAD = 'reload.' + NAMESPACE,
        EVENT_BOTTOMSHEET_BEFORESEND = 'bottomsheetBeforeSend.' + NAMESPACE,
        EVENT_BOTTOMSHEET_LOADED = 'bottomsheetLoaded.' + NAMESPACE,
        EVENT_BOTTOMSHEET_CLOSED = 'bottomsheetClosed.' + NAMESPACE,
        EVENT_BOTTOMSHEET_SUBMIT = 'bottomsheetSubmitComplete.' + NAMESPACE,
        EVENT_HIDDEN = 'hidden.' + NAMESPACE,
        EVENT_KEYUP = 'keyup.' + NAMESPACE,
        CLASS_OPEN = 'qor-bottomsheets-open',
        CLASS_IS_SHOWN = 'is-shown',
        CLASS_IS_SLIDED = 'is-slided',
        CLASS_MAIN_CONTENT = '.mdl-layout__content.qor-page',
        CLASS_BODY_CONTENT = '.qor-page__body',
        CLASS_BODY_HEAD = '.qor-page__header',
        CLASS_BOTTOMSHEETS_FILTER = '.qor-bottomsheet__filter',
        CLASS_BOTTOMSHEETS_BUTTON = '.qor-bottomsheets__search-button',
        CLASS_BOTTOMSHEETS_INPUT = '.qor-bottomsheets__search-input',
        URL_GETQOR = 'http://www.getqor.com/';

    function getUrlParameter(name, search) {
        name = name.replace(/[\[]/, '\\[').replace(/[\]]/, '\\]');
        let regex = new RegExp('[\\?&]' + name + '=([^&#]*)'),
            results = regex.exec(search);
        return results === null ? '' : decodeURIComponent(results[1].replace(/\+/g, ' '));
    }

    function updateQueryStringParameter(key, value, uri) {
        let escapedkey = String(key).replace(/[\\^$*+?.()|[\]{}]/g, '\\$&'),
            re = new RegExp('([?&])' + escapedkey + '=.*?(&|$)', 'i'),
            separator = uri.indexOf('?') !== -1 ? '&' : '?';

        if (uri.match(re)) {
            if (value) {
                return uri.replace(re, '$1' + key + '=' + value + '$2');
            } else {
                if (RegExp.$1 === '?' || RegExp.$1 === RegExp.$2) {
                    return uri.replace(re, '$1');
                } else {
                    return uri.replace(re, '');
                }
            }
        } else if (value) {
            return uri + separator + key + '=' + value;
        }
    }

    function pushArrary($ele, isScript) {
        let array = [],
            prop = 'href';

        isScript && (prop = 'src');
        $ele.each(function() {
            array.push($(this).attr(prop));
        });
        return _.uniq(array);
    }

    function execSlideoutEvents(url, response) {
        // exec qorSliderAfterShow after script loaded
        let qorSliderAfterShow = $.fn.qorSliderAfterShow;
        for (let name in qorSliderAfterShow) {
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

            script.onload = function() {
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
        ss.onload = function() {
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

    function QorBottomSheets(element, options) {
        this.$element = $(element);
        this.options = $.extend({}, QorBottomSheets.DEFAULTS, $.isPlainObject(options) && options);
        this.disabled = false;
        this.resourseData = {};
        this.init();
    }

    QorBottomSheets.prototype = {
        constructor: QorBottomSheets,

        init: function() {
            this.filterURL = '';
            this.searchParams = '';
            this.renderedOpt = {};
            this.build();
            this.bind();
        },

        build: function() {
            let $bottomsheets;

            this.$bottomsheets = $bottomsheets = $(QorBottomSheets.TEMPLATE).appendTo('body');
            this.$body = $bottomsheets.find('.qor-bottomsheets__body');
            this.$title = $bottomsheets.find('.qor-bottomsheets__title');
            this.$header = $bottomsheets.find('.qor-bottomsheets__header');
            this.$bodyClass = $('body').prop('class');
        },

        bind: function() {
            this.$bottomsheets
                .on(EVENT_CLICK, '[data-dismiss="bottomsheets"]', this.hide.bind(this))
                .on(EVENT_CLICK, '.qor-pagination a', this.pagination.bind(this))
                .on(EVENT_CLICK, CLASS_BOTTOMSHEETS_BUTTON, this.search.bind(this))
                .on(EVENT_KEYUP, this.keyup.bind(this))
                .on('selectorChanged.qor.selector', this.selectorChanged.bind(this))
                .on('filterChanged.qor.filter', this.filterChanged.bind(this))
                .on(EVENT_CLICK, '[data-dismiss="fullscreen"]', this.toggleMode.bind(this));
        },

        unbind: function() {
            this.$bottomsheets
                .off(EVENT_CLICK, '[data-dismiss="bottomsheets"]', this.hide.bind(this))
                .off(EVENT_CLICK, '.qor-pagination a', this.pagination.bind(this))
                .off(EVENT_CLICK, CLASS_BOTTOMSHEETS_BUTTON, this.search.bind(this))
                .off('selectorChanged.qor.selector', this.selectorChanged.bind(this))
                .off('filterChanged.qor.filter', this.filterChanged.bind(this))
                .on(EVENT_CLICK, '[data-dismiss="fullscreen"]', this.toggleMode.bind(this));
        },

        toggleMode: function () {
            this.$bottomsheets
                .toggleClass('qor-bottomsheets__fullscreen')
                .find('[data-dismiss="fullscreen"] span')
                .toggle();
        },

        bindActionData: function(actiondData) {
            let $form = this.$body.find('[data-toggle="qor-action-slideout"]').find('form'),
                pkValues = [];
            for (let i = actiondData.length - 1; i >= 0; i--) {
                pkValues.push(actiondData[i])
            }
            if (pkValues.length > 0) {
                $form.prepend('<input type="hidden" name=":pk" value="' + pkValues.join(":") + '" />');
            }
        },

        filterChanged: function(e, search, key) {
            // if this event triggered:
            // search: ?locale_mode=locale, ?filters[Color].Value=2
            // key: search param name: locale_mode

            let loadUrl;

            loadUrl = this.constructloadURL(search, key);
            loadUrl && this.reload(loadUrl);
            return false;
        },

        selectorChanged: function(e, url, key) {
            // if this event triggered:
            // url: /admin/!remote_data_searcher/products/Collections?locale=en-US
            // key: search param key: locale

            let loadUrl;

            loadUrl = this.constructloadURL(url, key);
            loadUrl && this.reload(loadUrl);
            return false;
        },

        keyup: function(e) {
            let searchInput = this.$bottomsheets.find(CLASS_BOTTOMSHEETS_INPUT);

            if (e.which === 13 && searchInput.length && searchInput.is(':focus')) {
                this.search();
            }
        },

        search: function() {
            let $bottomsheets = this.$bottomsheets,
                param = 'keyword=',
                baseUrl = $bottomsheets.data().url,
                searchValue = $.trim($bottomsheets.find(CLASS_BOTTOMSHEETS_INPUT).val() || ''),
                url = baseUrl + (baseUrl.indexOf('?') === -1 ? '?' : '&') + param + encodeURIComponent(searchValue);

            this.reload(url);
        },

        pagination: function(e) {
            let $ele = $(e.target),
                url = $ele.prop('href');
            if (url) {
                this.reload(url);
            }
            return false;
        },

        reload: function(url) {
            let $content = this.$bottomsheets.find(CLASS_BODY_CONTENT);

            this.addLoading($content);
            this.fetchPage(url);
        },

        fetchPage: function(url) {
            let $bottomsheets = this.$bottomsheets;

            $.ajax({
                url: url,
                headers: {
                    'X-Layout': 'lite'
                },
                success: function (response) {
                    let $response = $(response).find(CLASS_MAIN_CONTENT),
                        $responseHeader = $response.find(CLASS_BODY_HEAD),
                        $responseBody = $response.find(CLASS_BODY_CONTENT);

                    if ($responseBody.length) {
                        $bottomsheets.find(CLASS_BODY_CONTENT).html($responseBody.html()).trigger('enable');

                        if ($responseHeader.length) {
                            $bottomsheets.removeAttr('data-src');
                            this.$body
                                .find(CLASS_BODY_HEAD)
                                .html($responseHeader.html())
                                .attr('data-src', url)
                                .trigger('enable');
                            this.addHeaderClass();
                        }
                        // will trigger this event(relaod.qor.bottomsheets) when bottomsheets reload complete: like pagination, filter, action etc.
                        $bottomsheets.trigger(EVENT_RELOAD);
                    } else {
                        this.reload(url);
                    }
                }.bind(this),
                error: QOR.ajaxError
            });
        },

        constructloadURL: function(url, key) {
            let fakeURL,
                value,
                filterURL = this.filterURL,
                bindUrl = this.$bottomsheets.data().url;

            if (!filterURL) {
                if (bindUrl) {
                    filterURL = bindUrl;
                } else {
                    return;
                }
            }

            fakeURL = new URL(URL_GETQOR + url);
            value = getUrlParameter(key, fakeURL.search);
            filterURL = this.filterURL = updateQueryStringParameter(key, value, filterURL);

            return filterURL;
        },

        addHeaderClass: function() {
            this.$body.find(CLASS_BODY_HEAD).remove();
            if (this.$bottomsheets.find(CLASS_BODY_HEAD).children(CLASS_BOTTOMSHEETS_FILTER).length) {
                this.$body
                    .addClass('has-header')
                    .find(CLASS_BODY_HEAD)
                    .show();
            }
        },

        addLoading: function($element) {
            $element.html('');
            let $loading = $(QorBottomSheets.TEMPLATE_LOADING).appendTo($element);
            window.componentHandler.upgradeElement($loading.children()[0]);
        },

        loadExtraResource: function(data) {
            let styleDiff = compareLinks(data.$links),
                scriptDiff = compareScripts(data.$scripts);

            if (styleDiff.length) {
                loadStyles(styleDiff);
            }

            if (scriptDiff.length) {
                loadScripts(scriptDiff, data);
            }
        },

        loadMedialibraryJS: function($response) {
            let $script = $response.filter('script'),
                theme = /theme=media_library/g,
                src,
                _this = this;

            $script.each(function() {
                src = $(this).prop('src');
                if (theme.test(src)) {
                    let script = document.createElement('script');
                    script.src = src;
                    document.body.appendChild(script);
                    _this.scriptAdded = true;
                }
            });
        },

        submit: function(e) {
            let $body = this.$body,
                form = e.target,
                $form = $(form),
                _this = this,
                url = $form.prop('action'),
                formData = QOR.FormData(form).formData(),
                $bottomsheets = $form.closest('.qor-bottomsheets'),
                resourseData = $bottomsheets.data(),
                ajaxType = resourseData.ajaxType,
                $submit = $form.find(':submit');

            // will ingore submit event if need handle with other submit event: like select one, many...
            if (resourseData.ignoreSubmit) {
                return;
            }

            // will submit form as normal,
            // if you need download file after submit form or other things, please add
            // data-use-normal-submit="true" to form tag
            // <form action="/admin/products/!action/localize" method="POST" enctype="multipart/form-data" data-normal-submit="true"></form>
            let normalSubmit = $form.data().normalSubmit;

            if (normalSubmit) {
                return;
            }

            $(document).trigger(EVENT_BOTTOMSHEET_BEFORESEND);
            e.preventDefault();

            $.ajax(url, {
                method: $form.prop('method'),
                data: formData,
                dataType: ajaxType ? ajaxType : 'html',
                processData: false,
                contentType: false,
                headers: {
                    'X-Layout': 'lite'
                },
                beforeSend: function() {
                    $submit.prop('disabled', true);
                },
                success: function(data, textStatus, jqXHR) {
                    if (resourseData.ajaxMute) {
                        $bottomsheets.remove();
                        return;
                    }

                    if (resourseData.ajaxTakeover) {
                        resourseData.$target.parent().trigger(EVENT_SUBMITED, [data, $bottomsheets]);
                        return;
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

                        window.QOR.qorAjaxHandleFile(url, contentType, fileName, formData);
                        $submit.prop('disabled', false);

                        return;
                    }

                    $('.qor-error').remove();

                    let returnUrl = $form.data('returnUrl'),
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
                        _this.refresh();
                    }

                    $(document).trigger(EVENT_BOTTOMSHEET_SUBMIT);
                },
                error: function(xhr, textStatus, errorThrown) {
                    if (xhr.status === 422) {
                        $body.find('.qor-error').remove();
                        let $error = $(xhr.responseText).find('.qor-error');
                        $form.before($error);
                        $('.qor-bottomsheets .qor-page__body').scrollTop(0);
                        QOR.alert($error)
                    } else {
                        QOR.ajaxError.apply(this, arguments)
                    }
                },
                complete: function() {
                    $submit.prop('disabled', false);
                }
            });
        },

        render: function(title, body) {
            this.$header.find('.qor-bottomsheets__title').html(title || '');
            if (typeof body === "string") {
                this.$body.html(body);
            } else {
                this.$body.html('');
                this.$body.append(body);
            }
            if (this.options.persistent) {
                this.$bottomsheets.trigger('enable');
            } else {
                this.show();
                this.$bottomsheets
                    .one(EVENT_HIDDEN, function () {
                        $(this).trigger('disable');
                    })
                    .trigger('enable');
            }
        },

        renderResponse: function(response, opt) {
            this.renderedOpt = opt = $.extend({}, opt || {});
            let $response = $(response),
                $content,
                bodyClass,
                loadExtraResourceData = {
                    $scripts: $response.filter('script'),
                    $links: $response.filter('link'),
                    url: opt.url,
                    response: response
                },
                bodyHtml = response.match(/<\s*body.*>[\s\S]*<\s*\/body\s*>/gi);

            $content = $response.find(CLASS_MAIN_CONTENT);

            if (bodyHtml) {
                bodyHtml = bodyHtml
                    .join('')
                    .replace(/<\s*body/gi, '<div')
                    .replace(/<\s*\/body/gi, '</div');
                bodyClass = $(bodyHtml).prop('class');
                $('body').addClass(bodyClass);
            }

            if (!$content.length) {
                return;
            }

            this.loadExtraResource(loadExtraResourceData);

            if (opt.removeHeader) {
                $content.find(CLASS_BODY_HEAD).remove();
            }

            $content.find('.qor-button--cancel').attr('data-dismiss', 'bottomsheets');

            this.$body.removeAttr('data-src').html($content.html());
            this.$title.html(opt.title ? opt.title : (opt.titleSelector ? $response.find(opt.titleSelector).html() : ''));
            this.show();
            this.$bottomsheets
                .one(EVENT_HIDDEN, function () {
                    $(this).trigger('disable');
                })
                .attr('data-src', opt.url || '')
                .trigger('enable');
        },

        load: function(url, data, callback) {
            data = data || {};

            let options = this.options,
                method,
                dataType,
                load,
                actionData = data.actionData,
                resourseData = this.resourseData,
                selectModal = resourseData.selectModal,
                ignoreSubmit = resourseData.ignoreSubmit,
                $bottomsheets = this.$bottomsheets,
                $header = this.$header,
                $body = this.$body;

            if (!url) {
                return;
            }

            if (data.image) {
                $body.css({overflow: 'auto', padding: 0});
                this.render(data.title, '<img style="max-width: inherit" src="' + url + '">');
                return;
            }

            if (data.$element) {
                url = QOR.Xurl(url, data.$element).toString();
            }

            this.show();
            this.addLoading($body);

            this.filterURL = url;
            $body.removeClass('has-header has-hint');

            data = $.isPlainObject(data) ? data : {};

            method = data.method ? data.method : 'GET';
            dataType = data.datatype ? data.datatype : 'html';

            if (actionData && actionData.length) {
                url += (url.indexOf('?') === -1 ? '?' : '&') + (':pk='+actionData.join(':'));
            }

            const onLoad = function(response) {
                if (method === 'GET') {
                    let $response = $(response),
                        $content,
                        bodyClass,
                        loadExtraResourceData = {
                            $scripts: $response.filter('script'),
                            $links: $response.filter('link'),
                            url: url,
                            response: response
                        },
                        hasSearch = selectModal && $response.find('.qor-search-container').length,
                        bodyHtml = response.match(/<\s*body.*>[\s\S]*<\s*\/body\s*>/gi);

                    $content = $response.find(CLASS_MAIN_CONTENT);

                    if (bodyHtml) {
                        bodyHtml = bodyHtml
                            .join('')
                            .replace(/<\s*body/gi, '<div')
                            .replace(/<\s*\/body/gi, '</div');
                        bodyClass = $(bodyHtml).prop('class');
                        $('body').addClass(bodyClass);
                    }

                    if (!$content.length) {
                        return;
                    }

                    this.loadExtraResource(loadExtraResourceData);

                    if (ignoreSubmit) {
                        $content.find(CLASS_BODY_HEAD).remove();
                    }

                    $content.find('.qor-button--cancel').attr('data-dismiss', 'bottomsheets');
                    $content.find('[data-order-by]').removeAttr('data-order-by').removeClass('is-not-sorted');

                    $body.html($content.html());
                    this.$title.html($response.find(options.title).html());

                    if (data.selectDefaultCreating) {
                        this.$title.append(
                            `<button class="mdl-button mdl-button--primary" type="button" data-load-inline="true" data-select-nohint="${data.selectNohint}" data-select-modal="${data.selectModal}" data-select-listing-url="${data.selectListingUrl}">${data.selectBacktolistTitle}</button>`
                        );
                    }

                    if (selectModal) {
                        $body
                            .find('.qor-button--new')
                            .data('ignoreSubmit', true)
                            .data('selectId', resourseData.selectId)
                            .data('loadInline', true)
                            // TODO: fix duplicate new values on submit
                            .remove();
                        if (
                            selectModal !== 'one' &&
                            !data.selectNohint &&
                            (typeof resourseData.maxItem === 'undefined' || resourseData.maxItem !== '1')
                        ) {
                            $body.addClass('has-hint');
                        }
                        if (selectModal === 'mediabox' && !this.scriptAdded) {
                            this.loadMedialibraryJS($response);
                        }
                    }

                    $header.find('.qor-button--new').remove();
                    this.$title.after($body.find('.qor-button--new'));

                    if (hasSearch) {
                        $bottomsheets.addClass('has-search');
                        $header.find('.qor-bottomsheets__search').remove();
                        $header.append(QorBottomSheets.TEMPLATE_SEARCH);
                    }

                    if (actionData && actionData.length) {
                        this.bindActionData(actionData);
                    }

                    if (resourseData.bottomsheetClassname) {
                        $bottomsheets.addClass(resourseData.bottomsheetClassname);
                    }

                    this.addHeaderClass();

                    $bottomsheets.trigger('enable');

                    let $form = $bottomsheets.find('.qor-bottomsheets__body form');
                    if ($form.length) {
                        this.$bottomsheets.qorSelectCore('newFormHandle', $form, this.load.bind(this));
                        $form.qorAsyncFormSubmiter('onSuccess', function(data, textStatus, jqXHR) {
                            $(document).trigger(EVENT_BOTTOMSHEET_SUBMIT);
                            if (jqXHR.getResponseHeader('X-Frame-Reload') === 'render-body') {
                                onLoad(data)
                                return
                            }
                            this.hide({target:$form[0]});
                            if (this.resourseData.windowReload) {
                                location.href = location.href
                            }
                        }.bind(this));
                    }

                    $body.find('form .qor-button--cancel:last').click(function (e){
                        this.hide(e);
                        return false;
                    }.bind(this));

                    $bottomsheets.one(EVENT_HIDDEN, function() {
                        $(this).trigger('disable');
                    });
                    $bottomsheets.data(data);

                    // handle after opened callback
                    if (callback && $.isFunction(callback)) {
                        callback(this.$bottomsheets);
                    }

                    // callback for after bottomSheets loaded HTML
                    $bottomsheets.trigger(EVENT_BOTTOMSHEET_LOADED, [url, response]);
                } else {
                    if (data.returnUrl) {
                        this.load(data.returnUrl);
                    } else {
                        this.refresh();
                    }
                }
            }.bind(this);

            load = $.proxy(function() {
                $.ajax(url, {
                    headers: {
                        'X-Layout': 'lite'
                    },
                    method: method,
                    dataType: dataType,
                    success: onLoad,

                    error: (function(jqXHR) {
                        let hasErr = false;
                        if (jqXHR.responseText) {
                            const $resp = $(jqXHR.responseText),
                                $errors = $resp.find('.qor-error span');
                            if ($errors.length > 0) {
                                hasErr = true;
                                let errors = $errors
                                    .map(function () {
                                        return $(this).text();
                                    })
                                    .get()
                                    .join(', ');
                                QOR.alert(errors);
                            }
                        }
                        if (!hasErr) {
                            QOR.ajaxError.apply(jqXHR, arguments)
                        }
                    }).bind(this)
                });
            }, this);

            load();
        },

        open: function(options, callback) {
            if (!options.loadInline) {
                this.init();
            }
            this.resourseData = options;
            this.load(options.url, options, callback);
        },

        show: function() {
            this.$bottomsheets.addClass(CLASS_IS_SHOWN + " " + CLASS_IS_SLIDED);
            //$('body').addClass(CLASS_OPEN);
        },

        hide: function(e) {
            const $bottomsheets = $(e.target).closest('.qor-bottomsheets');

            if (this.options.persistent) {
                $bottomsheets.removeClass(CLASS_IS_SHOWN + " " + CLASS_IS_SLIDED);
            } else {
                $bottomsheets.qorSelectCore('destroy');
                $bottomsheets.trigger(EVENT_BOTTOMSHEET_CLOSED).trigger('disable').remove();
                if (!$('.qor-bottomsheets').is(':visible')) {
                    $('body').removeClass(CLASS_OPEN);
                }
                this.destroy();
            }
            return false;
        },

        refresh: function() {
            this.$bottomsheets.remove();
            $('body').removeClass(CLASS_OPEN);

            setTimeout(function() {
                window.location.reload();
            }, 350);
        },

        destroy: function() {
            this.unbind();
            this.$element.removeData(NAMESPACE);
        }
    };

    QorBottomSheets.DEFAULTS = {
        title: '.qor-form-title, .mdl-layout-title',
        content: false
    };

    QorBottomSheets.TEMPLATE_ERROR = `<ul class="qor-error"><li><label><i class="material-icons">error</i><span>[[error]]</span></label></li></ul>`;
    QorBottomSheets.TEMPLATE_LOADING = `<div style="text-align: center; margin-top: 30px;"><div class="mdl-spinner mdl-js-spinner is-active qor-layout__bottomsheet-spinner"></div></div>`;
    QorBottomSheets.TEMPLATE_SEARCH = `<div class="qor-bottomsheets__search">
            <input autocomplete="off" type="text" class="mdl-textfield__input qor-bottomsheets__search-input" placeholder="Search" />
            <button class="mdl-button mdl-js-button mdl-button--icon qor-bottomsheets__search-button" type="button"><i class="material-icons">search</i></button>
        </div>`;

    QorBottomSheets.TEMPLATE = `<div class="qor-bottomsheets">
            <div class="qor-bottomsheets__header">
                <div class="qor-bottomsheets__header-control">
                    <h3 class="qor-bottomsheets__title"></h3>
                    <button type="button" class="mdl-button mdl-button--icon mdl-js-button mdl-js-repple-effect" data-dismiss="fullscreen">
                        <span class="material-icons">fullscreen</span>
                        <span class="material-icons" style="display: none;">fullscreen_exit</span>
                    </button>
                    <button type="button" class="mdl-button mdl-button--icon mdl-js-button mdl-js-repple-effect qor-bottomsheets__close" data-dismiss="bottomsheets">
                        <span class="material-icons">close</span>
                    </button>
                </div>
            </div>
            <div class="qor-bottomsheets__body"></div>
        </div>`;

    QorBottomSheets.plugin = function(options) {
        let args = Array.prototype.slice.call(arguments, 1);
        return this.each(function() {
            let $this = $(this),
                data = $this.data(NAMESPACE),
                fn;

            if (!data) {
                if (/destroy/.test(options)) {
                    return;
                }

                $this.data(NAMESPACE, (data = new QorBottomSheets(this, options)));
            }

            if (typeof options === 'string' && $.isFunction((fn = data[options]))) {
                fn.apply(data, args);
            } else if ($.isFunction(options)) {
                options.call({}, $this, data);
            } else if (options && (fn = options['do']) && $.isFunction(fn)) {
                fn.call(options, $this, data);
            }
        });
    };

    $.fn.qorBottomSheets = QorBottomSheets.plugin;

    return QorBottomSheets;
});
