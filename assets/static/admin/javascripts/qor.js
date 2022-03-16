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
    let QOR = {
        messages: {
            common: {
                recordNotFoundError: 'Record not found',
                ajaxError: 'Server error, please try again later!',
                alert: {
                    ok: 'Ok'
                },
                confirm: {
                    ok: 'Ok',
                    cancel: 'Cancel'
                }
            }
        }
    };
    window.QOR = QOR;

    QOR.showDialog = function (options, msg) {
        options = $.extend({
            ok: true,
            cancel: false,
            okText: QOR.messages.common.confirm.ok,
            cancelText: QOR.messages.common.confirm.cancel,
            level: 'warning',
            severity: true,
        }, options)

        const icon = ({
            warning: 'warning',
            error: 'error',
            info: 'info',
            notice: 'info'
        })[options.level];

        let $dialog = $(`<div id="dialog">
                  <div class="mdl-dialog-bg"></div>
                  <div class="mdl-dialog">
                      <div class="mdl-dialog__content` + (options.severity ? ` severity_${options.level} severity--text` : '') + `">
                        <p><i class="material-icons">${icon}</i></p>
                        <p class="mdl-dialog__message dialog-message">${msg}</p>
                      </div>
                      <div class="mdl-dialog__actions">
                        ` + (options.ok ? `<button data-ok type="button" class="mdl-button mdl-button--raised mdl-button--colored dialog-ok dialog-button">ok</button>` : '') + `
                        ` + (options.cancel ? `<button type="button" class="mdl-button dialog-cancel dialog-button" data-cancel>cancel</button>` : '') + `
                      </div>
                    </div>
                </div>`);
        if (options.ok) {
            $('[data-ok]', $dialog).one('click', function (e) {
                $dialog.remove();
                if (options.ok !== true) {
                    options.ok(e);
                }
            })
        }
        if (options.cancel) {
            $('[data-cancel]', $dialog).one('click', function (e) {
                $dialog.remove();
                if (options.cancel !== true) {
                    options.cancel(e);
                }
            })
        }

        $($dialog).appendTo('body');
        $dialog.show();
    };

    $(function () {
        let html = `<div id="dialog">
                  <div class="mdl-dialog-bg"></div>
                  <div class="mdl-dialog">
                      <div class="mdl-dialog__content">
                        <p><i class="material-icons" icon>warning</i></p>
                        <p class="mdl-dialog__message dialog-message">
                        </p>
                      </div>
                      <div class="mdl-dialog__actions">
                        <button type="button" class="mdl-button mdl-button--raised mdl-button--colored dialog-ok dialog-button" data-type="confirm">
                          ok
                        </button>
                        <button type="button" class="mdl-button dialog-cancel dialog-button" data-type="">
                          cancel
                        </button>
                      </div>
                    </div>
                </div>`,
            _ = window._,
            $dialog = $(html).appendTo('body');

        // ************************************ Refactor window.confirm ************************************
        $(document)
            .on('keyup.qor.confirm', function (e) {
                if (e.which === 27) {
                    if ($dialog.is(':visible')) {
                        setTimeout(function () {
                            $dialog.hide();
                        }, 100);
                    }
                }
            })
            .on('click.qor.confirm', '.dialog-button', function () {
                let value = $(this).data('type'),
                    callback = QOR.qorConfirmCallback;

                $dialog.hide();
                QOR.qorConfirmCallback = undefined;
                $.isFunction(callback) && callback(value);
                return false;
            });

        QOR.qorConfirm = function (data, callback) {
            let okBtn = $dialog.find('.dialog-ok'),
                cancelBtn = $dialog.find('.dialog-cancel').show();

            if (_.isString(data)) {
                $dialog.find('.dialog-message').text(data);
                okBtn.text(QOR.messages.common.confirm.ok);
                cancelBtn.text(QOR.messages.common.confirm.cancel);
            } else if (_.isObject(data)) {
                if (data.confirmOk && data.confirmCancel) {
                    okBtn.text(data.confirmOk);
                    cancelBtn.text(data.confirmCancel);
                } else {
                    okBtn.text(QOR.messages.common.confirm.ok);
                    cancelBtn.text(QOR.messages.common.confirm.cancel);
                }

                $dialog.find('.dialog-message').text(data.confirm);
            }

            $dialog.show();
            QOR.qorConfirmCallback = callback;
            return false;
        };

        QOR.alert = function (data, callback) {
            let okBtn = $dialog.find('.dialog-ok');
            $dialog.find('.dialog-cancel').hide();

            if (_.isString(data)) {
                $dialog.find('.dialog-message').html(data);
                okBtn.text(QOR.messages.common.alert.ok);
            } else if (data.jquery) {
                okBtn.text(QOR.messages.common.alert.ok);
                $dialog.find('.dialog-message').html(data);
            } else if (_.isObject(data)) {
                if (data.ok) {
                    okBtn.text(data.ok);
                } else {
                    okBtn.text(QOR.messages.common.alert.ok);
                }
                $dialog.find('.dialog-message').html(data.message);
            }

            $dialog.show();
            QOR.qorConfirmCallback = callback;
            return false;
        };

        QOR.ajaxError = function (xhr, textStatus, errorThrown) {
            QOR.alert(QOR.ajaxErrorString.apply(this, arguments));
        };

        QOR.ajaxErrorString = function (xhr, textStatus, errorThrown) {
            return "<strong>" + QOR.messages.common.ajaxError + "<strong></strong>:<br/>" + [textStatus, errorThrown].join(': ')
        };

        QOR.Xurl = function (url, $this) {
            return new Xurl(url, $this.valuerOf());
        };

        QOR.submitContinueEditing = function (e) {
            let $form = $(e).parents('form'),
                action = $form.attr('action') || window.location.href;
            if (!/(\?|&)continue_editing=/.test(action)) {
                let param = 'continue_editing=true';
                if (action.indexOf('?') === -1) {
                    action += '?' + param
                } else if (action[action.length - 1] !== '?') {
                    action += '&' + param
                } else {
                    action += param
                }
            }
            $form.attr('action', action);
            $form.submit();
            return false;
        };

        QOR.submitValues = function (e) {
            let $form = $(e).parents('form'),
                action = $form.attr('action') || window.location.href,
                pairs = Array.prototype.slice.call(arguments, 1),
                key, values;
            $form.attr('action', action);
            $form.children(':not(.qor-form__actions,input[name="_method"])').trigger('DISABLE').remove();
            for (let i = 0; i < pairs.length; i += 2) {
                key = pairs[i];
                values = pairs[i + 1];
                values.forEach(function (value) {
                    $form.append($(`<input type="hidden" name="${pairs[i]}" value="${value}">`))
                })
            }
            $form.submit();
            return false;
        };

        QOR.prepareFormData = function (formData) {
            let prefixes = {},
                removes = {},
                pair;
            for (pair of formData.entries()) {
                if (/\.@enabled$/.test(pair[0])) {
                    let prefix = pair[0].substring(0, pair[0].length - 8)
                    if (!prefixes.hasOwnProperty(prefix)) {
                        prefixes[prefix] = pair[1] === "false";
                    }
                }
            }

            for (let prefix in prefixes) {
                if (!prefixes[prefix]) {
                    continue
                }
                for (let key of formData.keys()) {
                    if (key.indexOf(prefix) === 0) {
                        removes[key] = 1
                    }
                }
            }

            for (let key in removes) {
                formData.delete(key)
            }

            for (let prefix in prefixes) {
                if (prefixes[prefix]) {
                    for (let prefix2 in prefixes) {
                        if (prefix !== prefix2 && prefix2.indexOf(prefix) === 0) {
                            delete prefixes[prefix2]
                        }
                    }
                } else {
                    delete prefixes[prefix]
                }
            }

            for (let prefix in prefixes) {
                formData.append(prefix + "@enabled", "false")
            }
        };

        QOR.SUBMITER = "qor.submiter";


        // *******************************************************************************

        // ****************Handle download file from AJAX POST****************************
        let objectToFormData = function (obj, form) {
            let formdata = form || new FormData(),
                key;

            for (var variable in obj) {
                if (obj.hasOwnProperty(variable) && obj[variable]) {
                    key = variable;
                }

                if (obj[variable] instanceof Date) {
                    formdata.append(key, obj[variable].toISOString());
                } else if (typeof obj[variable] === 'object' && !(obj[variable] instanceof File)) {
                    objectToFormData(obj[variable], formdata);
                } else {
                    formdata.append(key, obj[variable]);
                }
            }

            return formdata;
        };

        QOR.qorAjaxHandleFile = function (url, contentType, fileName, data) {
            let request = new XMLHttpRequest();

            request.responseType = 'arraybuffer';
            request.open('POST', url, true);
            request.onload = function () {
                if (this.status === 200) {
                    let blob = new Blob([this.response], {
                            type: contentType
                        }),
                        url = window.URL.createObjectURL(blob),
                        a = document.createElement('a');

                    document.body.appendChild(a);
                    a.href = url;
                    a.download = fileName || 'download-' + $.now();
                    a.click();
                } else {
                    window.alert('server error, please try again!');
                }
            };

            if (_.isObject(data)) {
                if (Object.prototype.toString.call(data) != '[object FormData]') {
                    data = objectToFormData(data);
                }

                request.send(data);
            }
        };

        // ********************************convert video link********************
        // linkyoutube: /https?:\/\/(?:[0-9A-Z-]+\.)?(?:youtu\.be\/|youtube\.com\S*[^\w\-\s])([\w\-]{11})(?=[^\w\-]|$)(?![?=&+%\w.\-]*(?:['"][^<>]*>|<\/a>))[?=&+%\w.-]*/ig,
        // linkvimeo: /https?:\/\/(www\.)?vimeo.com\/(\d+)($|\/)/,

        let converVideoLinks = function () {
            let $ele = $('.qor-linkify-object'),
                linkyoutube = /https?:\/\/(?:[0-9A-Z-]+\.)?(?:youtu\.be\/|youtube\.com\S*[^\w\-\s])([\w\-]{11})(?=[^\w\-]|$)(?![?=&+%\w.\-]*(?:['"][^<>]*>|<\/a>))[?=&+%\w.-]*/gi;

            if (!$ele.length) {
                return;
            }

            $ele.each(function () {
                let url = $(this).data('video-link');
                if (url.match(linkyoutube)) {
                    $(this).html(`<iframe width="100%" height="100%" src="//www.youtube.com/embed/${url.replace(linkyoutube, '$1')}" frameborder="0" allowfullscreen></iframe>`);
                }
            });
        };

        $.fn.qorSliderAfterShow.converVideoLinks = converVideoLinks;
        converVideoLinks();

        /**
         * Fire an event handler to the specified node. Event handlers can detect that the event was fired programatically
         * by testing for a 'synthetic=true' property on the event object
         * @param {HTMLNode} node The node to fire the event handler on.
         * @param {String} eventName The name of the event without the "on" (e.g., "focus")
         */
        QOR.fireEvent = function (node, eventName) {
            // Make sure we use the ownerDocument from the provided node to avoid cross-window problems
            var doc;
            if (node.ownerDocument) {
                doc = node.ownerDocument;
            } else if (node.nodeType == 9) {
                // the node may be the document itself, nodeType 9 = DOCUMENT_NODE
                doc = node;
            } else {
                throw new Error("Invalid node passed to fireEvent: " + node.id);
            }

            if (node.dispatchEvent) {
                // Gecko-style approach (now the standard) takes more work
                var eventClass = "";

                // Different events have different event classes.
                // If this switch statement can't map an eventName to an eventClass,
                // the event firing is going to fail.
                switch (eventName) {
                    case "click": // Dispatching of 'click' appears to not work correctly in Safari. Use 'mousedown' or 'mouseup' instead.
                    case "mousedown":
                    case "mouseup":
                        eventClass = "MouseEvents";
                        break;

                    case "focus":
                    case "change":
                    case "blur":
                    case "select":
                        eventClass = "HTMLEvents";
                        break;

                    default:
                        throw "fireEvent: Couldn't find an event class for event '" + eventName + "'.";
                        break;
                }
                var event = doc.createEvent(eventClass);

                var bubbles = eventName == "change" ? false : true;
                event.initEvent(eventName, bubbles, true); // All events created as bubbling and cancelable.

                event.synthetic = true; // allow detection of synthetic events
                node.dispatchEvent(event, true);
            } else if (node.fireEvent) {
                // IE-old school style
                var event = doc.createEventObject();
                event.synthetic = true; // allow detection of synthetic events
                node.fireEvent("on" + eventName, event);
            }
        };
    });
});

(function (factory) {
    if (typeof define === 'function' && define.amd) {
        // AMD. Register as anonymous module.
        define('datepicker', ['jquery'], factory);
    } else if (typeof exports === 'object') {
        // Node / CommonJS
        factory(require('jquery'));
    } else {
        // Browser globals.
        factory(jQuery);
    }
})(function ($) {

    'use strict';

    var $window = $(window);
    var document = window.document;
    var $document = $(document);
    var Number = window.Number;
    var NAMESPACE = 'datepicker';

    // Events
    var EVENT_CLICK = 'click.' + NAMESPACE;
    var EVENT_KEYUP = 'keyup.' + NAMESPACE;
    var EVENT_FOCUS = 'focus.' + NAMESPACE;
    var EVENT_RESIZE = 'resize.' + NAMESPACE;
    var EVENT_SHOW = 'show.' + NAMESPACE;
    var EVENT_HIDE = 'hide.' + NAMESPACE;
    var EVENT_PICK = 'pick.' + NAMESPACE;

    // RegExps
    var REGEXP_FORMAT = /(y|m|d)+/g;
    var REGEXP_DIGITS = /\d+/g;
    var REGEXP_YEAR = /^\d{2,4}$/;

    // Classes
    var CLASS_INLINE = NAMESPACE + '-inline';
    var CLASS_DROPDOWN = NAMESPACE + '-dropdown';
    var CLASS_TOP_LEFT = NAMESPACE + '-top-left';
    var CLASS_TOP_RIGHT = NAMESPACE + '-top-right';
    var CLASS_BOTTOM_LEFT = NAMESPACE + '-bottom-left';
    var CLASS_BOTTOM_RIGHT = NAMESPACE + '-bottom-right';
    var CLASS_PLACEMENTS = [
        CLASS_TOP_LEFT,
        CLASS_TOP_RIGHT,
        CLASS_BOTTOM_LEFT,
        CLASS_BOTTOM_RIGHT
    ].join(' ');
    var CLASS_HIDE = NAMESPACE + '-hide';

    // Maths
    var min = Math.min;

    // Utilities
    var toString = Object.prototype.toString;

    function typeOf(obj) {
        return toString.call(obj).slice(8, -1).toLowerCase();
    }

    function isString(str) {
        return typeof str === 'string';
    }

    function isNumber(num) {
        return typeof num === 'number' && !isNaN(num);
    }

    function isUndefined(obj) {
        return typeof obj === 'undefined';
    }

    function isDate(date) {
        return typeOf(date) === 'date';
    }

    function toArray(obj, offset) {
        var args = [];

        if (Array.from) {
            return Array.from(obj).slice(offset || 0);
        }

        // This is necessary for IE8
        if (isNumber(offset)) {
            args.push(offset);
        }

        return args.slice.apply(obj, args);
    }

    // Custom proxy to avoid jQuery's guid
    function proxy(fn, context) {
        var args = toArray(arguments, 2);

        return function () {
            return fn.apply(context, args.concat(toArray(arguments)));
        };
    }

    function isLeapYear(year) {
        return (year % 4 === 0 && year % 100 !== 0) || year % 400 === 0;
    }

    function getDaysInMonth(year, month) {
        return [31, (isLeapYear(year) ? 29 : 28), 31, 30, 31, 30, 31, 31, 30, 31, 30, 31][month];
    }

    function parseFormat(format) {
        var source = String(format).toLowerCase();
        var parts = source.match(REGEXP_FORMAT);
        var length;
        var i;

        if (!parts || parts.length === 0) {
            throw new Error('Invalid date format.');
        }

        format = {
            source: source,
            parts: parts
        };

        length = parts.length;

        for (i = 0; i < length; i++) {
            switch (parts[i]) {
                case 'dd':
                case 'd':
                    format.hasDay = true;
                    break;

                case 'mm':
                case 'm':
                    format.hasMonth = true;
                    break;

                case 'yyyy':
                case 'yy':
                    format.hasYear = true;
                    break;

                // No default
            }
        }

        return format;
    }

    function parseDate(format, date) {
        var parts = [];
        var length;
        var year;
        var day;
        var month;
        var val;
        var i;

        if (isDate(date)) {
            return new Date(date.getFullYear(), date.getMonth(), date.getDate());
        } else if (isString(date)) {
            parts = date.match(REGEXP_DIGITS) || [];
        }

        date = new Date();
        year = date.getFullYear();
        day = date.getDate();
        month = date.getMonth();
        length = format.parts.length;

        if (parts.length === length) {
            for (i = 0; i < length; i++) {
                val = parseInt(parts[i], 10) || 1;

                switch (format.parts[i]) {
                    case 'dd':
                    case 'd':
                        day = val;
                        break;

                    case 'mm':
                    case 'm':
                        month = val - 1;
                        break;

                    case 'yy':
                        year = 2000 + val;
                        break;

                    case 'yyyy':
                        year = val;
                        break;

                    // No default
                }
            }
        }

        return new Date(year, month, day);
    }

    $.extend(true, QOR.messages, {
        datepicker: {
            // The date string format
            format: 'yyyy-mm-dd',

            // Days' name of the week.
            days: ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday'],

            // Shorter days' name
            daysShort: ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'],

            // Shortest days' name
            daysMin: ['Su', 'Mo', 'Tu', 'We', 'Th', 'Fr', 'Sa'],

            // Months' name
            months: ['January', 'February', 'March', 'April', 'May', 'June', 'July', 'August', 'September', 'October', 'November', 'December'],

            // Shorter months' name
            monthsShort: ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'],
        }
    });

    function Datepicker(element, options) {
        options = $.isPlainObject(options) ? options : {};
        this.$element = $(element);
        this.options = $.extend({}, Datepicker.DEFAULTS, QOR.messages.datepicker, options);
        this.isBuilt = false;
        this.isShown = false;
        this.isInput = false;
        this.isInline = false;
        this.initialValue = '';
        this.initialDate = null;
        this.startDate = null;
        this.endDate = null;
        this.init();
    }

    Datepicker.formatDate = function (date, format) {
        if (!format) {
            format = QOR.messages.datepicker.format
        }
        if (isString(format)) {
            format = parseFormat(format)
        }
        var formated = '';
        var length;
        var year;
        var part;
        var val;
        var i;

        if (isDate(date)) {
            formated = format.source;
            year = date.getFullYear();
            val = {
                d: date.getDate(),
                m: date.getMonth() + 1,
                yy: year.toString().substring(2),
                yyyy: year
            };

            val.dd = (val.d < 10 ? '0' : '') + val.d;
            val.mm = (val.m < 10 ? '0' : '') + val.m;
            length = format.parts.length;

            for (i = 0; i < length; i++) {
                part = format.parts[i];
                formated = formated.replace(part, val[part]);
            }
        }

        return formated;
    };

    Datepicker.prototype = {
        constructor: Datepicker,

        init: function () {
            var options = this.options;
            var $this = this.$element;
            var startDate = options.startDate;
            var endDate = options.endDate;
            var date = options.date;

            this.$trigger = $(options.trigger || $this);
            this.isInput = $this.is('input') || $this.is('textarea');
            this.isInline = options.inline && (options.container || !this.isInput);
            this.format = parseFormat(options.format);
            this.initialValue = this.getValue();
            date = this.parseDate(date || this.initialValue);

            if (startDate) {
                startDate = this.parseDate(startDate);

                if (date.getTime() < startDate.getTime()) {
                    date = new Date(startDate);
                }

                this.startDate = startDate;
            }

            if (endDate) {
                endDate = this.parseDate(endDate);

                if (startDate && endDate.getTime() < startDate.getTime()) {
                    endDate = new Date(startDate);
                }

                if (date.getTime() > endDate.getTime()) {
                    date = new Date(endDate);
                }

                this.endDate = endDate;
            }

            this.date = date;
            this.viewDate = new Date(date);
            this.initialDate = new Date(this.date);

            this.bind();

            if (options.autoshow || this.isInline) {
                this.show();
            }

            if (options.autopick) {
                this.pick();
            }
        },

        build: function () {
            var options = this.options;
            var $this = this.$element;
            var $picker;

            if (this.isBuilt) {
                return;
            }

            this.isBuilt = true;

            this.$picker = $picker = $(options.template);
            this.$week = $picker.find('[data-view="week"]');

            // Years view
            this.$yearsPicker = $picker.find('[data-view="years picker"]');
            this.$yearsPrev = $picker.find('[data-view="years prev"]');
            this.$yearsNext = $picker.find('[data-view="years next"]');
            this.$yearsCurrent = $picker.find('[data-view="years current"]');
            this.$years = $picker.find('[data-view="years"]');

            // Months view
            this.$monthsPicker = $picker.find('[data-view="months picker"]');
            this.$yearPrev = $picker.find('[data-view="year prev"]');
            this.$yearNext = $picker.find('[data-view="year next"]');
            this.$yearCurrent = $picker.find('[data-view="year current"]');
            this.$months = $picker.find('[data-view="months"]');

            // Days view
            this.$daysPicker = $picker.find('[data-view="days picker"]');
            this.$monthPrev = $picker.find('[data-view="month prev"]');
            this.$monthNext = $picker.find('[data-view="month next"]');
            this.$monthCurrent = $picker.find('[data-view="month current"]');
            this.$days = $picker.find('[data-view="days"]');

            if (this.isInline) {
                $(options.container || $this).append($picker.addClass(CLASS_INLINE));
            } else {
                $(document.body).append($picker.addClass(CLASS_DROPDOWN));
                $picker.addClass(CLASS_HIDE);
            }

            this.fillWeek();
        },

        unbuild: function () {
            if (!this.isBuilt) {
                return;
            }

            this.isBuilt = false;
            this.$picker.remove();
        },

        bind: function () {
            var options = this.options;
            var $this = this.$element;

            if ($.isFunction(options.show)) {
                $this.on(EVENT_SHOW, options.show);
            }

            if ($.isFunction(options.hide)) {
                $this.on(EVENT_HIDE, options.hide);
            }

            if ($.isFunction(options.pick)) {
                $this.on(EVENT_PICK, options.pick);
            }

            if (this.isInput) {
                $this.on(EVENT_KEYUP, $.proxy(this.keyup, this));

                if (!options.trigger) {
                    $this.on(EVENT_FOCUS, $.proxy(this.show, this));
                }
            }

            this.$trigger.on(EVENT_CLICK, $.proxy(this.show, this));
        },

        unbind: function () {
            var options = this.options;
            var $this = this.$element;

            if ($.isFunction(options.show)) {
                $this.off(EVENT_SHOW, options.show);
            }

            if ($.isFunction(options.hide)) {
                $this.off(EVENT_HIDE, options.hide);
            }

            if ($.isFunction(options.pick)) {
                $this.off(EVENT_PICK, options.pick);
            }

            if (this.isInput) {
                $this.off(EVENT_KEYUP, this.keyup);

                if (!options.trigger) {
                    $this.off(EVENT_FOCUS, this.show);
                }
            }

            this.$trigger.off(EVENT_CLICK, this.show);
        },

        showView: function (view) {
            var $yearsPicker = this.$yearsPicker;
            var $monthsPicker = this.$monthsPicker;
            var $daysPicker = this.$daysPicker;
            var format = this.format;

            if (format.hasYear || format.hasMonth || format.hasDay) {
                switch (Number(view)) {
                    case 2:
                    case 'years':
                        $monthsPicker.addClass(CLASS_HIDE);
                        $daysPicker.addClass(CLASS_HIDE);

                        if (format.hasYear) {
                            this.fillYears();
                            $yearsPicker.removeClass(CLASS_HIDE);
                        } else {
                            this.showView(0);
                        }

                        break;

                    case 1:
                    case 'months':
                        $yearsPicker.addClass(CLASS_HIDE);
                        $daysPicker.addClass(CLASS_HIDE);

                        if (format.hasMonth) {
                            this.fillMonths();
                            $monthsPicker.removeClass(CLASS_HIDE);
                        } else {
                            this.showView(2);
                        }

                        break;

                    // case 0:
                    // case 'days':
                    default:
                        $yearsPicker.addClass(CLASS_HIDE);
                        $monthsPicker.addClass(CLASS_HIDE);

                        if (format.hasDay) {
                            this.fillDays();
                            $daysPicker.removeClass(CLASS_HIDE);
                        } else {
                            this.showView(1);
                        }
                }
            }
        },

        hideView: function () {
            if (this.options.autohide) {
                this.hide();
            }
        },

        place: function () {
            var options = this.options;
            var $this = this.$element;
            var $picker = this.$picker;
            var containerWidth = $document.outerWidth();
            var containerHeight = $document.outerHeight();
            var elementWidth = $this.outerWidth();
            var elementHeight = $this.outerHeight();
            var width = $picker.width();
            var height = $picker.height();
            var offsets = $this.offset();
            var left = offsets.left;
            var top = offsets.top;
            var offset = parseFloat(options.offset) || 10;
            var placement = CLASS_TOP_LEFT;

            if (top > height && top + elementHeight + height > containerHeight) {
                top -= height + offset;
                placement = CLASS_BOTTOM_LEFT;
            } else {
                top += elementHeight + offset;
            }

            if (left + width > containerWidth) {
                left = left + elementWidth - width;
                placement = placement.replace('left', 'right');
            }

            $picker.removeClass(CLASS_PLACEMENTS).addClass(placement).css({
                top: top,
                left: left,
                zIndex: parseInt(options.zIndex, 10)
            });
        },

        // A shortcut for triggering custom events
        trigger: function (type, data) {
            var e = $.Event(type, data);

            this.$element.trigger(e);

            return e;
        },

        createItem: function (data) {
            var options = this.options;
            var itemTag = options.itemTag;
            var defaults = {
                text: '',
                view: '',
                muted: false,
                picked: false,
                disabled: false
            };

            $.extend(defaults, data);

            return (
                '<' + itemTag + ' ' +
                (defaults.disabled ? 'class="' + options.disabledClass + '"' :
                    defaults.picked ? 'class="' + options.pickedClass + '"' :
                        defaults.muted ? 'class="' + options.mutedClass + '"' : '') +
                (defaults.view ? ' data-view="' + defaults.view + '"' : '') +
                '>' +
                defaults.text +
                '</' + itemTag + '>'
            );
        },

        fillAll: function () {
            this.fillYears();
            this.fillMonths();
            this.fillDays();
        },

        fillWeek: function () {
            var options = this.options;
            var weekStart = parseInt(options.weekStart, 10) % 7;
            var days = options.daysMin;
            var list = '';
            var i;

            days = $.merge(days.slice(weekStart), days.slice(0, weekStart));

            for (i = 0; i <= 6; i++) {
                list += this.createItem({
                    text: days[i]
                });
            }

            this.$week.html(list);
        },

        fillYears: function () {
            var options = this.options;
            var disabledClass = options.disabledClass || '';
            var suffix = options.yearSuffix || '';
            var filter = $.isFunction(options.filter) && options.filter;
            var startDate = this.startDate;
            var endDate = this.endDate;
            var viewDate = this.viewDate;
            var viewYear = viewDate.getFullYear();
            var viewMonth = viewDate.getMonth();
            var viewDay = viewDate.getDate();
            var date = this.date;
            var year = date.getFullYear();
            var isPrevDisabled = false;
            var isNextDisabled = false;
            var isDisabled = false;
            var isPicked = false;
            var isMuted = false;
            var list = '';
            var start = -5;
            var end = 6;
            var i;

            for (i = start; i <= end; i++) {
                date = new Date(viewYear + i, viewMonth, viewDay);
                isMuted = i === start || i === end;
                isPicked = (viewYear + i) === year;
                isDisabled = false;

                if (startDate) {
                    isDisabled = date.getFullYear() < startDate.getFullYear();

                    if (i === start) {
                        isPrevDisabled = isDisabled;
                    }
                }

                if (!isDisabled && endDate) {
                    isDisabled = date.getFullYear() > endDate.getFullYear();

                    if (i === end) {
                        isNextDisabled = isDisabled;
                    }
                }

                if (!isDisabled && filter) {
                    isDisabled = filter.call(this.$element, date) === false;
                }

                list += this.createItem({
                    text: viewYear + i,
                    view: isDisabled ? 'year disabled' : isPicked ? 'year picked' : 'year',
                    muted: isMuted,
                    picked: isPicked,
                    disabled: isDisabled
                });
            }

            this.$yearsPrev.toggleClass(disabledClass, isPrevDisabled);
            this.$yearsNext.toggleClass(disabledClass, isNextDisabled);
            this.$yearsCurrent.toggleClass(disabledClass, true).html((viewYear + start) + suffix + ' - ' + (viewYear + end) + suffix);
            this.$years.html(list);
        },

        fillMonths: function () {
            var options = this.options;
            var disabledClass = options.disabledClass || '';
            var months = options.monthsShort;
            var filter = $.isFunction(options.filter) && options.filter;
            var startDate = this.startDate;
            var endDate = this.endDate;
            var viewDate = this.viewDate;
            var viewYear = viewDate.getFullYear();
            var viewDay = viewDate.getDate();
            var date = this.date;
            var year = date.getFullYear();
            var month = date.getMonth();
            var isPrevDisabled = false;
            var isNextDisabled = false;
            var isDisabled = false;
            var isPicked = false;
            var list = '';
            var i;

            for (i = 0; i <= 11; i++) {
                date = new Date(viewYear, i, viewDay);
                isPicked = viewYear === year && i === month;
                isDisabled = false;

                if (startDate) {
                    isPrevDisabled = date.getFullYear() === startDate.getFullYear();
                    isDisabled = isPrevDisabled && date.getMonth() < startDate.getMonth();
                }

                if (!isDisabled && endDate) {
                    isNextDisabled = date.getFullYear() === endDate.getFullYear();
                    isDisabled = isNextDisabled && date.getMonth() > endDate.getMonth();
                }

                if (!isDisabled && filter) {
                    isDisabled = filter.call(this.$element, date) === false;
                }

                list += this.createItem({
                    index: i,
                    text: months[i],
                    view: isDisabled ? 'month disabled' : isPicked ? 'month picked' : 'month',
                    picked: isPicked,
                    disabled: isDisabled
                });
            }

            this.$yearPrev.toggleClass(disabledClass, isPrevDisabled);
            this.$yearNext.toggleClass(disabledClass, isNextDisabled);
            this.$yearCurrent.toggleClass(disabledClass, isPrevDisabled && isNextDisabled).html(viewYear + options.yearSuffix || '');
            this.$months.html(list);
        },

        fillDays: function () {
            var options = this.options;
            var disabledClass = options.disabledClass || '';
            var suffix = options.yearSuffix || '';
            var months = options.monthsShort;
            var weekStart = parseInt(options.weekStart, 10) % 7;
            var filter = $.isFunction(options.filter) && options.filter;
            var startDate = this.startDate;
            var endDate = this.endDate;
            var viewDate = this.viewDate;
            var viewYear = viewDate.getFullYear();
            var viewMonth = viewDate.getMonth();
            var prevViewYear = viewYear;
            var prevViewMonth = viewMonth;
            var nextViewYear = viewYear;
            var nextViewMonth = viewMonth;
            var date = this.date;
            var year = date.getFullYear();
            var month = date.getMonth();
            var day = date.getDate();
            var isPrevDisabled = false;
            var isNextDisabled = false;
            var isDisabled = false;
            var isPicked = false;
            var prevItems = [];
            var nextItems = [];
            var items = [];
            var total = 42; // 6 rows and 7 columns on the days picker
            var length;
            var i;
            var n;

            // Days of previous month
            // -----------------------------------------------------------------------

            if (viewMonth === 0) {
                prevViewYear -= 1;
                prevViewMonth = 11;
            } else {
                prevViewMonth -= 1;
            }

            // The length of the days of previous month
            length = getDaysInMonth(prevViewYear, prevViewMonth);

            // The first day of current month
            date = new Date(viewYear, viewMonth, 1);

            // The visible length of the days of previous month
            // [0,1,2,3,4,5,6] - [0,1,2,3,4,5,6] => [-6,-5,-4,-3,-2,-1,0,1,2,3,4,5,6]
            n = date.getDay() - weekStart;

            // [-6,-5,-4,-3,-2,-1,0,1,2,3,4,5,6] => [1,2,3,4,5,6,7]
            if (n <= 0) {
                n += 7;
            }

            if (startDate) {
                isPrevDisabled = date.getTime() <= startDate.getTime();
            }

            for (i = length - (n - 1); i <= length; i++) {
                date = new Date(prevViewYear, prevViewMonth, i);
                isDisabled = false;

                if (startDate) {
                    isDisabled = date.getTime() < startDate.getTime();
                }

                if (!isDisabled && filter) {
                    isDisabled = filter.call(this.$element, date) === false;
                }

                prevItems.push(this.createItem({
                    text: i,
                    view: 'day prev',
                    muted: true,
                    disabled: isDisabled
                }));
            }

            // Days of next month
            // -----------------------------------------------------------------------

            if (viewMonth === 11) {
                nextViewYear += 1;
                nextViewMonth = 0;
            } else {
                nextViewMonth += 1;
            }

            // The length of the days of current month
            length = getDaysInMonth(viewYear, viewMonth);

            // The visible length of next month
            n = total - (prevItems.length + length);

            // The last day of current month
            date = new Date(viewYear, viewMonth, length);

            if (endDate) {
                isNextDisabled = date.getTime() >= endDate.getTime();
            }

            for (i = 1; i <= n; i++) {
                date = new Date(nextViewYear, nextViewMonth, i);
                isDisabled = false;

                if (endDate) {
                    isDisabled = date.getTime() > endDate.getTime();
                }

                if (!isDisabled && filter) {
                    isDisabled = filter.call(this.$element, date) === false;
                }

                nextItems.push(this.createItem({
                    text: i,
                    view: 'day next',
                    muted: true,
                    disabled: isDisabled
                }));
            }

            // Days of current month
            // -----------------------------------------------------------------------

            for (i = 1; i <= length; i++) {
                date = new Date(viewYear, viewMonth, i);
                isPicked = viewYear === year && viewMonth === month && i === day;
                isDisabled = false;

                if (startDate) {
                    isDisabled = date.getTime() < startDate.getTime();
                }

                if (!isDisabled && endDate) {
                    isDisabled = date.getTime() > endDate.getTime();
                }

                if (!isDisabled && filter) {
                    isDisabled = filter.call(this.$element, date) === false;
                }

                items.push(this.createItem({
                    text: i,
                    view: isDisabled ? 'day disabled' : isPicked ? 'day picked' : 'day',
                    picked: isPicked,
                    disabled: isDisabled
                }));
            }

            // Render days picker
            // -----------------------------------------------------------------------

            this.$monthPrev.toggleClass(disabledClass, isPrevDisabled);
            this.$monthNext.toggleClass(disabledClass, isNextDisabled);
            this.$monthCurrent.toggleClass(disabledClass, isPrevDisabled && isNextDisabled).html(
                options.yearFirst ?
                    viewYear + suffix + ' ' + months[viewMonth] :
                    months[viewMonth] + ' ' + viewYear + suffix
            );
            this.$days.html(prevItems.join('') + items.join(' ') + nextItems.join(''));
        },

        click: function (e) {
            var $target = $(e.target);
            var viewDate = this.viewDate;
            var viewYear;
            var viewMonth;
            var viewDay;
            var isYear;
            var year;
            var view;

            e.stopPropagation();
            e.preventDefault();

            if ($target.hasClass('disabled')) {
                return;
            }

            viewYear = viewDate.getFullYear();
            viewMonth = viewDate.getMonth();
            viewDay = viewDate.getDate();
            view = $target.data('view');

            switch (view) {
                case 'years prev':
                case 'years next':
                    viewYear = view === 'years prev' ? viewYear - 10 : viewYear + 10;
                    year = $target.text();
                    isYear = REGEXP_YEAR.test(year);

                    if (isYear) {
                        viewYear = parseInt(year, 10);
                        this.date = new Date(viewYear, viewMonth, min(viewDay, 28));
                    }

                    this.viewDate = new Date(viewYear, viewMonth, min(viewDay, 28));
                    this.fillYears();

                    if (isYear) {
                        this.showView(1);
                        this.pick('year');
                    }

                    break;

                case 'year prev':
                case 'year next':
                    viewYear = view === 'year prev' ? viewYear - 1 : viewYear + 1;
                    this.viewDate = new Date(viewYear, viewMonth, min(viewDay, 28));
                    this.fillMonths();
                    break;

                case 'year current':

                    if (this.format.hasYear) {
                        this.showView(2);
                    }

                    break;

                case 'year picked':

                    if (this.format.hasMonth) {
                        this.showView(1);
                    } else {
                        this.hideView();
                    }

                    break;

                case 'year':
                    viewYear = parseInt($target.text(), 10);
                    this.date = new Date(viewYear, viewMonth, min(viewDay, 28));
                    this.viewDate = new Date(viewYear, viewMonth, min(viewDay, 28));

                    if (this.format.hasMonth) {
                        this.showView(1);
                    } else {
                        this.hideView();
                    }

                    this.pick('year');
                    break;

                case 'month prev':
                case 'month next':
                    viewMonth = view === 'month prev' ? viewMonth - 1 : view === 'month next' ? viewMonth + 1 : viewMonth;
                    this.viewDate = new Date(viewYear, viewMonth, min(viewDay, 28));
                    this.fillDays();
                    break;

                case 'month current':

                    if (this.format.hasMonth) {
                        this.showView(1);
                    }

                    break;

                case 'month picked':

                    if (this.format.hasDay) {
                        this.showView(0);
                    } else {
                        this.hideView();
                    }

                    break;

                case 'month':
                    viewMonth = $.inArray($target.text(), this.options.monthsShort);
                    this.date = new Date(viewYear, viewMonth, min(viewDay, 28));
                    this.viewDate = new Date(viewYear, viewMonth, min(viewDay, 28));

                    if (this.format.hasDay) {
                        this.showView(0);
                    } else {
                        this.hideView();
                    }

                    this.pick('month');
                    break;

                case 'day prev':
                case 'day next':
                case 'day':
                    viewMonth = view === 'day prev' ? viewMonth - 1 : view === 'day next' ? viewMonth + 1 : viewMonth;
                    viewDay = parseInt($target.text(), 10);
                    this.date = new Date(viewYear, viewMonth, viewDay);
                    this.viewDate = new Date(viewYear, viewMonth, viewDay);
                    this.fillDays();

                    if (view === 'day') {
                        this.hideView();
                    }

                    this.pick('day');
                    break;

                case 'day picked':
                    this.hideView();
                    this.pick('day');
                    break;

                // No default
            }
        },

        clickDoc: function (e) {
            var target = e.target;
            var trigger = this.$trigger[0];
            var ignored;

            while (target !== document) {
                if (target === trigger) {
                    ignored = true;
                    break;
                }

                target = target.parentNode;
            }

            if (!ignored) {
                this.hide();
            }
        },

        keyup: function () {
            this.update();
        },

        getValue: function () {
            var $this = this.$element;
            var val = '';

            if (this.isInput) {
                val = $this.val();
            } else if (this.isInline) {
                if (this.options.container) {
                    val = $this.text();
                }
            } else {
                val = $this.text();
            }

            return val;
        },

        setValue: function (val) {
            var $this = this.$element;

            val = isString(val) ? val : '';

            if (this.isInput) {
                $this.val(val);
            } else if (this.isInline) {
                if (this.options.container) {
                    $this.text(val);
                }
            } else {
                $this.text(val);
            }
        },


        // Methods
        // -------------------------------------------------------------------------

        // Show the datepicker
        show: function () {
            if (!this.isBuilt) {
                this.build();
            }

            if (this.isShown) {
                return;
            }

            if (this.trigger(EVENT_SHOW).isDefaultPrevented()) {
                return;
            }

            this.isShown = true;
            this.$picker.removeClass(CLASS_HIDE).on(EVENT_CLICK, $.proxy(this.click, this));
            this.showView(this.options.startView);

            if (!this.isInline) {
                $window.on(EVENT_RESIZE, (this._place = proxy(this.place, this)));
                $document.on(EVENT_CLICK, (this._clickDoc = proxy(this.clickDoc, this)));
                this.place();
            }
        },

        // Hide the datepicker
        hide: function () {
            if (!this.isShown) {
                return;
            }

            if (this.trigger(EVENT_HIDE).isDefaultPrevented()) {
                return;
            }

            this.isShown = false;
            this.$picker.addClass(CLASS_HIDE).off(EVENT_CLICK, this.click);

            if (!this.isInline) {
                $window.off(EVENT_RESIZE, this._place);
                $document.off(EVENT_CLICK, this._clickDoc);
            }
        },

        // Update the datepicker with the current input value
        update: function () {
            this.setDate(this.getValue(), true);
        },

        /**
         * Pick the current date to the element
         *
         * @param {String} _view (private)
         */
        pick: function (_view) {
            var $this = this.$element;
            var date = this.date;

            if (this.trigger(EVENT_PICK, {
                view: _view || '',
                date: date
            }).isDefaultPrevented()) {
                return;
            }

            this.setValue(date = this.formatDate(this.date));

            if (this.isInput) {
                $this.trigger('change');
            }
        },

        // Reset the datepicker
        reset: function () {
            this.setDate(this.initialDate, true);
            this.setValue(this.initialValue);

            if (this.isShown) {
                this.showView(this.options.startView);
            }
        },

        /**
         * Get the month name with given argument or the current date
         *
         * @param {Number} month (optional)
         * @param {Boolean} short (optional)
         * @return {String} (month name)
         */
        getMonthName: function (month, short) {
            var options = this.options;
            var months = options.months;

            if ($.isNumeric(month)) {
                month = Number(month);
            } else if (isUndefined(short)) {
                short = month;
            }

            if (short === true) {
                months = options.monthsShort;
            }

            return months[isNumber(month) ? month : this.date.getMonth()];
        },

        /**
         * Get the day name with given argument or the current date
         *
         * @param {Number} day (optional)
         * @param {Boolean} short (optional)
         * @param {Boolean} min (optional)
         * @return {String} (day name)
         */
        getDayName: function (day, short, min) {
            var options = this.options;
            var days = options.days;

            if ($.isNumeric(day)) {
                day = Number(day);
            } else {
                if (isUndefined(min)) {
                    min = short;
                }

                if (isUndefined(short)) {
                    short = day;
                }
            }

            days = min === true ? options.daysMin : short === true ? options.daysShort : days;

            return days[isNumber(day) ? day : this.date.getDay()];
        },

        /**
         * Get the current date
         *
         * @param {Boolean} formated (optional)
         * @return {Date|String} (date)
         */
        getDate: function (formated) {
            var date = this.date;

            return formated ? this.formatDate(date) : new Date(date);
        },

        /**
         * Get the current date format
         *
         * @return {String} (format)
         */
        getDateFormat: function () {
            return this.options.format;
        },

        /**
         * Set the current date with a new date
         *
         * @param {Date} date
         * @param {Boolean} _isUpdated (private)
         */
        setDate: function (date, _isUpdated) {
            var filter = this.options.filter;

            if (isDate(date) || isString(date)) {
                date = this.parseDate(date);

                if ($.isFunction(filter) && filter.call(this.$element, date) === false) {
                    return;
                }

                this.date = date;
                this.viewDate = new Date(date);

                if (!_isUpdated) {
                    this.pick();
                }

                if (this.isBuilt) {
                    this.fillAll();
                }
            }
        },

        /**
         * Set the start view date with a new date
         *
         * @param {Date} date
         */
        setStartDate: function (date) {
            if (isDate(date) || isString(date)) {
                this.startDate = this.parseDate(date);

                if (this.isBuilt) {
                    this.fillAll();
                }
            }
        },

        /**
         * Set the end view date with a new date
         *
         * @param {Date} date
         */
        setEndDate: function (date) {
            if (isDate(date) || isString(date)) {
                this.endDate = this.parseDate(date);

                if (this.isBuilt) {
                    this.fillAll();
                }
            }
        },

        /**
         * Parse a date string with the set date format
         *
         * @param {String} date
         * @return {Date} (parsed date)
         */
        parseDate: function (date) {
            return parseDate(this.format, date)
        },

        /**
         * Format a date object to a string with the set date format
         *
         * @param {Date} date
         * @return {String} (formated date)
         */
        formatDate: function (date) {
            return Datepicker.formatDate(date, this.format);
        },

        // Destroy the datepicker and remove the instance from the target element
        destroy: function () {
            this.unbind();
            this.unbuild();
            this.$element.removeData(NAMESPACE);
        }
    };

    Datepicker.DEFAULTS = {
        // Show the datepicker automatically when initialized
        autoshow: false,

        // Hide the datepicker automatically when picked
        autohide: false,

        // Pick the initial date automatically when initialized
        autopick: false,

        // Enable inline mode
        inline: false,

        // A element (or selector) for putting the datepicker
        container: null,

        // A element (or selector) for triggering the datepicker
        trigger: null,

        // The initial date
        date: null,

        // The start view date
        startDate: null,

        // The end view date
        endDate: null,

        // The start view when initialized
        startView: 0, // 0 for days, 1 for months, 2 for years

        // The start day of the week
        weekStart: 0, // 0 for Sunday, 1 for Monday, 2 for Tuesday, 3 for Wednesday, 4 for Thursday, 5 for Friday, 6 for Saturday

        // Show year before month on the datepicker header
        yearFirst: false,

        // A string suffix to the year number.
        yearSuffix: '',

        // A element tag for each item of years, months and days
        itemTag: 'li',

        // A class (CSS) for muted date item
        mutedClass: 'muted',

        // A class (CSS) for picked date item
        pickedClass: 'picked',

        // A class (CSS) for disabled date item
        disabledClass: 'disabled',

        // The template of the datepicker
        template: (
            '<div class="datepicker-container">' +
            '<div class="datepicker-panel" data-view="years picker">' +
            '<ul>' +
            '<li data-view="years prev">&lsaquo;</li>' +
            '<li data-view="years current"></li>' +
            '<li data-view="years next">&rsaquo;</li>' +
            '</ul>' +
            '<ul data-view="years"></ul>' +
            '</div>' +
            '<div class="datepicker-panel" data-view="months picker">' +
            '<ul>' +
            '<li data-view="year prev">&lsaquo;</li>' +
            '<li data-view="year current"></li>' +
            '<li data-view="year next">&rsaquo;</li>' +
            '</ul>' +
            '<ul data-view="months"></ul>' +
            '</div>' +
            '<div class="datepicker-panel" data-view="days picker">' +
            '<ul>' +
            '<li data-view="month prev">&lsaquo;</li>' +
            '<li data-view="month current"></li>' +
            '<li data-view="month next">&rsaquo;</li>' +
            '</ul>' +
            '<ul data-view="week"></ul>' +
            '<ul data-view="days"></ul>' +
            '</div>' +
            '</div>'
        ),

        // The offset top or bottom of the datepicker from the element
        offset: 10,

        // The `z-index` of the datepicker
        zIndex: 1000,

        // Filter each date item (return `false` to disable a date item)
        filter: null,

        // Event shortcuts
        show: null,
        hide: null,
        pick: null
    };

    Datepicker.setDefaults = function (options) {
        $.extend(Datepicker.DEFAULTS, $.isPlainObject(options) && options);
    };

    // Save the other datepicker
    Datepicker.other = $.fn.qorDatepicker;

    // Register as jQuery plugin
    $.fn.qorDatepicker = function (option) {
        var args = toArray(arguments, 1);
        var result;

        this.each(function () {
            var $this = $(this);
            var data = $this.data(NAMESPACE);
            var options;
            var fn;

            if (!data) {
                if (/destroy/.test(option)) {
                    return;
                }

                options = $.extend({}, $this.data(), $.isPlainObject(option) && option);
                $this.data(NAMESPACE, (data = new Datepicker(this, options)));
            }

            if (isString(option) && $.isFunction(fn = data[option])) {
                result = fn.apply(data, args);
            }
        });

        return isUndefined(result) ? this : result;
    };

    $.fn.qorDatepicker.Constructor = Datepicker;
    $.fn.qorDatepicker.setDefaults = Datepicker.setDefaults;
    $.fn.qorDatepicker.formatDate = Datepicker.formatDate;
    $.fn.qorDatepicker.parseDate = function (format, date) {
        return parseDate(parseFormat(format), date)
    };

    // No conflict
    $.fn.qorDatepicker.noConflict = function () {
        $.fn.qorDatepicker = Datepicker.other;
        return this;
    };

    $document.data('datepicker', function (cb) {
        return cb.call($.fn.qorDatepicker)
    })
});

(function (factory) {
    if (typeof define === 'function' && define.amd) {
        // AMD. Register as anonymous module.
        define('formValidator', ['jquery'], factory);
    } else if (typeof exports === 'object') {
        // Node / CommonJS
        factory(require('jquery'));
    } else {
        // Browser globals.
        factory(jQuery);
    }
})(function ($) {

    'use strict';

    let NAMESPACE = 'remoteFormValidator',
        EVENT_SUBMIT = 'submit.' + NAMESPACE,
        VALIDATING = 1,
        OK = 2,
        FAILED = 3;


    function FormValidator(element, options) {
        options = $.isPlainObject(options) ? options : {};
        this.$el = $(element);
        this.options = $.extend({}, FormValidator.DEFAULTS, options);
        this.key = 0;
        this.validators = {};
        this.validating = {};
        this.phase = 0;
        this.running = 0;
        this.destroyed = false;
        this.errors = [];
        this.init();
    }

    FormValidator.prototype = {
        constructor: FormValidator,

        init: function () {
            this.bind();
        },

        bind: function () {
            this.$el.on(EVENT_SUBMIT, this.validateOnSubmit.bind(this));
        },

        submit: function (e) {
            let submiter = this.$el.data(QOR.SUBMITER);
            if (submiter) {
                submiter(e)
            } else {
                this.unbind();
                this.$el.submit();
                this.bind();
            }
        },

        unbind: function () {
            this.$el.off(EVENT_SUBMIT);
        },
        validateOnSubmit: function(e) {
            if (Object.keys(this.validators).length === 0) {
                return true;
            }
            return this.validate(this.submit.bind(this, e), e)
        },
        validate: function(okCallback) {
            let key, count = 0, ok;
            if (Object.keys(this.validating).length > 0) {
                return false;
            }
            for (key in this.validators) {
                count++;
                this.validating[key] = true;
            }
            ok = count === 0;
            if (ok) {
                // must submit
                return true;
            }
            this.phase = VALIDATING;

            for (key in this.validators) {
                this.validating[key] = true;
                this._callValidator(key, okCallback);
            }

            return false;
        },

        callValidator: function (key, validatorCallback) {
            this._callValidator(key, null, validatorCallback)
        },

        _callValidator: function (key, mainOkCallback, validatorCallback) {
            if (this.destroyed) {
                return;
            }
            this.validators[key](this.validatorDone.bind(this, key, mainOkCallback, validatorCallback));
        },

        // validation done
        done: function (mainOkCallback) {
            if (this.errors.length === 0) {
                if (mainOkCallback) {
                    mainOkCallback()
                }
            } else {
                QOR.alert(this.errors.join('<hr />'));
                this.errors = [];
            }
            //this.$el.show();
        },

        validatorDone: function (key, mainOkCallback, validatorCallback, err) {
            if (this.destroyed) {
                return;
            }
            if (err) {
                if ((key in this.validating)) {
                    this.errors.push(err);
                }
            }
            // anonymous validate
            if (validatorCallback) {
                validatorCallback(err);
                return;
            }
            // all validate
            delete this.validating[key];
            for (key in this.validating) {
                // is running
                return
            }
            // done
            this.done(mainOkCallback);
        },

        // Methods
        // -------------------------------------------------------------------------


        // Register register new validator and return done callback
        register: function (validator, callback) {
            let key = this.key++;
            this.validators[key] = validator;
            if (callback) {
                callback(this.callValidator.bind(this, key))
            }
        },

        // Destroy the datepicker and remove the instance from the target element
        destroy: function () {
            this.destroyed = true;
            this.unbind();
            this.$el.removeData(NAMESPACE);
        }
    };

    FormValidator.DEFAULTS = {
    };

    FormValidator.setDefaults = function (options) {
        $.extend(FormValidator.DEFAULTS, $.isPlainObject(options) && options);
    };

    // Save the other
    FormValidator.other = $.fn.formValidator;

    // Register as jQuery plugin
    $.fn.formValidator = function (option) {
        let args = Array.prototype.slice.call(arguments, 1),
            result;

        this.each(function () {
            var $this = $(this);
            var data = $this.data(NAMESPACE);
            var options;
            var fn;

            if (!data) {
                if (/destroy/.test(option)) {
                    return;
                }

                options = $.extend({}, $this.data(), $.isPlainObject(option) && option);
                $this.data(NAMESPACE, (data = new FormValidator(this, options)));
            }

            if ((typeof option === "string") && $.isFunction(fn = data[option])) {
                result = fn.apply(data, args);
            }
        });

        return (typeof result === 'undefined') ? this : result;
    };

    $.fn.formValidator.Constructor = FormValidator;
    $.fn.formValidator.setDefaults = FormValidator.setDefaults;

    // No conflict
    $.fn.formValidator.noConflict = function () {
        $.fn.formValidator = FormValidator.other;
        return this;
    };
});

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

    $.extend({}, QOR.messages, {
        action: {
            bulk: {
                pleaseSelectAnItem: "You must select at least one item."
            },
            form: {
                areYouSure: "This action may not be undone. Are you sure you want to run anyway?"
            }
        }
    }, true);

    let Mustache = window.Mustache,
        NAMESPACE = 'qor.action',
        EVENT_ENABLE = 'enable.' + NAMESPACE,
        EVENT_DISABLE = 'disable.' + NAMESPACE,
        EVENT_CLICK = 'click.' + NAMESPACE,
        EVENT_UNDO = 'undo.' + NAMESPACE,
        ACTION_FORMS = '.qor-actions-bulk',
        ACTION_HEADER = '.qor-page__header',
        ACTION_BODY = '.qor-page__body',
        ACTION_BUTTON = '.qor-action-button',
        MDL_BODY = '.mdl-layout__content',
        ACTION_SELECTORS = '.qor-actions-default',
        ACTION_LINK = 'a.qor-action--button',
        MENU_ACTIONS = '.qor-table__actions a[data-url],[data-url][data-method=POST],[data-url][data-method=PUT],[data-url][data-method=DELETE]',
        BUTTON_BULKS = '.qor-action-bulk-buttons',
        QOR_TABLE = '.qor-table-container',
        QOR_TABLE_BULK = '.qor-table--bulking',
        QOR_SEARCH = '.qor-search-container',
        CLASS_IS_UNDO = 'is_undo',
        CLASS_BULK_EXIT = '.qor-action--exit-bulk',
        QOR_SLIDEOUT = '.qor-slideout',
        ACTION_FORM_DATA = ':pk';

    function QorAction(element, options) {
        this.$element = $(element);
        this.$wrap = $(ACTION_FORMS);
        this.options = $.extend({}, QorAction.DEFAULTS, $.isPlainObject(options) && options);
        this.ajaxForm = {};
        this.init();
    }

    QorAction.prototype = {
        constructor: QorAction,

        init: function () {
            this.bind();
            this.initActions();
        },

        bind: function () {
            this.$element.on(EVENT_CLICK, this.click.bind(this));
            this.$wrap.find(CLASS_BULK_EXIT).on(EVENT_CLICK, this.click.bind(this));
            $(document)
                .on(EVENT_CLICK, '.qor-table--bulking tr', this.click.bind(this))
                .on(EVENT_CLICK, ACTION_LINK, this.actionLink.bind(this));
        },

        unbind: function () {
            this.$element.off(EVENT_CLICK, this.click);

            $(document)
                .off(EVENT_CLICK, '.qor-table--bulking tr', this.click)
                .off(EVENT_CLICK, ACTION_LINK, this.actionLink);
        },

        initActions: function () {
            this.tables = $(QOR_TABLE).find('table').length;

            if (!this.tables) {
                $(BUTTON_BULKS)
                    .find('button')
                    .attr('disabled', true);
                $(ACTION_LINK).attr('disabled', true);
            }
        },

        collectFormData: function () {
            let checkedInputs = $(QOR_TABLE_BULK).find('.mdl-checkbox__input:checked'),
                formData = [],
                normalFormData = [],
                tempObj;

            if (checkedInputs.length) {
                checkedInputs.each(function () {
                    let id = $(this)
                        .closest('tr')
                        .data('primary-key');

                    tempObj = {};
                    if (id) {
                        formData.push({
                            name: ACTION_FORM_DATA,
                            value: id.toString()
                        });

                        tempObj[ACTION_FORM_DATA] = id.toString();
                        normalFormData.push(tempObj);
                    }
                });
            }
            this.ajaxForm.formData = formData;
            this.ajaxForm.normalFormData = normalFormData;
            return this.ajaxForm;
        },

        actionLink: function () {
            // if not in index page
            if (!$(QOR_TABLE).find('table').length) {
                return false;
            }
        },

        actionSubmit: function ($action) {
            let $target = $($action);
            this.$actionButton = $target;
            if ($target.data().method) {
                if ($target.data().ajaxForm) {
                    this.collectFormData();
                    this.ajaxForm.properties = $target.data();
                }
                this.submit();
                return false;
            }
        },

        click: function (e) {
            let $target = $(e.target),
                $pageHeader = $('.qor-page > .qor-page__header'),
                $pageBody = $('.qor-page > .qor-page__body'),
                triggerHeight = $pageHeader.find('.qor-page-subnav__header').length ? 96 : 48;

            this.$actionButton = $target;

            if ($target.data().ajaxForm) {
                return;
            }

            if ($target.is('.qor-action--bulk')) {
                this.$wrap.removeClass('hidden');
                $('.qor-table__inner-list').remove();
                this.appendTableCheckbox();
                $(QOR_TABLE).addClass('qor-table--bulking');
                $(ACTION_HEADER)
                    .find(ACTION_SELECTORS)
                    .addClass('hidden');
                $(ACTION_HEADER)
                    .find(QOR_SEARCH)
                    .addClass('hidden');
                if ($pageHeader.height() > triggerHeight) {
                    $pageBody.css('padding-top', $pageHeader.height());
                }
            }

            if ($target.is(CLASS_BULK_EXIT)) {
                this.$wrap.addClass('hidden');
                this.removeTableCheckbox();
                $(QOR_TABLE).removeClass('qor-table--bulking');
                $(ACTION_HEADER)
                    .find(ACTION_SELECTORS)
                    .removeClass('hidden');
                $(ACTION_HEADER)
                    .find(QOR_SEARCH)
                    .removeClass('hidden');
                if (parseInt($pageBody.css('padding-top')) > triggerHeight) {
                    $pageBody.css('padding-top', '');
                }
            }

            if ($(this).is('tr') && !$target.is('a')) {
                let $firstTd = $(this)
                    .find('td')
                    .first();

                // Manual make checkbox checked or not
                if ($firstTd.find('.mdl-checkbox__input').get(0)) {
                    let hasPopoverForm = $('body').hasClass('qor-bottomsheets-open') || $('body').hasClass('qor-slideout-open'),
                        $checkbox = $firstTd.find('.mdl-js-checkbox'),
                        slideroutActionForm = $('[data-toggle="qor-action-slideout"]').find('form'),
                        formValueInput = slideroutActionForm.find('.js-primary-value'),
                        primaryValue = $(this).data('primary-key'),
                        $alreadyHaveValue = formValueInput.filter('[value="' + primaryValue + '"]'),
                        isChecked;

                    $checkbox.toggleClass('is-checked');
                    $firstTd.parents('tr').toggleClass('is-selected');

                    isChecked = $checkbox.hasClass('is-checked');

                    $firstTd.find('input').prop('checked', isChecked);

                    if (slideroutActionForm.length && hasPopoverForm) {
                        if (isChecked && !$alreadyHaveValue.length) {
                            slideroutActionForm.prepend(
                                '<input class="js-primary-value" type="hidden" name=":pk" value="' + primaryValue + '" />'
                            );
                        }

                        if (!isChecked && $alreadyHaveValue.length) {
                            $alreadyHaveValue.remove();
                        }
                    }

                    return false;
                }
            }
        },

        renderFlashMessage: function (data) {
            let flashMessageTmpl = QorAction.FLASHMESSAGETMPL,
                msg;
            Mustache.parse(flashMessageTmpl);
            msg = Mustache.render(flashMessageTmpl, data)
            return msg;
        },

        submit: function () {
            let _this = this,
                $parent,
                $element = this.$element,
                $actionButton = this.$actionButton,
                ajaxForm = this.ajaxForm || {},
                properties = ajaxForm.properties || $actionButton.data(),
                url = properties.url,
                undoUrl = properties.undoUrl,
                isUndo = $actionButton.hasClass(CLASS_IS_UNDO),
                isInSlideout = $actionButton.closest(QOR_SLIDEOUT).length,
                isOne = properties.one,
                needDisableButtons = $element && !isInSlideout,
                headers = {
                    'X-Error-Body': 'true'
                };

            if (properties.disableSuccessRedirection) {
                headers['X-Disabled-Success-Redirection'] = 'true';
            }

            if (!properties.optional && (properties.fromIndex && (!ajaxForm.formData || !ajaxForm.formData.length))) {
                QOR.alert(QOR.messages.action.bulk.pleaseSelectAnItem);
                return;
            }

            if (properties.passCurrentQuery) {
                if (location.search.length) {
                    url += url.indexOf('?') !== -1 ? '&' : '?'
                    url += location.search.substring(1)
                }
            }

            if (properties.confirm && properties.ajaxForm && !properties.fromIndex) {
                QOR.qorConfirm(properties, function (confirm) {
                    if (confirm) {
                        let success = function (data, status, response) {
                            if (properties.reloadDisabled) {
                                return
                            }

                            // TODO: self reload if in resource object page (check slideout)

                            let xLocation = response.getResponseHeader('X-Location');

                            if (xLocation) {
                                window.location.href = xLocation;
                                return
                            }

                            let refreshUrl = properties.refreshUrl;

                            if (refreshUrl) {
                                window.location.href = refreshUrl;
                                return;
                            }

                            if (isOne) {
                                $actionButton.remove();
                            }
                            window.location.reload();
                        };
                        $.ajax(url, {
                            method: "POST",
                            data: {_method: properties.method},
                            // TODO: process JSON data type {message?}
                            //dataType: properties.datatype,
                            headers: headers,
                            success: success,
                            error: function (res) {
                                QOR.showDialog({level: 'error', ok: true}, res.responseText)
                            }
                        })
                    } else {

                    }
                });
            } else {
                if (isUndo) {
                    url = properties.undoUrl;
                }

                if (properties.targetWindow) {
                    let formS = '<form action="' + url + '" method="POST" target="_blank">';
                    this.ajaxForm.formData.forEach(function (el) {
                        formS += '<input type="hidden" name="' + el.name + '" value="' + el.value + '">'
                    })
                    formS += "</form>";
                    let $form = $(formS);
                    $('body').append($form);
                    $form.submit();
                    $form.remove();
                    return;
                }

                $.ajax(url, {
                    method: properties.method,
                    data: ajaxForm.formData,
                    dataType: properties.datatype,
                    headers: headers,
                    beforeSend: function () {
                        if (undoUrl) {
                            $actionButton.prop('disabled', true);
                        } else if (needDisableButtons) {
                            _this.switchButtons($element, 1);
                        }
                    },
                    success: function (data, status, response) {
                        if (properties.readOnly) {
                            let title = this.$actionButton.text();
                            $('body').qorBottomSheets('renderResponse', data, {url: url, title: title});
                            return
                        }
                        if (properties.reloadDisabled) {
                            return
                        }
                        let xLocation = response.getResponseHeader('X-Location');

                        if (xLocation) {
                            window.location.href = xLocation;
                            return
                        }

                        let refreshUrl = properties.refreshUrl;

                        if (refreshUrl) {
                            window.location.href = refreshUrl;
                            return;
                        }

                        let contentType = response.getResponseHeader('content-type'),
                            // handle file download from form submit
                            disposition = response.getResponseHeader('Content-Disposition');

                        if (disposition && disposition.indexOf('attachment') !== -1) {
                            let fileNameRegex = /filename[^;=\n]*=((['"]).*?\2|[^;\n]*)/,
                                matches = fileNameRegex.exec(disposition),
                                fileData = {},
                                fileName = '';

                            if (matches != null && matches[1]) {
                                fileName = matches[1].replace(/['"]/g, '');
                            }

                            if (properties.method) {
                                fileData = $.extend({}, ajaxForm.normalFormData, {
                                    _method: properties.method
                                });
                            }

                            window.QOR.qorAjaxHandleFile(url, contentType, fileName, fileData);

                            if (undoUrl) {
                                $actionButton.prop('disabled', false);
                            } else {
                                _this.switchButtons($element);
                            }

                            return;
                        }

                        // has undo action
                        if (undoUrl) {
                            $element.triggerHandler(EVENT_UNDO, [$actionButton, isUndo, data]);
                            isUndo ? $actionButton.removeClass(CLASS_IS_UNDO) : $actionButton.addClass(CLASS_IS_UNDO);
                            $actionButton.prop('disabled', false);
                            return;
                        }

                        if (contentType.indexOf('json') > -1) {
                            // render notification
                            $('.qor-alert').remove();
                            needDisableButtons && _this.switchButtons($element);
                            isInSlideout ? ($parent = $(QOR_SLIDEOUT)) : ($parent = $(MDL_BODY));
                            $parent.find(ACTION_BODY).prepend(_this.renderFlashMessage(data));
                        } else {
                            // properties.fromIndex || properties.fromMenu
                            window.location.reload();
                        }
                    }.bind(this),
                    error: function (xhr, textStatus, errorThrown) {
                        if (undoUrl) {
                            $actionButton.prop('disabled', false);
                        } else if (needDisableButtons) {
                            _this.switchButtons($element);
                        }
                        QOR.alert([textStatus, errorThrown].join(': '));
                    }.bind(this)
                });
            }
        },

        switchButtons: function ($element, disable) {
            let needDisbale = !!disable;
            $element.find(ACTION_BUTTON).prop('disabled', needDisbale);
        },

        destroy: function () {
            this.unbind();
            this.$element.removeData(NAMESPACE);
        },

        // Helper
        removeTableCheckbox: function () {
            $('.qor-page__body .mdl-data-table__select')
                .each(function (i, e) {
                    $(e)
                        .parents('td')
                        .remove();
                })
                .each(function (i, e) {
                    $(e)
                        .parents('th')
                        .remove();
                });
            $('.qor-table-container tr.is-selected').removeClass('is-selected');
            $('.qor-page__body table.mdl-data-table--selectable').removeClass('mdl-data-table--selectable');
            $('.qor-page__body tr.is-selected').removeClass('is-selected');
        },

        appendTableCheckbox: function () {
            // Only value change and the table isn't selectable will add checkboxes
            $('.qor-page__body .mdl-data-table__select')
                .each(function (i, e) {
                    $(e)
                        .parents('td')
                        .remove();
                })
                .each(function (i, e) {
                    $(e)
                        .parents('th')
                        .remove();
                });
            $('.qor-table-container tr.is-selected').removeClass('is-selected');

            let $tb = $('.qor-page__body table:eq(0)').addClass('mdl-data-table--selectable'),
                headRows = $tb.find('thead tr').length;

            // init google material
            new window.MaterialDataTable($tb.get(0));

            $tb.find('thead :checkbox').closest('th').attr('rowspan', headRows);

            let $selection = $tb.find('thead.is-hidden tr:first th:first').remove();

            $selection.clone().attr('rowspan', headRows).prependTo($tb.find('thead.is-hidden tr:last'));
            $selection.prependTo($tb.find('thead:not(.is-hidden) tr:last'));
            /*
            $('<th />').prependTo($('thead:not(".is-hidden") tr,thead.is-hidden tr:not(:last)'));

            $('thead.is-hidden tr').each(function (){
                $('<th />').insertAfter($(this).find('th:first'));
            })*/

            let $fixedHeadCheckBox = $tb.find(`thead:not(.is-fixed) .mdl-checkbox__input`),
                isMediaLibrary = $tb.is('.qor-table--medialibrary'),
                hasPopoverForm = $('body').hasClass('qor-bottomsheets-open') || $('body').hasClass('qor-slideout-open');

            isMediaLibrary && ($fixedHeadCheckBox = $('thead .mdl-checkbox__input'));

            $fixedHeadCheckBox.on('click', function () {
                if (!isMediaLibrary) {
                    $($tb, 'thead.is-fixed tr th')
                        .eq(0)
                        .find('label')
                        .click();
                    $(this)
                        .closest('label')
                        .toggleClass('is-checked');
                }

                let slideroutActionForm = $('[data-toggle="qor-action-slideout"]').find('form'),
                    slideroutActionFormPrimaryValues = slideroutActionForm.find('.js-primary-value');

                if (slideroutActionForm.length && hasPopoverForm) {
                    if ($(this).is(':checked')) {
                        let allPrimaryValues = $($tb, 'tbody tr'),
                            pkValues = [];

                        allPrimaryValues.each(function () {
                            let primaryValue = $(this).data('primary-key');
                            if (primaryValue) {
                                pkValues[pkValues.length] = primaryValue
                            }
                        });

                        if (pkValues.length > 0) {
                            slideroutActionForm.prepend(
                                '<input class="js-primary-value" type="hidden" name=":pk" value="' + pkValues.join(":") + '" />'
                            );
                        }
                    } else {
                        slideroutActionFormPrimaryValues.remove();
                    }
                }
            });
        }
    };
    QorAction.FLASHMESSAGETMPL = `<div class="qor-alert qor-action-alert qor-alert--success [[#error]]qor-alert--error[[/error]]" [[#message]]data-dismissible="true"[[/message]] role="alert">
          <button type="button" class="mdl-button mdl-button--icon" data-dismiss="alert">
            <i class="material-icons">close</i>
          </button>
          <span class="qor-alert-message">
            [[#message]]
              [[message]]
            [[/message]]
            [[#error]]
              [[error]]
            [[/error]]
          </span>
        </div>`;

    QorAction.DEFAULTS = {};

    $.fn.qorSliderAfterShow.qorActionInit = function (url, html) {
        let hasAction = $(html).find('[data-toggle="qor-action-slideout"]').length,
            $actionForm = $('[data-toggle="qor-action-slideout"]').find('form'),
            $checkedItem = $('.qor-page__body .mdl-checkbox__input:checked');

        if (hasAction && $checkedItem.length) {
            let pkValues = [];
            // insert checked value into sliderout form
            $checkedItem.each(function (i, e) {
                let id = $(e)
                    .parents('tbody tr')
                    .data('primary-key');
                if (id) {
                    pkValues[pkValues.length] = id;
                }
            });
            if (pkValues.length) {
                $actionForm.prepend('<input class="js-primary-value" type="hidden" name=":pk" value="' + pkValues.join(":") + '" />');
            }

        }
    };

    QorAction.plugin = function (options) {
        return this.each(function () {
            let $this = $(this),
                data = $this.data(NAMESPACE),
                fn;

            if (!data) {
                $this.data(NAMESPACE, (data = new QorAction(this, options)));
            }

            if (typeof options === 'string' && $.isFunction((fn = data[options]))) {
                fn.call(data);
            }
        });
    };

    $(function () {
        let options = {},
            selector = '[data-toggle="qor.action.bulk"]';

        $(document)
            .on(EVENT_DISABLE, function (e) {
                QorAction.plugin.call($(selector, e.target), 'destroy');
            })
            .on(EVENT_ENABLE, function (e) {
                QorAction.plugin.call($(selector, e.target), options);
            })
            .on(EVENT_CLICK, MENU_ACTIONS, function () {
                new QorAction().actionSubmit(this);
                return false;
            })
            .triggerHandler(EVENT_ENABLE);
    });

    return QorAction;
});

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

    let location = window.location,
        QOR = window.QOR,
        NAMESPACE = 'qor.advancedsearch',
        EVENT_ENABLE = 'enable.' + NAMESPACE,
        EVENT_DISABLE = 'disable.' + NAMESPACE,
        EVENT_CLICK = 'click.' + NAMESPACE,
        EVENT_SHOWN = 'shown.qor.modal',
        EVENT_SUBMIT = 'submit.' + NAMESPACE;

    function getExtraPairs(names) {
        let pairs = decodeURIComponent(location.search.substr(1)).split('&'),
            pairsObj = {},
            pair,
            i;

        if (pairs.length == 1 && pairs[0] == '') {
            return false;
        }

        for (i in pairs) {
            if (pairs[i] === '') continue;

            pair = pairs[i].split('=');
            pairsObj[pair[0]] = pair[1];
        }

        names.forEach(function(item) {
            delete pairsObj[item];
        });

        return pairsObj;
    }

    function QorAdvancedSearch(element, options) {
        this.$element = $(element);
        this.options = $.extend({}, QorAdvancedSearch.DEFAULTS, $.isPlainObject(options) && options);
        this.init();
    }

    QorAdvancedSearch.prototype = {
        constructor: QorAdvancedSearch,

        init: function() {
            this.$form = this.$element.find('form');
            this.$modal = $(QorAdvancedSearch.MODAL).appendTo('body');
            this.bind();
        },

        bind: function() {
            this.$element
                .on(EVENT_SUBMIT, 'form', this.submit.bind(this))
                .on(EVENT_CLICK, '.qor-advanced-filter__save', this.showSaveFilter.bind(this))
                .on(EVENT_CLICK, '.qor-advanced-filter__toggle', this.toggleFilterContent)
                .on(EVENT_CLICK, '.qor-advanced-filter__close', this.closeFilter)
                .on(EVENT_CLICK, '.qor-advanced-filter__delete', this.deleteSavedFilter);

            this.$modal.on(EVENT_SHOWN, this.start.bind(this));
        },

        closeFilter: function() {
            $('.qor-advanced-filter__dropdown').hide();
        },

        toggleFilterContent: function(e) {
            $(e.target)
                .closest('.qor-advanced-filter__toggle')
                .parent()
                .find('>[advanced-search-toggle]')
                .toggle();
        },

        showSaveFilter: function() {
            this.$modal.qorModal('show');
        },

        deleteSavedFilter: function(e) {
            let $target = $(e.target).closest('.qor-advanced-filter__delete'),
                $savedFilter = $target.closest('.qor-advanced-filter__savedfilter'),
                name = $target.data('filter-name'),
                url = location.pathname,
                message = {
                    confirm: 'Are you sure you want to delete this saved filter?'
                };

            QOR.qorConfirm(message, function(confirm) {
                if (confirm) {
                    $.get(url, $.param({delete_saved_filter: name}))
                        .done(function() {
                            $target.closest('li').remove();
                            if ($savedFilter.find('li').length === 0) {
                                $savedFilter.remove();
                            }
                        })
                        .fail(function() {
                            QOR.qorConfirm('Server error, please try again!');
                        });
                }
            });
            return false;
        },

        start: function() {
            this.$modal.trigger('enable.qor.material').on(EVENT_CLICK, '.qor-advanced-filter__savefilter', this.saveFilter.bind(this));
        },

        saveFilter: function() {
            let name = this.$modal.find('#qor-advanced-filter__savename').val();

            if (!name) {
                return;
            }

            this.$form.prepend(`<input type="hidden" name="filter_saving_name" value=${name}  />`).submit();
        },

        submit: function() {
            let $form = this.$form,
                formArr = $form.find('input[name],select[name]'),
                names = [],
                extraPairs;

            formArr.each(function() {
                names.push($(this).attr('name'));
            });

            extraPairs = getExtraPairs(names);

            if (!$.isEmptyObject(extraPairs)) {
                for (let key in extraPairs) {
                    if (extraPairs.hasOwnProperty(key)) {
                        $form.prepend(`<input type="hidden" name=${key} value=${extraPairs[key]}  />`);
                    }
                }
            }

            this.$element.find('.qor-advanced-filter__dropdown').hide();

            this.removeEmptyPairs($form);
        },

        removeEmptyPairs: function($form) {
            $form.find('advanced-filter-group').each(function() {
                let $this = $(this),
                    $input = $this.find('[filter-required]');
                if ($input.val() == '') {
                    $this.remove();
                }
            });
        },

        unbind: function() {},

        destroy: function() {
            this.unbind();
            this.$element.removeData(NAMESPACE);
        }
    };

    QorAdvancedSearch.DEFAULTS = {};

    QorAdvancedSearch.MODAL = `<div class="qor-modal fade" tabindex="-1" role="dialog" aria-hidden="true">
            <div class="mdl-card mdl-shadow--2dp" role="document">
                <div class="mdl-card__title">
                    <h2 class="mdl-card__title-text">Save advanced filter</h2>
                </div>
                <div class="mdl-card__supporting-text">
                        
                    <div class="mdl-textfield mdl-textfield--full-width mdl-js-textfield">
                        <input class="mdl-textfield__input" type="text" id="qor-advanced-filter__savename">
                        <label class="mdl-textfield__label" for="qor-advanced-filter__savename">Please enter name for this filter</label>
                    </div>
                </div>
                <div class="mdl-card__actions">
                    <a class="mdl-button mdl-button--colored mdl-button--raised qor-advanced-filter__savefilter">Save This Filter</a>
                    <a class="mdl-button mdl-button--colored" data-dismiss="modal">Cancel</a>
                </div>
                <div class="mdl-card__menu">
                    <button class="mdl-button mdl-button--icon" data-dismiss="modal" aria-label="close">
                        <i class="material-icons">close</i>
                    </button>
                </div>
            </div>
        </div>`;

    QorAdvancedSearch.plugin = function(options) {
        return this.each(function() {
            let $this = $(this),
                data = $this.data(NAMESPACE),
                fn;

            if (!data) {
                if (/destroy/.test(options)) {
                    return;
                }

                $this.data(NAMESPACE, (data = new QorAdvancedSearch(this, options)));
            }

            if (typeof options === 'string' && $.isFunction((fn = data[options]))) {
                fn.apply(data);
            }
        });
    };

    $(function() {
        let selector = '[data-toggle="qor.advancedsearch"]',
            options;

        $(document)
            .on(EVENT_DISABLE, function(e) {
                QorAdvancedSearch.plugin.call($(selector, e.target), 'destroy');
            })
            .on(EVENT_ENABLE, function(e) {
                QorAdvancedSearch.plugin.call($(selector, e.target), options);
            })
            .triggerHandler(EVENT_ENABLE);
    });

    return QorAdvancedSearch;
});
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

  var NAMESPACE = 'qor.autoheight';
  var EVENT_ENABLE = 'enable.' + NAMESPACE;
  var EVENT_DISABLE = 'disable.' + NAMESPACE;
  var EVENT_INPUT = 'input';

  function QorAutoheight(element, options) {
    this.$element = $(element);
    this.options = $.extend({}, QorAutoheight.DEFAULTS, $.isPlainObject(options) && options);
    this.init();
  }

  QorAutoheight.prototype = {
    constructor: QorAutoheight,

    init: function () {
      var $this = this.$element;

      this.overflow = $this.css('overflow');
      this.paddingTop = parseInt($this.css('padding-top'), 10);
      this.paddingBottom = parseInt($this.css('padding-bottom'), 10);
      $this.css('overflow', 'hidden');
      this.resize();
      this.bind();
    },

    bind: function () {
      this.$element.on(EVENT_INPUT, $.proxy(this.resize, this));
    },

    unbind: function () {
      this.$element.off(EVENT_INPUT, this.resize);
    },

    resize: function () {
      var $this = this.$element;

      if ($this.is(':hidden')) {
        return;
      }

      $this.height('auto').height($this.prop('scrollHeight') - this.paddingTop - this.paddingBottom);
    },

    destroy: function () {
      this.unbind();
      this.$element.css('overflow', this.overflow).removeData(NAMESPACE);
    }
  };

  QorAutoheight.DEFAULTS = {};

  QorAutoheight.plugin = function (options) {
    return this.each(function () {
      var $this = $(this);
      var data = $this.data(NAMESPACE);
      var fn;

      if (!data) {
        if (/destroy/.test(options)) {
          return;
        }

        $this.data(NAMESPACE, (data = new QorAutoheight(this, options)));
      }

      if (typeof options === 'string' && $.isFunction(fn = data[options])) {
        fn.apply(data);
      }
    });
  };

  $(function () {
    var selector = 'textarea.qor-js-autoheight';

    $(document).
      on(EVENT_DISABLE, function (e) {
        QorAutoheight.plugin.call($(selector, e.target), 'destroy');
      }).
      on(EVENT_ENABLE, function (e) {
        QorAutoheight.plugin.call($(selector, e.target));
      }).
      triggerHandler(EVENT_ENABLE);
  });

  return QorAutoheight;

});

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
        FormData = window.FormData,
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
                formData = new FormData(form),
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

    // https://github.com/samdutton/simpl/blob/gh-pages/getusermedia/sources/js/mai

    function getDevices() {
        // AFAICT in Safari this only gets default devices until gUM is called :/
        return navigator.mediaDevices.enumerateDevices();
    }

    function handleError(error) {
        console.error('Error: ', error);
    }

    function QorMediaDevices ($audioSelect, $videoSelect, $videoElement) {
        this.$audioSelect = $audioSelect;
        this.$videoSelect = $videoSelect;
        this.$videoElement = $videoElement;

        this.init();
    }

    QorMediaDevices.Available = function () {
        if (!navigator.mediaDevices || !navigator.mediaDevices.enumerateDevices) {
            console.log("enumerateDevices() not supported.");
            return false;
        }

// List cameras and microphones.

        navigator.mediaDevices.enumerateDevices()
            .then(function(devices) {
                devices.forEach(function(device) {
                    console.log(device.kind + ": " + device.label +
                        " id = " + device.deviceId);
                });
            })
            .catch(function(err) {
                console.log(err.name + ": " + err.message);
            });
    }

    QorMediaDevices.prototype = {
        constructor: QorMediaDevices,

        init : function () {

        },

        gotDevices: function (deviceInfos) {
            this.$audioSelect.html('');
            this.$videoSelect.html('');

            window.deviceInfos = deviceInfos; // make available to console
            console.log('Available input and output devices:', deviceInfos);
            for (const deviceInfo of deviceInfos) {
                const option = document.createElement('option');
                option.value = deviceInfo.deviceId;
                if (deviceInfo.kind === 'audioinput') {
                    if (this.$audioSelect) {
                        option.text = deviceInfo.label || `Microphone ${audioSelect.length + 1}`;
                        this.$audioSelect.appendChild(option);
                    }
                } else if (deviceInfo.kind === 'videoinput') {
                    if (this.$videoSelect) {
                        option.text = deviceInfo.label || `Camera ${videoSelect.length + 1}`;
                        this.$videoSelect.appendChild(option);
                    }
                }
            }
        },

        getStream: function () {
            if (window.stream) {
                window.stream.getTracks().forEach(track => {
                    track.stop();
                });
            }
            const audioSource = this.$audioSelect ? this.$audioSelect.val() : null;
            const videoSource = this.$videoSelect ? this.$videoSelect.val() : null;
            const constraints = {
                audio: {deviceId: audioSource ? {exact: audioSource} : undefined},
                video: {deviceId: videoSource ? {exact: videoSource} : undefined}
            };
            return navigator.mediaDevices.getUserMedia(constraints).
            then(this.gotStream.bind(this)).catch(handleError);
        },

        gotStream: function (stream) {
            window.stream = stream; // make stream available to console
            this.$audioSelect.selectedIndex = [...audioSelect.options].
            findIndex(option => option.text === stream.getAudioTracks()[0].label);
            this.$videoSelect.selectedIndex = [...videoSelect.options].
            findIndex(option => option.text === stream.getVideoTracks()[0].label);
            this.$videoElement.srcObject = stream;
        }
    }

    window.QorMediaDevices = QorMediaDevices;
});
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

    let location = window.location,
        NAMESPACE = 'qor.content_copy',
        EVENT_ENABLE = 'enable.' + NAMESPACE,
        EVENT_DISABLE = 'disable.' + NAMESPACE,
        EVENT_CLICK = 'click.' + NAMESPACE;

    function QorContentCopy(element, options) {
        this.$element = $(element);
        this.options = $.extend({}, QorContentCopy.DEFAULTS, $.isPlainObject(options) && options);
        this.init();
    }

    QorContentCopy.prototype = {
        constructor: QorContentCopy,

        init: function() {
            this.bind();
        },

        bind: function() {
            var options = this.options;
            this.$element.attr('href', 'javascript:void(0);')
                .html('<i class="material-icons">content_copy</i>')
                .on(EVENT_CLICK, options.label, $.proxy(this.do, this));
        },

        unbind: function() {
            this.$element.off(EVENT_CLICK, this.do);
        },

        do: function(e) {
            let $this = $(e.currentTarget),
                value = $this.data('value');
            if (!value) {
                value = $this.data('value-b64');
                if (value) {
                    value = atob(value)
                }
            }
            if (!value) {
                let $el = $(this).parent().find("[data-content-copy-value]");
                if ($el.length === 0) {
                    return
                }
                value = $el.text()
            }
            let $temp = $("<input style='position: absolute; top: -200px'>");
            $("body").append($temp);
            $temp.val(value).select();
            document.execCommand("copy");
            $temp.remove();
        },

        destroy: function() {
            this.unbind();
            this.$element.removeData(NAMESPACE);
        }
    };

    QorContentCopy.DEFAULTS = {};

    QorContentCopy.plugin = function(options) {
        return this.each(function() {
            var $this = $(this);
            var data = $this.data(NAMESPACE);
            var fn;

            if (!data) {
                if (/destroy/.test(options)) {
                    return;
                }

                $this.data(NAMESPACE, (data = new QorContentCopy(this, options)));
            }

            if (typeof options === 'string' && $.isFunction((fn = data[options]))) {
                fn.apply(data);
            }
        });
    };

    $(function() {
        var selector = '[data-content-copy]';
        var options = {};

        $(document)
            .on(EVENT_DISABLE, function(e) {
                QorContentCopy.plugin.call($(selector, e.target), 'destroy');
            })
            .on(EVENT_ENABLE, function(e) {
                QorContentCopy.plugin.call($(selector, e.target), options);
            })
            .triggerHandler(EVENT_ENABLE);
    });

    return QorContentCopy;
});
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

    let URL = window.URL || window.webkitURL,
        NAMESPACE = 'qor.cropper',
        // Events
        EVENT_ENABLE = 'enable.' + NAMESPACE,
        EVENT_DISABLE = 'disable.' + NAMESPACE,
        EVENT_CHANGE = 'change.' + NAMESPACE,
        EVENT_CLICK = 'click.' + NAMESPACE,
        EVENT_SHOWN = 'shown.qor.modal',
        EVENT_HIDDEN = 'hidden.qor.modal',
        // Classes
        CLASS_TOGGLE = '.qor-cropper__toggle',
        CLASS_CANVAS = '.qor-cropper__canvas',
        CLASS_WRAPPER = '.qor-cropper__wrapper',
        CLASS_OPTIONS = '.qor-cropper__options',
        CLASS_SAVE = '.qor-cropper__save',
        CLASS_DELETE = '.qor-cropper__toggle--delete',
        CLASS_CROP = '.qor-cropper__toggle--crop',
        CLASS_UNDO = '.qor-fieldset__undo',
        HIDDEN_DATA_INPUT = 'input[name="QorResource.MediaOption"]:hidden';

    function capitalize(str) {
        if (typeof str === 'string') {
            str = str.charAt(0).toUpperCase() + str.substr(1);
        }

        return str;
    }

    function getLowerCaseKeyObject(obj) {
        let newObj = {},
            key;

        if ($.isPlainObject(obj)) {
            for (key in obj) {
                if (obj.hasOwnProperty(key)) {
                    newObj[String(key).toLowerCase()] = obj[key];
                }
            }
        }

        return newObj;
    }

    function getValueByNoCaseKey(obj, key) {
        let originalKey = String(key),
            lowerCaseKey = originalKey.toLowerCase(),
            upperCaseKey = originalKey.toUpperCase(),
            capitalizeKey = capitalize(originalKey);

        if ($.isPlainObject(obj)) {
            return obj[lowerCaseKey] || obj[capitalizeKey] || obj[upperCaseKey];
        }
    }

    function replaceText(str, data) {
        if (typeof str === 'string') {
            if (typeof data === 'object') {
                $.each(data, function(key, val) {
                    str = str.replace('$[' + String(key).toLowerCase() + ']', val);
                });
            }
        }

        return str;
    }

    function isSVG(url) {
        return /.svg$/.test(url);
    }



    function QorCropper(element, options) {
        this.$element = $(element);
        this.options = $.extend(true, {}, QorCropper.DEFAULTS, $.isPlainObject(options) && options);
        this.data = null;
        this.init();
    }

    QorCropper.prototype = {
        constructor: QorCropper,

        init: function() {
            let options = this.options,
                $this = this.$element,
                $parent = $this.closest(options.parent),
                $form = $parent.closest('form'),
                // $takePicture = $form.find(`#${$this.attr('id')}-take-picture`),
                data,
                outputValue,
                fetchUrl,
                imageData;

            if (!$parent.length) {
                $parent = $this.parent();
            }

            this.$parent = $parent;
            this.$takePicture = null //this.$takePicture = $takePicture.length && vigator.mediaDevices.getUserMedia ? $takePicture : null;
            this.$output = $parent.find(options.output);
            this.$formCropInput = $form.find(HIDDEN_DATA_INPUT);
            this.$list = $parent.find(options.list);

            fetchUrl = this.$output.data('fetchSizedata');

            if (fetchUrl) {
                $.getJSON(fetchUrl, function(data) {
                    imageData = JSON.parse(data.MediaOption);
                    this.$output.val(JSON.stringify(data));
                    this.$formCropInput.val(JSON.stringify(data));
                    this.data = imageData || {};
                    if (isSVG(imageData.URL || imageData.Url)) {
                        this.resetImage();
                    }
                    this.build();
                    this.bind();
                }.bind(this));
            } else {
                outputValue = $.trim(this.$output.val());
                if (outputValue) {
                    data = JSON.parse(outputValue);
                    if (isSVG(data.URL || data.Url)) {
                        this.resetImage();
                    }
                }

                this.data = data || {};

                this.build();
                this.bind();
            }
        },

        resetImage: function() {
            this.$parent.addClass('is-svg');
        },

        build: function() {
            let textData = this.$output.data(),
                text = {},
                replaceTexts;

            if (textData) {
                text = {
                    title: textData.cropperTitle,
                    ok: textData.cropperOk,
                    cancel: textData.cropperCancel
                };
                replaceTexts = this.options.text;
            }

            if (text.ok && text.title && text.cancel) {
                replaceTexts = text;
            }

            this.wrap();
            this.$modal = $(replaceText(QorCropper.MODAL, replaceTexts)).appendTo('body');

            if (this.$takePicture) {
                let text;
                if (textData) {
                    text = {
                        title: textData.tpTitle,
                        ok: textData.tpOk,
                        cancel: textData.tpCancel
                    };
                    replaceTexts = this.options.takePictureText;
                }

                if (text.ok && text.title && text.cancel) {
                    replaceTexts = text;
                }
                this.$takePictureModal = $(replaceText(QorCropper.TAKE_PICTURE, replaceTexts)).appendTo('body');
                this.$takePicture.show();
            }
        },

        unbuild: function() {
            this.$modal.remove();
            this.unwrap();
        },

        wrap: function() {
            let $list = this.$list,
                $img;

            $img = $list.find('img').not('.is-svg');

            if ($img.length) {
                $list.find('li').append(QorCropper.TOGGLE);
                $img.wrap(QorCropper.CANVAS);
                this.center($img);
            } else {
                $list.find(CLASS_CROP).remove();
            }
        },

        unwrap: function() {
            let $list = this.$list;

            $list.find(CLASS_TOGGLE).remove();
            $list.find(CLASS_CANVAS).each(function() {
                let $this = $(this);

                $this.before($this.html()).remove();
            });
        },

        bind: function() {
            this.$element.on(EVENT_CHANGE, $.proxy(this.read, this));
            this.$list.on(EVENT_CLICK, $.proxy(this.click, this));
            this.$modal.on(EVENT_SHOWN, $.proxy(this.start, this)).on(EVENT_HIDDEN, $.proxy(this.stop, this));
        },

        unbind: function() {
            this.$element.off(EVENT_CHANGE, this.read);
            this.$list.off(EVENT_CLICK, this.click);
            this.$modal.off(EVENT_SHOWN, this.start).off(EVENT_HIDDEN, this.stop);
        },

        click: function(e) {
            let target = e.target,
                $target,
                data = this.data,
                $alert;

            if (target === this.$list[0]) {
                return;
            }

            $target = $(target);

            if ($target.closest(CLASS_DELETE).length) {
                data.Delete = true;

                this.$output.val(JSON.stringify(data));
                this.$formCropInput.val(JSON.stringify(data));

                this.$list.hide();
                this.$parent.find('label').show();

                $alert = $(QorCropper.ALERT);
                $alert.find(CLASS_UNDO).one(
                    EVENT_CLICK,
                    function() {
                        $alert.remove();
                        this.$list.show();
                        delete data.Delete;
                        this.$output.val(JSON.stringify(data));
                        this.$formCropInput.val(JSON.stringify(data));
                    }.bind(this)
                );
                this.$parent.find('.qor-fieldset').append($alert);
            }

            if ($target.closest(CLASS_CROP).length) {
                $target = $target.closest('li').find('img');
                this.$target = $target;
                this.$modal.qorModal('show');
            }
        },

        read: function(e) {
            let files = e.target.files,
                file,
                $list = this.$list,
                $alert = this.$parent.find('.qor-fieldset__alert');

            $list.show();

            if ($alert.length) {
                $alert.remove();
            }

            if (files && files.length) {
                file = files[0];

                if (/^image\//.test(file.type) && URL) {
                    this.fileType = file.type;
                    this.load(URL.createObjectURL(file));
                    this.$parent.find('.qor-medialibrary__image-desc').show();
                } else {
                    $list.empty().html(QorCropper.FILE_LIST.replace('{{filename}}', file.name));
                }
            }
        },

        takePicture: function(e) {
            let files = e.target.files,
                file,
                $list = this.$list,
                $alert = this.$parent.find('.qor-fieldset__alert');

            $list.show();

            if ($alert.length) {
                $alert.remove();
            }

            if (files && files.length) {
                file = files[0];

                if (/^image\//.test(file.type) && URL) {
                    this.fileType = file.type;
                    this.load(URL.createObjectURL(file));
                    this.$parent.find('.qor-medialibrary__image-desc').show();
                } else {
                    $list.empty().html(QorCropper.FILE_LIST.replace('{{filename}}', file.name));
                }
            }
        },

        load: function(url, fromExternal, callback) {
            let options = this.options,
                _this = this,
                $list = this.$list,
                $ul = $(QorCropper.LIST),
                data = this.data || {},
                fileType = this.fileType,
                $image,
                imageLength;

            // media box will use load method, has it's own html structure.
            if (!fromExternal) {
                $list.find('ul').remove();
                $list.html($ul);
            }

            $image = $list.find('img');
            this.wrap();

            imageLength = $image.length;
            $image
                .one('load', function() {
                    if (fileType === 'image/svg+xml') {
                        $list.find(CLASS_TOGGLE).remove();
                        return false;
                    }

                    let $this = $(this),
                        naturalWidth = this.naturalWidth,
                        naturalHeight = this.naturalHeight,
                        sizeData = $this.data(),
                        sizeResolution = sizeData.sizeResolution,
                        sizeName = sizeData.sizeName,
                        emulateImageData = {},
                        emulateCropData = {},
                        aspectRatio,
                        width = sizeData.sizeResolutionWidth,
                        height = sizeData.sizeResolutionHeight;

                    if (sizeResolution) {
                        if (!width && !height) {
                            width = getValueByNoCaseKey(sizeResolution, 'width');
                            height = getValueByNoCaseKey(sizeResolution, 'height');
                        }
                        aspectRatio = width / height;

                        if (naturalHeight * aspectRatio > naturalWidth) {
                            width = naturalWidth;
                            height = width / aspectRatio;
                        } else {
                            height = naturalHeight;
                            width = height * aspectRatio;
                        }

                        emulateImageData = {
                            naturalWidth: naturalWidth,
                            naturalHeight: naturalHeight
                        };

                        emulateCropData = {
                            x: Math.round((naturalWidth - width) / 2),
                            y: Math.round((naturalHeight - height) / 2),
                            width: Math.round(width),
                            height: Math.round(height)
                        };

                        _this.preview($this, emulateImageData, emulateCropData);

                        if (sizeName) {
                            data.Crop = true;

                            if (!data[options.key]) {
                                data[options.key] = {};
                            }

                            if (sizeName != 'original') {
                                data[options.key][sizeName] = emulateCropData;
                            }
                        }
                    } else {
                        _this.center($this);
                    }

                    // Crop, CropOptions and Delete should be BOOL type, if empty should delete,
                    if (data.Crop === '' || !fromExternal) {
                        delete data.Crop;
                    }

                    if (!fromExternal) {
                        data.CropOptions = null;
                        delete data.Sizes;
                    }

                    delete data.Delete;

                    _this.$output.val(JSON.stringify(data));
                    _this.$formCropInput.val(JSON.stringify(data));

                    // callback after load complete
                    if (sizeName && data[options.key] && Object.keys(data[options.key]).length >= imageLength) {
                        if (callback && $.isFunction(callback)) {
                            callback();
                        }
                    }
                })
                .attr('data-cropper', 'data-cropper')
                .attr('src', url)
                .data('originalUrl', url);

            $list.show();
            this.$parent.find('label').hide();
        },

        start: function() {
            let options = this.options,
                $modal = this.$modal,
                $target = this.$target,
                sizeData = $target.data(),
                sizeName = sizeData.sizeName || 'original',
                sizeResolution = sizeData.sizeResolution,
                $clone = $(`<img data-cropper src=${sizeData.originalUrl}>`),
                data = this.data || {},
                _this = this,
                sizeAspectRatio = NaN,
                sizeWidth = sizeData.sizeResolutionWidth,
                sizeHeight = sizeData.sizeResolutionHeight,
                list;

            if (sizeResolution) {
                if (!sizeWidth && !sizeHeight) {
                    sizeWidth = getValueByNoCaseKey(sizeResolution, 'width');
                    sizeHeight = getValueByNoCaseKey(sizeResolution, 'height');
                }
                sizeAspectRatio = sizeWidth / sizeHeight;
            }

            if (!data[options.key]) {
                data[options.key] = {};
            }

            $modal
                .trigger('enable.qor.material')
                .find(CLASS_WRAPPER)
                .html($clone);

            list = this.getList(sizeAspectRatio);

            if (list) {
                $modal
                    .find(CLASS_OPTIONS)
                    .show()
                    .append(list);
            }

            $clone.cropper({
                aspectRatio: sizeAspectRatio,
                data: getLowerCaseKeyObject(data[options.key][sizeName]),
                background: false,
                movable: false,
                zoomable: false,
                scalable: false,
                rotatable: false,
                autoCropArea: 1,

                ready: function() {
                    $modal
                        .find('.qor-cropper__options-toggle')
                        .on(EVENT_CLICK, function() {
                            $modal.find('.qor-cropper__options-input').prop('checked', $(this).prop('checked'));
                        })
                        .prop('checked', true);

                    $modal.find(CLASS_SAVE).one(EVENT_CLICK, function() {
                        let cropData = $clone.cropper('getData', true),
                            croppedCanvas = $clone.cropper('getCroppedCanvas'),
                            syncData = [],
                            url;

                        data.Crop = true;
                        data[options.key][sizeName] = cropData;
                        _this.imageData = $clone.cropper('getImageData');
                        _this.cropData = cropData;

                        if (croppedCanvas) {
                            try {
                                url = croppedCanvas.toDataURL();
                            } catch (error) {
                                console.log(error);
                                console.log('Please check image Cross-origin setting');
                            }
                        }

                        $modal.find(CLASS_OPTIONS + ' input').each(function() {
                            let $this = $(this);

                            if ($this.prop('checked')) {
                                syncData.push($this.attr('name'));
                            }
                        });

                        _this.output(url, syncData);
                        $modal.qorModal('hide');
                    });
                }
            });
        },

        stop: function() {
            this.$modal
                .trigger('disable.qor.material')
                .find(CLASS_WRAPPER + ' > img')
                .cropper('destroy')
                .remove()
                .end()
                .find(CLASS_OPTIONS)
                .hide()
                .find('ul')
                .remove();
        },

        getList: function(aspectRatio) {
            let list = [];

            this.$list
                .find('img')
                .not(this.$target)
                .each(function() {
                    let data = $(this).data(),
                        resolution = data.sizeResolution,
                        name = data.sizeName,
                        width = data.sizeResolutionWidth,
                        height = data.sizeResolutionHeight;

                    if (resolution) {
                        if (!width && !height) {
                            width = getValueByNoCaseKey(resolution, 'width');
                            height = getValueByNoCaseKey(resolution, 'height');
                        }

                        if (width / height === aspectRatio) {
                            list.push(
                                '<label>' +
                                '<input class="qor-cropper__options-input" type="checkbox" name="' +
                                name +
                                '" checked> ' +
                                '<span>' +
                                name +
                                '<small>(' +
                                width +
                                '&times;' +
                                height +
                                ' px)</small>' +
                                '</span>' +
                                '</label>'
                            );
                        }
                    }
                });

            return list.length ? '<ul><li>' + list.join('</li><li>') + '</li></ul>' : '';
        },

        output: function(url, data) {
            let $target = this.$target;

            if (url) {
                this.center($target.attr('src', url), true);
            } else {
                this.preview($target);
            }

            if ($.isArray(data) && data.length) {
                this.autoCrop(url, data);
            }

            this.$output.val(JSON.stringify(this.data)).trigger(EVENT_CHANGE);
            this.$formCropInput.val(JSON.stringify(this.data));
        },

        preview: function($target, emulateImageData, emulateCropData) {
            let $canvas = $target.parent(),
                $container = $canvas.parent(),
                containerWidth = $container.width(),
                containerHeight = $container.height(),
                imageData = emulateImageData || this.imageData,
                cropData = $.extend({}, emulateCropData || this.cropData), // Clone one to avoid changing it
                aspectRatio = cropData.width / cropData.height,
                canvasWidth = containerWidth,
                scaledRatio;

            if (canvasWidth == 0 || imageData.naturalWidth == 0 || imageData.naturalHeight == 0) {
                return;
            }

            if (containerHeight * aspectRatio <= containerWidth) {
                canvasWidth = containerHeight * aspectRatio;
            }

            scaledRatio = cropData.width / canvasWidth;

            $target.css({
                maxWidth: imageData.naturalWidth / scaledRatio,
                maxHeight: imageData.naturalHeight / scaledRatio
            });

            this.center($target);
        },

        center: function($target, reset) {
            $target.each(function() {
                let $this = $(this),
                    $canvas = $this.parent(),
                    $container = $canvas.parent();

                function center() {
                    let containerHeight = $container.height(),
                        canvasHeight = $canvas.height(),
                        marginTop = 'auto';

                    if (canvasHeight < containerHeight) {
                        marginTop = (containerHeight - canvasHeight) / 2;
                    }

                    $canvas.css('margin-top', marginTop);
                }

                if (reset) {
                    $canvas.add($this).removeAttr('style');
                }

                if (this.complete) {
                    center.call(this);
                } else {
                    this.onload = center;
                }
            });
        },

        autoCrop: function(url, data) {
            let cropData = this.cropData,
                cropOptions = this.data[this.options.key],
                _this = this;

            this.$list
                .find('img')
                .not(this.$target)
                .each(function() {
                    let $this = $(this),
                        sizeName = $this.data('sizeName');

                    if ($.inArray(sizeName, data) > -1) {
                        cropOptions[sizeName] = $.extend({}, cropData);

                        if (url) {
                            _this.center($this.attr('src', url), true);
                        } else {
                            _this.preview($this);
                        }
                    }
                });
        },

        destroy: function() {
            if (!isSVG) {
                this.unbind();
                this.unbuild();
            }
            this.$element.removeData(NAMESPACE);
        }
    };

    QorCropper.DEFAULTS = {
        parent: false,
        output: false,
        list: false,
        key: 'data',
        data: null,
        text: {
            title: 'Crop the image',
            ok: 'OK',
            cancel: 'Cancel'
        }
    };

    QorCropper.TOGGLE = `<div class="qor-cropper__toggle">
            <div class="qor-cropper__toggle--crop"><i class="material-icons">crop</i></div>
            <div class="qor-cropper__toggle--delete"><i class="material-icons">delete</i></div>
        </div>`;

    QorCropper.ALERT = `<div class="qor-fieldset__alert">
            <button class="mdl-button mdl-button--accent qor-fieldset__undo" type="button">Undo delete</button>
        </div>`;

    QorCropper.CANVAS = '<div class="qor-cropper__canvas"></div>';
    QorCropper.LIST = '<ul><li><img data-cropper></li></ul>';
    QorCropper.FILE_LIST = `<div class="qor-file__list-item">
                                <span><span>{{filename}}</span></span>
                                <div class="qor-cropper__toggle">
                                    <div class="qor-cropper__toggle--delete"><i class="material-icons">delete</i></div>
                                </div>
                            </div>`;
    QorCropper.MODAL = `<div class="qor-modal fade" tabindex="-1" role="dialog" aria-hidden="true">
            <div class="mdl-card mdl-shadow--2dp" role="document">
                <div class="mdl-card__title">
                    <h2 class="mdl-card__title-text">$[title]</h2>
                </div>
                <div class="mdl-card__supporting-text">
                    <div class="qor-cropper__wrapper"></div>
                    <div class="qor-cropper__options">
                        <p>Sync cropping result to: <label><input type="checkbox" class="qor-cropper__options-toggle" checked/> All</label></p>
                    </div>
                </div>
                <div class="mdl-card__actions mdl-card--border">
                    <a class="mdl-button mdl-button--colored mdl-button--raised qor-cropper__save">$[ok]</a>
                    <a class="mdl-button mdl-button--colored" data-dismiss="modal">$[cancel]</a>
                </div>
                <div class="mdl-card__menu">
                    <button class="mdl-button mdl-button--icon" data-dismiss="modal" aria-label="close">
                        <i class="material-icons">close</i>
                    </button>
                </div>
            </div>
        </div>`;
    QorCropper.TAKE_PICTURE = `<div class="qor-modal fade" tabindex="-1" role="dialog" aria-hidden="true">
            <div class="mdl-card mdl-shadow--2dp" role="document">
                <div class="mdl-card__title">
                    <h2 class="mdl-card__title-text">$[title]</h2>
                </div>
                <div class="mdl-card__supporting-text">
                    <video autoplay="true" />
                </div>
                <div class="mdl-card__actions mdl-card--border">
                    <a class="mdl-button mdl-button--colored mdl-button--raised qor-cropper__save">$[ok]</a>
                    <a class="mdl-button mdl-button--colored" data-dismiss="modal">$[cancel]</a>
                </div>
                <div class="mdl-card__menu">
                    <button class="mdl-button mdl-button--icon" data-dismiss="modal" aria-label="close">
                        <i class="material-icons">close</i>
                    </button>
                </div>
            </div>
        </div>`;

    QorCropper.plugin = function(option) {
        return this.each(function() {
            let $this = $(this),
                data = $this.data(NAMESPACE),
                options,
                fn;

            if (!data) {
                if (!$.fn.cropper) {
                    return;
                }

                if (/destroy/.test(option)) {
                    return;
                }

                options = $.extend(true, {}, $this.data(), typeof option === 'object' && option);
                $this.data(NAMESPACE, (data = new QorCropper(this, options)));
            }

            if (typeof option === 'string' && $.isFunction((fn = data[option]))) {
                fn.apply(data);
            }
        });
    };

    $(function() {
        let selector = ".qor-file__input:not([data-cropper='disabled'])",
            options = {
                parent: '.qor-file',
                output: '.qor-file__options',
                list: '.qor-file__list',
                key: 'CropOptions'
            };

        $(document)
            .on(EVENT_ENABLE, function(e) {
                QorCropper.plugin.call($(selector, e.target), options);
            })
            .on(EVENT_DISABLE, function(e) {
                QorCropper.plugin.call($(selector, e.target), 'destroy');
            })
            .triggerHandler(EVENT_ENABLE);
    });

    return QorCropper;
});

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

    var NAMESPACE = 'qor.datepicker';
    var EVENT_ENABLE = 'enable.' + NAMESPACE;
    var EVENT_DISABLE = 'disable.' + NAMESPACE;
    var EVENT_CHANGE = 'pick.' + NAMESPACE;
    var EVENT_CLICK = 'click.' + NAMESPACE;

    var CLASS_EMBEDDED = '.qor-datepicker__embedded';
    var CLASS_SAVE = '.qor-datepicker__save';
    var CLASS_PARENT = '[data-picker-type]';

    function replaceText(str, data) {
        if (typeof str === 'string') {
            if (typeof data === 'object') {
                $.each(data, function (key, val) {
                    str = str.replace('$[' + String(key).toLowerCase() + ']', val);
                });
            }
        }

        return str;
    }

    $.extend(true, QOR.messages, {
        datepicker: {
            datepicker: {
                title: 'Pick a date',
                ok: 'OK',
                cancel: 'Cancel'
            }
        }
    });

    function QorDatepicker(element, options) {
        this.$element = $(element);
        this.options = $.extend(true, {}, {
                text: QOR.messages.datepicker.datepicker
            },
            ($.isPlainObject(options) && options)
        );
        this.date = null;
        this.formatDate = null;
        this.built = false;
        this.pickerData = this.$element.data();
        this.init();
    }

    QorDatepicker.prototype = {
        init: function () {
            this.bind();
        },

        bind: function () {
            this.$element.on(EVENT_CLICK, $.proxy(this.show, this));
        },

        unbind: function () {
            this.$element.off(EVENT_CLICK, this.show);
        },

        build: function () {
            let $modal,
                $ele = this.$element,
                data = this.pickerData,
                date = $ele.val() ? new Date($ele.val()) : new Date(),
                datepickerOptions = {
                    date: date,
                    inline: true
                },
                parent = $ele.closest(CLASS_PARENT);
            if (data.format) {
                datepickerOptions.format = data.format
            }

            let $targetInput = parent.find(data.targetInput);

            if (this.built) {
                return;
            }

            this.$modal = $modal = $(replaceText(QorDatepicker.TEMPLATE, this.options.text)).appendTo('body');

            if ($targetInput.length) {
                let val = $targetInput.val(), date;
                if (val.length) {
                    datepickerOptions.date = datepickerOptions.format ? $.fn.qorDatepicker.parseDate(datepickerOptions.format, val): new Date(val);
                } else {
                    datepickerOptions.date = new Date();
                }
            }

            if (data.targetInput && $targetInput.data('start-date')) {
                datepickerOptions.startDate = new Date();
            }

            $modal.find(CLASS_EMBEDDED).on(EVENT_CHANGE, $.proxy(this.change, this)).qorDatepicker(datepickerOptions).triggerHandler(EVENT_CHANGE);

            $modal.find(CLASS_SAVE).on(EVENT_CLICK, $.proxy(this.pick, this));

            this.built = true;
        },

        unbuild: function () {
            if (!this.built) {
                return;
            }

            this.$modal.find(CLASS_EMBEDDED).off(EVENT_CHANGE, this.change).qorDatepicker('destroy').end().find(CLASS_SAVE).off(EVENT_CLICK, this.pick).end().remove();
        },

        change: function (e) {
            var $modal = this.$modal;
            var $target = $(e.target);
            var date;

            this.date = date = $target.qorDatepicker('getDate');
            this.formatDate = $target.qorDatepicker('getDate', true);

            $modal.find('.qor-datepicker__picked-year').text(date.getFullYear());
            $modal
                .find('.qor-datepicker__picked-date')
                .text(
                    [$target.qorDatepicker('getDayName', date.getDay(), true) + ',', String($target.qorDatepicker('getMonthName', date.getMonth(), true)), date.getDate()].join(' ')
                );
        },

        show: function () {
            if (!this.built) {
                this.build();
            }

            this.$modal.qorModal('show');
        },

        pick: function () {
            let targetInputClass = this.pickerData.targetInput,
                $element = this.$element,
                $parent = $element.closest(CLASS_PARENT),
                $targetInput = targetInputClass ? $parent.find(targetInputClass) : $element,
                pickerType = $parent.data('picker-type'),
                newValue = this.formatDate;

            if (pickerType === 'datetime') {
                var regDate = /^\d{4}-\d{1,2}-\d{1,2}/;
                var oldValue = $targetInput.val();
                var hasDate = regDate.test(oldValue);

                if (hasDate) {
                    newValue = oldValue.replace(regDate, newValue);
                } else {
                    newValue = newValue + ' 00:00';
                }
            }
            $targetInput.val(newValue).trigger('change');
            this.$modal.qorModal('hide');
        },

        destroy: function () {
            this.unbind();
            this.unbuild();
            this.$element.removeData(NAMESPACE);
        }
    };

    QorDatepicker.TEMPLATE = `<div class="qor-modal fade qor-datepicker" tabindex="-1" role="dialog" aria-hidden="true">
            <div class="mdl-card mdl-shadow--2dp" role="document">
                <div class="mdl-card__title">
                    <h2 class="mdl-card__title-text">$[title]</h2>
                </div>
                <div class="mdl-card__supporting-text">
                    <div class="qor-datepicker__picked">
                        <div class="qor-datepicker__picked-year"></div>
                        <div class="qor-datepicker__picked-date"></div>
                    </div>
                    <div class="qor-datepicker__embedded"></div>
                </div>
                <div class="mdl-card__actions">
                    <a class="mdl-button mdl-button--colored  mdl-button--raised qor-datepicker__save">$[ok]</a>
                    <a class="mdl-button mdl-button--colored " data-dismiss="modal">$[cancel]</a>
                </div>
            </div>
        </div>`;

    QorDatepicker.plugin = function (option) {
        return this.each(function () {
            var $this = $(this);
            var data = $this.data(NAMESPACE);
            var options;
            var fn;

            if (!data) {
                if (!$.fn.qorDatepicker) {
                    return;
                }

                if (/destroy/.test(option)) {
                    return;
                }

                options = $.extend(true, {}, $this.data(), typeof option === 'object' && option);
                $this.data(NAMESPACE, (data = new QorDatepicker(this, options)));
            }

            if (typeof option === 'string' && $.isFunction((fn = data[option]))) {
                fn.apply(data);
            }
        });
    };

    let $document = $(document);

    $document.data('qor.datepicker', function (cb) {
        return cb.call(QorDatepicker);
    });

    $(function () {
        var selector = '[data-toggle="qor.datepicker"]';

        $document
            .on(EVENT_DISABLE, function (e) {
                QorDatepicker.plugin.call($(selector, e.target), 'destroy');
            })
            .on(EVENT_ENABLE, function (e) {
                QorDatepicker.plugin.call($(selector, e.target));
            })
            .triggerHandler(EVENT_ENABLE);
    });

    return QorDatepicker;
});

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

    const NAMESPACE = 'qor.dialog',
        EVENT_ENABLE = 'enable.' + NAMESPACE,
        EVENT_DISABLE = 'disable.' + NAMESPACE,
        EVENT_CLICK = 'click.' + NAMESPACE,
        SELECTOR_DIALOG = 'dialog.mdl-dialog',
        SELECTOR_DIALOG_SHOW_BTN = '[data-dialog]';

    function Dialog(el) {
        this.dialog = el;
        this.$el = $(el);
        this.init();
    }

    Dialog.prototype = {
        init: function () {
            if (!this.dialog.showModal) {
                dialogPolyfill.registerDialog(this.dialog);
            }
            this.$closers = this.$el.find('> .mdl-dialog__actions .close');
            this.bind();
        },

        bind: function () {
            this.$closers.on(EVENT_CLICK, this.hide.bind(this))
        },

        unbind: function () {
            if (this.$closers) this.$closers.off(EVENT_CLICK);
        },

        hide: function () {
            this.dialog.close()
        },

        show: function () {
            this.dialog.showModal()
        },

        destroy: function () {
            this.unbind();
            this.$el = this.$closers = this.dialog = null;
        }
    }

    Dialog.plugin = function (option) {
        return this.each(function () {
            const $this = $(this);
            let data = $this.data(NAMESPACE),
                options,
                fn;

            if (!data) {
                if (/destroy/.test(option)) {
                    return;
                }

                options = $.extend(true, {}, $this.data(), typeof option === 'object' && option);
                $this.data(NAMESPACE, (data = new Dialog(this, options)));
            } else if (/destroy/.test(option)) {
                $this.removeData(NAMESPACE)
            }

            if (typeof option === 'string' && $.isFunction((fn = data[option]))) {
                fn.apply(data);
            }
        });
    };


    function OpenDialog(el, options) {
        this.$el = $(el);
        this.init();
    }

    OpenDialog.prototype = {
        init: function (options) {
            let dialog = options && options.target || this.$el.data().dialog,
                $dialog = dialog && $(dialog) || null;
            if (!$dialog) {
                return
            }
            this.$dialog = $dialog;
            this.bind();
        },

        bind: function () {
            this.$el.on(EVENT_CLICK, this.show.bind(this))
        },

        unbind: function () {
            this.$el.off(EVENT_CLICK);
        },

        show: function () {
            this.$dialog.data(NAMESPACE).show()
        },

        destroy: function () {
            this.unbind();
            this.$el = this.$dialog = null;
        }
    }

    OpenDialog.plugin = function (option) {
        return this.each(function () {
            const $this = $(this);
            let data = $this.data(NAMESPACE),
                options,
                fn;

            if (!data) {
                if (/destroy/.test(option)) {
                    return;
                }

                options = $.extend(true, {}, $this.data(), typeof option === 'object' && option);
                $this.data(NAMESPACE, (data = new OpenDialog(this, options)));
            } else if (/destroy/.test(option)) {
                $this.removeData(NAMESPACE)
            }

            if (typeof option === 'string' && $.isFunction((fn = data[option]))) {
                fn.apply(data);
            }
        });
    };

    $(function () {
        $(document)
            .on(EVENT_ENABLE, function (e) {
                Dialog.plugin.call($(SELECTOR_DIALOG, e.target));
                OpenDialog.plugin.call($(SELECTOR_DIALOG_SHOW_BTN, e.target));
            })
            .on(EVENT_DISABLE, function (e) {
                Dialog.plugin.call($(SELECTOR_DIALOG, e.target), 'destroy');
                OpenDialog.plugin.call($(SELECTOR_DIALOG_SHOW_BTN, e.target), 'destroy');
            })
            .triggerHandler(EVENT_ENABLE)
    });
});

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

    var dirtyForm = function(ele, options) {
        var hasChangedObj = false;

        if (this instanceof jQuery) {
            options = ele;
            ele = this;
        } else if (!(ele instanceof jQuery)) {
            ele = $(ele);
        }

        ele.each(function(item, element) {
            var $ele = $(element);

            if ($ele.is('form')) {
                if ($ele.hasClass('ignore-dirtyform')) {
                    return false;
                }
                hasChangedObj = dirtyForm(
                    $ele.find(
                        'input:not([type="hidden"]):not(".search-field input"):not(".chosen-search input"):not(".ignore-dirtyform"), textarea, select'
                    ),
                    options
                );
                if (hasChangedObj) {
                    return false;
                }
            } else if ($ele.is(':checkbox') || $ele.is(':radio')) {
                if ($ele.hasClass('ignore-dirtyform')) {
                    return false;
                }

                if (element.checked != element.defaultChecked) {
                    hasChangedObj = true;
                    return false;
                }
            } else if ($ele.is('input') || $ele.is('textarea')) {
                if ($ele.hasClass('ignore-dirtyform')) {
                    return false;
                }

                if (element.value != element.defaultValue) {
                    hasChangedObj = true;
                    return false;
                }
            } else if ($ele.is('select')) {
                if ($ele.hasClass('ignore-dirtyform')) {
                    return false;
                }

                var option;
                var defaultSelectedIndex = 0;
                var numberOfOptions = element.options.length;

                for (var i = 0; i < numberOfOptions; i++) {
                    option = element.options[i];
                    hasChangedObj = hasChangedObj || option.selected != option.defaultSelected;
                    if (option.defaultSelected) {
                        defaultSelectedIndex = i;
                    }
                }

                if (hasChangedObj && !element.multiple) {
                    hasChangedObj = defaultSelectedIndex != element.selectedIndex;
                }

                if (hasChangedObj) {
                    return false;
                }
            }
        });

        return hasChangedObj;
    };

    $.fn.extend({
        dirtyForm: dirtyForm
    });

    $(function() {
        $(document).on('submit', 'form', function() {
            window.onbeforeunload = null;
            $.fn.qorSlideoutBeforeHide = null;
        });

        $(document).on('change', 'form', function() {
            if ($(this).dirtyForm()) {
                $.fn.qorSlideoutBeforeHide = true;
                window.onbeforeunload = function() {
                    return 'You have unsaved changes on this page. If you leave this page, you will lose all unsaved changes.';
                };
            } else {
                $.fn.qorSlideoutBeforeHide = null;
                window.onbeforeunload = null;
            }
        });
    });
});

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

    const NAMESPACE = 'qor.layout_drawer_fixer',
        SELECTOR = '.mdl-layout__drawer .qor-layout__sidebar',
        EVENT_ENABLE = 'enable.'+NAMESPACE,
        EVENT_DISABLE = 'disable.'+NAMESPACE,
        EVENT_RESIZE = 'resize.' + NAMESPACE,
        $WINDOW = $(window);

    function QorLayoutDrawerFixer(element) {
        this.$el = $(element);
        this.init();
    }

    QorLayoutDrawerFixer.prototype = {
        init: function () {
            this.$header = this.$el.find('.sidebar-header')
            this.$body = this.$el.find('.sidebar-body')
            this.$footer = this.$el.find('.sidebar-footer')

            this.bind();
            this.resize();
        },

        bind: function () {
            $WINDOW.on(EVENT_RESIZE, this.resize.bind(this));
        },

        unbind: function() {
            $WINDOW.off(EVENT_RESIZE, this.resize)
        },

        destroy: function () {
            this.unbind();
            this.$el = this.$header = $this.$body = this.$footer = null;
        },

        resize : function () {
            this.$body.height(this.$el.height()-this.$header.height()-this.$footer.height());
            this.$body.css('top', this.$header.height());
        }
    }


    QorLayoutDrawerFixer.plugin = function (options) {
        let args = Array.prototype.slice.call(arguments, 1),
            result;

        this.each(function () {
            let $this = $(this),
                data = $this.data(NAMESPACE),
                fn;

            if (!data) {
                if (/destroy/.test(options)) {
                    return;
                }

                if (typeof options === "object")
                    options = $.extend({}, options, true)

                data = new QorLayoutDrawerFixer(this, options);
                $this.data(NAMESPACE, data)
            }

            if (typeof options === 'string' && $.isFunction((fn = data[options]))) {
                result = fn.apply(data, args);
            }
        });

        return (typeof result === 'undefined') ? this : result;
    };

    $(function () {
        let options = {};

        $(document)
            .on(EVENT_DISABLE, function (e) {
                QorLayoutDrawerFixer.plugin.call($(SELECTOR, e.target), 'destroy');
            })
            .on(EVENT_ENABLE, function (e) {
                QorLayoutDrawerFixer.plugin.call($(SELECTOR, e.target), options);
            })
            .triggerHandler(EVENT_ENABLE);
    });
});
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

    let location = window.location,
        $document = $(document),
        NAMESPACE = 'qor.filter',
        EVENT_FILTER_CHANGE = 'filterChanged.' + NAMESPACE,
        EVENT_ENABLE = 'enable.' + NAMESPACE,
        EVENT_DISABLE = 'disable.' + NAMESPACE,
        EVENT_CLICK = 'click.' + NAMESPACE,
        CLASS_BOTTOMSHEETS = '.qor-bottomsheets',
        CLASS_DATE_START = '.qor-filter__start',
        CLASS_DATE_END = '.qor-filter__end',
        CLASS_SEARCH_PARAM = '[data-search-param]',
        CLASS_FILTER_SELECTOR = '.qor-filter__dropdown',
        CLASS_FILTER_TOGGLE = '.qor-filter-toggle',
        CLASS_IS_SELECTED = 'is-selected';

    function QorFilterTime(element, options) {
        this.$element = $(element);
        this.options = $.extend({}, QorFilterTime.DEFAULTS, $.isPlainObject(options) && options);
        this.init();
    }

    QorFilterTime.prototype = {
        constructor: QorFilterTime,

        init: function() {
            this.bind();
            let $element = this.$element,
                lcoal_moment = window.moment();

            this.$timeStart = $element.find(CLASS_DATE_START);
            this.$timeEnd = $element.find(CLASS_DATE_END);
            this.$searchParam = $element.find(CLASS_SEARCH_PARAM);
            this.$searchButton = $element.find(this.options.button);

            this.startWeekDate = lcoal_moment.startOf('isoweek').toDate();
            this.endWeekDate = lcoal_moment.endOf('isoweek').toDate();

            this.startMonthDate = lcoal_moment.startOf('month').toDate();
            this.endMonthDate = lcoal_moment.endOf('month').toDate();
            this.initActionTemplate();
        },

        bind: function() {
            var options = this.options;

            this.$element
                .on(EVENT_CLICK, options.trigger, this.show.bind(this))
                .on(EVENT_CLICK, options.label, this.setFilterTime.bind(this))
                .on(EVENT_CLICK, options.clear, this.clear.bind(this))
                .on(EVENT_CLICK, options.button, this.search.bind(this));

            $document.on(EVENT_CLICK, this.close);
        },

        unbind: function() {
            this.$element.off(EVENT_CLICK);
        },

        initActionTemplate: function() {
            var scheduleStartAt = this.getUrlParameter('schedule_start_at'),
                scheduleEndAt = this.getUrlParameter('schedule_end_at'),
                $filterToggle = $(this.options.trigger);

            if (scheduleStartAt || scheduleEndAt) {
                this.$timeStart.val(scheduleStartAt);
                this.$timeEnd.val(scheduleEndAt);

                scheduleEndAt = !scheduleEndAt ? '' : ' - ' + scheduleEndAt;
                $filterToggle
                    .addClass('active clearable')
                    .find('.qor-selector-label')
                    .html(scheduleStartAt + scheduleEndAt);
                $filterToggle.append('<i class="material-icons qor-selector-clear">clear</i>');
            }
        },

        show: function() {
            this.$element.find(CLASS_FILTER_SELECTOR).toggle();
        },

        close: function(e) {
            var $target = $(e.target),
                $filter = $(CLASS_FILTER_SELECTOR),
                filterVisible = $filter.is(':visible'),
                isInFilter = $target.closest(CLASS_FILTER_SELECTOR).length,
                isInToggle = $target.closest(CLASS_FILTER_TOGGLE).length,
                isInModal = $target.closest('.qor-modal').length,
                isInTimePicker = $target.closest('.ui-timepicker-wrapper').length;

            if (filterVisible && (isInFilter || isInToggle || isInModal || isInTimePicker)) {
                return;
            }
            $filter.hide();
        },

        setFilterTime: function(e) {
            let $target = $(e.target),
                data = $target.data(),
                range = data.filterRange,
                startTime,
                endTime,
                startDate,
                endDate;

            if (!range) {
                return false;
            }

            $(this.options.label).removeClass(CLASS_IS_SELECTED);
            $target.addClass(CLASS_IS_SELECTED);

            if (range == 'events') {
                this.$timeStart.val(data.scheduleStartAt || '');
                this.$timeEnd.val(data.scheduleEndAt || '');
                this.$searchButton.click();
                return false;
            }

            switch (range) {
                case 'today':
                    startDate = endDate = new Date();
                    break;
                case 'week':
                    startDate = this.startWeekDate;
                    endDate = this.endWeekDate;
                    break;
                case 'month':
                    startDate = this.startMonthDate;
                    endDate = this.endMonthDate;
                    break;
            }

            if (!startDate || !endDate) {
                return false;
            }

            startTime = this.getTime(startDate) + ' 00:00';
            endTime = this.getTime(endDate) + ' 23:59';

            this.$timeStart.val(startTime);
            this.$timeEnd.val(endTime);
            this.$searchButton.click();
        },

        getTime: function(dateNow) {
            var month = dateNow.getMonth() + 1,
                date = dateNow.getDate();

            month = month < 8 ? '0' + month : month;
            date = date < 10 ? '0' + date : date;

            return dateNow.getFullYear() + '-' + month + '-' + date;
        },

        clear: function() {
            var $trigger = $(this.options.trigger),
                $label = $trigger.find('.qor-selector-label');

            $trigger.removeClass('active clearable');
            $label.html($label.data('label'));
            this.$timeStart.val('');
            this.$timeEnd.val('');

            this.$searchButton.click();
            return false;
        },

        getUrlParameter: function(name) {
            let search = location.search,
                parameterName = name.replace(/[\[]/, '\\[').replace(/[\]]/, '\\]'),
                regex = new RegExp('[\\?&]' + parameterName + '=([^&#]*)'),
                results = regex.exec(search);

            return results === null ? '' : decodeURIComponent(results[1].replace(/\+/g, ' '));
        },

        updateQueryStringParameter: function(key, value, url) {
            let href = url || location.href,
                local_hash = href.match(/#\S*$/) || '',
                escapedkey = String(key).replace(/[\\^$*+?.()|[\]{}]/g, '\\$&'),
                re = new RegExp('([?&])' + escapedkey + '=.*?(&|$)', 'i'),
                separator = href.indexOf('?') !== -1 ? '&' : '?';

            if (local_hash) {
                local_hash = local_hash[0];
                href = href.replace(local_hash, '');
            }

            if (href.match(re)) {
                if (value) {
                    href = href.replace(re, '$1' + key + '=' + value + '$2');
                } else {
                    if (RegExp.$1 === '?' || RegExp.$1 === RegExp.$2) {
                        href = href.replace(re, '$1');
                    } else {
                        href = href.replace(re, '');
                    }
                }
            } else if (value) {
                href = href + separator + key + '=' + value;
            }

            return href + local_hash;
        },

        search: function() {
            var $searchParam = this.$searchParam,
                href = location.href,
                _this = this,
                type = 'qor.filter.time';

            if (!$searchParam.length) {
                return;
            }

            $searchParam.each(function() {
                var $this = $(this),
                    searchParam = $this.data().searchParam,
                    val = $this.val();

                href = _this.updateQueryStringParameter(searchParam, val, href);
            });

            if (this.$element.closest(CLASS_BOTTOMSHEETS).length) {
                $(CLASS_BOTTOMSHEETS).trigger(EVENT_FILTER_CHANGE, [href, type]);
            } else {
                location.href = href;
            }
        },

        destroy: function() {
            this.unbind();
            this.$element.removeData(NAMESPACE);
        }
    };

    QorFilterTime.DEFAULTS = {
        label: false,
        trigger: false,
        button: false,
        clear: false
    };

    QorFilterTime.plugin = function(options) {
        return this.each(function() {
            var $this = $(this);
            var data = $this.data(NAMESPACE);
            var fn;

            if (!data) {
                if (/destroy/.test(options)) {
                    return;
                }

                $this.data(NAMESPACE, (data = new QorFilterTime(this, options)));
            }

            if (typeof options === 'string' && $.isFunction((fn = data[options]))) {
                fn.apply(data);
            }
        });
    };

    $(function() {
        var selector = '[data-toggle="qor.filter.time"]';
        var options = {
            label: '.qor-filter__block-buttons button',
            trigger: 'a.qor-filter-toggle',
            button: '.qor-filter__button-search',
            clear: '.qor-selector-clear'
        };

        $(document)
            .on(EVENT_DISABLE, function(e) {
                QorFilterTime.plugin.call($(selector, e.target), 'destroy');
            })
            .on(EVENT_ENABLE, function(e) {
                QorFilterTime.plugin.call($(selector, e.target), options);
            })
            .triggerHandler(EVENT_ENABLE);
    });

    return QorFilterTime;
});

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

    let location = window.location,
        NAMESPACE = 'qor.filter',
        EVENT_FILTER_CHANGE = 'filterChanged.' + NAMESPACE,
        EVENT_ENABLE = 'enable.' + NAMESPACE,
        EVENT_DISABLE = 'disable.' + NAMESPACE,
        EVENT_CLICK = 'click.' + NAMESPACE,
        EVENT_CHANGE = 'change.' + NAMESPACE,
        CLASS_IS_ACTIVE = 'is-active',
        CLASS_BOTTOMSHEETS = '.qor-bottomsheets';

    let re = /([^&=]+)(=([^&]*))?/g;
    let decodeRE = /\+/g;  // Regex for replacing addition symbol with a space

    function decode(str) {
        return decodeURIComponent( str.replace(decodeRE, " ") )
    }


    function decodeSearch(search) {
        let data = [];

        if (search && search.indexOf('?') > -1) {
            search = search.replace(/\+/g, ' ').split('?')[1];

            if (search && search.indexOf('#') > -1) {
                search = search.split('#')[0];
            }

            if (search) {
                // search = search.toLowerCase();
                data = $.map(search.split('&'), function(n) {
                    let param = [];
                    let value;

                    n = n.split('=');
                    if (/page/.test(n[0])) {
                        return;
                    }
                    value = n[1];
                    param.push(n[0]);

                    if (value) {
                        value = $.trim(decodeURIComponent(value));

                        if (value) {
                            param.push(value);
                        }
                    }

                    return param.join('=');
                });
            }
        }

        return data;
    }

    function parseParams(data) {
        let query = decodeURI(data === undefined ? location.search : data),
            params = {}, e, search;
        if (query && query[0] === '?') {
            query = query.substring(1)
        }
        while ( e = re.exec(query) ) {
            let k = decode( e[1] ), v = decode( e[3] );
            if (k.substring(k.length - 2) === '[]') {
                (params[k] || (params[k] = [])).push(v);
            }
            else params[k] = v;
        }
        return {
            params: params,
            isArray: function(key) {
                return key.substring(key.length - 2) === '[]'
            },
            remove: function (key) {
                delete (this.params[key])
            },
            set: function(key, value) {
                if (key.substring(key.length - 2) === '[]') {
                    this.params[key] = this.params[key] || []
                    this.params[key].push(value);
                }
                else this.params[key] = value;
            },
            removeAny: function (key, values) {
                if (!this.params.hasOwnProperty(key)) return;
                this.params[key] = this.params[key].filter(val => values.indexOf(val) < 0);
                if (!this.params[key].length)
                    this.remove(key)
            },
            removeItem: function (key, item) {
                if (!this.params.hasOwnProperty(key)) return;
                this.params[key] = this.params[key].filter(val => val !== item);
                if (!this.params[key].length)
                    this.remove(key)
            },
            encode: function () {
                const search = $.param(this.params);
                return search.length ? '?' + search : '';
            }
        }
    }

    function QorFilter(element, options) {
        this.$element = $(element);
        this.options = $.extend({}, QorFilter.DEFAULTS, $.isPlainObject(options) && options);
        this.init();
    }

    QorFilter.prototype = {
        constructor: QorFilter,

        init: function() {
            // this.parse();
            this.bind();
        },

        bind: function() {
            var options = this.options;

            this.$element
                .on(EVENT_CLICK, options.label, $.proxy(this.toggle, this))
                .on(EVENT_CHANGE, options.group, $.proxy(this.toggle, this));
        },

        unbind: function() {
            this.$element.off(EVENT_CLICK, this.toggle).off(EVENT_CHANGE, this.toggle);
        },

        toggle: function(e) {
            let $target = $(e.currentTarget),
                params = parseParams(),
                paramName,
                value,
                search;

            if ($target.is('select')) {
                paramName = $target.attr('name');
                value = $target.val();
                if (params.isArray(paramName)) {
                    let values = [];
                    $target.children().each((_, el) => values[values.length] = $(el).prop('value'));
                    params.removeAny(paramName, values);
                    if (value) {
                        params.set(paramName, value)
                    }
                } else {
                    if (value) params.set(paramName, value)
                    else params.remove(paramName)
                }
                search = params.encode()
            } else if ($target.is('a')) {
                e.preventDefault();
                let uri = $target.attr('href'),
                    pos = uri.indexOf('?');
                if (pos >= 0) {
                    search = uri.substring(0, pos)
                } else {
                    search = "?"
                }
            } else if ($target.is('input')) {
                paramName = $target.attr('name');
                value = $target.val();
                if (value)
                    params.set(paramName, value);
                else
                    params.remove(paramName);
                search = params.encode()
            }
            this.applySearch(search, paramName)
        },

        applySearch: function(search, paramName) {
            if (this.$element.closest(CLASS_BOTTOMSHEETS).length) {
                $(CLASS_BOTTOMSHEETS).trigger(EVENT_FILTER_CHANGE, [search, paramName]);
            } else {
                location.search = search;
            }
        },

        destroy: function() {
            this.unbind();
            this.$element.removeData(NAMESPACE);
        }
    };

    QorFilter.DEFAULTS = {
        label: false,
        group: false
    };

    QorFilter.plugin = function(options) {
        return this.each(function() {
            var $this = $(this);
            var data = $this.data(NAMESPACE);
            var fn;

            if (!data) {
                if (/destroy/.test(options)) {
                    return;
                }

                $this.data(NAMESPACE, (data = new QorFilter(this, options)));
            }

            if (typeof options === 'string' && $.isFunction((fn = data[options]))) {
                fn.apply(data);
            }
        });
    };

    $(function() {
        var selector = '[data-toggle="qor.filter"]';
        var options = {
            label: 'a',
            group: 'select,input'
        };

        $(document)
            .on(EVENT_DISABLE, function(e) {
                QorFilter.plugin.call($(selector, e.target), 'destroy');
            })
            .on(EVENT_ENABLE, function(e) {
                QorFilter.plugin.call($(selector, e.target), options);
            })
            .triggerHandler(EVENT_ENABLE);
    });

    return QorFilter;
});

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

    const $window = $(window),
        NAMESPACE = 'qor.flex-row-wraped-fixer',
        EVENT_ENABLE = 'enable.' + NAMESPACE,
        EVENT_DISABLE = 'disable.' + NAMESPACE,
        EVENT_RESIZE = 'resize.' + NAMESPACE,
        EVENT_BEFORE_PRINT = 'beforeprint.' + NAMESPACE,
        EVENT_AFTER_PRINT = 'afterprint.' + NAMESPACE,
        SELECTOR = '.qor-flex-row-wrap';

    function Fixer(element, options) {
        this.$el = $(element);
        this.options = $.extend({}, Fixer.DEFAULTS, $.isPlainObject(options) && options);
        this.$children = this.$el.children();
        if (!this.$children.length) {
            return
        }
        this.init();
    }

    Fixer.prototype = {
        constructor: Fixer,

        init: function () {
            this.bind();
            this.fix();
        },

        bind: function () {
            $window.on(EVENT_RESIZE, this.fix.bind(this))
                .on(EVENT_AFTER_PRINT, this.afterPrint.bind(this))
                .on(EVENT_BEFORE_PRINT, this.beforePrint.bind(this));
        },

        unbind: function () {
            $window.off(EVENT_RESIZE, this.fix)
                .off(EVENT_AFTER_PRINT, this.afterPrint)
                .off(EVENT_BEFORE_PRINT, this.beforePrint);
        },

        beforePrint: function () {
            this.fix();
        },

        afterPrint: function () {
            this.fix()
        },

        fix: function () {
            const distance = this.$el.css('--item-distance'),
                $children = this.$children,
                dw = parseInt(this.$el.width());

            if (!distance) return;

            let row = [],
                rows = [row],
                rw = 0;

            $children.each(function () {
                $(this).children().css({marginLeft: 0, marginRight: 0, width: 'auto'});
            })

            $children.each(function (i) {
                let $el = $(this),
                    w = parseInt($el.width());
                if (rw > 0 && (rw + w) > dw) {
                    row = [];
                    rows[rows.length] = row
                    rw = w
                } else {
                    rw += w
                }
                row[row.length] = [i, w];
            });

            rows.forEach((row) => {
                row.forEach((el, i) => {
                    let $el = $($children[el[0]]),
                        rdw = $el.width();

                    if (row.length > 1) {
                        if (i === 0) {
                            $el.children().css({marginRight: distance / 2, width: rdw - distance / 2});
                        } else if (i === row.length - 1) {
                            $el.children().css({marginLeft: distance / 2, width: rdw - distance / 2});
                        } else {
                            $el.children().css({
                                marginLeft: distance / 2,
                                marginRight: distance / 2,
                                width: rdw - distance
                            })
                        }
                    }
                })
            })
        },

        destroy: function () {
            this.unbind();
            this.$el.removeData(NAMESPACE);
        }
    };

    Fixer.DEFAULTS = {
        header: false,
        content: false
    };

    Fixer.plugin = function (options) {
        return this.each(function () {
            var $this = $(this);
            var data = $this.data(NAMESPACE);
            var fn;

            if (!data) {
                $this.data(NAMESPACE, (data = new Fixer(this, options)));
            }

            if (typeof options === 'string' && $.isFunction(fn = data[options])) {
                fn.call(data);
            }
        });
    };

    return;

    $(function () {
        $(document).on(EVENT_ENABLE, function (e) {
            Fixer.plugin.call($(SELECTOR, e.target));
        }).on(EVENT_DISABLE, function (e) {
            Fixer.plugin.call($(SELECTOR, e.target), 'destroy');
        }).triggerHandler(EVENT_ENABLE);
    });

    return Fixer;

});
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

    const $window = $(window),
        _ = window._,
        NAMESPACE = 'qor.fixer',
        EVENT_ENABLE = 'enable.' + NAMESPACE,
        EVENT_DISABLE = 'disable.' + NAMESPACE;

    $(function() {
        $(document).
        on(EVENT_ENABLE, function(e) {
            $('.collection-edit-tabled', e.target).each(function (){
               const $el = $(this), $p2 = $el.parent().parent();
               if ($p2.is('.sec-col')) {
                   $p2.css({marginRight: 0, marginLeft: 0})
               }
            });
        }).
        on(EVENT_DISABLE, function(e) {
        }).
        triggerHandler(EVENT_ENABLE);
    });
});
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

    var $window = $(window);
    var _ = window._;
    var NAMESPACE = 'qor.head-fixer';
    var EVENT_ENABLE = 'enable.' + NAMESPACE;
    var EVENT_DISABLE = 'disable.' + NAMESPACE;
    var EVENT_RESIZE = 'resize.' + NAMESPACE;
    var EVENT_BEFORE_PRINT = 'beforeprint.' + NAMESPACE;
    var EVENT_AFTER_PRINT = 'afterprint.' + NAMESPACE;
    var CLASS_HEADER = '.qor-page__header';
    var CLASS_BODY = '.qor-page__body';

    function QorHeadFixer(element, options) {
        this.$element = $(element);
        this.options = $.extend({}, QorHeadFixer.DEFAULTS, $.isPlainObject(options) && options);
        this.$clone = null;
        this.init();
    }

    QorHeadFixer.prototype = {
        constructor: QorHeadFixer,

        init: function() {
            this.bind();
            this.fix();
        },

        bind: function() {
            $window.on(EVENT_RESIZE, this.fix.bind(this))
                .on(EVENT_AFTER_PRINT, this.afterPrint.bind(this))
                .on(EVENT_BEFORE_PRINT, this.beforePrint.bind(this));
        },

        unbind: function() {
            $window.off(EVENT_RESIZE, this.fix)
                .off(EVENT_AFTER_PRINT, this.afterPrint)
                .off(EVENT_BEFORE_PRINT, this.beforePrint);
        },

        beforePrint: function() {
            $(CLASS_BODY).each(function () {
                $(this).removeAttr('style');
            })
        },

        afterPrint: function() {
            this.fix()
        },

        fix: function() {
            $(CLASS_BODY).each(function () {
                let $this = $(this),
                    $header = $this.siblings(CLASS_HEADER);
                if ($header.length === 0) return;
                $this.css('paddingTop', 0);
                if ($header.css('position') !== 'fixed') {
                    $this.css('marginTop', 0);
                } else {
                    $this.css('marginTop', $header.height());
                }
                if ($header.children(':visible').length) {
                    $header.removeClass('no-visibile-items', true).show()
                } else {
                    $header.addClass('no-visibile-items', true).hide()
                }
            })
        },

        destroy: function() {
            this.unbind();
            this.$element.removeData(NAMESPACE);
        }
    };

    QorHeadFixer.DEFAULTS = {
        header: false,
        content: false
    };

    QorHeadFixer.plugin = function(options) {
        return this.each(function() {
            var $this = $(this);
            var data = $this.data(NAMESPACE);
            var fn;

            if (!data) {
                $this.data(NAMESPACE, (data = new QorHeadFixer(this, options)));
            }

            if (typeof options === 'string' && $.isFunction(fn = data[options])) {
                fn.call(data);
            }
        });
    };


    $('.qor-page > .qor-page__header').each(function (){
        const $thead = $(this).siblings('.qor-page__body').find('> .qor-table-container > table > thead');
        if (!$thead.length) return;

        const resize_ob = new ResizeObserver(function(entries) {
            // since we are observing only a single element, so we access the first element in entries array
            let rect = entries[0].contentRect;

            // current width & height
            let width = rect.width;
            let height = rect.height;
            $thead.css({top:rect.height})
        });
        resize_ob.observe(this);
    })

// start observing for resize

    return;

    $(function() {
        if (/[?&]prin(t&|t$)/.test(location.search)) {
            return
        }
        var selector = '.qor-js-table';
        var options = {
            header: '.mdl-layout__header',
            subHeader: '.qor-page__header',
            content: '.mdl-layout__content',
            paddingHeight: 2 // Fix sub header height bug
        };

        $(document).
        on(EVENT_DISABLE, function(e) {
            QorHeadFixer.plugin.call($(e.target), 'destroy');
        }).
        on(EVENT_ENABLE, function(e) {
            QorHeadFixer.plugin.call($(e.target), options);
        }).
        triggerHandler(EVENT_ENABLE);
    });

    return QorHeadFixer;

});
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

    const NAMESPACE = 'qor.inlineEdit',
        EVENT_ENABLE = 'enable.' + NAMESPACE,
        EVENT_DISABLE = 'disable.' + NAMESPACE,
        EVENT_CLICK = 'click.' + NAMESPACE,
        EVENT_MOUSEENTER = 'mouseenter.' + NAMESPACE,
        EVENT_MOUSELEAVE = 'mouseleave.' + NAMESPACE,
        CLASS_FIELD = '.qor-field',
        CLASS_FIELD_SHOW = '.qor-field__show',
        CLASS_FIELD_SHOW_INNER = '.qor-field__show-inner',
        CLASS_EDIT = '.qor-inlineedit__edit',
        CLASS_SAVE = '.qor-inlineedit__save',
        CLASS_BUTTONS = '.qor-inlineedit__buttons',
        CLASS_CANCEL = '.qor-inlineedit__cancel',
        CLASS_CONTAINER = 'qor-inlineedit__field';

    function QorInlineEdit(element, options) {
        this.$element = $(element);
        this.options = $.extend({}, QorInlineEdit.DEFAULTS, $.isPlainObject(options) && options);
        this.init();
    }

    function getJsonData(names, data) {
        let key,
            value = data[names[0].slice(1)];

        if (names.length > 1) {
            for (let i = 1; i < names.length; i++) {
                key = names[i].slice(1);
                value = $.isArray(value) ? value[0][key] : value[key];
            }
        }

        return value;
    }

    QorInlineEdit.prototype = {
        constructor: QorInlineEdit,

        init: function() {
            let $element = this.$element,
                saveButton = $element.data('button-save'),
                cancelButton = $element.data('button-cancel');

            this.TEMPLATE_SAVE = `<div class="qor-inlineedit__buttons">
                                        <button class="mdl-button mdl-button--colored mdl-js-button qor-button--small qor-inlineedit__cancel" type="button">${cancelButton}</button>
                                        <button class="mdl-button mdl-button--colored mdl-js-button qor-button--small qor-inlineedit__save" type="button">${saveButton}</button>
                                      </div>`;
            this.bind();
        },

        bind: function() {
            this.$element
                .on(EVENT_MOUSEENTER, CLASS_FIELD_SHOW, this.showEditButton)
                .on(EVENT_MOUSELEAVE, CLASS_FIELD_SHOW, this.hideEditButton)
                .on(EVENT_CLICK, CLASS_CANCEL, this.hideEdit)
                .on(EVENT_CLICK, CLASS_SAVE, this.saveEdit)
                .on(EVENT_CLICK, CLASS_EDIT, this.showEdit.bind(this));
        },

        unbind: function() {
            this.$element
                .off(EVENT_MOUSEENTER, CLASS_FIELD_SHOW, this.showEditButton)
                .off(EVENT_MOUSELEAVE, CLASS_FIELD_SHOW, this.hideEditButton)
                .off(EVENT_CLICK, CLASS_CANCEL, this.hideEdit)
                .off(EVENT_CLICK, CLASS_SAVE, this.saveEdit)
                .off(EVENT_CLICK, CLASS_EDIT, this.showEdit);
        },

        showEditButton: function(e) {
            let $edit = $(QorInlineEdit.TEMPLATE_EDIT);

            if ($(e.target).closest(CLASS_FIELD).find('input:disabled, textarea:disabled,select:disabled').length) {
                return false;
            }

            $edit.appendTo($(this));
        },

        hideEditButton: function() {
            $('.qor-inlineedit__edit').remove();
        },

        showEdit: function(e) {
            let $parent = $(e.target).closest(CLASS_EDIT).hide().closest(CLASS_FIELD).addClass(CLASS_CONTAINER),
                $save = $(this.TEMPLATE_SAVE);

            $save.appendTo($parent);
        },

        hideEdit: function() {
            let $parent = $(this).closest(CLASS_FIELD).removeClass(CLASS_CONTAINER);
            $parent.find(CLASS_BUTTONS).remove();
        },

        saveEdit: function() {
            let $btn = $(this),
                $parent = $btn.closest(CLASS_FIELD),
                $form = $btn.closest('form'),
                $hiddenInput = $parent.closest('.qor-fieldset').find('input.qor-hidden__primary_key[type="hidden"]'),
                $input = $parent.find('input[name*="QorResource"],textarea[name*="QorResource"],select[name*="QorResource"]'),
                $method = $form.find('input[name=_method]'),
                names = $input.length && $input.prop('name').match(/\.\w+/g),
                inputData = $input.serialize();

            if ($hiddenInput.length) {
                inputData = `${inputData}&${$hiddenInput.serialize()}`;
            }

            inputData += "&qorInlineEdit=true";

            if ($method.length > 0) {
                inputData += "&_method=" + $method.val();
            }

            if (names.length) {
                $.ajax($form.prop('action'), {
                    method: $form.prop('method'),
                    data: inputData,
                    dataType: 'json',
                    beforeSend: function() {
                        $btn.prop('disabled', true);
                    },
                    success: function(data) {
                        let newValue = getJsonData(names, data),
                            $show = $parent.removeClass(CLASS_CONTAINER).find(CLASS_FIELD_SHOW);

                        if ($show.find(CLASS_FIELD_SHOW_INNER).length) {
                            $show.find(CLASS_FIELD_SHOW_INNER).html(newValue);
                        } else {
                            $show.html(newValue);
                        }

                        $parent.find(CLASS_BUTTONS).remove();
                        $btn.prop('disabled', false);
                    },
                    error: function(xhr, textStatus, errorThrown) {
                        window.alert([textStatus, errorThrown].join(': '));
                        $btn.prop('disabled', false);
                    }
                });
            }
        },

        destroy: function() {
            this.unbind();
            this.$element.removeData(NAMESPACE);
        }
    };

    QorInlineEdit.DEFAULTS = {};

    QorInlineEdit.TEMPLATE_EDIT = `<button class="mdl-button mdl-js-button mdl-button--icon mdl-button--colored qor-inlineedit__edit" type="button"><i class="material-icons">mode_edit</i></button>`;

    QorInlineEdit.plugin = function(options) {
        return this.each(function() {
            var $this = $(this);
            var data = $this.data(NAMESPACE);
            var fn;

            if (!data) {
                $this.data(NAMESPACE, (data = new QorInlineEdit(this, options)));
            }

            if (typeof options === 'string' && $.isFunction((fn = data[options]))) {
                fn.call(data);
            }
        });
    };

    $(function() {
        if (/\/admin\//.test(location.href)) {
            return this
        }

        let selector = '[data-toggle="qor.inlineEdit"]',
            options = {};

        $(document)
            .on(EVENT_DISABLE, function(e) {
                QorInlineEdit.plugin.call($(selector, e.target), 'destroy');
            })
            .on(EVENT_ENABLE, function(e) {
                QorInlineEdit.plugin.call($(selector, e.target), options);
            })
            .triggerHandler(EVENT_ENABLE);
    });

    return QorInlineEdit;
});

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

    let NAMESPACE = 'qor.input_mask',
        EVENT_ENABLE = 'enable.' + NAMESPACE,
        EVENT_DISABLE = 'disable.' + NAMESPACE;

    function QorInputMask(element, options) {
        let $el = $(element);
        let value = $el.data('masker');
        if (value) {
            this.$el = $el;
            this.masker = atob(value);
            this.options = $.extend({}, QorInputMask.DEFAULTS, $.isPlainObject(options) && options);
            this.init();
        } else {
            this.maker = null;
        }
    }

    QorInputMask.prototype = {
        constructor: QorInputMask,

        init: function() {
            this.bind();
        },

        bind: function() {
            (function (masker) {
                eval(masker)
            }).call(this.$el, this.masker);
        },

        unbind: function() {
            this.$el.unmask();
        },

        destroy: function() {
            this.unbind();
            this.$el.removeData(NAMESPACE);
        }
    };

    QorInputMask.DEFAULTS = {};

    QorInputMask.plugin = function(options) {
        return this.each(function() {
            let $this = $(this),
                data = $this.data(NAMESPACE),
                fn;

            if (!data) {
                if (/destroy/.test(options)) {
                    return;
                }
                data = new QorInputMask(this, options);
                if (("masker" in data)) {
                    $this.data(NAMESPACE, data);
                } else {
                    return
                }
            }

            if (typeof options === 'string' && $.isFunction((fn = data[options]))) {
                fn.apply(data);
            }
        });
    };

    $(function() {
        let selector = '[data-masker]',
            options = {};

        $(document)
            .on(EVENT_DISABLE, function(e) {
                QorInputMask.plugin.call($(selector, e.target), 'destroy');
            })
            .on(EVENT_ENABLE, function(e) {
                QorInputMask.plugin.call($(selector, e.target), options);
            })
            .triggerHandler(EVENT_ENABLE);
    });

    return QorInputMask;
});
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

    let NAMESPACE = 'qor.input_money',
        SELECTOR = 'input.input-money',
        EVENT_ENABLE = 'enable.' + NAMESPACE,
        EVENT_DISABLE = 'disable.' + NAMESPACE,
        EVENT_CHANGE = 'change.' + NAMESPACE;

    function QorInputMoney(element, options) {
        this.$el = $(element);
        let $p = this.$el.closest('.mdl-js-textfield');
        this.MDL = $p.length ? $p[0].MaterialTextfield : null;
        this.init();
    }

    QorInputMoney.prototype = {
        constructor: QorInputMoney,

        init: function () {
            this.bind();
        },

        bind: function () {
            this.$el.maskMoney();
            this.$el.on(EVENT_CHANGE, this.changed.bind(this))
        },

        destroy: function () {
            this.$el.maskMoney('destroy');
            this.$el.removeData(NAMESPACE);
            this.$el.off(EVENT_CHANGE, this.changed);
            this.MDL = null;
        },

        changed: function (e) {
            if (!this.MDL) {
                let $p = this.$el.closest('.mdl-js-textfield');
                this.MDL = $p.length ? $p[0].MaterialTextfield : null;
            }
            if (this.MDL) {
                this.MDL.checkDirty()
            }
        }
    };

    QorInputMoney.DEFAULTS = {};

    QorInputMoney.plugin = function (options) {
        let args = Array.prototype.slice.call(arguments, 1),
            result;

        return this.each(function () {
            var $this = $(this);
            var data = $this.data(NAMESPACE);
            var fn;

            if (!data) {
                if (/destroy/.test(options)) {
                    return;
                }
                data = new QorInputMoney(this, options);
                if (("$el" in data)) {
                    $this.data(NAMESPACE, data);
                } else {
                    return
                }
            }

            if (typeof options === 'string' && $.isFunction((fn = data[options]))) {
                result = fn.apply(data, args);
            }
        });

        return (typeof result === 'undefined') ? this : result;
    };

    $(function () {
        var options = {};

        $(document)
            .on(EVENT_DISABLE, function (e) {
                QorInputMoney.plugin.call($(SELECTOR, e.target), 'destroy');
            })
            .on(EVENT_ENABLE, function (e) {
                QorInputMoney.plugin.call($(SELECTOR, e.target), options);
            })
            .triggerHandler(EVENT_ENABLE);
    });

    return QorInputMoney;
});
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

    const NAMESPACE = 'qor.input_timer',
        EVENT_ENABLE = 'enable.' + NAMESPACE;

    $(function() {
        const selector = 'input.mdl-textfield__input[type="datetime-local"],input.mdl-textfield__input[type="date"]';

        $(document)
            .on(EVENT_ENABLE, function(e) {
                $(selector, e.target).each(function () {
                    const field = $(this).closest('.mdl-textfield');
                    field.length === 1 && new MaterialTextfield(field[0])
                })
            })
            .triggerHandler(EVENT_ENABLE);
    });
});
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

    let NAMESPACE = 'qor.input_remote_validation',
        EVENT_ENABLE = 'enable.' + NAMESPACE,
        EVENT_DISABLE = 'disable.' + NAMESPACE,
        EVENT_CHANGE = 'change.' + NAMESPACE;

    function QorInputValidator(element, options) {
        let $el = $(element),
            $form = $el.parents('form'),
            validator = $el.data('validator');
        options = $.extend({}, QorInputValidator.DEFAULTS, $.isPlainObject(options) && options)
        
        if (!(validator)) {
            return
        }
        eval('validator = function(){'+atob(validator)+'};');
        this.validator = validator.call(this);
        if (!$.isFunction(this.validator)) {
            alert("ERROR: BAD input validator for '"+$el.attr('name')+"'");
            return;
        }
        this.$el = $el;
        this.$form = $form;
        this.options = options;
        this.init();
    }

    QorInputValidator.prototype = {
        constructor: QorInputValidator,

        init: function () {
            this.bind();
        },

        _do: function() {},

        bind: function () {
            this.$el.on(EVENT_CHANGE, this.validate.bind(this));
            this.$form.formValidator('register', this._validateByForm.bind(this), function (validator) {
                this._do = validator;
            }.bind(this));
        },

        _validateByForm: function(done, e) {
            this.validator(done, e)
        },

        validator: function(e, done) {},

        validate: function(e) {
            if (this.$el.val() === this.$el[0].defaultValue) {
                return true;
            }
            this._do(function (err) {
                if (err) {
                    QOR.alert(err, function () {
                        this.$el.focus();
                    }.bind(this));
                }
            }.bind(this), e);
            return false
        },

        unbind: function () {
            this.$el.off(EVENT_CHANGE);
        },

        destroy: function () {
            this.unbind();
            this.$el.removeData(NAMESPACE);
        }
    };

    QorInputValidator.DEFAULTS = {};

    QorInputValidator.plugin = function (options) {
        let args = Array.prototype.slice.call(arguments, 1),
            result;

        return this.each(function () {
            var $this = $(this);
            var data = $this.data(NAMESPACE);
            var fn;

            if (!data) {
                if (/destroy/.test(options)) {
                    return;
                }
                data = new QorInputValidator(this, options);
                if (("$el" in data)) {
                    $this.data(NAMESPACE, data);
                } else {
                    return
                }
            }

            if (typeof options === 'string' && $.isFunction((fn = data[options]))) {
                result = fn.apply(data, args);
            }
        });

        return (typeof result === 'undefined') ? this : result;
    };

    $(function () {
        var selector = '[data-validator]';
        var options = {};

        $(document)
            .on(EVENT_DISABLE, function (e) {
                QorInputValidator.plugin.call($(selector, e.target), 'destroy');
            })
            .on(EVENT_ENABLE, function (e) {
                QorInputValidator.plugin.call($(selector, e.target), options);
            })
            .triggerHandler(EVENT_ENABLE);
    });

    return QorInputValidator;
});
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

    let NAMESPACE = 'qor.logo',
        SELECTOR = '.qor-logo[data-src]',
        EVENT_LOAD = 'load.'+NAMESPACE,
        EVENT_ERROR = 'error.'+NAMESPACE,
        EVENT_ENABLE = 'enable.' + NAMESPACE,
        EVENT_DISABLE = 'disable.' + NAMESPACE;

    function QorLogo(element, options) {
        this.$el = $(element);
        this.data = this.$el.data();
        this.$img = $("<img />");
        this.fallback = false;
        this.init();
    }

    QorLogo.prototype = {
        constructor: QorLogo,

        init: function () {
            this.build();
        },

        build: function () {
            this.$img
                .attr('src', this.data.src)
                .on(EVENT_LOAD, this.onload.bind(this))
                .on(EVENT_ERROR, this.onerror.bind(this));

            if (this.data.alt) {
                this.$img.attr('alt', this.data.alt)
            }
            if (this.data.title) {
                this.$img.attr('title', this.data.title)
            }

            this.$el.show().append(this.$img);
        },

        onerror: function(e) {
            if (this.fallback) {
                this.destroy();
                return
            }
            if (this.data.fallback) {
                this.fallback = true;
                this.$img.attr('src', this.data.fallback);
            } else {
                this.destroy();
            }
        },

        onload: function(e) {
            this.unbind();
        },

        unbind: function () {
            this.$img
                .off(EVENT_LOAD, this.onload)
                .off(EVENT_ERROR, this.onerror);
        },

        destroy: function () {
            this.unbind();
            this.$el.removeData(NAMESPACE);
            this.$el.hide();
            this.$img.remove();
        }
    };

    QorLogo.DEFAULTS = {};

    QorLogo.plugin = function (options) {
        let args = Array.prototype.slice.call(arguments, 1),
            result;

        return this.each(function () {
            var $this = $(this);
            var data = $this.data(NAMESPACE);
            var fn;

            if (!data) {
                if (/destroy/.test(options)) {
                    return;
                }
                data = new QorLogo(this, options);
                if (("$el" in data)) {
                    $this.data(NAMESPACE, data);
                } else {
                    return
                }
            }

            if (typeof options === 'string' && $.isFunction((fn = data[options]))) {
                result = fn.apply(data, args);
            }
        });

        return (typeof result === 'undefined') ? this : result;
    };

    $(function () {
        var options = {};

        $(document)
            .on(EVENT_DISABLE, function (e) {
                QorLogo.plugin.call($(SELECTOR, e.target), 'destroy');
            })
            .on(EVENT_ENABLE, function (e) {
                QorLogo.plugin.call($(SELECTOR, e.target), options);
            })
            .triggerHandler(EVENT_ENABLE);
    });

    return QorLogo;
});
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

    const componentHandler = window.componentHandler,
        NAMESPACE = 'qor.material',
        EVENT_ENABLE = 'enable.' + NAMESPACE,
        EVENT_DISABLE = 'disable.' + NAMESPACE,
        EVENT_UPDATE = 'update.' + NAMESPACE,
        SELECTOR_COMPONENT = '[class*="mdl-js"],[class*="mdl-tooltip"]';

    function enable(target) {
        const $el = $(target);
        /*jshint undef:false */
        if (componentHandler) {
            // Enable all MDL (Material Design Lite) components within the target element
            if ($el.is(SELECTOR_COMPONENT)) {
                componentHandler.upgradeElements(target);
            } else {
                componentHandler.upgradeElements($(SELECTOR_COMPONENT, target).toArray());
            }
        }
    }

    function disable(target) {
        const $el = $(target);
        /*jshint undef:false */
        if (componentHandler) {
            // Destroy all MDL (Material Design Lite) components within the target element
            if ($el.is(SELECTOR_COMPONENT)) {
                componentHandler.downgradeElements(target);
            } else {
                componentHandler.downgradeElements($(SELECTOR_COMPONENT, target).toArray());
            }
        }
    }

    $(function() {
        $(document)
            .on(EVENT_ENABLE, function(e) {
                enable(e.target);
            })
            .on(EVENT_DISABLE, function(e) {
                disable(e.target);
            })
            .on(EVENT_UPDATE, function(e) {
                disable(e.target);
                enable(e.target);
            });
    });
});

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

    const $document = $(document),
        NAMESPACE = 'qor.modal',
        EVENT_ENABLE = 'enable.' + NAMESPACE,
        EVENT_DISABLE = 'disable.' + NAMESPACE,
        EVENT_CLICK = 'click.' + NAMESPACE,
        EVENT_KEYUP = 'keyup.' + NAMESPACE,
        EVENT_SHOW = 'show.' + NAMESPACE,
        EVENT_SHOWN = 'shown.' + NAMESPACE,
        EVENT_HIDE = 'hide.' + NAMESPACE,
        EVENT_HIDDEN = 'hidden.' + NAMESPACE,
        EVENT_TRANSITION_END = 'transitionend',
        CLASS_OPEN = 'qor-modal-open',
        CLASS_SHOWN = 'shown',
        CLASS_FADE = 'fade',
        CLASS_IN = 'in',
        ARIA_HIDDEN = 'aria-hidden';

    function QorModal(element, options) {
        this.$element = $(element);
        this.options = $.extend({}, QorModal.DEFAULTS, $.isPlainObject(options) && options);
        this.transitioning = false;
        this.fadable = false;
        this.init();
    }

    QorModal.prototype = {
        constructor: QorModal,

        init: function () {
            this.fadable = this.$element.hasClass(CLASS_FADE);

            if (this.options.show) {
                this.show();
            } else {
                this.toggle();
            }
        },

        bind: function () {
            this.$element.on(EVENT_CLICK, $.proxy(this.click, this));

            if (this.options.keyboard) {
                $document.on(EVENT_KEYUP, $.proxy(this.keyup, this));
            }
        },

        unbind: function () {
            this.$element.off(EVENT_CLICK, this.click);

            if (this.options.keyboard) {
                $document.off(EVENT_KEYUP, this.keyup);
            }
        },

        click: function (e) {
            const element = this.$element[0];
            let target = e.target;

            if (target === element && this.options.backdrop) {
                this.hide();
                return;
            }

            while (target !== element) {
                if ($(target).data('dismiss') === 'modal') {
                    this.hide();
                    break;
                }

                target = target.parentNode;
            }
        },

        keyup: function (e) {
            if (e.which === 27) {
                this.hide();
            }
        },

        show: function (noTransition) {
            var $this = this.$element,
                showEvent;

            if (this.transitioning || $this.hasClass(CLASS_IN)) {
                return;
            }

            showEvent = $.Event(EVENT_SHOW);
            $this.trigger(showEvent);

            if (showEvent.isDefaultPrevented()) {
                return;
            }

            $document.find('body').addClass(CLASS_OPEN);

            /*jshint expr:true */
            $this.addClass(CLASS_SHOWN).scrollTop(0).get(0).offsetHeight; // reflow for transition
            this.transitioning = true;

            if (noTransition || !this.fadable) {
                $this.addClass(CLASS_IN);
                this.shown();
                return;
            }

            $this.one(EVENT_TRANSITION_END, $.proxy(this.shown, this));
            $this.addClass(CLASS_IN);
        },

        shown: function () {
            this.transitioning = false;
            this.bind();
            this.$element.attr(ARIA_HIDDEN, false).trigger(EVENT_SHOWN).focus();
        },

        hide: function (noTransition) {
            var $this = this.$element,
                hideEvent;

            if (this.transitioning || !$this.hasClass(CLASS_IN)) {
                return;
            }

            hideEvent = $.Event(EVENT_HIDE);
            $this.trigger(hideEvent);

            if (hideEvent.isDefaultPrevented()) {
                return;
            }

            $document.find('body').removeClass(CLASS_OPEN);
            this.transitioning = true;

            if (noTransition || !this.fadable) {
                $this.removeClass(CLASS_IN);
                this.hidden();
                return;
            }

            $this.one(EVENT_TRANSITION_END, $.proxy(this.hidden, this));
            $this.removeClass(CLASS_IN);
        },

        hidden: function () {
            this.transitioning = false;
            this.unbind();
            this.$element.removeClass(CLASS_SHOWN).attr(ARIA_HIDDEN, true).trigger(EVENT_HIDDEN);
        },

        toggle: function () {
            if (this.$element.hasClass(CLASS_IN)) {
                this.hide();
            } else {
                this.show();
            }
        },

        destroy: function () {
            this.$element.removeData(NAMESPACE);
        }
    };

    QorModal.DEFAULTS = {
        backdrop: false,
        keyboard: true,
        show: true
    };

    QorModal.plugin = function (options) {
        return this.each(function () {
            var $this = $(this);
            var data = $this.data(NAMESPACE);
            var fn;

            if (!data) {
                if (/destroy/.test(options)) {
                    return;
                }

                $this.data(NAMESPACE, (data = new QorModal(this, options)));
            }

            if (typeof options === 'string' && $.isFunction(fn = data[options])) {
                fn.apply(data);
            }
        });
    };

    $.fn.qorModal = QorModal.plugin;

    $(function () {
        var selector = '.qor-modal';

        $(document).
        on(EVENT_CLICK, '[data-toggle="qor.modal"]', function () {
            var $this = $(this);
            var data = $this.data();
            var $target = $(data.target || $this.attr('href'));

            QorModal.plugin.call($target, $target.data(NAMESPACE) ? 'toggle' : data);
        }).
        on(EVENT_DISABLE, function (e) {
            QorModal.plugin.call($(selector, e.target), 'destroy');
        }).
        on(EVENT_ENABLE, function (e) {
            QorModal.plugin.call($(selector, e.target));
        });
    });

    return QorModal;

});

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

    let NAMESPACE = 'qor.password_visibility',
        SELECTOR = '[data-toggle="qor.password_visibility"]',
        EVENT_ENABLE = 'enable.' + NAMESPACE,
        EVENT_DISABLE = 'disable.' + NAMESPACE,
        EVENT_CLICK = 'click.' + NAMESPACE;

    function QorPasswordVisibility(element, options) {
        this.$el = $(element);
        this.init();
    }

    QorPasswordVisibility.prototype = {
        constructor: QorPasswordVisibility,

        init: function () {
            this.flag = false;
            this.$icon = this.$el.find('i');
            this.$target = this.$el.parents('div:eq(0)').children('input[type=password]');
            this.icons = [this.$icon.text(), this.$el.data('toggleIcon')];
            this.bind();
        },

        bind: function () {
            this.$el.bind(EVENT_CLICK, this.toggle.bind(this));
        },

        toggle: function () {
            this.flag = !this.flag;
            this.$icon.html(this.icons[+this.flag]);
            this.$target.attr('type', this.flag?'text':'password');
        },

        destroy: function () {
            this.$el.off(EVENT_CLICK, this.toggle);
            this.$el.removeData(NAMESPACE);
        }
    };

    QorPasswordVisibility.DEFAULTS = {};

    QorPasswordVisibility.plugin = function (options) {
        let args = Array.prototype.slice.call(arguments, 1),
            result;

        return this.each(function () {
            var $this = $(this);
            var data = $this.data(NAMESPACE);
            var fn;

            if (!data) {
                if (/destroy/.test(options)) {
                    return;
                }
                data = new QorPasswordVisibility(this, options);
                if (("$el" in data)) {
                    $this.data(NAMESPACE, data);
                } else {
                    return
                }
            }

            if (typeof options === 'string' && $.isFunction((fn = data[options]))) {
                result = fn.apply(data, args);
            }
        });

        return (typeof result === 'undefined') ? this : result;
    };

    $(function () {
        var options = {};

        $(document)
            .on(EVENT_DISABLE, function (e) {
                QorPasswordVisibility.plugin.call($(SELECTOR, e.target), 'destroy');
            })
            .on(EVENT_ENABLE, function (e) {
                QorPasswordVisibility.plugin.call($(SELECTOR, e.target), options);
            })
            .triggerHandler(EVENT_ENABLE);
    });

    return QorPasswordVisibility;
});
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

  var NAMESPACE = 'qor.tabbar.radio';
  var EVENT_ENABLE = 'enable.' + NAMESPACE;
  var EVENT_DISABLE = 'disable.' + NAMESPACE;
  var EVENT_CLICK = 'click.' + NAMESPACE;
  var EVENT_SWITCHED = 'switched.' + NAMESPACE;
  var CLASS_TAB = '[data-tab-target]';
  var CLASS_TAB_SOURCE = '[data-tab-source]';
  var CLASS_ACTIVE = 'is-active';

  function QorTabRadio(element, options) {
    this.$element = $(element);
    this.options = $.extend({}, QorTabRadio.DEFAULTS, $.isPlainObject(options) && options);
    this.init();
  }

  QorTabRadio.prototype = {
    constructor: QorTabRadio,

    init: function () {
      this.bind();
    },

    bind: function () {
      this.$element.on(EVENT_CLICK, CLASS_TAB, this.switchTab.bind(this));
    },

    unbind: function () {
      this.$element.off(EVENT_CLICK, CLASS_TAB, this.switchTab);
    },

    switchTab: function (e) {
      var $target = $(e.target),
          $element = this.$element,
          $tabs = $element.find(CLASS_TAB),
          $tabSources = $element.find(CLASS_TAB_SOURCE),
          data = $target.data(),
          tabTarget = data.tabTarget;

      if ($target.hasClass(CLASS_ACTIVE)){
        return;
      }

      $tabs.removeClass(CLASS_ACTIVE);
      $target.addClass(CLASS_ACTIVE);

      $tabSources.hide().filter('[data-tab-source="' + tabTarget + '"]').show();
      $element.trigger(EVENT_SWITCHED, [$element, tabTarget]);

    },

    destroy: function () {
      this.unbind();
    }
  };

  QorTabRadio.DEFAULTS = {};

  QorTabRadio.plugin = function (options) {
    return this.each(function () {
      var $this = $(this);
      var data = $this.data(NAMESPACE);
      var fn;

      if (!data) {
        if (/destroy/.test(options)) {
          return;
        }

        $this.data(NAMESPACE, (data = new QorTabRadio(this, options)));
      }

      if (typeof options === 'string' && $.isFunction(fn = data[options])) {
        fn.apply(data);
      }
    });
  };

  $(function () {
    var selector = '[data-toggle="qor.tab.radio"]';

    $(document)
      .on(EVENT_DISABLE, function (e) {
        QorTabRadio.plugin.call($(selector, e.target), 'destroy');
      })
      .on(EVENT_ENABLE, function (e) {
        QorTabRadio.plugin.call($(selector, e.target));
      })
      .triggerHandler(EVENT_ENABLE);
  });

  return QorTabRadio;

});

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

    let NAMESPACE = 'qor.redactor',
        EVENT_ENABLE = 'enable.' + NAMESPACE,
        EVENT_DISABLE = 'disable.' + NAMESPACE,
        EVENT_CLICK = 'click.' + NAMESPACE,
        EVENT_KEYUP = 'keyup.' + NAMESPACE,
        EVENT_ADD_CROP = 'addCrop.' + NAMESPACE,
        EVENT_REMOVE_CROP = 'removeCrop.' + NAMESPACE,
        EVENT_SHOWN = 'shown.qor.modal',
        EVENT_HIDDEN = 'hidden.qor.modal',
        EVENT_SCROLL = 'scroll.' + NAMESPACE,
        CLASS_WRAPPER = '.qor-cropper__wrapper',
        CLASS_SAVE = '.qor-cropper__save',
        CLASS_CROPPER_TOGGLE = '.qor-cropper__toggle--redactor',
        ID_REDACTOR_LINK_TITLE = '#redactor-link-title',
        ID_REDACTOR_LINK_TEXT = '#redactor-link-url-text',
        ID_REDACTOR_MODAL_BUTTON_CANCEL = '#redactor-modal-button-cancel';

    function encodeCropData(data) {
        var nums = [];

        if ($.isPlainObject(data)) {
            $.each(data, function() {
                nums.push(arguments[1]);
            });
        }

        return nums.join();
    }

    function decodeCropData(data) {
        var nums = data && data.split(',');

        data = null;

        if (nums && nums.length === 4) {
            data = {
                x: Number(nums[0]),
                y: Number(nums[1]),
                width: Number(nums[2]),
                height: Number(nums[3])
            };
        }

        return data;
    }

    function capitalize(str) {
        if (typeof str === 'string') {
            str = str.charAt(0).toUpperCase() + str.substr(1);
        }

        return str;
    }

    function getCapitalizeKeyObject(obj) {
        var newObj = {},
            key;

        if ($.isPlainObject(obj)) {
            for (key in obj) {
                if (obj.hasOwnProperty(key)) {
                    newObj[capitalize(key)] = obj[key];
                }
            }
        }

        return newObj;
    }

    function replaceText(str, data) {
        if (typeof str === 'string') {
            if (typeof data === 'object') {
                $.each(data, function(key, val) {
                    str = str.replace('$[' + String(key).toLowerCase() + ']', val);
                });
            }
        }

        return str;
    }

    function escapeHTML(unsafe_str) {
        return unsafe_str.replace(/&/g, ' ').replace(/</g, ' ').replace(/>/g, ' ').replace(/\"/g, ' ').replace(/\'/g, ' ').replace(/\`/g, ' ');
    }

    function redactorToolbarSrcoll($editor, toolbarFixedTopOffset) {
        let $toolbar = $editor.find('.redactor-toolbar'),
            offsetTop = $editor.offset().top,
            editorHeight = $editor.height(),
            normallCSS = {
                position: 'relative',
                top: 'auto',
                width: 'auto',
                boxShadow: 'none'
            },
            fixedCSS = {
                position: 'fixed',
                boxShadow: '0 2px 4px rgba(0,0,0,.1)',
                top: toolbarFixedTopOffset,
                width: $editor.width()
            };

        if ($toolbar.css('position') === 'relative') {
            editorHeight = $editor.height() - 50;
        }

        if (offsetTop < toolbarFixedTopOffset) {
            if (editorHeight - 50 - toolbarFixedTopOffset < Math.abs(offsetTop)) {
                $toolbar.css(normallCSS);
            } else {
                $toolbar.css(fixedCSS);
            }
        } else {
            $toolbar.css(normallCSS);
        }
    }

    function QorRedactor(element, options) {
        this.$element = $(element);
        this.options = $.extend(true, {}, QorRedactor.DEFAULTS, $.isPlainObject(options) && options);
        this.init();
    }

    QorRedactor.prototype = {
        constructor: QorRedactor,

        init: function() {
            var options = this.options;
            var $this = this.$element;
            var $parent = $this.closest(options.parent);

            if (!$parent.length) {
                $parent = $this.parent();
            }

            this.$parent = $parent;
            this.$button = $(QorRedactor.BUTTON);
            this.$modal = $(replaceText(QorRedactor.MODAL, options.text)).appendTo('body');
            this.bind();
        },

        bind: function() {
            this.$element.on(EVENT_ADD_CROP, $.proxy(this.addButton, this)).on(EVENT_REMOVE_CROP, $.proxy(this.removeButton, this));
        },

        unbind: function() {
            this.$element.off(EVENT_ADD_CROP).off(EVENT_REMOVE_CROP).off(EVENT_SCROLL);
        },

        addButton: function(e, image) {
            var $image = $(image);

            this.$button.css('left', $(image).width() / 2).prependTo($image.parent()).find(CLASS_CROPPER_TOGGLE).one(EVENT_CLICK, $.proxy(this.crop, this, $image));
        },

        removeButton: function() {
            this.$button.find(CLASS_CROPPER_TOGGLE).off(EVENT_CLICK);
            this.$button.detach();
        },

        crop: function($image) {
            var options = this.options;
            var url = $image.attr('src');
            var originalUrl = url;
            var $clone = $('<img>');
            var $modal = this.$modal;

            if ($.isFunction(options.replace)) {
                originalUrl = options.replace(originalUrl);
            }

            $clone.attr('src', originalUrl);
            $modal
                .one(EVENT_SHOWN, function() {
                    $clone.cropper({
                        data: decodeCropData($image.attr('data-crop-options')),
                        background: false,
                        movable: false,
                        zoomable: false,
                        scalable: false,
                        rotatable: false,
                        checkImageOrigin: false,

                        built: function() {
                            $modal.find(CLASS_SAVE).one(EVENT_CLICK, function() {
                                var cropData = $clone.cropper('getData', true);

                                $.ajax(options.remote, {
                                    type: 'POST',
                                    contentType: 'application/json',
                                    data: JSON.stringify({
                                        Url: url,
                                        CropOptions: {
                                            original: getCapitalizeKeyObject(cropData)
                                        },
                                        Crop: true
                                    }),
                                    dataType: 'json',

                                    success: function(response) {
                                        if ($.isPlainObject(response) && response.url) {
                                            $image.attr('src', response.url).attr('data-crop-options', encodeCropData(cropData)).removeAttr('style').removeAttr('rel');

                                            if ($.isFunction(options.complete)) {
                                                options.complete();
                                            }
                                            $modal.qorModal('hide');
                                        }
                                    }
                                });
                            });
                        }
                    });
                })
                .one(EVENT_HIDDEN, function() {
                    $clone.cropper('destroy').remove();
                })
                .qorModal('show')
                .find(CLASS_WRAPPER)
                .append($clone);
        },

        destroy: function() {
            this.unbind();
            this.$modal.qorModal('hide').remove();
            this.$element.removeData(NAMESPACE);
        }
    };

    QorRedactor.DEFAULTS = {
        remote: false,
        parent: false,
        toggle: false,
        replace: null,
        complete: null,
        text: {
            title: 'Crop the image',
            ok: 'OK',
            cancel: 'Cancel'
        }
    };

    QorRedactor.BUTTON = `<div class="qor-redactor__image--buttons">
            <span class="qor-redactor__image--edit" contenteditable="false">Edit</span>
            <span class="qor-cropper__toggle--redactor" contenteditable="false">Crop</span>
        </div>`;

    QorRedactor.MODAL = `<div class="qor-modal fade" tabindex="-1" role="dialog" aria-hidden="true">
            <div class="mdl-card mdl-shadow--2dp" role="document">
              <div class="mdl-card__title">
                <h2 class="mdl-card__title-text">$[title]</h2>
              </div>
              <div class="mdl-card__supporting-text">
                <div class="qor-cropper__wrapper"></div>
              </div>
              <div class="mdl-card__actions mdl-card--border">
                <a class="mdl-button mdl-button--colored mdl-js-button mdl-js-ripple-effect qor-cropper__save">$[ok]</a>
                <a class="mdl-button mdl-button--colored mdl-js-button mdl-js-ripple-effect" data-dismiss="modal">$[cancel]</a>
              </div>
              <div class="mdl-card__menu">
                <button class="mdl-button mdl-button--icon mdl-js-button mdl-js-ripple-effect" data-dismiss="modal" aria-label="close">
                  <i class="material-icons">close</i>
                  </button>
              </div>
            </div>
        </div>`;

    QorRedactor.plugin = function(option) {
        return this.each(function() {
            let $this = $(this),
                data = $this.data(NAMESPACE),
                config,
                fn;

            if (!data) {
                if (!$.fn.redactor) {
                    return;
                }

                if (/destroy/.test(option)) {
                    return;
                }

                $this.data(NAMESPACE, (data = {}));

                let editorButtons = ['html', 'format', 'bold', 'italic', 'deleted', 'lists', 'image', 'file', 'link', 'table'];

                config = {
                    imageUpload: $this.data('uploadUrl'),
                    fileUpload: $this.data('uploadUrl'),
                    imageResizable: true,
                    imagePosition: true,
                    toolbarFixed: false,
                    buttons: editorButtons,

                    callbacks: {
                        init: function() {
                            let button,
                                $editor = this.core.box(),
                                isInSlideout = $('.qor-slideout').is(':visible'),
                                toolbarFixedTarget,
                                toolbarFixedTopOffset = 64;

                            editorButtons.forEach(function(item) {
                                button = this.button.get(item);
                                this.button.setIcon(button, '<i class="material-icons ' + item + '"></i>');
                            }, this);

                            if (isInSlideout) {
                                toolbarFixedTarget = '.qor-slideout';
                                toolbarFixedTopOffset = $('.qor-slideout__header').height();
                            } else {
                                toolbarFixedTarget = '.qor-layout main.qor-page';
                                toolbarFixedTopOffset = toolbarFixedTopOffset + $(toolbarFixedTarget).find('.qor-page__header').height();
                            }

                            $(toolbarFixedTarget).on(EVENT_SCROLL, function() {
                                redactorToolbarSrcoll($editor, toolbarFixedTopOffset);
                            });

                            if (!$this.data('cropUrl')) {
                                return;
                            }

                            $this.data(
                                NAMESPACE,
                                (data = new QorRedactor($this, {
                                    remote: $this.data('cropUrl'),
                                    text: $this.data('text'),
                                    parent: '.qor-field',
                                    toggle: '.qor-cropper__toggle--redactor',
                                    replace: function(url) {
                                        return url.replace(/\.\w+$/, function(extension) {
                                            return '.original' + extension;
                                        });
                                    },
                                    complete: $.proxy(function() {
                                        this.code.sync();
                                    }, this)
                                }))
                            );
                        },

                        imageUpload: function(image, json) {
                            var $image = $(image);
                            json.filelink && $image.prop('src', json.filelink);
                        },

                        click: function(e) {
                            var $currentTag = $(this.selection.parent());

                            if ($currentTag.is('.redactor-layer')) {
                                $currentTag = $(this.selection.current());
                            }
                            this.selection.$currentTag = $currentTag;
                            this.link.linkDescription = '';
                            this.link.insertedTriggered = false;
                            this.link.valueChanged = false;

                            if (this.link.is()) {
                                this.link.linkDescription = this.link.get().prop('title');
                                this.link.$linkHtml = $(e.target);
                            }
                        },

                        modalOpened: function(name, modal) {
                            var _this = this;
                            if (name == 'link') {
                                $(modal)
                                    .find('#redactor-link-url-text')
                                    .closest('section')
                                    .after(
                                        '<section><label>Description for Accessibility</label><input value="' +
                                            this.link.linkDescription +
                                            '" type="text" id="redactor-link-title" placeholder="If blank, will use Text value above" /></section>'
                                    );

                                this.link.linkUrlText = $(ID_REDACTOR_LINK_TEXT).val();
                                this.link.description = $(ID_REDACTOR_LINK_TITLE).val();

                                $(ID_REDACTOR_LINK_TITLE).off(EVENT_KEYUP);
                                $(ID_REDACTOR_LINK_TEXT).off(EVENT_KEYUP);
                                $(ID_REDACTOR_MODAL_BUTTON_CANCEL).off(EVENT_CLICK);

                                $(ID_REDACTOR_MODAL_BUTTON_CANCEL).on(EVENT_CLICK, function() {
                                    _this.link.clickCancel = true;
                                });

                                $(ID_REDACTOR_LINK_TITLE).on(EVENT_KEYUP, function() {
                                    _this.link.valueChanged = true;
                                    _this.link.description = escapeHTML($(this).val());
                                });

                                $(ID_REDACTOR_LINK_TEXT).on(EVENT_KEYUP, function() {
                                    _this.link.valueChanged = true;
                                    _this.link.linkUrlText = escapeHTML($(this).val());
                                });
                            }
                        },

                        modalClosed: function(name) {
                            var $linkHtml = this.link.$linkHtml,
                                description = this.link.description;

                            if (name == 'link' && !this.link.insertedTriggered && $linkHtml && $linkHtml.length && this.link.valueChanged && !this.link.clickCancel) {
                                if (description) {
                                    $linkHtml.prop('title', description);
                                } else {
                                    $linkHtml.prop('title', this.link.linkUrlText);
                                }
                            }

                            this.link.description = '';
                            this.link.linkUrlText = '';
                            this.link.insertedTriggered = false;
                            this.link.valueChanged = false;
                            this.link.clickCancel = false;
                        },

                        insertedLink: function(link) {
                            var $link = $(link),
                                description = this.link.description;

                            $link.prop('title', description ? description : $link.text());
                            this.link.description = '';
                            this.link.linkUrlText = '';
                            this.link.insertedTriggered = true;
                        },

                        fileUpload: function(link, json) {
                            $(link).prop('href', json.filelink).html(json.filename);
                        }
                    }
                };

                $.extend(config, $this.data('redactorSettings'));
                $this.redactor(config);
            } else {
                if (/destroy/.test(option)) {
                    $this.redactor('destroy');
                }
            }

            if (typeof option === 'string' && $.isFunction((fn = data[option]))) {
                fn.apply(data);
            }
        });
    };

    $(function() {
        var selector = 'textarea[data-toggle="qor.redactor"]';

        $(document)
            .on(EVENT_DISABLE, function(e) {
                QorRedactor.plugin.call($(selector, e.target), 'destroy');
            })
            .on(EVENT_ENABLE, function(e) {
                QorRedactor.plugin.call($(selector, e.target));
            })
            .triggerHandler(EVENT_ENABLE);
    });

    return QorRedactor;
});

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

    const _ = window._,
        counter = {v:0},
        NAMESPACE = 'qor.replicator',
        EVENT_ENABLE = 'enable.' + NAMESPACE,
        EVENT_DISABLE = 'disable.' + NAMESPACE,
        EVENT_SUBMIT = 'submit.' + NAMESPACE,
        EVENT_CLICK = 'click.' + NAMESPACE,
        EVENT_SLIDEOUTBEFORESEND = 'slideoutBeforeSend.qor.slideout.replicator',
        EVENT_SELECTCOREBEFORESEND = 'selectcoreBeforeSend.qor.selectcore.replicator bottomsheetBeforeSend.qor.bottomsheets.replicator',
        EVENT_REPLICATOR_ADDED = 'added.' + NAMESPACE,
        EVENT_REPLICATORS_ADDED = 'addedMultiple.' + NAMESPACE,
        EVENT_REPLICATORS_ADDED_DONE = 'addedMultipleDone.' + NAMESPACE,
        CLASS_CONTAINER = '.qor-fieldset-container,.qor-replicator-container';

    function nextId() {
        counter.v++
        return counter.v
    }

    function QorReplicator(element, options) {
        this.$element = $(element);
        this.options = $.extend({}, QorReplicator.DEFAULTS, {}, $.isPlainObject(options) && options);
        this.index = 0;
        this.init();
    }

    QorReplicator.prototype = {
        constructor: QorReplicator,

        init: function() {
            let $element = this.$element,
                $template = $element.find('> template[type="qor-collection-edit-new/html"]'),
                data = $element.data(),
                $root = $element.find('> .qor-field__block ' + (this.options.rootSelector || data.rootSelector || '')),
                fieldsetName,
                alertTemplate = $.extend({}, this.options.alertTemplate, true),
                $alert;

            this.$root = $root;
            this.baseClass = data.baseClass ? data.baseClass : 'fieldset';
            this.itemSelector = data.itemSelector ? data.itemSelector : this.options.itemClass;
            this.isInSlideout = $element.closest('.qor-slideout').length;
            this.hasInlineReplicator = $element.find(CLASS_CONTAINER).length;
            this.maxitems = data.maxItem;
            this.isSortable = $element.hasClass(this.options.sortableClass);
            this.$target = $element.find('.' + this.baseClass + '__target');
            if (!this.$target.length) {
                this.$target = null
            }

            if (data.alertTag) alertTemplate.tag = data.alertTag

            $alert = $('<' + alertTemplate.tag + '>');
            for (let tagName in alertTemplate.attrs)
                $alert.attr(tagName, alertTemplate.attrs[tagName])
            $alert.html(alertTemplate.body);
            this.alertTemplate = $alert.wrapAll('<div>').parent().html().replace('UNDO_DELETE_MESSAGE', QOR.messages.replicator.undoDelete);

            this.canCreate = $template.length > 0;

            // if have isMultiple data value or template length large than 1
            this.isMultipleTemplate = $element.data('isMultiple');

            if (this.canCreate) {
                if (this.isMultipleTemplate) {
                    this.fieldsetName = [];
                    this.template = {};
                    this.index = [];

                    $template.each((i, ele) => {
                        fieldsetName = $(ele).data('fieldsetName');
                        if (fieldsetName) {
                            this.template[fieldsetName] = $(ele).html();
                            this.fieldsetName.push(fieldsetName);
                        }
                    });

                    this.parseMultiple();
                } else {
                    this.template = $template.html();
                    this.prefix = $template.attr('data-prefix');
                    this.index = $template.data("next-index");
                }
            }

            this.id = `${nextId()}`;
            this.$element.attr('data-qor-replicator', this.id);
            this.bind();
            this.resetButton();
            this.resetPositionButton();

            let deletedHandler = this.del.bind(this);
            $(this.itemSelector+'[data-deleted]', $root).each(function (){
                deletedHandler({target:this})
            })
        },

        resetPositionButton: function() {
            let sortableButton = this.$element.find('> .qor-sortable__button');

            if (this.isSortable) {
                if (this.getCurrentItems() > 1) {
                    sortableButton.show();
                } else {
                    sortableButton.hide();
                }
            }
        },

        getCurrentItems: function() {
            return this.$root.find(`> ${this.itemSelector}`).not('.is-deleted').length;
        },

        toggleButton: function(isHide) {
            let $button = this.$element.find(this.options.addClass);

            if (isHide) {
                $button.hide();
            } else {
                $button.show();
            }
        },

        resetButton: function() {
            if (this.maxitems <= this.getCurrentItems()) {
                this.toggleButton(true);
            } else {
                this.toggleButton();
            }
        },

        parseMultiple: function() {
            let template,
                name,
                fieldsetName = this.fieldsetName;

            for (let i = 0, len = fieldsetName.length; i < len; i++) {
                name = fieldsetName[i];
                template = this.initTemplate(this.template[name]);
                this.template[name] = template.template;
                this.index.push(template.index);
            }

            this.multipleIndex = _.max(this.index);
        },

        bind: function() {
            let options = this.options;

            if (this.canCreate) {
                this.$element
                    .on(EVENT_CLICK, options.addClass, this.add.bind(this))
            }

            this.$element
                .on(EVENT_CLICK, options.delClass, this.del.bind(this));
        },

        unbind: function() {
            this.$element.off(EVENT_CLICK);

            !this.isInSlideout && $(document).off(EVENT_SUBMIT, 'form');
            $(document)
                .off(EVENT_SLIDEOUTBEFORESEND, '.qor-slideout')
                .off(EVENT_SELECTCOREBEFORESEND);
        },

        add: function(e, data, isAutomatically) {
            if (!this.accept(e.target)) {
                return;
            }

            let options = this.options,
                $item,
                template;

            if (this.maxitems <= this.getCurrentItems()) {
                return false;
            }

            if (!isAutomatically) {
                var $target = $(e.target).closest(options.addClass),
                    templateName = $target.data('template'),
                    parents = $target.closest(this.$element),
                    parentsChildren = parents.children(options.childrenClass),
                    $fieldset = $target.closest(options.childrenClass).children('fieldset');
            }

            if (this.isMultipleTemplate) {
                this.parseNestTemplate(templateName);
                template = this.template[templateName];

                $item = $(template.replace(/(["\s][^"\s]+?)(\{\{index\}\})/g, '$1'+this.multipleIndex)).data(NAMESPACE+".new", true);

                for (let dataKey in $target.data()) {
                    if (dataKey.match(/^sync/)) {
                        let k = dataKey.replace(/^sync/, '');
                        $item.find("[name*='." + k + "']").val($target.data(dataKey));
                    }
                }

                if (this.$target) {
                    this.$target.append($item.show())
                } else {
                    if ($fieldset.length) {
                        $fieldset.last().after($item.show());
                    } else {
                        parentsChildren.prepend($item.show());
                    }
                }
                $item.data('itemIndex', this.multipleIndex).removeClass(this.options.newClass.substr(1));
                this.multipleIndex++;
            } else {
                if (!isAutomatically) {
                    $item = this.addSingle();
                    if (this.$target) {
                        this.$target.append($item.show())
                    } else {
                        $target.before($item.show());
                    }
                    this.index++;
                } else {
                    if (data && data.length) {
                        this.addMultiple(data);
                        $(document).trigger(EVENT_REPLICATORS_ADDED_DONE);
                    }
                }
            }

            if (!isAutomatically) {
                $item.trigger('enable');
                $(document).trigger(EVENT_REPLICATOR_ADDED, [$item]);
                e.stopPropagation();
            }

            this.resetPositionButton();
            this.resetButton();
        },

        addMultiple: function(data) {
            let $item;

            for (let i = 0, len = data.length; i < len; i++) {
                $item = this.addSingle();
                this.index++;
                $(document).trigger(EVENT_REPLICATORS_ADDED, [$item, data[i]]);
            }
        },

        addSingle: function() {
            let $item;

            $item = $(this.template.replace(/(="\S*?)(\{\{index\}\})/g, (input, prefix) => prefix+this.index));

            // add order property for sortable fieldset
            if (this.isSortable) {
                let order = this.$root.find('> .qor-sortable__item').length;
                $item.attr('order-index', order).css('order', order);
            }

            $item.data('itemIndex', this.index).removeClass(this.options.newClass.substr('.'));

            return $item.data(NAMESPACE+".new", true);
        },

        accept: function (el) {
            const $target = $(el),
                currentID = $target.parents('[data-qor-replicator]:eq(0)').attr('data-qor-replicator');

            return currentID === this.id
        },

        del: function(e) {
            const $target = $(e.target);

            if (!this.accept($target)) {
                return;
            }

            let $item = $target.is(this.itemSelector) ? $target : $target.closest(this.itemSelector),
                options = this.options,
                name = this.parseName($item),
                $alert;

            if ($item.data(NAMESPACE+".new")) {
                $item.remove();
            } else {
                $item
                    .addClass('is-deleted')
                    .children(':visible')
                    .addClass('hidden')
                    .hide();

                $alert = $(this.alertTemplate.replaceAll('{{name}}', name).replaceAll('{{id}}', $item.data().primaryKey));
                $alert.find(options.undoClass).one(
                    EVENT_CLICK,
                    function () {
                        if (this.maxitems <= this.getCurrentItems()) {
                            window.QOR.qorConfirm(this.$element.data('maxItemHint'));
                            return false;
                        }

                        $item.find('> ' + this.options.alertClass).remove();
                        $item
                            .removeClass('is-deleted')
                            .children('.hidden')
                            .removeClass('hidden')
                            .show();
                        this.resetButton();
                        this.resetPositionButton();
                    }.bind(this)
                );
                $item.append($alert);
            }
            this.resetButton();
            this.resetPositionButton();
        },

        parseNestTemplate: function(templateType) {
            let $element = this.$element,
                parentForm = $element.parents('.qor-fieldset-container'),
                index;

            if (parentForm.length) {
                index = $element.closest(this.itemSelector).data('itemIndex');
                if (index) {
                    if (templateType) {
                        this.template[templateType] = this.template[templateType].replace(/\[\d+\]/g, '[' + index + ']');
                    } else {
                        this.template = this.template.replace(/\[\d+\]/g, '[' + index + ']');
                    }
                }
            }
        },

        parseName: function($item) {
            let name = $item.find('input[name],select[name]').eq(0).attr('name') || $item.find('textarea[name]').attr('name');
            if (name) {
                name = name.split(".").slice(0, -1).join(".")
            }
            if (this.prefix) {
                name = this.prefix + name.substr(this.prefix.length).replace(/^(.+)[.|\[].+$/i, '$1')
            }
            return name
        },

        destroy: function() {
            this.unbind();
            this.$element.removeData(NAMESPACE);
        }
    };


    $.extend({}, QOR.messages, {
        replicator:{
            undoDelete: 'Undo Delete'
        }
    }, true);

    QorReplicator.DEFAULTS = {
        rootSelector: '',
        itemClass: '.qor-fieldset',
        newClass: '.qor-fieldset--new',
        addClass: '.qor-fieldset__add',
        delClass: '.qor-fieldset__delete',
        childrenClass: '.qor-field__block',
        undoClass: '.qor-fieldset__undo',
        sortableClass: '.qor-fieldset-sortable',
        alertClass: '.qor-fieldset__alert',
        alertTemplate: {
            tag: 'div',
            attrs: {
                class: 'qor-fieldset__alert',
            },
            body: '<input type="hidden" name="{{name}}._destroy" value="1"><input type="hidden" name="{{name}}.id" value="{{id}}">' +
                '<button class="mdl-button mdl-button--accent mdl-js-button mdl-js-ripple-effect qor-fieldset__undo" type="button" title="UNDO_DELETE_MESSAGE"><span class="material-icons">undo</span> </button>',
        }
    };

    QorReplicator.plugin = function(options) {
        return this.each(function() {
            let $this = $(this),
                data = $this.data(NAMESPACE),
                fn;

            if (!data) {
                if (!options) {
                    return;
                }
                $this.data(NAMESPACE, (data = new QorReplicator(this, options)));
            }

            if (!options) {
                if(data) {
                    return data
                }
            } else if (typeof options === 'string') {
                if ($.isFunction((fn = data[options]))) {
                    const res = fn.apply(data, Array.prototype.slice.call(arguments, 1));
                    if (res !== undefined) {
                        return res
                    }
                } else if ((options in data)) {
                    return data[options]
                }
            }
        });
    };

    $.fn.qorReplicator = QorReplicator.plugin

    $(function() {
        let selector = CLASS_CONTAINER;
        let options = {};

        $(document)
            .on(EVENT_DISABLE, function(e) {
                QorReplicator.plugin.call($(selector, e.target), 'destroy');
            })
            .on(EVENT_ENABLE, function(e) {
                QorReplicator.plugin.call($(selector, e.target), options);
            })
            .triggerHandler(EVENT_ENABLE);
    });

    return QorReplicator;
});

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

    let NAMESPACE = "qor.load_resource",
        SELECTOR = `input[data-toggle="${NAMESPACE}"]`,
        EVENT_ENABLE = 'enable.' + NAMESPACE,
        EVENT_DISABLE = 'disable.' + NAMESPACE,
        EVENT_CHANGE = 'change.' + NAMESPACE;

    function ResourceId(element, options) {
        this.$el = $(element);
        this.options = $.extend({}, this.$el.data(), options || {});
        this.init();
    }

    ResourceId.prototype = {
        constructor: ResourceId,

        init: function () {
            let $scope = this.$el.closest('form'),
                find = (expr) => {
                    if (!expr) return null;
                    let $el = $scope.find(expr);
                    return $el.length ? $el : null
                };
            if (!$scope.length) {
                $scope = $('body');
            }
            this.$scope = $scope;
            this.url = this.options.resourceUrl;
            this.param = this.options.param || "id";
            this.loading = false;
            this.$string = find(this.options.targetString);
            this.$error = find(this.options.targetError);
            this.$val = find(this.options.targetVal);
            if (this.$error && this.$error.length) {
                if (this.$error.is('.mdl-textfield__error')) {
                    this.errorHandler = {
                        show: function () {
                            this.$el.parents('.mdl-textfield:eq(0)').addClass('is-invalid')
                        }.bind(this),
                        hide: function () {
                            this.$el.parents('.mdl-textfield:eq(0)').removeClass('is-invalid')
                        }.bind(this),
                    }
                } else {
                    this.errorHandler = {
                        show: function () {
                            this.$error.show()
                        }.bind(this),
                        hide: function () {
                            this.$error.hide()
                        }.bind(this),
                    }
                }
            }
            this.$loading = this.options.targetLoading ? $scope.find(this.options.targetLoading).hide() : null;
            this.selectedTemplate = this.$el.siblings('template[name="selected-template"]').html().replace(/\[\[ *&amp;/g, '[[&');
            this.bind();
        },

        bind: function () {
            this.$el.bind(EVENT_CHANGE, this.find.bind(this));
        },

        enabledAll: function () {
        },

        setError: function (msg) {
            if(this.errorHandler) {
                this.$error.html(msg);
                this.errorHandler.show();
                this.$string.addClass('').html('').hide();
            } else {
                this.$string.addClass('mdl-textfield__error').html(msg);
            }
            if(this.$val) this.$val.val('')
        },

        setStringValue: function (original, result) {
            this.$string.removeClass('mdl-textfield__error').html(result).show();
            if(this.errorHandler) {
                this.$error.html('');
                this.errorHandler.hide();
            }

            const pk = original[this.options.primaryField || 'ID'];
            if (pk) {
                this.$el[0].value = pk;
            }
            if(this.$val) this.$val.val(pk || this.$el.val())
        },

        setResult: function (result) {
            if ($.isArray(result)) {
                switch (result.length) {
                    case 1:
                        result = result[0]
                        break
                    default:
                        result = null
                }
            }

            if (result === null) {
                this.setError(window.QOR.messages.common.recordNotFoundError)
            } else {
                this.setStringValue(result, Mustache.render(this.selectedTemplate, result));
            }
        },

        onResponse: function(response) {
            this.findDone();

            if (!response.ok) {
                this.setError("HTTP ERROR: " + response.status);
                return;
            }
            return response.json();
        },

        findDone: function(err) {
            this.$loading.hide();
            this.$el.removeAttr('disabled');
            this.loading = false;
            if (err) {
                console.log(err)
            }
        },

        find: function () {
            if (this.loading) {
                return
            }
            this.$el.removeClass('is-invalid');

            let val = this.$el.val().replace(/(^\s+|\s+$)/gms, ''),
                uri = this.url+(this.url.indexOf('?') !== -1 ? '&' : '?') +this.param+'='+encodeURI(val);
            uri = QOR.Xurl(uri, this.$val).build().url;

            if (val === "") {
                this.setResult(null)
                return;
            }

            this.loading = true;

            if (this.$loading)
                this.$loading.show();
            this.$el.attr('disabled', 'disabled');

            fetch(uri, { method: 'GET',
                headers: new Headers({
                    'Accept': 'application/json'
                }),
                cache: 'no-store' })
                .then(this.onResponse.bind(this))
                .then(this.setResult.bind(this))
                .catch(this.findDone.bind(this));
        },

        destroy: function () {
            this.$el.off(EVENT_CHANGE, this.find);
            this.$el.removeData(NAMESPACE);
            this.$el.unmask();
        }
    };

    ResourceId.DEFAULTS = {};

    ResourceId.plugin = function (options) {
        let args = Array.prototype.slice.call(arguments, 1),
            result;

        this.each(function () {
            let $this = $(this),
                data = $this.data(NAMESPACE),
                fn;

            if (!data) {
                if (/destroy/.test(options)) {
                    return;
                }

                if (typeof options === "object")
                    options = $.extend({}, options, true)

                data = new ResourceId(this, options);
                if (("$el" in data)) {
                    $this.data(NAMESPACE, data);
                } else {
                    return
                }
            }

            if (typeof options === 'string' && $.isFunction((fn = data[options]))) {
                result = fn.apply(data, args);
            }
        });

        return (typeof result === 'undefined') ? this : result;
    };

    $(function () {
        let options = {};

        $(document)
            .on(EVENT_DISABLE, function (e) {
                ResourceId.plugin.call($(SELECTOR, e.target), 'destroy');
            })
            .on(EVENT_ENABLE, function (e) {
                ResourceId.plugin.call($(SELECTOR, e.target), options);
            })
            .triggerHandler(EVENT_ENABLE);
    });

    return ResourceId;
});
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
  var location = window.location;
  var componentHandler = window.componentHandler;
  var history = window.history;
  var NAMESPACE = 'qor.globalSearch';
  var EVENT_ENABLE = 'enable.' + NAMESPACE;
  var EVENT_DISABLE = 'disable.' + NAMESPACE;
  var EVENT_CLICK = 'click.' + NAMESPACE;

  var SEARCH_RESOURCE = '.qor-global-search--resource';
  var SEARCH_RESULTS = '.qor-global-search--results';
  var QOR_TABLE = '.qor-table';
  var IS_ACTIVE = 'is-active';

  function QorSearchCenter(element, options) {
    this.$element = $(element);
    this.options = $.extend({}, QorSearchCenter.DEFAULTS, $.isPlainObject(options) && options);
    this.init();
  }

  QorSearchCenter.prototype = {
    constructor: QorSearchCenter,

    init: function () {
      this.bind();
      this.initTab();
    },

    bind: function () {
      this.$element.on(EVENT_CLICK, $.proxy(this.click, this));
    },

    unbind: function () {
      this.$element.off(EVENT_CLICK, this.check);
    },

    initTab: function () {
      var locationSearch = location.search;
      var resourceName;
      if (/resource_name/.test(locationSearch)){
        resourceName = locationSearch.match(/resource_name=\w+/g).toString().split('=')[1];
        $(SEARCH_RESOURCE).removeClass(IS_ACTIVE);
        $('[data-resource="' + resourceName + '"]').addClass(IS_ACTIVE);
      }
    },

    click : function (e) {
      let $target = $(e.target),
        data = $target.data();

      if ($target.is(SEARCH_RESOURCE)){
        let url = QOR.Xurl(location.href, this.$element),
          newUrl;

        if (data.resource) {
          url.qset('resource_name', data.resource)
        } else {
          url.qdel('resource_name')
        }

        url.qdel('keyword');

        newUrl = url.toString();

        if (history.pushState){
          this.fetchSearch(newUrl, $target);
        } else {
          location.href = newUrl;
        }
      }
    },

    fetchSearch: function (url,$target) {
      var title = document.title;

      $.ajax(url, {
        method: 'GET',
        dataType: 'html',
        beforeSend: function () {
          $('.mdl-spinner').remove();
          $(SEARCH_RESULTS).prepend('<div class="mdl-spinner mdl-js-spinner is-active"></div>').find('.qor-section').hide();
          componentHandler.upgradeElement(document.querySelector('.mdl-spinner'));
        },
        success: function (html) {
          var result = $(html).find(SEARCH_RESULTS).html();
          $(SEARCH_RESOURCE).removeClass(IS_ACTIVE);
          $target.addClass(IS_ACTIVE);
          // change location URL without refresh page
          history.pushState({ Page: url, Title: title }, title, url);
          $('.mdl-spinner').remove();
          $(SEARCH_RESULTS).removeClass('loading').html(result);
          componentHandler.upgradeElements(document.querySelectorAll(QOR_TABLE));
        },
        error: function (xhr, textStatus, errorThrown) {
          $(SEARCH_RESULTS).find('.qor-section').show();
          $('.mdl-spinner').remove();
          window.alert([textStatus, errorThrown].join(': '));
        }
      });
    },

    destroy: function () {
      this.unbind();
      this.$element.removeData(NAMESPACE);
    }

  };

  QorSearchCenter.DEFAULTS = {
  };

  QorSearchCenter.plugin = function (options) {
    return this.each(function () {
      var $this = $(this);
      var data = $this.data(NAMESPACE);
      var fn;

      if (!data) {
        $this.data(NAMESPACE, (data = new QorSearchCenter(this, options)));
      }

      if (typeof options === 'string' && $.isFunction(fn = data[options])) {
        fn.call(data);
      }
    });
  };

  $(function () {
    var selector = '[data-toggle="qor.global.search"]';
    var options = {};

    $(document).
      on(EVENT_DISABLE, function (e) {
        QorSearchCenter.plugin.call($(selector, e.target), 'destroy');
      }).
      on(EVENT_ENABLE, function (e) {
        QorSearchCenter.plugin.call($(selector, e.target), options);
      }).
      triggerHandler(EVENT_ENABLE);
  });

  return QorSearchCenter;

});

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

    let FormData = window.FormData,
        NAMESPACE = 'qor.selectcore',
        EVENT_SELECTCORE_BEFORESEND = 'selectcoreBeforeSend.' + NAMESPACE,
        EVENT_ONSELECT = 'afterSelected.' + NAMESPACE,
        EVENT_ONSUBMIT = 'afterSubmitted.' + NAMESPACE,
        EVENT_CLICK = 'click.' + NAMESPACE,
        EVENT_SUBMIT = 'submit.' + NAMESPACE,
        CLASS_CLICK_TABLE = '.qor-table-container tbody tr';

    function QorSelectCore(element, options) {
        this.$element = $(element);
        this.options = $.extend({}, QorSelectCore.DEFAULTS, $.isPlainObject(options) && options);
        this.init();
    }

    QorSelectCore.prototype = {
        constructor: QorSelectCore,

        init: function () {
            this.bind();
        },

        bind: function () {
            this.$element.on(EVENT_CLICK, CLASS_CLICK_TABLE, this.processingData.bind(this));
            //.on(EVENT_SUBMIT, 'form', this.submit.bind(this));
            //this.$element.on('enable.qor.asyncFormSubmiter', 'form', this.newFormHandle.bind(this))
        },

        unbind: function () {
            this.$element.off(EVENT_CLICK, '.qor-table tbody tr').off(EVENT_SUBMIT, 'form');
        },

        newFormHandle: function ($form, openPage) {
            $form.qorAsyncFormSubmiter('updateOptions', {
                onBeforeSubmit: function (jqXHR, cfg) {
                    if (!cfg.continueEditing) {
                        cfg.headers['X-Redirection-Disabled'] = true;
                        cfg.headers['X-Accept'] = 'json';
                    }
                    cfg.headers['X-Flash-Messages-Disabled'] = true;
                }.bind(this),
                onSubmitSuccess: function (data, statusText, jqXHR) {
                    let onSelect = this.options.onSelect, jsonData;
                    try {
                        jsonData = JSON.parse(data);
                    } catch (e) {
                        return
                    }

                    if (onSelect && $.isFunction(onSelect)) {
                        onSelect(jsonData, undefined);
                        $(document).trigger(EVENT_ONSELECT);
                    }
                }.bind(this),
                openPage: openPage
            });
        },

        processingData: function (e) {
            let $this = $(e.target).closest('tr'),
                data = {},
                url,
                options = this.options,
                onSelect = options.onSelect;

            data = $this.data();
            data = $.extend({}, {}, data);
            data.$clickElement = $this;

            url = data.mediaLibraryUrl || data.url;

            if (url) {
                $.getJSON(url, function (json) {
                    json.MediaOption && (json.MediaOption = JSON.parse(json.MediaOption));
                    data = $.extend({}, json, data);
                    if (onSelect && $.isFunction(onSelect)) {
                        onSelect(data, e);
                        $(document).trigger(EVENT_ONSELECT);
                    }
                });
            } else {
                if (onSelect && $.isFunction(onSelect)) {
                    onSelect(data, e);
                    $(document).trigger(EVENT_ONSELECT);
                }
            }
            return false;
        },

        submit: function (e) {
            let form = e.target,
                $form = $(form),
                _this = this,
                $submit = $form.find(':submit'),
                data,
                onSubmit = this.options.onSubmit;

            if ($form.parents('.qor-page__new').length) {
                return
            }

            $(document).trigger(EVENT_SELECTCORE_BEFORESEND);

            if (FormData) {
                e.preventDefault();

                $.ajax($form.prop('action'), {
                    method: $form.prop('method'),
                    data: new FormData(form),
                    dataType: 'json',
                    processData: false,
                    contentType: false,
                    headers: {
                        'X-Layout': 'lite',
                        'X-Redirection-Disabled': true
                    },
                    beforeSend: function () {
                        $form
                            .parent()
                            .find('.qor-error')
                            .remove();
                        $submit.prop('disabled', true);
                    },
                    success: function (json) {
                        data = json;
                        data.primaryKey = data.ID;

                        $('.qor-error').remove();

                        if (onSubmit && $.isFunction(onSubmit)) {
                            onSubmit(data, e);
                            $(document).trigger(EVENT_ONSUBMIT);
                        } else {
                            _this.refresh();
                        }
                    },
                    error: function (xhr, textStatus, errorThrown) {
                        let error;

                        if (xhr.responseJSON) {
                            error = `<ul class="qor-error">
                                        <li><label>
                                            <i class="material-icons">error</i>
                                            <span>${xhr.responseJSON.errors[0]}</span>
                                        </label></li>
                                    </ul>`;
                        } else {
                            error = `<ul class="qor-error">${$(xhr.responseText)
                                .find('#errors')
                                .html()}</ul>`;
                        }

                        $('.qor-bottomsheets .qor-page__body').scrollTop(0);

                        if (xhr.status === 422 && error) {
                            $form.before(error);
                        } else {
                            window.alert([textStatus, errorThrown].join(': '));
                        }
                    },
                    complete: function () {
                        $submit.prop('disabled', false);
                    }
                });
            }
        },

        refresh: function () {
            setTimeout(function () {
                window.location.reload();
            }, 350);
        },

        destroy: function () {
            this.unbind();
        }
    };

    QorSelectCore.plugin = function (options) {
        let args = Array.prototype.slice.call(arguments, 1);
        return this.each(function () {
            let $this = $(this),
                data = $this.data(NAMESPACE),
                fn;

            if (!data) {
                if (/destroy/.test(options)) {
                    return;
                }
                $this.data(NAMESPACE, (data = new QorSelectCore(this, options)));
            }

            if (typeof options === 'string' && $.isFunction((fn = data[options]))) {
                fn.apply(data, args);
            }
        });
    };

    $.fn.qorSelectCore = QorSelectCore.plugin;

    return QorSelectCore;
});

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

    let $body = $('body'),
        $document = $(document),
        Mustache = window.Mustache,
        NAMESPACE = 'qor.selectone',
        PARENT_NAMESPACE = 'qor.bottomsheets',
        EVENT_CLICK = 'click.' + NAMESPACE,
        EVENT_ENABLE = 'enable.' + NAMESPACE,
        EVENT_DISABLE = 'disable.' + NAMESPACE,
        EVENT_RELOAD = 'reload.' + PARENT_NAMESPACE,
        CLASS_CLEAR_SELECT = '.qor-selected-many__remove',
        CLASS_UNDO_DELETE = '.qor-selected-many__undo',
        CLASS_DELETED_ITEM = 'qor-selected-many__deleted',
        CLASS_SELECT_FIELD = '.qor-field__selected-many',
        CLASS_SELECT_INPUT = '.qor-field__selectmany-input',
        CLASS_SELECT_ICON = '.qor-select__select-icon',
        CLASS_SELECT_HINT = '.qor-selectmany__hint',
        CLASS_PARENT = '.qor-field__selectmany',
        CLASS_SELECTED = 'is_selected',
        CLASS_MANY = 'qor-bottomsheets__select-many';

    function QorSelectMany(element, options) {
        this.$element = $(element);
        this.options = $.extend({}, QorSelectMany.DEFAULTS, $.isPlainObject(options) && options);
        this.init();
    }

    var lock = {lock: false};

    QorSelectMany.prototype = {
        constructor: QorSelectMany,

        init: function() {
            this.bind();
        },

        bind: function() {
            $document
                .on(EVENT_RELOAD, `.${CLASS_MANY}`, this.reloadData.bind(this));

            this.$element
                .on(EVENT_CLICK, CLASS_CLEAR_SELECT, this.clearSelect.bind(this))
                .on(EVENT_CLICK, '[data-select-modal="many"]', this.openBottomSheets.bind(this))
                .on(EVENT_CLICK, CLASS_UNDO_DELETE, this.undoDelete.bind(this));
        },

        unbind: function() {
            $document.off(EVENT_RELOAD, `.${CLASS_MANY}`);
            this.$element.off(EVENT_CLICK, CLASS_CLEAR_SELECT).off(EVENT_CLICK, '[data-select-modal="many"]').off(EVENT_CLICK, CLASS_UNDO_DELETE);
        },

        clearSelect: function(e) {
            var $target = $(e.target),
                $selectFeild = $target.closest(CLASS_PARENT);

            $target.closest('[data-primary-key]').addClass(CLASS_DELETED_ITEM);
            this.updateSelectInputData($selectFeild);

            return false;
        },

        undoDelete: function(e) {
            var $target = $(e.target),
                $selectFeild = $target.closest(CLASS_PARENT);

            $target.closest('[data-primary-key]').removeClass(CLASS_DELETED_ITEM);
            this.updateSelectInputData($selectFeild);

            return false;
        },

        openBottomSheets: function(e) {
            if (lock.lock) {
                e.preventDefault();
                return false;
            }

            lock.lock = true;
            setTimeout(function () {lock.lock = false}, 1000*3);

            let $this = $(e.target),
                data = $this.data();

            this.BottomSheets = $body.data('qor.bottomsheets');
            this.bottomsheetsData = data;

            this.$selector = data.selectId ? $(data.selectId) : $this.closest(CLASS_PARENT).find('select');
            this.$selectFeild = this.$selector.closest(CLASS_PARENT).find(CLASS_SELECT_FIELD);

            // select many templates
            this.SELECT_MANY_SELECTED_ICON = $('[name="select-many-selected-icon"]').html();
            this.SELECT_MANY_UNSELECTED_ICON = $('[name="select-many-unselected-icon"]').html();
            this.SELECT_MANY_HINT = $('[name="select-many-hint"]').html();
            this.SELECT_MANY_TEMPLATE = $('[name="select-many-template"]').html();

            data.url = data.selectListingUrl;

            if (data.selectDefaultCreating) {
                data.url = data.selectCreatingUrl;
            }

            this.BottomSheets.open(data, this.handleSelectMany.bind(this));
        },

        reloadData: function() {
            this.initItems();
        },

        renderSelectMany: function(data) {
            const res = Mustache.render(this.SELECT_MANY_TEMPLATE, data)
            return res;
        },

        renderHint: function(data) {
            const res = Mustache.render(this.SELECT_MANY_HINT, data);
            return res
        },

        initItems: function() {
            var $tr = this.$bottomsheets.find('tbody tr'),
                selectedIconTmpl = this.SELECT_MANY_SELECTED_ICON,
                unSelectedIconTmpl = this.SELECT_MANY_UNSELECTED_ICON,
                selectedIDs = [],
                primaryKey,
                $selectedItems = this.$selectFeild.find('[data-primary-key]').not('.' + CLASS_DELETED_ITEM);

            $selectedItems.each(function() {
                selectedIDs.push($(this).data().primaryKey);
            });

            $tr.each(function() {
                var $this = $(this),
                    $td = $this.find('td:first');

                primaryKey = $this.data().primaryKey;

                if (selectedIDs.indexOf(primaryKey) != '-1') {
                    $this.addClass(CLASS_SELECTED);
                    $td.append(selectedIconTmpl);
                } else {
                    $td.append(unSelectedIconTmpl);
                }
            });

            this.updateHint(this.getSelectedItemData());
        },

        getSelectedItemData: function() {
            var selecedItems = this.$selectFeild.find('[data-primary-key]').not('.' + CLASS_DELETED_ITEM);
            return {
                selectedNum: selecedItems.length
            };
        },

        updateHint: function(data) {
            var template;

            $.extend(data, this.bottomsheetsData);
            template = this.renderHint(data);

            this.$bottomsheets.find(CLASS_SELECT_HINT).remove();
            this.$bottomsheets.find('.qor-page__body').before(template);
        },

        updateSelectInputData: function($selectFeild) {
            var $selectList = $selectFeild ? $selectFeild : this.$selectFeild,
                $selectedItems = $selectList.find('[data-primary-key]').not('.' + CLASS_DELETED_ITEM),
                $selector = $selectFeild ? $selectFeild.find(CLASS_SELECT_INPUT) : this.$selector,
                $options = $selector.find('option'),
                $option,
                data,
                primaryKey;

            $options.prop('selected', false);

            $selectedItems.each(function() {
                primaryKey = $(this).data().primaryKey;
                $option = $options.filter('[value="' + primaryKey + '"]');

                if (!$option.length) {
                    data = {
                        primaryKey: primaryKey,
                        displayName: ''
                    };
                    const res = Mustache.render(QorSelectMany.SELECT_MANY_OPTION_TEMPLATE, data)
                    $option = $(res);
                    $selector.append($option);
                }

                $option.prop('selected', true);
            });
        },

        changeIcon: function($ele, template) {
            $ele.find(CLASS_SELECT_ICON).remove();
            $ele.find('td:first').prepend(template);
        },

        removeItem: function(data) {
            var primaryKey = data.primaryKey;

            this.$selectFeild.find('[data-primary-key="' + primaryKey + '"]').find(CLASS_CLEAR_SELECT).click();
            this.changeIcon(data.$clickElement, this.SELECT_MANY_UNSELECTED_ICON);
        },

        addItem: function(data, isNewData) {
            var template = this.renderSelectMany(data),
                $option,
                $list = this.$selectFeild.find('[data-primary-key="' + data.primaryKey + '"]');

            if ($list.length) {
                if ($list.hasClass(CLASS_DELETED_ITEM)) {
                    $list.removeClass(CLASS_DELETED_ITEM);
                    this.updateSelectInputData();
                    this.changeIcon(data.$clickElement, this.SELECT_MANY_SELECTED_ICON);
                    return;
                } else {
                    return;
                }
            }

            this.$selectFeild.append(template);

            if (isNewData) {
                const res = Mustache.render(QorSelectMany.SELECT_MANY_OPTION_TEMPLATE, data);
                $option = $(res);
                $option.appendTo(this.$selector);
                $option.prop('selected', true);
                this.$bottomsheets.remove();
                if (!$('.qor-bottomsheets').is(':visible')) {
                    $('body').removeClass('qor-bottomsheets-open');
                }
                return;
            }

            this.changeIcon(data.$clickElement, this.SELECT_MANY_SELECTED_ICON);
        },

        handleSelectMany: function($bottomsheets) {
            let options = {
                onSelect: this.onSelectResults.bind(this), // render selected item after click item lists
                onSubmit: this.onSubmitResults.bind(this) // render new items after new item form submitted
            };

            $bottomsheets.qorSelectCore(options).addClass(CLASS_MANY);
            this.$bottomsheets = $bottomsheets;
            this.initItems();
        },

        onSelectResults: function(data) {
            this.handleResults(data);
        },

        onSubmitResults: function(data) {
            this.handleResults(data, true);
        },

        handleResults: function(data, isNewData) {
            var firstKey = function() {
                var keys = Object.keys(data);
                if (keys.length > 1 && keys[0] == "ID") {
                    return keys[1];
                }
                return keys[0];
            };

            data.displayName = data.Text || data.Name || data.Title || data.Code || data.Value || data[firstKey()];

            if (isNewData) {
                this.addItem(data, true);
                return;
            }

            var $element = data.$clickElement,
                isSelected;

            $element.toggleClass(CLASS_SELECTED);
            isSelected = $element.hasClass(CLASS_SELECTED);

            if (isSelected) {
                this.addItem(data);
            } else {
                this.removeItem(data);
            }

            this.updateHint(this.getSelectedItemData());
            this.updateSelectInputData();
        },

        destroy: function() {
            this.unbind();
            this.$element.removeData(NAMESPACE);
        }
    };

    QorSelectMany.SELECT_MANY_OPTION_TEMPLATE = '<option value="[[ primaryKey ]]" >[[ displayName ]]</option>';

    QorSelectMany.plugin = function(options) {
        return this.each(function() {
            var $this = $(this);
            var data = $this.data(NAMESPACE);
            var fn;

            if (!data) {
                if (/destroy/.test(options)) {
                    return;
                }

                $this.data(NAMESPACE, (data = new QorSelectMany(this, options)));
            }

            if (typeof options === 'string' && $.isFunction((fn = data[options]))) {
                fn.apply(data);
            }
        });
    };

    $(function() {
        var selector = '[data-toggle="qor.selectmany"]';
        $(document)
            .on(EVENT_DISABLE, function(e) {
                QorSelectMany.plugin.call($(selector, e.target), 'destroy');
            })
            .on(EVENT_ENABLE, function(e) {
                QorSelectMany.plugin.call($(selector, e.target));
            })
            .triggerHandler(EVENT_ENABLE);
    });

    return QorSelectMany;
});

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

  let $body = $('body'),
    $document = $(document),
    Mustache = window.Mustache,
    NAMESPACE = 'qor.selectone',
    PARENT_NAMESPACE = 'qor.bottomsheets',
    EVENT_CLICK = 'click.' + NAMESPACE,
    EVENT_ENABLE = 'enable.' + NAMESPACE,
    EVENT_DISABLE = 'disable.' + NAMESPACE,
    EVENT_RELOAD = 'reload.' + PARENT_NAMESPACE,
    CLASS_CLEAR_SELECT = '.qor-selected__remove',
    CLASS_CHANGE_SELECT = '.qor-selected__change',
    CLASS_SELECT_FIELD = '.qor-field__selected',
    CLASS_SELECT_INPUT = '.qor-field__selectone-input',
    CLASS_SELECT_TRIGGER = '.qor-field__selectone-trigger',
    CLASS_PARENT = '.qor-field__selectone',
    CLASS_SELECTED = 'is_selected',
    CLASS_ONE = 'qor-bottomsheets__select-one';

  function QorSelectOne(element, options) {
    this.$element = $(element);
    this.options = $.extend({}, QorSelectOne.DEFAULTS, $.isPlainObject(options) && options);
    this.selectedRender = null;
    this.init();
  }

  function firstTextKey(obj) {
    var keys = Object.keys(obj);
    if (keys.length > 1 && keys[0] === "ID") {
      return keys[1];
    }
    return keys[0];
  }

  var glock = {
    lock: false,
    id: 0
  }

  function gunlock() {
    glock.lock = false
  }

  QorSelectOne.prototype = {
    constructor: QorSelectOne,

    id: undefined,

    init: function() {
      let selectedRender = this.$element.data().selectedRender;
      if (selectedRender) {
        eval('this.selectedRender = function(data){'+atob(selectedRender)+'};');
      }
      this.id = glock.id++;
      this.$selectOneSelectedTemplate = this.$element.find('[name="select-one-selected-template"]');
      this.$selectOneSelectedIconTemplate = this.$element.find('[name="select-one-selected-icon"]');
      this.lock = {
        $parent: null,
        $select: null,
      };
      this.bind();
    },

    bind: function() {
      $document
        .on(EVENT_RELOAD, `.${CLASS_ONE}`, this.reloadData.bind(this));
      this.$element
        .on(EVENT_CLICK, CLASS_CLEAR_SELECT, this.clearSelect.bind(this))
        .on(EVENT_CLICK, '[data-selectone-url],[data-selectone-url] .material-icons', this.openBottomSheets.bind(this))
        .on(EVENT_CLICK, CLASS_CHANGE_SELECT, this.changeSelect);
    },

    unbind: function() {
      $document.off(EVENT_RELOAD, `.${CLASS_ONE}`);
      this.$element.off(EVENT_CLICK, CLASS_CLEAR_SELECT).off(EVENT_CLICK, '[data-selectone-url]').off(EVENT_CLICK, CLASS_CHANGE_SELECT);
    },

    clearSelect: function(e) {
      var $target = $(e.target),
        $parent = $target.closest(CLASS_PARENT);

      $parent.find(CLASS_SELECT_FIELD).remove();
      $parent.find(CLASS_SELECT_INPUT).html('<option value="" selected></option>');
      $parent.find(CLASS_SELECT_INPUT)[0].value = '';
      $parent.find(CLASS_SELECT_TRIGGER).show();

      $parent.trigger('qor.selectone.unselected');
      return false;
    },

    changeSelect: function() {
      var $target = $(this),
          $parent = $target.closest(CLASS_PARENT);

      $parent.find(CLASS_SELECT_TRIGGER).trigger('click');
    },

    openBottomSheets: function (e) {
      if (glock.lock) {
        e.preventDefault();
        return false;
      }

      glock.lock = true;
      setTimeout(gunlock, 1000*3);
      let $this = $(e.target);
      if ($this.is('.material-icons')) {
        $this = $this.parent()
      }
      this.lock.currentData = $this.data();

      this.lock.$parent = $this.closest(CLASS_PARENT);
      this.lock.$select = this.lock.$parent.find('select');

      this.lock.currentData.url = this.lock.currentData.selectoneUrl;
      this.lock.primaryField = this.lock.currentData.remoteDataPrimaryKey;
      this.lock.displayField = this.lock.currentData.remoteDataDisplayKey;
      this.lock.iconField = this.lock.currentData.remoteDataIconKey;

      this.lock.SELECT_ONE_SELECTED_ICON = this.$selectOneSelectedIconTemplate.html();
      let data = $.extend({}, this.lock.currentData);
      if (this.lock.$select.length) {
        data.$element = this.lock.$select;
      }
      $('body').qorBottomSheets('open', data, this.handleSelectOne.bind(this));
    },

    initItem: function() {
      var $selectField = this.lock.$parent.find(CLASS_SELECT_FIELD),
          recordeUrl = this.lock.currentData.remoteRecordeUrl,
          selectedID;

      if (recordeUrl) {
        this.lock.$bottomsheets.find('tr[data-primary-key]').each(function () {
          var $this = $(this), data = $this.data();
          data.url = recordeUrl.replace("{ID}", data.primaryKey)
        })
      }

      if (!$selectField.length) {
        return;
      }

      selectedID = $selectField.data().primaryKey;

      if (selectedID) {
        this.lock.$bottomsheets
          .find('tr[data-primary-key="' + selectedID + '"]')
          .addClass(CLASS_SELECTED)
          .find('td:first')
          .append(this.lock.SELECT_ONE_SELECTED_ICON);
      }
    },

    reloadData: function() {
      this.initItem();
    },

    renderSelectOne: function(data) {
      const res = Mustache.render(this.$selectOneSelectedTemplate.html().replace(/\[\[ *&amp;/g, '[[&'), data);
      return res;
    },

    handleSelectOne: function($bottomsheets) {
      var options = {
        onSelect: this.onSelectResults.bind(this), //render selected item after click item lists
        onSubmit: this.onSubmitResults.bind(this) //render new items after new item form submitted
      };

      $bottomsheets.qorSelectCore(options).addClass(CLASS_ONE);
      this.lock.$bottomsheets = $bottomsheets;
      this.initItem();
    },

    onSelectResults: function(data) {
      this.handleResults(data);
    },

    onSubmitResults: function(data) {
      this.handleResults(data, true);
    },

    handleResults: function(data) {
      let template,
          $parent = this.lock.$parent,
          $selectField = $parent.find(CLASS_SELECT_FIELD);

      data.displayName = this.lock.displayField ? data[this.lock.displayField] :
          (data.Text || data.Name || data.Title || data.Value || data.Code || firstTextKey(data));
      data.selectoneValue = this.lock.primaryField ? data[this.lock.primaryField] : (data.primaryKey || data.ID);

      if (this.lock.iconField) {
          data.icon = data[this.lock.iconField];
      }

      if (data.icon && /\.svg/.test(data.icon)) {
          data.iconSVG = true;
      }

      if (!this.lock.$select.length) {
        return;
      }

      if (this.selectedRender) {
        data.displayText = this.selectedRender(data)
      }
      template = this.renderSelectOne(data);

      if ($selectField.length) {
        $selectField.remove();
      }

      $parent.prepend(template);
      $parent.find(CLASS_SELECT_TRIGGER).hide();

      const res = Mustache.render(QorSelectOne.SELECT_ONE_OPTION_TEMPLATE, data);
      this.lock.$select.html(res);
      // this.lock.$select[0].value = data.primaryKey || data.ID;

      $parent.trigger('qor.selectone.selected', [data]);

      this.lock.$bottomsheets.qorSelectCore('destroy').remove();
      if (!$('.qor-bottomsheets').is(':visible')) {
        $('body').removeClass('qor-bottomsheets-open');
      }
    },

    destroy: function() {
      this.unbind();
      this.$element.removeData(NAMESPACE);
      this.lock = undefined;
    }
  };

  QorSelectOne.SELECT_ONE_OPTION_TEMPLATE = '<option value="[[ selectoneValue ]]" selected>[[ displayName ]]</option>';

  QorSelectOne.plugin = function(options) {
    let args = Array.prototype.slice.call(arguments, 1);
    return this.each(function() {
      let $this = $(this),
        data = $this.data(NAMESPACE),
        fn;

      if (!data) {
        if (/destroy/.test(options)) {
          return;
        }

        $this.data(NAMESPACE, (data = new QorSelectOne(this, options)));
      }

      if (typeof options === 'string' && $.isFunction((fn = data[options]))) {
        fn.apply(data, args);
      }
    });
  };

  $(function() {
    const selector = '[data-toggle="qor.selectone"]';
    $(document)
      .on(EVENT_DISABLE, function(e) {
        QorSelectOne.plugin.call($(selector, e.target), 'destroy');
      })
      .on(EVENT_ENABLE, function(e) {
        QorSelectOne.plugin.call($(selector, e.target));
      })
      .triggerHandler(EVENT_ENABLE);
  });

  return QorSelectOne;
});

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

    var $document = $(document);
    var NAMESPACE = 'qor.selector';
    var EVENT_ENABLE = 'enable.' + NAMESPACE;
    var EVENT_DISABLE = 'disable.' + NAMESPACE;
    var EVENT_CLICK = 'click.' + NAMESPACE;
    var EVENT_SELECTOR_CHANGE = 'selectorChanged.' + NAMESPACE;
    var CLASS_OPEN = 'open';
    var CLASS_ACTIVE = 'active';
    var CLASS_HOVER = 'hover';
    var CLASS_SELECTED = 'selected';
    var CLASS_DISABLED = 'disabled';
    var CLASS_CLEARABLE = 'clearable';
    var SELECTOR_SELECTED = '.' + CLASS_SELECTED;
    var SELECTOR_TOGGLE = '.qor-selector-toggle';
    var SELECTOR_LABEL = '.qor-selector-label';
    var SELECTOR_CLEAR = '.qor-selector-clear';
    var SELECTOR_MENU = '.qor-selector-menu';
    var CLASS_BOTTOMSHEETS = '.qor-bottomsheets';

    function QorSelector(element, options) {
        this.options = options;
        this.$element = $(element);
        this.init();
    }

    QorSelector.prototype = {
        constructor: QorSelector,

        init: function() {
            var $this = this.$element;

            this.placeholder = $this.attr('placeholder') || $this.attr('name') || 'Select';
            this.build();
        },

        build: function() {
            var $this = this.$element;
            var $selector = $(QorSelector.TEMPLATE);
            var alignedClass = this.options.aligned + '-aligned';
            var data = {};
            var eleData = $this.data();
            var hover = eleData.hover;
            var paramName = $this.attr('name');

            this.isBottom = eleData.position == 'bottom';

            hover && $selector.addClass(CLASS_HOVER);

            $selector.addClass(alignedClass).find(SELECTOR_MENU).html(function() {
                var list = [];

                $this.children().each(function() {
                    var $this = $(this);
                    var selected = $this.attr('selected');
                    var disabled = $this.attr('disabled');
                    var value = $this.attr('value');
                    var label = $this.text();
                    var classNames = [];

                    if (selected) {
                        classNames.push(CLASS_SELECTED);
                        data.value = value;
                        data.label = label;
                        data.paramName = paramName;
                    }

                    if (disabled) {
                        classNames.push(CLASS_DISABLED);
                    }

                    list.push(
                        '<li' +
                        (classNames.length ? ' class="' + classNames.join(' ') + '"' : '') +
                        ' data-value="' + value + '"' +
                        ' data-label="' + label + '"' +
                        ' data-param-name="' + paramName + '"' +
                        '>' +
                        label +
                        '</li>'
                    );
                });

                return list.join('');
            });

            this.$selector = $selector;
            $this.hide().after($selector);
            $selector.find(SELECTOR_TOGGLE).data('paramName', paramName);
            this.pick(data, true);
            this.bind();
        },

        unbuild: function() {
            this.unbind();
            this.$selector.remove();
            this.$element.show();
        },

        bind: function() {
            this.$selector.on(EVENT_CLICK, $.proxy(this.click, this));
            $document.on(EVENT_CLICK, $.proxy(this.close, this));
        },

        unbind: function() {
            this.$selector.off(EVENT_CLICK, this.click);
            $document.off(EVENT_CLICK, this.close);
        },

        click: function(e) {
            var $target = $(e.target);

            e.stopPropagation();

            if ($target.is(SELECTOR_CLEAR)) {
                this.clear();
            } else if ($target.is('li')) {
                if (!$target.hasClass(CLASS_SELECTED) && !$target.hasClass(CLASS_DISABLED)) {
                    this.pick($target.data());
                }

                this.close();
            } else if ($target.closest(SELECTOR_TOGGLE).length) {
                this.open();
            }
        },

        pick: function(data, initialized) {
            var $selector = this.$selector;
            var selected = !!data.value;
            var $element = this.$element;

            $selector.
            find(SELECTOR_TOGGLE).
            toggleClass(CLASS_ACTIVE, selected).
            toggleClass(CLASS_CLEARABLE, selected && this.options.clearable).
            find(SELECTOR_LABEL).
            text(data.label || this.placeholder);

            if (!initialized) {
                $selector.
                find(SELECTOR_MENU).
                children('[data-value="' + data.value + '"]').
                addClass(CLASS_SELECTED).
                siblings(SELECTOR_SELECTED).
                removeClass(CLASS_SELECTED);

                $element.val(data.value);


                if ($element.closest(CLASS_BOTTOMSHEETS).length && !$element.closest('[data-toggle="qor.filter"]').length) {
                    // If action is in bottom sheet, will trigger filterChanged.qor.selector event, add passed data.value parameter to event.
                    $(CLASS_BOTTOMSHEETS).trigger(EVENT_SELECTOR_CHANGE, [data.value, data.paramName]);
                } else {
                    $element.trigger('change');
                }
            }
        },

        clear: function() {
            var $element = this.$element;

            this.$selector.
            find(SELECTOR_TOGGLE).
            removeClass(CLASS_ACTIVE).
            removeClass(CLASS_CLEARABLE).
            find(SELECTOR_LABEL).
            text(this.placeholder).
            end().
            end().
            find(SELECTOR_MENU).
            children(SELECTOR_SELECTED).
            removeClass(CLASS_SELECTED);

            $element.val('').trigger('change');
        },

        open: function() {

            // Close other opened dropdowns first
            $document.triggerHandler(EVENT_CLICK);
            $('.qor-filter__dropdown').hide();

            // Open the current dropdown
            this.$selector.addClass(CLASS_OPEN);
            if (this.isBottom) {
                this.$selector.addClass('bottom');
            }
        },

        close: function() {
            this.$selector.removeClass(CLASS_OPEN);
            if (this.isBottom) {
                this.$selector.removeClass('bottom');
            }
        },

        destroy: function() {
            this.unbuild();
            this.$element.removeData(NAMESPACE);
        }
    };

    QorSelector.DEFAULTS = {
        aligned: 'left',
        clearable: false
    };

    QorSelector.TEMPLATE = (
        '<div class="qor-selector">' +
        '<a class="qor-selector-toggle">' +
        '<span class="qor-selector-label"></span>' +
        '<i class="material-icons qor-selector-arrow">arrow_drop_down</i>' +
        '<i class="material-icons qor-selector-clear">clear</i>' +
        '</a>' +
        '<ul class="qor-selector-menu"></ul>' +
        '</div>'
    );

    QorSelector.plugin = function(option) {
        return this.each(function() {
            var $this = $(this);
            var data = $this.data(NAMESPACE);
            var options;
            var fn;

            if (!data) {
                if (/destroy/.test(option)) {
                    return;
                }

                options = $.extend({}, QorSelector.DEFAULTS, $this.data(), typeof option === 'object' && option);
                $this.data(NAMESPACE, (data = new QorSelector(this, options)));
            }

            if (typeof option === 'string' && $.isFunction(fn = data[option])) {
                fn.apply(data);
            }
        });
    };

    $(function() {
        var selector = '[data-toggle="qor.selector"]';

        $(document).
        on(EVENT_DISABLE, function(e) {
            QorSelector.plugin.call($(selector, e.target), 'destroy');
        }).
        on(EVENT_ENABLE, function(e) {
            QorSelector.plugin.call($(selector, e.target));
        }).
        triggerHandler(EVENT_ENABLE);
    });

    return QorSelector;

});
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

    let NAMESPACE = 'qor.single_edit',
        EVENT_ENABLE = 'enable.' + NAMESPACE,
        EVENT_DISABLE = 'disable.' + NAMESPACE,
        EVENT_CHANGE = 'change.' + NAMESPACE;

    function QorSingleEdit(element, options) {
        let $el = $(element);
        let value = $el.data('name');
        if (value) {
            this.$el = $el;
            this.$block = $el.find('.qor-field__block:eq(0)');
            this.$toggle = this.$el.find("[name='"+value+".@enabled']")
            this.init();
        } else {
            this.maker = null;
        }
    }

    QorSingleEdit.prototype = {
        constructor: QorSingleEdit,

        init: function() {
            this.bind();
            if (this.$toggle.length) {
                this.toggle()
            }
        },

        bind: function() {
            this.$toggle.on(EVENT_CHANGE, this.toggle.bind(this))
        },

        unbind: function() {
            this.$toggle.off(EVENT_CHANGE);
        },

        destroy: function() {
            this.unbind();
            this.$el.removeData(NAMESPACE);
        },

        toggle: function () {
            if (this.$toggle.is(':checked')) {
                this.$block.show()
            } else {
                this.$block.hide()
            }
        }
    };

    QorSingleEdit.DEFAULTS = {};

    QorSingleEdit.plugin = function(options) {
        return this.each(function() {
            let $this = $(this),
                data = $this.data(NAMESPACE),
                fn;

            if (!data) {
                if (/destroy/.test(options)) {
                    return;
                }
                data = new QorSingleEdit(this, options);
                if (("masker" in data)) {
                    $this.data(NAMESPACE, data);
                } else {
                    return
                }
            }

            if (typeof options === 'string' && $.isFunction((fn = data[options]))) {
                fn.apply(data);
            }
        });
    };

    $(function() {
        let selector = '.single-edit',
            options = {};

        $(document)
            .on(EVENT_DISABLE, function(e) {
                QorSingleEdit.plugin.call($(selector, e.target), 'destroy');
            })
            .on(EVENT_ENABLE, function(e) {
                QorSingleEdit.plugin.call($(selector, e.target), options);
            })
            .triggerHandler(EVENT_ENABLE);
    });

    return QorSingleEdit;
});
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
        let qorSliderAfterShow = $.fn.qorSliderAfterShow;
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
                    if (callback && $.isFunction(callback)) {
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
        this.$element = element ? $(element) : null;
        this.options = $.extend({}, QorSlideout.DEFAULTS, $.isPlainObject(options) && options);
        this.options.language = this.options.language || (this.$element ? this.$element.attr("lang") : null) || $('html').attr('lang');
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

            $document.on(EVENT_KEYUP, this.keyup.bind(this));
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
            if (this.$element) {
                this.$element.find('[data-url]').removeClass(CLASS_IS_SELECTED);
            }
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
            let $slideout = this.$slideout,
                form = e.target,
                $form = $(form),
                $submit = $form.find(':submit'),
                formData = new FormData(form);

            if (e.originalEvent && e.originalEvent.constructor === SubmitEvent) {
                const $submitter = $(e.originalEvent.submitter),
                    name = $submitter.attr('name'),
                    val = $submitter.attr('value');

                if (name && val) {
                    formData.append(name, val);
                }
            }

            $slideout.trigger(EVENT_SLIDEOUT_BEFORESEND);

            e.preventDefault();
            let action = $form.prop('action'),
                continueEditing = /[?|&]continue_editing=true/.test(action),
                headers = {
                    'X-Layout': 'lite'
                };

            QOR.prepareFormData(formData);

            if (continueEditing) {
                action = action.replace(/([?|&]continue_editing)=true/, '$1_url=true')
            } else {
                headers["X-Redirection-Disabled"] = "true"
            }

            $.ajax(action, {
                method: $form.prop('method'),
                data: formData,
                dataType: 'html',
                processData: false,
                contentType: false,
                headers: headers,
                beforeSend: function () {
                    $submit.prop('disabled', true);
                    $.fn.qorSlideoutBeforeHide = null;
                },
                success: (function (html, statusText, jqXHR) {
                    $form.parent().find('.qor-error').remove();
                    $slideout.trigger(EVENT_SLIDEOUT_SUBMIT_COMPLEMENT);
                    let xLocation = jqXHR.getResponseHeader('X-Location');

                    if (jqXHR.getResponseHeader('X-Frame-Reload') === 'render-body') {
                        this.hide(true);
                        this.$slideout.one(EVENT_HIDDEN, (function (){
                            this.render(this.options, action, html)
                        }).bind(this));
                        return
                    }

                    if (xLocation) {
                        if (e.originalEvent && e.originalEvent.explicitOriginalTarget) {
                            let data = $(e.originalEvent.explicitOriginalTarget).data();
                            if (data.submitSuccessTarget === "window") {
                                window.location.href = xLocation;
                                return;
                            }
                        }
                        if (jqXHR.getResponseHeader('X-Steps-Done') === 'true') {
                            window.location.href = xLocation;
                            return;
                        }
                        this.load(xLocation);
                        return
                    }

                    let returnUrl = $form.data('returnUrl'),
                        refreshUrl = $form.data('refreshUrl');

                    if (refreshUrl) {
                        window.location.href = refreshUrl;
                        return;
                    }

                    if (returnUrl === 'refresh') {
                        this.refresh();
                        return;
                    }

                    if (returnUrl && returnUrl !== 'refresh') {
                        this.load(returnUrl);
                    } else {
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
                        this.refresh();
                    }
                }).bind(this),
                error: (function (xhr, textStatus, errorThrown) {
                    $form.parent().find('.qor-error').remove();

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
                }).bind(this),
                complete: function () {
                    $submit.prop('disabled', false);
                }
            });
        },

        showOnLoad: function(url, response) {
            // callback for after slider loaded HTML
            // this callback is deprecated, use slideoutLoaded.qor.slideout event.
            let qorSliderAfterShow = $.fn.qorSliderAfterShow;

            if (qorSliderAfterShow) {
                for (var name in qorSliderAfterShow) {
                    if (qorSliderAfterShow.hasOwnProperty(name) && $.isFunction(qorSliderAfterShow[name])) {
                        qorSliderAfterShow[name]['isLoaded'] = true;
                        qorSliderAfterShow[name].call(this, url, response);
                    }
                }
            }

            this.$slideout.one(EVENT_SLIDEOUT_LOADED, function (){
                this.show();
            }.bind(this));

            // will trigger slideoutLoaded.qor.slideout event after slideout loaded
            this.$slideout.trigger(EVENT_SLIDEOUT_LOADED, [url, response]);
        },

        render: function (options, url, response) {
            const $slideout = this.$slideout;
            let $response, $content, $qorFormContainer, $form, $scripts, $links, bodyClass;

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
                    url: url || window.location.href,
                    response: response
                };

                loadScripts($scripts, data);
            }

            if ($links.length) {
                loadStyles($links);
            }

            $content.find('script[src],link[href]').remove();

            // reset slideout header and body
            $slideout.trigger('disable');
            $slideout.html(this.$slideoutTemplate);
            this.$title = $slideout.find('.qor-slideout__title');

            this.$body = $slideout.find('.qor-slideout__body');

            this.$title.html($response.find(options.title).html());
            replaceHtml($slideout.find('.qor-slideout__body')[0], $content.html());
            this.$body.find(CLASS_HEADER_LOCALE).remove();

            $slideout
                .attr('data-src', url)
                .one(EVENT_SHOWN, function () {
                    $(this).trigger('enable');
                })
                .one(EVENT_HIDDEN, function () {
                    $(this).trigger('disable');
                });

            $slideout.find('.qor-slideout__opennew').attr('href', url);
            $slideout.find('.qor-slideout__print').attr('href', url + (url.indexOf('?') !== -1 ? '&' : '?') + 'print');
            this.showOnLoad(url, response);
        },

        load: function (url, data) {
            let method,
                dataType,
                load,
                $slideout = this.$slideout;

            if (!url) {
                return;
            }

            data = $.isPlainObject(data) ? data : {};

            if (data.image) {
                $slideout.trigger('disable');
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
                    .attr('data-src', url)
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

            load = (function () {
                $.ajax(url, {
                    method: method,
                    dataType: dataType,
                    cache: true,
                    ifModified: true,
                    headers: {
                        'X-Layout': 'lite',
                        'X-Requested-Frame': 'Action'
                    },
                    success: (function (response) {
                        $(CLASS_BODY_LOADING).remove();
                        if (method === 'GET') {
                            this.render(this.options, url, response)
                        } else {
                            if (data.returnUrl) {
                                this.load(data.returnUrl);
                            } else {
                                this.refresh();
                            }
                        }
                    }).bind(this),

                    error: (function () {
                        $(CLASS_BODY_LOADING).remove();
                        let errors = $('.qor-error span');
                        if (errors.length > 0) {
                            let errors = errors
                                .map(function () {
                                    return $(this).text();
                                })
                                .get()
                                .join(', ');
                            QOR.alert(errors)
                        } else {
                            QOR.ajaxError.apply(this, arguments)
                        }
                    }).bind(this)
                });
            }).bind(this);

            if (this.slided) {
                this.hide(true);
                this.$slideout.one(EVENT_HIDDEN, load);
            } else {
                load();
            }
        },

        open: function (options) {
            this.addLoading();
            if (typeof options === "string") {
                options = {url:options}
            }
            return this.load(options.url, options.data || options);
        },

        reload: function (url) {
            this.hide(true);
            this.load(url);
        },

        show: function () {
            let $slideout = this.$slideout,
                showEvent;

            if (this.slided) {
                return;
            }

            showEvent = $.Event(EVENT_SHOW);
            $slideout.trigger(showEvent);

            if (showEvent.isDefaultPrevented()) {
                return;
            }

            $slideout.removeClass(CLASS_MINI);
            this.slideoutType === 'mini' && $slideout.addClass(CLASS_MINI);

            $slideout.addClass(CLASS_IS_SHOWN).get(0).offsetWidth;
            $slideout
                .one(EVENT_TRANSITIONEND, this.shown.bind(this))
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
            let $slideout = this.$slideout,
                hideEvent,
                $datePicker = $('.qor-datepicker').not('.hidden');

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

            $slideout.one(EVENT_TRANSITIONEND, this.hidden.bind(this)).removeClass(`${CLASS_IS_SLIDED} qor-slideout__fullscreen`);
            !isReload && $slideout.trigger(EVENT_SLIDEOUT_CLOSED);
        },

        hidden: function () {
            this.slided = false;

            // Enable to scroll body element
            $('body').removeClass(CLASS_OPEN);

            this.$slideout.removeClass(CLASS_IS_SHOWN).trigger(EVENT_HIDDEN);
            this.$slideout.find('.qor-slideout__header').trigger('disable');
            this.$slideout.find('.qor-slideout__body').trigger('disable');
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
            if (this.$element) {
                this.$element.removeData(NAMESPACE)
            }
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
                    <a href="#" target="_blank" class="mdl-button mdl-button--icon mdl-js-button mdl-js-repple-effect qor-slideout__print"><i class="material-icons">print</i></a>
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

    QOR.slideout = function (options) {
        return new QorSlideout(null, options)
    };

    QorSlideout.plugin = function (options) {
        return this.each(function () {
            let $this = $(this),
                data = $this.data(NAMESPACE),
                fn;

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

  var location = window.location;
  var NAMESPACE = 'qor.sorter';
  var EVENT_ENABLE = 'enable.' + NAMESPACE;
  var EVENT_DISABLE = 'disable.' + NAMESPACE;
  var EVENT_CLICK = 'click.' + NAMESPACE;
  var CLASS_IS_SORTABLE = 'is-sortable';

  function QorSorter(element, options) {
    this.$element = $(element);
    this.options = $.extend({}, QorSorter.DEFAULTS, $.isPlainObject(options) && options);
    this.init();
  }

  QorSorter.prototype = {
    constructor: QorSorter,

    init: function () {
      this.$element.addClass(CLASS_IS_SORTABLE);
      this.bind();
    },

    bind: function () {
      if (window.PRINT_MODE) {
        return;
      }
      this.$element.on(EVENT_CLICK, '> thead > tr > th', this.sort.bind(this));
    },

    unbind: function () {
      this.$element.off(EVENT_CLICK, this.sort);
    },

    sort: function (e) {
      if (e.target !== e.currentTarget) {
        return;
      }

      var $target = $(e.currentTarget);
      var orderBy = $target.data('orderBy');
      var search = location.search;
      var param = 'order_by=' + orderBy;

      // Stop when it is not sortable
      if (!orderBy) {
        return;
      }

      if (/order_by/.test(search)) {
        search = search.replace(/order_by(=[.\w:]+)?/, function () {
          return param;
        });
      } else {
        search += search.indexOf('?') > -1 ? ('&' + param) : param;
      }

      location.search = search;
    },

    destroy: function () {
      this.unbind();
      this.$element.removeClass(CLASS_IS_SORTABLE).removeData(NAMESPACE);
    }
  };

  QorSorter.DEFAULTS = {};

  QorSorter.plugin = function (options) {
    return this.each(function () {
      var $this = $(this);
      var data = $this.data(NAMESPACE);
      var fn;

      if (!data) {
        if (/destroy/.test(options)) {
          return;
        }

        $this.data(NAMESPACE, (data = new QorSorter(this, options)));
      }

      if (typeof options === 'string' && $.isFunction(fn = data[options])) {
        fn.apply(data);
      }
    });
  };

  $(function () {
    var selector = '.qor-js-table';

    $(document)
      .on(EVENT_DISABLE, function (e) {
        QorSorter.plugin.call($(selector, e.target), 'destroy');
      })
      .on(EVENT_ENABLE, function (e) {
        QorSorter.plugin.call($(selector, e.target));
      })
      .triggerHandler(EVENT_ENABLE);
  });

  return QorSorter;

});

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

    const NAMESPACE = 'qor.table',
        EVENT_ENABLE = 'enable.' + NAMESPACE,
        EVENT_DISABLE = 'disable.' + NAMESPACE,
        SELECTOR = '.qor-table';


    function QorTable(el, options) {
        this.$el = $(el);
        this.init();
    }

    QorTable.prototype = {
        init: function (options) {
            this.$el.resizableColumns();
        },

        bind: function () {
        },

        unbind: function () {
        },

        destroy: function () {
            this.unbind();
            this.$el = null;
        }
    }

    QorTable.plugin = function (option) {
        return this.each(function () {
            const $this = $(this);
            let data = $this.data(NAMESPACE),
                options,
                fn;

            if (!data) {
                if (/destroy/.test(option)) {
                    return;
                }

                options = $.extend(true, {}, $this.data(), typeof option === 'object' && option);
                $this.data(NAMESPACE, (data = new QorTable(this, options)));
            } else if (/destroy/.test(option)) {
                $this.removeData(NAMESPACE)
            }

            if (typeof option === 'string' && $.isFunction((fn = data[option]))) {
                fn.apply(data);
            }
        });
    };

    $(function () {
        $(document)
            .on(EVENT_ENABLE, function (e) {
                QorTable.plugin.call($(SELECTOR, e.target));
            })
            .on(EVENT_DISABLE, function (e) {
                QorTable.plugin.call($(SELECTOR, e.target), 'destroy');
            })
            .triggerHandler(EVENT_ENABLE)
    });
});

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

  var _ = window._;
  var $body = $('body');
  var NAMESPACE = 'qor.tabbar';
  var EVENT_ENABLE = 'enable.' + NAMESPACE;
  var EVENT_DISABLE = 'disable.' + NAMESPACE;
  var EVENT_CLICK = 'click.' + NAMESPACE;
  var CLASS_TAB = '.qor-layout__tab-button';
  var CLASS_TAB_CONTENT = '.qor-layout__tab-content';
  var CLASS_TAB_BAR = '.mdl-layout__tab-bar-container';
  var CLASS_TAB_BAR_RIGHT = '.qor-layout__tab-right';
  var CLASS_TAB_BAR_LEFT = '.qor-layout__tab-left';
  var CLASS_ACTIVE = 'is-active';

  function QorTab(element, options) {
    this.$element = $(element);
    this.options = $.extend({}, QorTab.DEFAULTS, $.isPlainObject(options) && options);
    this.init();
  }

  QorTab.prototype = {
    constructor: QorTab,

    init: function () {
      this.initTab();
      this.bind();
    },

    bind: function () {
      this.$element.on(EVENT_CLICK, CLASS_TAB, this.switchTab.bind(this));
      this.$element.on(EVENT_CLICK, CLASS_TAB_BAR_RIGHT, this.scrollTabRight.bind(this));
      this.$element.on(EVENT_CLICK, CLASS_TAB_BAR_LEFT, this.scrollTabLeft.bind(this));
    },

    unbind: function () {
      this.$element.off(EVENT_CLICK, CLASS_TAB, this.switchTab);
      this.$element.off(EVENT_CLICK, CLASS_TAB_BAR_RIGHT, this.scrollTabRight);
      this.$element.off(EVENT_CLICK, CLASS_TAB_BAR_LEFT, this.scrollTabLeft);
    },

    initTab: function () {
      var data = this.$element.data();

      if (!data.scopeActive) {
        $(CLASS_TAB).first().addClass(CLASS_ACTIVE);
        $body.data('tabScopeActive',$(CLASS_TAB).first().data('name'));
      } else {
        $body.data('tabScopeActive',data.scopeActive);
      }

      this.tabWidth = 0;
      this.slideoutWidth = $(CLASS_TAB_CONTENT).outerWidth();

      _.each($(CLASS_TAB), function(ele) {
        this.tabWidth = this.tabWidth + $(ele).outerWidth();
      }.bind(this));

      if (this.tabWidth > this.slideoutWidth) {
        this.$element.find(CLASS_TAB_BAR).append(QorTab.ARROW_RIGHT);
      }

    },

    scrollTabLeft: function (e) {
      e.stopPropagation();

      var $scrollBar = $(CLASS_TAB_BAR),
          scrollLeft = $scrollBar.scrollLeft(),
          jumpDistance = scrollLeft - this.slideoutWidth;

      if (scrollLeft > 0){
        $scrollBar.animate({scrollLeft:jumpDistance}, 400, function () {

          $(CLASS_TAB_BAR_RIGHT).show();
          if ($scrollBar.scrollLeft() == 0) {
            $(CLASS_TAB_BAR_LEFT).hide();
          }

        });
      }
    },

    scrollTabRight: function (e) {
      e.stopPropagation();

      var $scrollBar = $(CLASS_TAB_BAR),
          scrollLeft = $scrollBar.scrollLeft(),
          tabWidth = this.tabWidth,
          slideoutWidth = this.slideoutWidth,
          jumpDistance = scrollLeft + slideoutWidth;

      if (jumpDistance < tabWidth){
        $scrollBar.animate({scrollLeft:jumpDistance}, 400, function () {

          $(CLASS_TAB_BAR_LEFT).show();
          if ($scrollBar.scrollLeft() + slideoutWidth >= tabWidth) {
            $(CLASS_TAB_BAR_RIGHT).hide();
          }

        });

        !$(CLASS_TAB_BAR_LEFT).length && this.$element.find(CLASS_TAB_BAR).prepend(QorTab.ARROW_LEFT);
      }
    },

    switchTab: function (e) {
      var $target = $(e.target),
          $element = this.$element,
          data = $target.data(),
          tabScopeActive = $body.data().tabScopeActive,
          isInSlideout = $('.qor-slideout').is(':visible');

      if (!isInSlideout) {
        return;
      }

      if ($target.hasClass(CLASS_ACTIVE)){
        return false;
      }

      $element.find(CLASS_TAB).removeClass(CLASS_ACTIVE);
      $target.addClass(CLASS_ACTIVE);

      $.ajax(data.tabUrl, {
          method: 'GET',
          dataType: 'html',
          processData: false,
          contentType: false,
          beforeSend: function () {
            $('.qor-layout__tab-spinner').remove();
            var $spinner = '<div class="mdl-spinner mdl-js-spinner is-active qor-layout__tab-spinner"></div>';
            $(CLASS_TAB_CONTENT).hide().before($spinner);
            window.componentHandler.upgradeElement($('.qor-layout__tab-spinner')[0]);
          },
          success: function (html) {
            $('.qor-layout__tab-spinner').remove();
            $body.data('tabScopeActive',$target.data('name'));
            var $content = $(html).find(CLASS_TAB_CONTENT).html();
            $(CLASS_TAB_CONTENT).show().html($content).trigger('enable');

          },
          error: function () {
            $('.qor-layout__tab-spinner').remove();
            $body.data('tabScopeActive',tabScopeActive);
          }
        });
      return false;
    },

    destroy: function () {
      this.unbind();
      $body.removeData('tabScopeActive');
    }
  };

  QorTab.ARROW_RIGHT = '<a href="javascript://" class="qor-layout__tab-right"></a>';
  QorTab.ARROW_LEFT = '<a href="javascript://" class="qor-layout__tab-left"></a>';

  QorTab.DEFAULTS = {};

  QorTab.plugin = function (options) {
    return this.each(function () {
      var $this = $(this);
      var data = $this.data(NAMESPACE);
      var fn;

      if (!data) {
        if (/destroy/.test(options)) {
          return;
        }

        $this.data(NAMESPACE, (data = new QorTab(this, options)));
      }

      if (typeof options === 'string' && $.isFunction(fn = data[options])) {
        fn.apply(data);
      }
    });
  };

  $(function () {
    var selector = '[data-toggle="qor.tab"]';

    $(document)
      .on(EVENT_DISABLE, function (e) {
        QorTab.plugin.call($(selector, e.target), 'destroy');
      })
      .on(EVENT_ENABLE, function (e) {
        QorTab.plugin.call($(selector, e.target));
      })
      .triggerHandler(EVENT_ENABLE);
  });

  return QorTab;

});

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

    let NAMESPACE = 'qor.take_picture',
        SELECTOR = '[data-take-picture]',
        EVENT_CLICK = 'click.' + NAMESPACE,
        EVENT_CHANGE = 'change.' + NAMESPACE,
        EVENT_ENABLE = 'enable.' + NAMESPACE,
        EVENT_DISABLE = 'disable.' + NAMESPACE;

    function QorTakePicture(element, options) {
        this.$el = $(element);
        this.message = QOR.messages.takePicture;
        this.init();
    }

    QorTakePicture.prototype = {
        constructor: QorTakePicture,

        init: function () {
            const $el = this.$el,
                data = $el.data();

            this.$target = this.$el.closest('form').find(data.takePicture);
            this.build();
            this.bind();
        },

        bind: function () {
            this.$el.bind(EVENT_CLICK, this.open.bind(this))
            this.$body.find('[take-picture__take]').bind(EVENT_CLICK, this.take.bind(this));
            this.$invert.bind(EVENT_CHANGE, this.toggleInverter.bind(this));
            this.$devs.bind(EVENT_CHANGE, this.chooseDevice.bind(this));
        },

        build: function () {
            const $el = this.$el;
            let template = QorTakePicture.DEFAULTS.TEMPLATE.VIDEO;
            this.$body = $(template);
            this.$body.find('[take-picture__show-video]').html(this.message.openCamera);
            this.$body.find('[take-picture__invert] span').html(this.message.invertImage);
            this.$body.find('[take-picture__dev] span').html(this.message.chooseDevice);
            this.$video = this.$body.find("video");
            this.$invert = this.$body.find('[take-picture__invert] input');
            this.$devs = this.$body.find('[take-picture__dev] select');
            this.$errors = this.$body.find("[take-picture__errors]");

            //$dialog.prependTo($form);
            $el.removeClass('hidden');
            $el.qorBottomSheets({
                persistent: true,
                do: (function ($el, bs) {
                    this.bs = bs;
                    bs.render(this.message.title, this.$body);
                }).bind(this)
            });
        },

        open: function (e) {
            this.bs.show();
            this.openCamera(e);
        },

        destroy: function () {
            this.$el.removeData(NAMESPACE);
            this.bs.destroy();
            this.bs = null;
        },

        errorMsg: function (msg, error) {
            const errorElement = document.querySelector('#errorMsg');
            if (typeof error !== 'undefined') {
                console.error(error);
                this.$errors.append($(`<p>${msg}</p>`))
            }
        },

        toggleInverter: function (e) {
            if (e.target.checked) {
                console.log('INVERT')
            } else {
                console.log('NO INVERT')
            }
        },

        chooseDevice: function (e) {
            this.deviceId = e.target.value;
            this.openCamera();
        },

        take: function (e) {

        },

        stop: function() {
            if (window.stream) {
                window.stream.getTracks().forEach(track => {
                    track.stop();
                });
                window.stream = null;
            }
        },

        openCamera: function (e) {
            this.stop();
            return;
            this.$devs.html('');
            this.$errors.html('');

            const constraints = {
                audio: false,
                video: {deviceId: this.deviceId ? {exact: this.deviceId} : true},
            };

            navigator.mediaDevices.getUserMedia(constraints)
                .then(function (stream) {
                    const video = this.$video[0];
                    const videoTracks = stream.getVideoTracks();
                    console.log(`Using video device: ${videoTracks[0].label}`);
                    window.stream = stream; // make variable available to browser console
                    video.srcObject = stream;
                    return navigator.mediaDevices.enumerateDevices();
                }.bind(this))
                .then(function (deviceInfos) {
                    if (!deviceInfos) {
                        return;
                    }
                    for (let i = 0; i !== deviceInfos.length; ++i) {
                        const deviceInfo = deviceInfos[i];
                        const option = document.createElement('option');
                        option.value = deviceInfo.deviceId;
                        if (deviceInfo.kind === 'videoinput') {
                            option.text = deviceInfo.label || this.messages.camera.replace('{}', `${i + 1}`);
                            this.$devs.append($(option));
                        }
                    }
                }.bind(this))
                .catch(function (error) {
                    if (error.name === 'ConstraintNotSatisfiedError') {
                        const v = constraints.video;
                        this.errorMsg(`The resolution ${v.width.exact}x${v.height.exact} px is not supported by your device.`);
                    } else if (error.name === 'PermissionDeniedError') {
                        this.errorMsg('Permissions have not been granted to use your camera and ' +
                            'microphone, you need to allow the page access to your devices in ' +
                            'order for the demo to work.');
                    }
                    this.errorMsg(`getUserMedia error: ${error.name}`, error);
                }.bind(this))
        }
    };

    $.extend(true, QOR.messages, {
        takePicture: {
            title: 'Capture Image',
            openCamera: 'Open Camera',
            invertImage: 'Invert Image',
            chooseDevice: 'Choose Device',
            camera: 'Camera {}'
        }
    });

    QorTakePicture.DEFAULTS = {
        TEMPLATE: {
            VIDEO: `<div class="center-text"><video id="gum-local" autoplay playsinline style="background: #222; --width: 100%;width: var(--width);height: calc(var(--width) * 0.75); margin-bottom: 20px"></video>
<div class="qor-field">

    <label take-picture__dev>
      <span class="mdl-checkbox__label">Device</span>
      <select>
      
       </select>
    </label>
    <button take-picture__take type="button" class="close mdl-button mdl-js-button mdl-button--fab mdl-button--mini-fab mdl-button--colored"><i class="material-icons">camera_alt</i></button>
    <label take-picture__invert class="mdl-checkbox mdl-js-checkbox mdl-js-ripple-effect">
      <input type="checkbox" class="mdl-checkbox__input">
      <span class="mdl-checkbox__label">Invert</span>
    </label>
    <div take-picture__errors style="color: darkred"></div>
    </div>`
        }
    };

    QorTakePicture.plugin = function (options) {
        let args = Array.prototype.slice.call(arguments, 1),
            result;

        return this.each(function () {
            const $this = $(this);
            let data = $this.data(NAMESPACE),
                opts = (typeof options === 'string') ? {} : ($.extend({}, options || {}, true)),
                funcName = typeof options === 'string' ? options : "",
                fn;

            if (!data) {
                if (/destroy/.test(options)) {
                    return;
                }
                data = new QorTakePicture(this, opts);
                if (("$el" in data)) {
                    $this.data(NAMESPACE, data);
                } else {
                    return
                }
            }

            if (funcName !== '' && $.isFunction((fn = data[options]))) {
                result = fn.apply(data, args);
            }
        });

        return (typeof result === 'undefined') ? this : result;
    };

    $(function () {
        $(document)
            .on(EVENT_DISABLE, function (e) {
                QorTakePicture.plugin.call($(SELECTOR, e.target), 'destroy');
            })
            .on(EVENT_ENABLE, function (e) {
                QorTakePicture.plugin.call($(SELECTOR, e.target), {});
            })
            .triggerHandler(EVENT_ENABLE);
    });

    return QorTakePicture;
});
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

    var NAMESPACE = 'qor.timepicker';
    var EVENT_ENABLE = 'enable.' + NAMESPACE;
    var EVENT_DISABLE = 'disable.' + NAMESPACE;
    var EVENT_CLICK = 'click.' + NAMESPACE;
    var EVENT_FOCUS = 'focus.' + NAMESPACE;
    var EVENT_KEYDOWN = 'keydown.' + NAMESPACE;
    var EVENT_BLUR = 'blur.' + NAMESPACE;
    var EVENT_CHANGE_TIME = 'selectTime.' + NAMESPACE;

    var CLASS_PARENT = '[data-picker-type]';
    var CLASS_TIME_SELECTED = '.ui-timepicker-selected';

    function QorTimepicker(element, options) {
        this.$element = $(element);
        this.options = $.extend(true, {}, QorTimepicker.DEFAULTS, $.isPlainObject(options) && options);
        this.formatDate = null;
        this.pickerData = this.$element.data();
        this.targetInputClass = this.pickerData.targetInput;
        this.parent = this.$element.closest(CLASS_PARENT);
        this.isDateTimePicker = this.targetInputClass && this.parent.length;
        this.$targetInput = this.parent.find(this.targetInputClass);
        this.init();
    }

    QorTimepicker.prototype = {
        init: function () {
            this.bind();
            this.oldValue = this.$targetInput.val();

            var dateNow = new Date();
            var month = dateNow.getMonth() + 1;
            var date = dateNow.getDate();

            month = (month < 8) ? ('0' + month) : month;
            date = (date < 10) ? ('0' + date) : date;

            this.dateValueNow = dateNow.getFullYear() + '-' + month + '-' + date;
        },

        bind: function () {

            var pickerOptions = {
                timeFormat: 'H:i',
                showOn: null,
                wrapHours: false,
                scrollDefault: 'now'
            };

            if (this.isDateTimePicker) {
                this.$targetInput
                    .qorTimepicker(pickerOptions)
                    .on(EVENT_CHANGE_TIME, $.proxy(this.changeTime, this))
                    .on(EVENT_BLUR, $.proxy(this.blur, this))
                    .on(EVENT_FOCUS, $.proxy(this.focus, this))
                    .on(EVENT_KEYDOWN, $.proxy(this.keydown, this));
            }

            this.$element.on(EVENT_CLICK, $.proxy(this.show, this));
        },

        unbind: function () {
            this.$element.off(EVENT_CLICK, this.show);

            if (this.isDateTimePicker) {
                this.$targetInput
                    .off(EVENT_CHANGE_TIME, this.changeTime)
                    .off(EVENT_BLUR, this.blur)
                    .off(EVENT_FOCUS, this.focus)
                    .off(EVENT_KEYDOWN, this.keydown);
            }
        },

        focus: function () {

        },

        blur: function () {
            var inputValue = this.$targetInput.val();
            var inputArr = inputValue.split(' ');
            var inputArrLen = inputArr.length;

            var tempValue;
            var newDateValue;
            var newTimeValue;
            var isDate;
            var isTime;
            var splitSym;

            var timeReg = /\d{1,2}:\d{1,2}/;
            var dateReg = /^\d{4}-\d{1,2}-\d{1,2}/;

            if (!inputValue) {
                return;
            }

            if (inputArrLen == 1) {
                if (dateReg.test(inputArr[0])) {
                    newDateValue = inputArr[0];
                    newTimeValue = '00:00';
                }

                if (timeReg.test(inputArr[0])) {
                    newDateValue = this.dateValueNow;
                    newTimeValue = inputArr[0];
                }

            } else {
                for (var i = 0; i < inputArrLen; i++) {
                    // check for date && time
                    isDate = dateReg.test(inputArr[i]);
                    isTime = timeReg.test(inputArr[i]);

                    if (isDate) {
                        newDateValue = inputArr[i];
                        splitSym = '-';
                    }

                    if (isTime) {
                        newTimeValue = inputArr[i];
                        splitSym = ':';
                    }

                    tempValue = inputArr[i].split(splitSym);

                    for (var j = 0; j < tempValue.length; j++) {
                        if (tempValue[j].length < 2) {
                            tempValue[j] = '0' + tempValue[j];
                        }
                    }

                    if (isDate) {
                        newDateValue = tempValue.join(splitSym);
                    }

                    if (isTime) {
                        newTimeValue = tempValue.join(splitSym);
                    }
                }

            }

            if (this.checkDate(newDateValue) && this.checkTime(newTimeValue)) {
                this.$targetInput.val(newDateValue + ' ' + newTimeValue);
                this.oldValue = this.$targetInput.val();
            } else {
                this.$targetInput.val(this.oldValue);
            }

        },

        keydown: function (e) {
            var keycode = e.keyCode;
            var keys = [48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 8, 37, 38, 39, 40, 27, 32, 20, 189, 16, 186, 96, 97, 98, 99, 100, 101, 102, 103, 104, 105];
            if (keys.indexOf(keycode) == -1) {
                e.preventDefault();
            }
        },

        checkDate: function (value) {
            var regCheckDate = /^(?:(?!0000)[0-9]{4}-(?:(?:0[1-9]|1[0-2])-(?:0[1-9]|1[0-9]|2[0-8])|(?:0[13-9]|1[0-2])-(?:29|30)|(?:0[13578]|1[02])-31)|(?:[0-9]{1,2}(?:0[48]|[2468][048]|[13579][26])|(?:0[48]|[2468][048]|[13579][26])00)-02-29)$/;
            return regCheckDate.test(value);
        },

        checkTime: function (value) {
            var regCheckTime = /^([01]\d|2[0-3]):?([0-5]\d)$/;
            return regCheckTime.test(value);
        },

        changeTime: function () {
            var $targetInput = this.$targetInput;

            var oldValue = this.oldValue;
            var timeReg = /\d{1,2}:\d{1,2}/;
            var hasTime = timeReg.test(oldValue);
            var selectedTime = $targetInput.data().timepickerList.find(CLASS_TIME_SELECTED).html();
            var newValue;

            if (!oldValue) {
                newValue = this.dateValueNow + ' ' + selectedTime;
            } else if (hasTime) {
                newValue = oldValue.replace(timeReg, selectedTime);
            } else {
                newValue = oldValue + ' ' + selectedTime;
            }

            $targetInput.val(newValue);

        },

        show: function () {
            if (!this.isDateTimePicker) {
                return;
            }

            this.$targetInput.qorTimepicker('show');
            this.oldValue = this.$targetInput.val();
        },

        destroy: function () {
            this.unbind();
            this.$targetInput.qorTimepicker('remove');
            this.$element.removeData(NAMESPACE);
        }
    };

    QorTimepicker.DEFAULTS = {};

    QorTimepicker.plugin = function (option) {
        return this.each(function () {
            var $this = $(this);
            var data = $this.data(NAMESPACE);
            var options;
            var fn;

            if (!data) {
                if (!$.fn.qorDatepicker) {
                    return;
                }

                if (/destroy/.test(option)) {
                    return;
                }

                options = $.extend(true, {}, $this.data(), typeof option === 'object' && option);
                $this.data(NAMESPACE, (data = new QorTimepicker(this, options)));
            }

            if (typeof option === 'string' && $.isFunction(fn = data[option])) {
                fn.apply(data);
            }
        });
    };

    $(function () {
        var selector = '[data-toggle="qor.timepicker"]';

        $(document).
        on(EVENT_DISABLE, function (e) {
            QorTimepicker.plugin.call($(selector, e.target), 'destroy');
        }).
        on(EVENT_ENABLE, function (e) {
            QorTimepicker.plugin.call($(selector, e.target));
        }).
        triggerHandler(EVENT_ENABLE);
    });

    return QorTimepicker;

});
