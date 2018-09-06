// init for slideout after show event
$.fn.qorSliderAfterShow = $.fn.qorSliderAfterShow || {};
window.QOR = {};

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
        remoteDataCache = !(select2Data.remoteDataCache === 'false');

    return {
        dataType: 'json',
        cache: remoteDataCache,
        delay: 250,
        data: function(params) {
            return {
                keyword: params.term, // search term
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
        result = data.text || data.Name || data.Title || data.Code || data[Object.keys(data)[0]];
    }

    // if is HTML
    if (/<(.*)(\/>|<\/.+>)/.test(result)) {
        return $(result);
    }
    return result;
};

$(function() {
    let html = `<div id="dialog" style="display: none;">
                  <div class="mdl-dialog-bg"></div>
                  <div class="mdl-dialog">
                      <div class="mdl-dialog__content">
                        <p><i class="material-icons">warning</i></p>
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
        QOR = window.QOR,
        $dialog = $(html).appendTo('body');

    // ************************************ Refactor window.confirm ************************************
    $(document)
        .on('keyup.qor.confirm', function(e) {
            if (e.which === 27) {
                if ($dialog.is(':visible')) {
                    setTimeout(function() {
                        $dialog.hide();
                    }, 100);
                }
            }
        })
        .on('click.qor.confirm', '.dialog-button', function() {
            let value = $(this).data('type'),
                callback = QOR.qorConfirmCallback;

            $.isFunction(callback) && callback(value);
            $dialog.hide();
            QOR.qorConfirmCallback = undefined;
            return false;
        });

    QOR.qorConfirm = function(data, callback) {
        let okBtn = $dialog.find('.dialog-ok'),
            cancelBtn = $dialog.find('.dialog-cancel');

        if (_.isString(data)) {
            $dialog.find('.dialog-message').text(data);
            okBtn.text('ok');
            cancelBtn.text('cancel');
        } else if (_.isObject(data)) {
            if (data.confirmOk && data.confirmCancel) {
                okBtn.text(data.confirmOk);
                cancelBtn.text(data.confirmCancel);
            } else {
                okBtn.text('ok');
                cancelBtn.text('cancel');
            }

            $dialog.find('.dialog-message').text(data.confirm);
        }

        $dialog.show();
        QOR.qorConfirmCallback = callback;
        return false;
    };

    // *******************************************************************************

    // ****************Handle download file from AJAX POST****************************
    let objectToFormData = function(obj, form) {
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

    QOR.qorAjaxHandleFile = function(url, contentType, fileName, data) {
        let request = new XMLHttpRequest();

        request.responseType = 'arraybuffer';
        request.open('POST', url, true);
        request.onload = function() {
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

    let converVideoLinks = function() {
        let $ele = $('.qor-linkify-object'),
            linkyoutube = /https?:\/\/(?:[0-9A-Z-]+\.)?(?:youtu\.be\/|youtube\.com\S*[^\w\-\s])([\w\-]{11})(?=[^\w\-]|$)(?![?=&+%\w.\-]*(?:['"][^<>]*>|<\/a>))[?=&+%\w.-]*/gi;

        if (!$ele.length) {
            return;
        }

        $ele.each(function() {
            let url = $(this).data('video-link');
            if (url.match(linkyoutube)) {
                $(this).html(`<iframe width="100%" height="100%" src="//www.youtube.com/embed/${url.replace(linkyoutube, '$1')}" frameborder="0" allowfullscreen></iframe>`);
            }
        });
    };

    $.fn.qorSliderAfterShow.converVideoLinks = converVideoLinks;
    converVideoLinks();
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

    function Datepicker(element, options) {
        options = $.isPlainObject(options) ? options : {};

        if (options.language) {
            options = $.extend({}, Datepicker.LANGUAGES[options.language], options);
        }

        this.$element = $(element);
        this.options = $.extend({}, Datepicker.DEFAULTS, options);
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
            this.$yearsCurrent.
            toggleClass(disabledClass, true).
            html((viewYear + start) + suffix + ' - ' + (viewYear + end) + suffix);
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
            this.$yearCurrent.
            toggleClass(disabledClass, isPrevDisabled && isNextDisabled).
            html(viewYear + options.yearSuffix || '');
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
            this.$monthCurrent.
            toggleClass(disabledClass, isPrevDisabled && isNextDisabled).
            html(
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
            var format = this.format;
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
        },

        /**
         * Format a date object to a string with the set date format
         *
         * @param {Date} date
         * @return {String} (formated date)
         */
        formatDate: function (date) {
            var format = this.format;
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
        },

        // Destroy the datepicker and remove the instance from the target element
        destroy: function () {
            this.unbind();
            this.unbuild();
            this.$element.removeData(NAMESPACE);
        }
    };

    Datepicker.LANGUAGES = {};

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

        // The ISO language code (built-in: en-US)
        language: '',

        // The date string format
        format: 'yyyy-mm-dd',

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
    $.fn.qorDatepicker.languages = Datepicker.LANGUAGES;
    $.fn.qorDatepicker.setDefaults = Datepicker.setDefaults;

    // No conflict
    $.fn.qorDatepicker.noConflict = function () {
        $.fn.qorDatepicker = Datepicker.other;
        return this;
    };

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
    let Mustache = window.Mustache,
        NAMESPACE = 'qor.action',
        EVENT_ENABLE = 'enable.' + NAMESPACE,
        EVENT_DISABLE = 'disable.' + NAMESPACE,
        EVENT_CLICK = 'click.' + NAMESPACE,
        EVENT_UNDO = 'undo.' + NAMESPACE,
        ACTION_FORMS = '.qor-action-forms',
        ACTION_HEADER = '.qor-page__header',
        ACTION_BODY = '.qor-page__body',
        ACTION_BUTTON = '.qor-action-button',
        MDL_BODY = '.mdl-layout__content',
        ACTION_SELECTORS = '.qor-actions',
        ACTION_LINK = 'a.qor-action--button',
        MENU_ACTIONS = '.qor-table__actions a[data-url],[data-url][data-method=POST],[data-url][data-method=PUT],[data-url][data-method=DELETE]',
        BUTTON_BULKS = '.qor-action-bulk-buttons',
        QOR_TABLE = '.qor-table-container',
        QOR_TABLE_BULK = '.qor-table--bulking',
        QOR_SEARCH = '.qor-search-container',
        CLASS_IS_UNDO = 'is_undo',
        QOR_SLIDEOUT = '.qor-slideout',
        ACTION_FORM_DATA = 'primary_values[]';

    function QorAction(element, options) {
        this.$element = $(element);
        this.$wrap = $(ACTION_FORMS);
        this.options = $.extend({}, QorAction.DEFAULTS, $.isPlainObject(options) && options);
        this.ajaxForm = {};
        this.init();
    }

    QorAction.prototype = {
        constructor: QorAction,

        init: function() {
            this.bind();
            this.initActions();
        },

        bind: function() {
            this.$element.on(EVENT_CLICK, $.proxy(this.click, this));
            $(document)
                .on(EVENT_CLICK, '.qor-table--bulking tr', this.click)
                .on(EVENT_CLICK, ACTION_LINK, this.actionLink);
        },

        unbind: function() {
            this.$element.off(EVENT_CLICK, this.click);

            $(document)
                .off(EVENT_CLICK, '.qor-table--bulking tr', this.click)
                .off(EVENT_CLICK, ACTION_LINK, this.actionLink);
        },

        initActions: function() {
            this.tables = $(QOR_TABLE).find('table').length;

            if (!this.tables) {
                $(BUTTON_BULKS)
                    .find('button')
                    .attr('disabled', true);
                $(ACTION_LINK).attr('disabled', true);
            }
        },

        collectFormData: function() {
            let checkedInputs = $(QOR_TABLE_BULK).find('.mdl-checkbox__input:checked'),
                formData = [],
                normalFormData = [],
                tempObj;

            if (checkedInputs.length) {
                checkedInputs.each(function() {
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

        actionLink: function() {
            // if not in index page
            if (!$(QOR_TABLE).find('table').length) {
                return false;
            }
        },

        actionSubmit: function($action) {
            let $target = $($action);
            this.$actionButton = $target;
            if ($target.data().method) {
                this.submit();
                return false;
            }
        },

        click: function(e) {
            let $target = $(e.target),
                $pageHeader = $('.qor-page > .qor-page__header'),
                $pageBody = $('.qor-page > .qor-page__body'),
                triggerHeight = $pageHeader.find('.qor-page-subnav__header').length ? 96 : 48;

            this.$actionButton = $target;

            if ($target.data().ajaxForm) {
                this.collectFormData();
                this.ajaxForm.properties = $target.data();
                this.submit();
                return false;
            }

            if ($target.is('.qor-action--bulk')) {
                this.$wrap.removeClass('hidden');
                $(BUTTON_BULKS)
                    .find('button')
                    .toggleClass('hidden');
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

            if ($target.is('.qor-action--exit-bulk')) {
                this.$wrap.addClass('hidden');
                $(BUTTON_BULKS)
                    .find('button')
                    .toggleClass('hidden');
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
                                '<input class="js-primary-value" type="hidden" name="primary_values[]" value="' + primaryValue + '" />'
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

        renderFlashMessage: function(data) {
            let flashMessageTmpl = QorAction.FLASHMESSAGETMPL;
            Mustache.parse(flashMessageTmpl);
            return Mustache.render(flashMessageTmpl, data);
        },

        submit: function() {
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
                needDisableButtons = $element && !isInSlideout;

            if (properties.fromIndex && (!ajaxForm.formData || !ajaxForm.formData.length)) {
                window.alert(ajaxForm.properties.errorNoItem);
                return;
            }

            if (properties.confirm && properties.ajaxForm && !properties.fromIndex) {
                window.QOR.qorConfirm(properties, function(confirm) {
                    if (confirm) {
                        $.post(properties.url, {_method: properties.method}, function() {
                            window.location.reload();
                        });
                    } else {
                        return;
                    }
                });
            } else {
                if (isUndo) {
                    url = properties.undoUrl;
                }

                $.ajax(url, {
                    method: properties.method,
                    data: ajaxForm.formData,
                    dataType: properties.datatype,
                    beforeSend: function() {
                        if (undoUrl) {
                            $actionButton.prop('disabled', true);
                        } else if (needDisableButtons) {
                            _this.switchButtons($element, 1);
                        }
                    },
                    success: function(data, status, response) {
                        let contentType = response.getResponseHeader('content-type'),
                            // handle file download from form submit
                            disposition = response.getResponseHeader('Content-Disposition');

                        if (disposition && disposition.indexOf('attachment') !== -1) {
                            var fileNameRegex = /filename[^;=\n]*=((['"]).*?\2|[^;\n]*)/,
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
                    },
                    error: function(xhr, textStatus, errorThrown) {
                        if (undoUrl) {
                            $actionButton.prop('disabled', false);
                        } else if (needDisableButtons) {
                            _this.switchButtons($element);
                        }
                        window.alert([textStatus, errorThrown].join(': '));
                    }
                });
            }
        },

        switchButtons: function($element, disbale) {
            let needDisbale = disbale ? true : false;
            $element.find(ACTION_BUTTON).prop('disabled', needDisbale);
        },

        destroy: function() {
            this.unbind();
            this.$element.removeData(NAMESPACE);
        },

        // Helper
        removeTableCheckbox: function() {
            $('.qor-page__body .mdl-data-table__select').each(function(i, e) {
                $(e)
                    .parents('td')
                    .remove();
            });
            $('.qor-page__body .mdl-data-table__select').each(function(i, e) {
                $(e)
                    .parents('th')
                    .remove();
            });
            $('.qor-table-container tr.is-selected').removeClass('is-selected');
            $('.qor-page__body table.mdl-data-table--selectable').removeClass('mdl-data-table--selectable');
            $('.qor-page__body tr.is-selected').removeClass('is-selected');
        },

        appendTableCheckbox: function() {
            // Only value change and the table isn't selectable will add checkboxes
            $('.qor-page__body .mdl-data-table__select').each(function(i, e) {
                $(e)
                    .parents('td')
                    .remove();
            });
            $('.qor-page__body .mdl-data-table__select').each(function(i, e) {
                $(e)
                    .parents('th')
                    .remove();
            });
            $('.qor-table-container tr.is-selected').removeClass('is-selected');
            $('.qor-page__body table').addClass('mdl-data-table--selectable');

            // init google material
            new window.MaterialDataTable($('.qor-page__body table').get(0));
            $('thead.is-hidden tr th:not(".mdl-data-table__cell--non-numeric")')
                .clone()
                .prependTo($('thead:not(".is-hidden") tr'));

            let $fixedHeadCheckBox = $('thead:not(".is-fixed") .mdl-checkbox__input'),
                isMediaLibrary = $('.qor-table--medialibrary').length,
                hasPopoverForm = $('body').hasClass('qor-bottomsheets-open') || $('body').hasClass('qor-slideout-open');

            isMediaLibrary && ($fixedHeadCheckBox = $('thead .mdl-checkbox__input'));

            $fixedHeadCheckBox.on('click', function() {
                if (!isMediaLibrary) {
                    $('thead.is-fixed tr th')
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
                        let allPrimaryValues = $('.qor-table--bulking tbody tr');
                        allPrimaryValues.each(function() {
                            let primaryValue = $(this).data('primary-key');
                            if (primaryValue) {
                                slideroutActionForm.prepend(
                                    '<input class="js-primary-value" type="hidden" name="primary_values[]" value="' + primaryValue + '" />'
                                );
                            }
                        });
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

    $.fn.qorSliderAfterShow.qorActionInit = function(url, html) {
        let hasAction = $(html).find('[data-toggle="qor-action-slideout"]').length,
            $actionForm = $('[data-toggle="qor-action-slideout"]').find('form'),
            $checkedItem = $('.qor-page__body .mdl-checkbox__input:checked');

        if (hasAction && $checkedItem.length) {
            // insert checked value into sliderout form
            $checkedItem.each(function(i, e) {
                let id = $(e)
                    .parents('tbody tr')
                    .data('primary-key');
                if (id) {
                    $actionForm.prepend('<input class="js-primary-value" type="hidden" name="primary_values[]" value="' + id + '" />');
                }
            });
        }
    };

    QorAction.plugin = function(options) {
        return this.each(function() {
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

    $(function() {
        let options = {},
            selector = '[data-toggle="qor.action.bulk"]';

        $(document)
            .on(EVENT_DISABLE, function(e) {
                QorAction.plugin.call($(selector, e.target), 'destroy');
            })
            .on(EVENT_ENABLE, function(e) {
                QorAction.plugin.call($(selector, e.target), options);
            })
            .on(EVENT_CLICK, MENU_ACTIONS, function() {
                new QorAction().actionSubmit(this);
                return false;
            })
            .triggerHandler(EVENT_ENABLE);
    });

    return QorAction;
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
        var regex = new RegExp('[\\?&]' + name + '=([^&#]*)');
        var results = regex.exec(search);
        return results === null ? '' : decodeURIComponent(results[1].replace(/\+/g, ' '));
    }

    function updateQueryStringParameter(key, value, uri) {
        var escapedkey = String(key).replace(/[\\^$*+?.()|[\]{}]/g, '\\$&'),
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
            this.filterURL = '';
            this.searchParams = '';
        },

        bind: function() {
            this.$bottomsheets
                .on(EVENT_SUBMIT, 'form', this.submit.bind(this))
                .on(EVENT_CLICK, '[data-dismiss="bottomsheets"]', this.hide.bind(this))
                .on(EVENT_CLICK, '.qor-pagination a', this.pagination.bind(this))
                .on(EVENT_CLICK, CLASS_BOTTOMSHEETS_BUTTON, this.search.bind(this))
                .on(EVENT_KEYUP, this.keyup.bind(this))
                .on('selectorChanged.qor.selector', this.selectorChanged.bind(this))
                .on('filterChanged.qor.filter', this.filterChanged.bind(this));
        },

        unbind: function() {
            this.$bottomsheets
                .off(EVENT_SUBMIT, 'form', this.submit.bind(this))
                .off(EVENT_CLICK, '[data-dismiss="bottomsheets"]', this.hide.bind(this))
                .off(EVENT_CLICK, '.qor-pagination a', this.pagination.bind(this))
                .off(EVENT_CLICK, CLASS_BOTTOMSHEETS_BUTTON, this.search.bind(this))
                .off('selectorChanged.qor.selector', this.selectorChanged.bind(this))
                .off('filterChanged.qor.filter', this.filterChanged.bind(this));
        },

        bindActionData: function(actiondData) {
            var $form = this.$body.find('[data-toggle="qor-action-slideout"]').find('form');
            for (var i = actiondData.length - 1; i >= 0; i--) {
                $form.prepend('<input type="hidden" name="primary_values[]" value="' + actiondData[i] + '" />');
            }
        },

        filterChanged: function(e, search, key) {
            // if this event triggered:
            // search: ?locale_mode=locale, ?filters[Color].Value=2
            // key: search param name: locale_mode

            var loadUrl;

            loadUrl = this.constructloadURL(search, key);
            loadUrl && this.reload(loadUrl);
            return false;
        },

        selectorChanged: function(e, url, key) {
            // if this event triggered:
            // url: /admin/!remote_data_searcher/products/Collections?locale=en-US
            // key: search param key: locale

            var loadUrl;

            loadUrl = this.constructloadURL(url, key);
            loadUrl && this.reload(loadUrl);
            return false;
        },

        keyup: function(e) {
            var searchInput = this.$bottomsheets.find(CLASS_BOTTOMSHEETS_INPUT);

            if (e.which === 13 && searchInput.length && searchInput.is(':focus')) {
                this.search();
            }
        },

        search: function() {
            var $bottomsheets = this.$bottomsheets,
                param = '?keyword=',
                baseUrl = $bottomsheets.data().url,
                searchValue = $.trim($bottomsheets.find(CLASS_BOTTOMSHEETS_INPUT).val()),
                url = baseUrl + param + searchValue;

            this.reload(url);
        },

        pagination: function(e) {
            var $ele = $(e.target),
                url = $ele.prop('href');
            if (url) {
                this.reload(url);
            }
            return false;
        },

        reload: function(url) {
            var $content = this.$bottomsheets.find(CLASS_BODY_CONTENT);

            this.addLoading($content);
            this.fetchPage(url);
        },

        fetchPage: function(url) {
            var $bottomsheets = this.$bottomsheets,
                _this = this;

            $.get(url, function(response) {
                var $response = $(response).find(CLASS_MAIN_CONTENT),
                    $responseHeader = $response.find(CLASS_BODY_HEAD),
                    $responseBody = $response.find(CLASS_BODY_CONTENT);

                if ($responseBody.length) {
                    $bottomsheets.find(CLASS_BODY_CONTENT).html($responseBody.html());

                    if ($responseHeader.length) {
                        _this.$body
                            .find(CLASS_BODY_HEAD)
                            .html($responseHeader.html())
                            .trigger('enable');
                        _this.addHeaderClass();
                    }
                    // will trigger this event(relaod.qor.bottomsheets) when bottomsheets reload complete: like pagination, filter, action etc.
                    $bottomsheets.trigger(EVENT_RELOAD);
                } else {
                    _this.reload(url);
                }
            }).fail(function() {
                window.alert('server error, please try again later!');
            });
        },

        constructloadURL: function(url, key) {
            var fakeURL,
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
            this.$body.find(CLASS_BODY_HEAD).hide();
            if (this.$bottomsheets.find(CLASS_BODY_HEAD).children(CLASS_BOTTOMSHEETS_FILTER).length) {
                this.$body
                    .addClass('has-header')
                    .find(CLASS_BODY_HEAD)
                    .show();
            }
        },

        addLoading: function($element) {
            $element.html('');
            var $loading = $(QorBottomSheets.TEMPLATE_LOADING).appendTo($element);
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
            var $script = $response.filter('script'),
                theme = /theme=media_library/g,
                src,
                _this = this;

            $script.each(function() {
                src = $(this).prop('src');
                if (theme.test(src)) {
                    var script = document.createElement('script');
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
            if (resourseData.ingoreSubmit) {
                return;
            }

            // will submit form as normal,
            // if you need download file after submit form or other things, please add
            // data-use-normal-submit="true" to form tag
            // <form action="/admin/products/!action/localize" method="POST" enctype="multipart/form-data" data-normal-submit="true"></form>
            var normalSubmit = $form.data().normalSubmit;

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
                    var disposition = jqXHR.getResponseHeader('Content-Disposition');
                    if (disposition && disposition.indexOf('attachment') !== -1) {
                        var fileNameRegex = /filename[^;=\n]*=((['"]).*?\2|[^;\n]*)/,
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

                    var returnUrl = $form.data('returnUrl');
                    var refreshUrl = $form.data('refreshUrl');

                    if (refreshUrl) {
                        window.location.href = refreshUrl;
                        return;
                    }

                    if (returnUrl == 'refresh') {
                        _this.refresh();
                        return;
                    }

                    if (returnUrl && returnUrl != 'refresh') {
                        _this.load(returnUrl);
                    } else {
                        _this.refresh();
                    }

                    $(document).trigger(EVENT_BOTTOMSHEET_SUBMIT);
                },
                error: function(xhr, textStatus, errorThrown) {
                    var $error;

                    if (xhr.status === 422) {
                        $body.find('.qor-error').remove();
                        $error = $(xhr.responseText).find('.qor-error');
                        $form.before($error);
                        $('.qor-bottomsheets .qor-page__body').scrollTop(0);
                    } else {
                        window.alert([textStatus, errorThrown].join(': '));
                    }
                },
                complete: function() {
                    $submit.prop('disabled', false);
                }
            });
        },

        load: function(url, data, callback) {
            var options = this.options,
                method,
                dataType,
                load,
                actionData = data.actionData,
                resourseData = this.resourseData,
                selectModal = resourseData.selectModal,
                ingoreSubmit = resourseData.ingoreSubmit,
                $bottomsheets = this.$bottomsheets,
                $header = this.$header,
                $body = this.$body;

            if (!url) {
                return;
            }

            this.show();
            this.addLoading($body);

            this.filterURL = url;
            $body.removeClass('has-header has-hint');

            data = $.isPlainObject(data) ? data : {};

            method = data.method ? data.method : 'GET';
            dataType = data.datatype ? data.datatype : 'html';

            load = $.proxy(function() {
                $.ajax(url, {
                    method: method,
                    dataType: dataType,
                    success: $.proxy(function(response) {
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

                            if (ingoreSubmit) {
                                $content.find(CLASS_BODY_HEAD).remove();
                            }

                            $content.find('.qor-button--cancel').attr('data-dismiss', 'bottomsheets');

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
                                    .data('ingoreSubmit', true)
                                    .data('selectId', resourseData.selectId)
                                    .data('loadInline', true);
                                if (
                                    selectModal != 'one' &&
                                    !data.selectNohint &&
                                    (typeof resourseData.maxItem === 'undefined' || resourseData.maxItem != '1')
                                ) {
                                    $body.addClass('has-hint');
                                }
                                if (selectModal == 'mediabox' && !this.scriptAdded) {
                                    this.loadMedialibraryJS($response);
                                }
                            }

                            $header.find('.qor-button--new').remove();
                            this.$title.after($body.find('.qor-button--new'));

                            if (hasSearch) {
                                $bottomsheets.addClass('has-search');
                                $header.find('.qor-bottomsheets__search').remove();
                                $header.prepend(QorBottomSheets.TEMPLATE_SEARCH);
                            }

                            if (actionData && actionData.length) {
                                this.bindActionData(actionData);
                            }

                            if (resourseData.bottomsheetClassname) {
                                $bottomsheets.addClass(resourseData.bottomsheetClassname);
                            }

                            $bottomsheets.trigger('enable');

                            $bottomsheets.one(EVENT_HIDDEN, function() {
                                $(this).trigger('disable');
                            });

                            this.addHeaderClass();
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
                    }, this),

                    error: $.proxy(function() {
                        this.$bottomsheets.remove();
                        if (!$('.qor-bottomsheets').is(':visible')) {
                            $('body').removeClass(CLASS_OPEN);
                        }
                        var errors;
                        if ($('.qor-error span').length > 0) {
                            errors = $('.qor-error span')
                                .map(function() {
                                    return $(this).text();
                                })
                                .get()
                                .join(', ');
                        } else {
                            errors = 'Server error, please try again later!';
                        }
                        window.alert(errors);
                    }, this)
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
            this.$bottomsheets.addClass(CLASS_IS_SHOWN).get(0).offsetHeight;
            this.$bottomsheets.addClass(CLASS_IS_SLIDED);
            $('body').addClass(CLASS_OPEN);
        },

        hide: function(e) {
            let $bottomsheets = $(e.target).closest('.qor-bottomsheets'),
                $datePicker = $('.qor-datepicker').not('.hidden');

            if ($datePicker.length) {
                $datePicker.addClass('hidden');
            }

            $bottomsheets.qorSelectCore('destroy');

            $bottomsheets.trigger(EVENT_BOTTOMSHEET_CLOSED).remove();
            if (!$('.qor-bottomsheets').is(':visible')) {
                $('body').removeClass(CLASS_OPEN);
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
            <h3 class="qor-bottomsheets__title"></h3>
            <button type="button" class="mdl-button mdl-button--icon mdl-js-button mdl-js-repple-effect qor-bottomsheets__close" data-dismiss="bottomsheets">
            <span class="material-icons">close</span>
            </button>
            </div>
            <div class="qor-bottomsheets__body"></div>
        </div>`;

    QorBottomSheets.plugin = function(options) {
        return this.each(function() {
            var $this = $(this);
            var data = $this.data(NAMESPACE);
            var fn;

            if (!data) {
                if (/destroy/.test(options)) {
                    return;
                }

                $this.data(NAMESPACE, (data = new QorBottomSheets(this, options)));
            }

            if (typeof options === 'string' && $.isFunction((fn = data[options]))) {
                fn.apply(data);
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
                };

            if (select2Data.remoteData) {
                option.ajax = $.fn.select2.ajaxCommonOptions(select2Data);

                option.templateResult = function(data) {
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
        CLASS_UNDO = '.qor-fieldset__undo';

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

    function clearObject(obj) {
        for (let prop in obj) {
            if (obj.hasOwnProperty(prop)) obj[prop] = '';
        }
        return obj;
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
                data,
                outputValue,
                fetchUrl,
                _this = this,
                imageData;

            if (!$parent.length) {
                $parent = $this.parent();
            }

            this.$parent = $parent;
            this.$output = $parent.find(options.output);
            this.$list = $parent.find(options.list);

            fetchUrl = this.$output.data('fetchSizedata');

            if (fetchUrl) {
                $.getJSON(fetchUrl, function(data) {
                    imageData = JSON.parse(data.MediaOption);
                    _this.$output.val(JSON.stringify(data));
                    _this.data = imageData || {};
                    if (isSVG(imageData.URL || imageData.Url)) {
                        _this.resetImage();
                    }
                    _this.build();
                    _this.bind();
                });
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
                this.$list.hide();

                $alert = $(QorCropper.ALERT);
                $alert.find(CLASS_UNDO).one(
                    EVENT_CLICK,
                    function() {
                        $alert.remove();
                        this.$list.show();
                        delete data.Delete;
                        this.$output.val(JSON.stringify(data));
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
                $alert = this.$parent.find('.qor-fieldset__alert');

            if ($alert.length) {
                $alert.remove();
                this.data = clearObject(this.data);
            }

            if (files && files.length) {
                file = files[0];

                if (/^image\//.test(file.type) && URL) {
                    this.fileType = file.type;
                    this.load(URL.createObjectURL(file));
                    this.$parent.find('.qor-medialibrary__image-desc').show();
                } else {
                    this.$list.empty().html(QorCropper.FILE_LIST.replace('{{filename}}', file.name));
                }
            }
        },

        load: function(url, callback) {
            let options = this.options,
                _this = this,
                $list = this.$list,
                $ul = $list.find('ul'),
                data = this.data || {},
                fileType = this.fileType,
                $image,
                imageLength;

            if (!$ul.length || !$ul.find('li').length) {
                $ul = $(QorCropper.LIST);
                $list.html($ul);
                this.wrap();
            }

            $ul.show(); // show ul when it is hidden

            $image = $list.find('img');
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

                            data[options.key][sizeName] = emulateCropData;
                        }
                    } else {
                        _this.center($this);
                    }

                    // Crop, CropOptions and Delete should be BOOL type, if empty should delete,
                    if (data.Crop === '') {
                        delete data.Crop;
                    }

                    if (data.CropOptions === '') {
                        delete data.CropOptions;
                    }

                    delete data.Delete;
                    _this.$output.val(JSON.stringify(data));

                    // callback after load complete
                    if (sizeName && data[options.key] && Object.keys(data[options.key]).length >= imageLength) {
                        if (callback && $.isFunction(callback)) {
                            callback();
                        }
                    }
                })
                .attr('src', url)
                .data('originalUrl', url);

            $list.show();
        },

        start: function() {
            let options = this.options,
                $modal = this.$modal,
                $target = this.$target,
                sizeData = $target.data(),
                sizeName = sizeData.sizeName || 'original',
                sizeResolution = sizeData.sizeResolution,
                $clone = $('<img>').attr('src', sizeData.originalUrl),
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
                checkImageOrigin: false,
                autoCropArea: 1,

                built: function() {
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
                            url = croppedCanvas.toDataURL();
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
    QorCropper.LIST = '<ul><li><img></li></ul>';
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
        let selector = '.qor-file__input',
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
                $.each(data, function(key, val) {
                    str = str.replace('$[' + String(key).toLowerCase() + ']', val);
                });
            }
        }

        return str;
    }

    function QorDatepicker(element, options) {
        this.$element = $(element);
        this.options = $.extend(true, {}, QorDatepicker.DEFAULTS, $.isPlainObject(options) && options);
        this.date = null;
        this.formatDate = null;
        this.built = false;
        this.pickerData = this.$element.data();
        this.init();
    }

    QorDatepicker.prototype = {
        init: function() {
            this.bind();
        },

        bind: function() {
            this.$element.on(EVENT_CLICK, $.proxy(this.show, this));
        },

        unbind: function() {
            this.$element.off(EVENT_CLICK, this.show);
        },

        build: function() {
            let $modal,
                $ele = this.$element,
                data = this.pickerData,
                date = $ele.val() ? new Date($ele.val()) : new Date(),
                datepickerOptions = {
                    date: date,
                    inline: true
                },
                parent = $ele.closest(CLASS_PARENT),
                $targetInput = parent.find(data.targetInput);

            if (this.built) {
                return;
            }

            this.$modal = $modal = $(replaceText(QorDatepicker.TEMPLATE, this.options.text)).appendTo('body');

            if ($targetInput.length) {
                datepickerOptions.date = $targetInput.val() ? new Date($targetInput.val()) : new Date();
            }

            if (data.targetInput && $targetInput.data('start-date')) {
                datepickerOptions.startDate = new Date();
            }

            $modal.find(CLASS_EMBEDDED).on(EVENT_CHANGE, $.proxy(this.change, this)).qorDatepicker(datepickerOptions).triggerHandler(EVENT_CHANGE);

            $modal.find(CLASS_SAVE).on(EVENT_CLICK, $.proxy(this.pick, this));

            this.built = true;
        },

        unbuild: function() {
            if (!this.built) {
                return;
            }

            this.$modal.find(CLASS_EMBEDDED).off(EVENT_CHANGE, this.change).qorDatepicker('destroy').end().find(CLASS_SAVE).off(EVENT_CLICK, this.pick).end().remove();
        },

        change: function(e) {
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

        show: function() {
            if (!this.built) {
                this.build();
            }

            this.$modal.qorModal('show');
        },

        pick: function() {
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

        destroy: function() {
            this.unbind();
            this.unbuild();
            this.$element.removeData(NAMESPACE);
        }
    };

    QorDatepicker.DEFAULTS = {
        text: {
            title: 'Pick a date',
            ok: 'OK',
            cancel: 'Cancel'
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

    QorDatepicker.plugin = function(option) {
        return this.each(function() {
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

    $(function() {
        var selector = '[data-toggle="qor.datepicker"]';

        $(document)
            .on(EVENT_DISABLE, function(e) {
                QorDatepicker.plugin.call($(selector, e.target), 'destroy');
            })
            .on(EVENT_ENABLE, function(e) {
                QorDatepicker.plugin.call($(selector, e.target));
            })
            .triggerHandler(EVENT_ENABLE);
    });

    return QorDatepicker;
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

    var location = window.location;
    var NAMESPACE = 'qor.filter';
    var EVENT_FILTER_CHANGE = 'filterChanged.' + NAMESPACE;
    var EVENT_ENABLE = 'enable.' + NAMESPACE;
    var EVENT_DISABLE = 'disable.' + NAMESPACE;
    var EVENT_CLICK = 'click.' + NAMESPACE;
    var EVENT_CHANGE = 'change.' + NAMESPACE;
    var CLASS_IS_ACTIVE = 'is-active';
    var CLASS_BOTTOMSHEETS = '.qor-bottomsheets';

    function QorFilter(element, options) {
        this.$element = $(element);
        this.options = $.extend({}, QorFilter.DEFAULTS, $.isPlainObject(options) && options);
        this.init();
    }

    function encodeSearch(data, detached) {
        var search = decodeURI(location.search);
        var params;

        if ($.isArray(data)) {
            params = decodeSearch(search);

            $.each(data, function(i, param) {
                i = $.inArray(param, params);

                if (i === -1) {
                    params.push(param);
                } else if (detached) {
                    params.splice(i, 1);
                }
            });

            search = '?' + params.join('&');
        }

        return search;
    }

    function decodeSearch(search) {
        var data = [];

        if (search && search.indexOf('?') > -1) {
            search = search.split('?')[1];

            if (search && search.indexOf('#') > -1) {
                search = search.split('#')[0];
            }

            if (search) {
                // search = search.toLowerCase();
                data = $.map(search.split('&'), function(n) {
                    var param = [];
                    var value;

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

    QorFilter.prototype = {
        constructor: QorFilter,

        init: function() {
            // this.parse();
            this.bind();
        },

        bind: function() {
            var options = this.options;

            this.$element.on(EVENT_CLICK, options.label, $.proxy(this.toggle, this)).on(EVENT_CHANGE, options.group, $.proxy(this.toggle, this));
        },

        unbind: function() {
            this.$element.off(EVENT_CLICK, this.toggle).off(EVENT_CHANGE, this.toggle);
        },

        toggle: function(e) {
            var $target = $(e.currentTarget);
            var data = [];
            var params;
            var param;
            var search;
            var name;
            var value;
            var index;
            var matched;
            var paramName;

            if ($target.is('select')) {
                params = decodeSearch(decodeURI(location.search));
                paramName = name = $target.attr('name');
                value = $target.val();
                param = [name];

                if (value) {
                    param.push(value);
                }

                param = param.join('=');

                if (value) {
                    data.push(param);
                }

                $target.children().each(function() {
                    var $this = $(this);
                    var param = [name];
                    var value = $.trim($this.prop('value'));

                    if (value) {
                        param.push(value);
                    }

                    param = param.join('=');
                    index = $.inArray(param, params);

                    if (index > -1) {
                        matched = param;
                        return false;
                    }
                });

                if (matched) {
                    data.push(matched);
                    search = encodeSearch(data, true);
                } else {
                    search = encodeSearch(data);
                }
            } else if ($target.is('a')) {
                e.preventDefault();
                paramName = $target.data().paramName;
                data = decodeSearch($target.attr('href'));
                if ($target.hasClass(CLASS_IS_ACTIVE)) {
                    search = encodeSearch(data, true); // set `true` to detach
                } else {
                    search = encodeSearch(data);
                }
            }

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
            group: 'select'
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
    var NAMESPACE = 'qor.fixer';
    var EVENT_ENABLE = 'enable.' + NAMESPACE;
    var EVENT_DISABLE = 'disable.' + NAMESPACE;
    var EVENT_CLICK = 'click.' + NAMESPACE;
    var EVENT_RESIZE = 'resize.' + NAMESPACE;
    var EVENT_SCROLL = 'scroll.' + NAMESPACE;
    var CLASS_IS_HIDDEN = 'is-hidden';
    var CLASS_IS_FIXED = 'is-fixed';
    var CLASS_HEADER = '.qor-page__header';

    function QorFixer(element, options) {
        this.$element = $(element);
        this.options = $.extend({}, QorFixer.DEFAULTS, $.isPlainObject(options) && options);
        this.$clone = null;
        this.init();
    }

    QorFixer.prototype = {
        constructor: QorFixer,

        init: function() {
            var options = this.options;
            var $this = this.$element;
            if (this.buildCheck()) {
                return;
            }
            this.$thead = $this.find('thead:first');
            this.$tbody = $this.find('tbody:first');
            this.$header = $(options.header);
            this.$subHeader = $(options.subHeader);
            this.$content = $(options.content);
            this.marginBottomPX = parseInt(this.$subHeader.css('marginBottom'));
            this.paddingHeight = options.paddingHeight;
            this.fixedHeaderWidth = [];
            this.isEqualed = false;

            this.resize();
            this.bind();
        },

        bind: function() {
            this.$element.on(EVENT_CLICK, $.proxy(this.check, this));

            this.$content.on(EVENT_SCROLL, $.proxy(this.toggle, this));
            $window.on(EVENT_RESIZE, $.proxy(this.resize, this));
        },

        unbind: function() {
            this.$element.off(EVENT_CLICK, this.check);

            this.$content.
            off(EVENT_SCROLL, this.toggle).
            off(EVENT_RESIZE, this.resize);
        },

        build: function() {
            if (!this.$content.length) {
                return;
            }

            var $this = this.$element;
            var $thead = this.$thead;
            var $clone = this.$clone;
            var self = this;
            var $items = $thead.find('> tr').children();
            var pageBodyTop = this.$content.offset().top + $(CLASS_HEADER).height();

            if (!$clone) {
                this.$clone = $clone = $thead.clone().prependTo($this).css({ top: pageBodyTop });
            }

            $clone.
            addClass([CLASS_IS_FIXED, CLASS_IS_HIDDEN].join(' ')).
            find('> tr').
            children().
            each(function(i) {
                $(this).outerWidth($items.eq(i).outerWidth());
                self.fixedHeaderWidth.push($(this).outerWidth());
            });
        },

        unbuild: function() {
            this.$clone.remove();
        },

        buildCheck: function() {
            var $this = this.$element;
            // disable fixer if have multiple tables or in search page or in media library list page
            if ($('.qor-page__body .qor-js-table').length > 1 || $('.qor-global-search--container').length > 0 || $this.hasClass('qor-table--medialibrary') || $this.is(':hidden') || $this.find('tbody > tr:visible').length <= 1) {
                return true;
            }
            return false;
        },

        check: function(e) {
            var $target = $(e.target);
            var checked;

            if ($target.is('.qor-js-check-all')) {
                checked = $target.prop('checked');

                $target.
                closest('thead').
                siblings('thead').
                find('.qor-js-check-all').prop('checked', checked).
                closest('.mdl-checkbox').toggleClass('is-checked', checked);
            }
        },

        toggle: function() {
            if (!this.$content.length) {
                return;
            }
            var self = this;
            var $clone = this.$clone;
            var $thead = this.$thead;
            var scrollTop = this.$content.scrollTop();
            var scrollLeft = this.$content.scrollLeft();
            var offsetTop = this.$subHeader.outerHeight() + this.paddingHeight + this.marginBottomPX;
            var headerHeight = $('.qor-page__header').outerHeight();

            if (!this.isEqualed) {
                this.headerWidth = [];
                var $items = $thead.find('> tr').children();
                $items.each(function() {
                    self.headerWidth.push($(this).outerWidth());
                });
                var notEqualWidth = _.difference(self.fixedHeaderWidth, self.headerWidth);
                if (notEqualWidth.length) {
                    $('thead.is-fixed').find('>tr').children().each(function(i) {
                        $(this).outerWidth(self.headerWidth[i]);
                    });
                    this.isEqualed = true;
                }
            }
            if (scrollTop > offsetTop - headerHeight) {
                $clone.css({ 'margin-left': -scrollLeft }).removeClass(CLASS_IS_HIDDEN);
            } else {
                $clone.css({ 'margin-left': '0' }).addClass(CLASS_IS_HIDDEN);
            }
        },

        resize: function() {
            this.build();
            this.toggle();
        },

        destroy: function() {
            if (this.buildCheck()) {
                return;
            }
            this.unbind();
            this.unbuild();
            this.$element.removeData(NAMESPACE);
        }
    };

    QorFixer.DEFAULTS = {
        header: false,
        content: false
    };

    QorFixer.plugin = function(options) {
        return this.each(function() {
            var $this = $(this);
            var data = $this.data(NAMESPACE);
            var fn;

            if (!data) {
                $this.data(NAMESPACE, (data = new QorFixer(this, options)));
            }

            if (typeof options === 'string' && $.isFunction(fn = data[options])) {
                fn.call(data);
            }
        });
    };

    $(function() {
        var selector = '.qor-js-table';
        var options = {
            header: '.mdl-layout__header',
            subHeader: '.qor-page__header',
            content: '.mdl-layout__content',
            paddingHeight: 2 // Fix sub header height bug
        };

        $(document).
        on(EVENT_DISABLE, function(e) {
            QorFixer.plugin.call($(selector, e.target), 'destroy');
        }).
        on(EVENT_ENABLE, function(e) {
            QorFixer.plugin.call($(selector, e.target), options);
        }).
        triggerHandler(EVENT_ENABLE);
    });

    return QorFixer;

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

    var componentHandler = window.componentHandler;
    var NAMESPACE = 'qor.material';
    var EVENT_ENABLE = 'enable.' + NAMESPACE;
    var EVENT_DISABLE = 'disable.' + NAMESPACE;
    var EVENT_UPDATE = 'update.' + NAMESPACE;
    var SELECTOR_COMPONENT = '[class*="mdl-js"],[class*="mdl-tooltip"]';

    function enable(target) {
        /*jshint undef:false */
        if (componentHandler) {
            // Enable all MDL (Material Design Lite) components within the target element
            if ($(target).is(SELECTOR_COMPONENT)) {
                componentHandler.upgradeElements(target);
            } else {
                componentHandler.upgradeElements($(SELECTOR_COMPONENT, target).toArray());
            }
        }
    }

    function disable(target) {
        /*jshint undef:false */
        if (componentHandler) {
            // Destroy all MDL (Material Design Lite) components within the target element
            if ($(target).is(SELECTOR_COMPONENT)) {
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

    var $document = $(document);
    var NAMESPACE = 'qor.modal';
    var EVENT_ENABLE = 'enable.' + NAMESPACE;
    var EVENT_DISABLE = 'disable.' + NAMESPACE;
    var EVENT_CLICK = 'click.' + NAMESPACE;
    var EVENT_KEYUP = 'keyup.' + NAMESPACE;
    var EVENT_SHOW = 'show.' + NAMESPACE;
    var EVENT_SHOWN = 'shown.' + NAMESPACE;
    var EVENT_HIDE = 'hide.' + NAMESPACE;
    var EVENT_HIDDEN = 'hidden.' + NAMESPACE;
    var EVENT_TRANSITION_END = 'transitionend';
    var CLASS_OPEN = 'qor-modal-open';
    var CLASS_SHOWN = 'shown';
    var CLASS_FADE = 'fade';
    var CLASS_IN = 'in';
    var ARIA_HIDDEN = 'aria-hidden';

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
            var element = this.$element[0];
            var target = e.target;

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
                    $this.redactor('core.destroy');
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

    let _ = window._,
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
        CLASS_CONTAINER = '.qor-fieldset-container';

    function QorReplicator(element, options) {
        this.$element = $(element);
        this.options = $.extend({}, QorReplicator.DEFAULTS, $.isPlainObject(options) && options);
        this.index = 0;
        this.init();
    }

    QorReplicator.prototype = {
        constructor: QorReplicator,

        init: function() {
            let $element = this.$element,
                $template = $element.find('> .qor-field__block > .qor-fieldset--new'),
                fieldsetName;

            this.isInSlideout = $element.closest('.qor-slideout').length;
            this.hasInlineReplicator = $element.find(CLASS_CONTAINER).length;
            this.maxitems = $element.data('maxItem');
            this.isSortable = $element.hasClass('qor-fieldset-sortable');

            if (!$template.length || $element.closest('.qor-fieldset--new').length) {
                return;
            }

            // Should destroy all components here
            $template.trigger('disable');

            // if have isMultiple data value or template length large than 1
            this.isMultipleTemplate = $element.data('isMultiple');

            if (this.isMultipleTemplate) {
                this.fieldsetName = [];
                this.template = {};
                this.index = [];

                $template.each((i, ele) => {
                    fieldsetName = $(ele).data('fieldsetName');
                    if (fieldsetName) {
                        this.template[fieldsetName] = $(ele).prop('outerHTML');
                        this.fieldsetName.push(fieldsetName);
                    }
                });

                this.parseMultiple();
            } else {
                this.template = $template.prop('outerHTML');
                this.parse();
            }

            $template.hide();
            this.bind();
            this.resetButton();
            this.resetPositionButton();
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
            return this.$element.find('> .qor-field__block > .qor-fieldset').not('.qor-fieldset--new,.is-deleted').length;
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

        parse: function() {
            let template;

            if (!this.template) {
                return;
            }
            template = this.initTemplate(this.template);
            this.template = template.template;
            this.index = template.index;
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

        initTemplate: function(template) {
            let i,
                hasInlineReplicator = this.hasInlineReplicator;

            template = template.replace(/(\w+)\="(\S*\[\d+\]\S*)"/g, function(attribute, name, value) {
                value = value.replace(/^(\S*)\[(\d+)\]([^\[\]]*)$/, function(input, prefix, index, suffix) {
                    if (input === value) {
                        if (name === 'name' && !i) {
                            i = index;
                        }

                        if (hasInlineReplicator && /\[\d+\]/.test(prefix)) {
                            return input.replace(/\[\d+\]/, '[{{index}}]');
                        } else {
                            return prefix + '[{{index}}]' + suffix;
                        }
                    }
                });

                return name + '="' + value + '"';
            });

            return {
                template: template,
                index: parseFloat(i)
            };
        },

        bind: function() {
            let options = this.options;

            this.$element.on(EVENT_CLICK, options.addClass, $.proxy(this.add, this)).on(EVENT_CLICK, options.delClass, $.proxy(this.del, this));

            !this.isInSlideout && $(document).on(EVENT_SUBMIT, 'form', this.removeData.bind(this));
            $(document)
                .on(EVENT_SLIDEOUTBEFORESEND, '.qor-slideout', this.removeData.bind(this))
                .on(EVENT_SELECTCOREBEFORESEND, this.removeData.bind(this));
        },

        unbind: function() {
            this.$element.off(EVENT_CLICK);

            !this.isInSlideout && $(document).off(EVENT_SUBMIT, 'form');
            $(document)
                .off(EVENT_SLIDEOUTBEFORESEND, '.qor-slideout')
                .off(EVENT_SELECTCOREBEFORESEND);
        },

        removeData: function() {
            $('.qor-fieldset--new').remove();
        },

        add: function(e, data, isAutomatically) {
            var options = this.options,
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

                $item = $(template.replace(/\{\{index\}\}/g, this.multipleIndex));

                for (var dataKey in $target.data()) {
                    if (dataKey.match(/^sync/)) {
                        var k = dataKey.replace(/^sync/, '');
                        $item.find("input[name*='." + k + "']").val($target.data(dataKey));
                    }
                }

                if ($fieldset.length) {
                    $fieldset.last().after($item.show());
                } else {
                    parentsChildren.prepend($item.show());
                }
                $item.data('itemIndex', this.multipleIndex).removeClass('qor-fieldset--new');
                this.multipleIndex++;
            } else {
                if (!isAutomatically) {
                    $item = this.addSingle();
                    $target.before($item.show());
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
            let $item,
                $element = this.$element;

            $item = $(this.template.replace(/\{\{index\}\}/g, this.index));
            // add order property for sortable fieldset
            if (this.isSortable) {
                let order = $element.find('> .qor-field__block > .qor-sortable__item').not('.qor-fieldset--new').length;
                $item.attr('order-index', order).css('order', order);
            }

            $item.data('itemIndex', this.index).removeClass('qor-fieldset--new');

            return $item;
        },

        del: function(e) {
            let options = this.options,
                $item = $(e.target).closest(options.itemClass),
                $alert;

            $item
                .addClass('is-deleted')
                .children(':visible')
                .addClass('hidden')
                .hide();
            $alert = $(options.alertTemplate.replace('{{name}}', this.parseName($item)));
            $alert.find(options.undoClass).one(
                EVENT_CLICK,
                function() {
                    if (this.maxitems <= this.getCurrentItems()) {
                        window.QOR.qorConfirm(this.$element.data('maxItemHint'));
                        return false;
                    }

                    $item.find('> .qor-fieldset__alert').remove();
                    $item
                        .removeClass('is-deleted')
                        .children('.hidden')
                        .removeClass('hidden')
                        .show();
                    this.resetButton();
                    this.resetPositionButton();
                }.bind(this)
            );
            this.resetButton();
            this.resetPositionButton();
            $item.append($alert);
        },

        parseNestTemplate: function(templateType) {
            let $element = this.$element,
                parentForm = $element.parents('.qor-fieldset-container'),
                index;

            if (parentForm.length) {
                index = $element.closest('.qor-fieldset').data('itemIndex');
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
            let name = $item.find('input[name]').attr('name') || $item.find('textarea[name]').attr('name');

            if (name) {
                return name.replace(/[^\[\]]+$/, '');
            }
        },

        destroy: function() {
            this.unbind();
            this.$element.removeData(NAMESPACE);
        }
    };

    QorReplicator.DEFAULTS = {
        itemClass: '.qor-fieldset',
        newClass: '.qor-fieldset--new',
        addClass: '.qor-fieldset__add',
        delClass: '.qor-fieldset__delete',
        childrenClass: '.qor-field__block',
        undoClass: '.qor-fieldset__undo',
        alertTemplate:
            '<div class="qor-fieldset__alert">' +
            '<input type="hidden" name="{{name}}._destroy" value="1">' +
            '<button class="mdl-button mdl-button--accent mdl-js-button mdl-js-ripple-effect qor-fieldset__undo" type="button">Undo delete</button>' +
            '</div>'
    };

    QorReplicator.plugin = function(options) {
        return this.each(function() {
            let $this = $(this),
                data = $this.data(NAMESPACE),
                fn;

            if (!data) {
                $this.data(NAMESPACE, (data = new QorReplicator(this, options)));
            }

            if (typeof options === 'string' && $.isFunction((fn = data[options]))) {
                fn.call(data);
            }
        });
    };

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
      var $target = $(e.target);
      var data = $target.data();

      if ($target.is(SEARCH_RESOURCE)){
        var oldUrl = location.href.replace(/#/g, '');
        var newUrl;
        var newResourceName = data.resource;
        var hasResource = /resource_name/.test(oldUrl);
        var hasKeyword = /keyword/.test(oldUrl);
        var resourceParam = 'resource_name=' + newResourceName;
        var searchSymbol = hasKeyword ? '&' : '?keyword=&';

        if (newResourceName){
          if (hasResource){
            newUrl = oldUrl.replace(/resource_name=\w+/g, resourceParam);
          } else {
            newUrl = oldUrl + searchSymbol + resourceParam;
          }
        } else {
          newUrl = oldUrl.replace(/&resource_name=\w+/g, '');
        }

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

        init: function() {
            this.bind();
        },

        bind: function() {
            this.$element.on(EVENT_CLICK, CLASS_CLICK_TABLE, this.processingData.bind(this)).on(EVENT_SUBMIT, 'form', this.submit.bind(this));
        },

        unbind: function() {
            this.$element.off(EVENT_CLICK, '.qor-table tbody tr').off(EVENT_SUBMIT, 'form');
        },

        processingData: function(e) {
            let $this = $(e.target).closest('tr'),
                data = {},
                url,
                options = this.options,
                onSelect = options.onSelect;

            data = $.extend({}, data, $this.data());
            data.$clickElement = $this;

            url = data.mediaLibraryUrl || data.url;

            if (url) {
                $.getJSON(url, function(json) {
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

        submit: function(e) {
            let form = e.target,
                $form = $(form),
                _this = this,
                $submit = $form.find(':submit'),
                data,
                onSubmit = this.options.onSubmit;

            $(document).trigger(EVENT_SELECTCORE_BEFORESEND);

            if (FormData) {
                e.preventDefault();

                $.ajax($form.prop('action'), {
                    method: $form.prop('method'),
                    data: new FormData(form),
                    dataType: 'json',
                    processData: false,
                    contentType: false,
                    beforeSend: function() {
                        $form
                            .parent()
                            .find('.qor-error')
                            .remove();
                        $submit.prop('disabled', true);
                    },
                    success: function(json) {
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
                    error: function(xhr, textStatus, errorThrown) {
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
                    complete: function() {
                        $submit.prop('disabled', false);
                    }
                });
            }
        },

        refresh: function() {
            setTimeout(function() {
                window.location.reload();
            }, 350);
        },

        destroy: function() {
            this.unbind();
        }
    };

    QorSelectCore.plugin = function(options) {
        return this.each(function() {
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
                fn.apply(data);
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
                .on(EVENT_CLICK, '[data-select-modal="many"]', this.openBottomSheets.bind(this))
                .on(EVENT_RELOAD, `.${CLASS_MANY}`, this.reloadData.bind(this));

            this.$element
                .on(EVENT_CLICK, CLASS_CLEAR_SELECT, this.clearSelect.bind(this))
                .on(EVENT_CLICK, CLASS_UNDO_DELETE, this.undoDelete.bind(this));
        },

        unbind: function() {
            $document.off(EVENT_CLICK, '[data-select-modal="many"]').off(EVENT_RELOAD, `.${CLASS_MANY}`);
            this.$element.off(EVENT_CLICK, CLASS_CLEAR_SELECT).off(EVENT_CLICK, CLASS_UNDO_DELETE);
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
            return Mustache.render(this.SELECT_MANY_TEMPLATE, data);
        },

        renderHint: function(data) {
            return Mustache.render(this.SELECT_MANY_HINT, data);
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
                    $option = $(Mustache.render(QorSelectMany.SELECT_MANY_OPTION_TEMPLATE, data));
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
                $option = $(Mustache.render(QorSelectMany.SELECT_MANY_OPTION_TEMPLATE, data));
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
            }

            data.displayName = data.Text || data.Name || data.Title || data.Code || data[firstKey()];

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
    this.init();
  }

  function firstTextKey(obj) {
    var keys = Object.keys(obj);
    if (keys.length > 1 && keys[0] === "ID") {
      return keys[1];
    }
    return keys[0];
  }

  var lock = {lock: false};

  QorSelectOne.prototype = {
    constructor: QorSelectOne,

    init: function() {
      this.$selectOneSelectedTemplate = this.$element.find('[name="select-one-selected-template"]');
      this.$selectOneSelectedIconTemplate = this.$element.find('[name="select-one-selected-icon"]');
      this.bind();
    },

    bind: function() {
      $document
        .on(EVENT_CLICK, '[data-selectone-url]', this.openBottomSheets.bind(this))
        .on(EVENT_RELOAD, `.${CLASS_ONE}`, this.reloadData.bind(this));
      this.$element
        .on(EVENT_CLICK, CLASS_CLEAR_SELECT, this.clearSelect.bind(this))
        .on(EVENT_CLICK, CLASS_CHANGE_SELECT, this.changeSelect);
    },

    unbind: function() {
      $document.off(EVENT_CLICK, '[data-selectone-url]').off(EVENT_RELOAD, `.${CLASS_ONE}`);
      this.$element.off(EVENT_CLICK, CLASS_CLEAR_SELECT).off(EVENT_CLICK, CLASS_CHANGE_SELECT);
    },

    clearSelect: function(e) {
      var $target = $(e.target),
        $parent = $target.closest(CLASS_PARENT);

      $parent.find(CLASS_SELECT_FIELD).remove();
      $parent.find(CLASS_SELECT_INPUT).html('');
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
      if (lock.lock) {
        e.preventDefault();
        return false;
      }

      lock.lock = true;
      setTimeout(function () {lock.lock = false}, 1000*3);
      var $this = $(e.target);
      this.currentData = $this.data();

      this.BottomSheets = $body.data('qor.bottomsheets');
      this.$parent = $this.closest(CLASS_PARENT);

      this.currentData.url = this.currentData.selectoneUrl;
      this.primaryField = this.currentData.remoteDataPrimaryKey;
      this.displayField = this.currentData.remoteDataDisplayKey;

      this.SELECT_ONE_SELECTED_ICON = this.$selectOneSelectedIconTemplate.html();
      this.BottomSheets.open(this.currentData, this.handleSelectOne.bind(this));
    },

    initItem: function() {
      var $selectField = this.$parent.find(CLASS_SELECT_FIELD),
          recordeUrl = this.currentData.remoteRecordeUrl,
          selectedID;

      if (recordeUrl) {
        this.$bottomsheets.find('tr[data-primary-key]').each(function () {
          var $this = $(this), data = $this.data();
          data.url = recordeUrl.replace("{ID}", data.primaryKey)
        })
      }

      if (!$selectField.length) {
        return;
      }

      selectedID = $selectField.data().primaryKey;

      if (selectedID) {
        this.$bottomsheets
          .find('tr[data-primary-key="' + selectedID + '"]')
          .addClass(CLASS_SELECTED)
          .find('td:first')
          .append(this.SELECT_ONE_SELECTED_ICON);
      }
    },

    reloadData: function() {
      this.initItem();
    },

    renderSelectOne: function(data) {
      return Mustache.render(this.$selectOneSelectedTemplate.html(), data);
    },

    handleSelectOne: function($bottomsheets) {
      var options = {
        onSelect: this.onSelectResults.bind(this), //render selected item after click item lists
        onSubmit: this.onSubmitResults.bind(this) //render new items after new item form submitted
      };

      $bottomsheets.qorSelectCore(options).addClass(CLASS_ONE);
      this.$bottomsheets = $bottomsheets;
      this.initItem();
    },

    onSelectResults: function(data) {
      this.handleResults(data);
    },

    onSubmitResults: function(data) {
      this.handleResults(data, true);
    },

    handleResults: function(data) {
      var template,
          $parent = this.$parent,
          $select = $parent.find('select'),
          $selectFeild = $parent.find(CLASS_SELECT_FIELD);

      data.displayName = this.displayField ? data[this.displayField] :
          (data.Text || data.Name || data.Title || data.Code || firstTextKey(data));
      data.selectoneValue = this.primaryField ? data[this.primaryField] : (data.primaryKey || data.ID);

      if (!$select.length) {
        return;
      }

      template = this.renderSelectOne(data);

      if ($selectFeild.length) {
        $selectFeild.remove();
      }

      $parent.prepend(template);
      $parent.find(CLASS_SELECT_TRIGGER).hide();

      $select.html(Mustache.render(QorSelectOne.SELECT_ONE_OPTION_TEMPLATE, data));
      $select[0].value = data.primaryKey || data.ID;

      $parent.trigger('qor.selectone.selected', [data]);

      this.$bottomsheets.qorSelectCore('destroy').remove();
      if (!$('.qor-bottomsheets').is(':visible')) {
        $('body').removeClass('qor-bottomsheets-open');
      }
    },

    destroy: function() {
      this.unbind();
      this.$element.removeData(NAMESPACE);
    }
  };

  QorSelectOne.SELECT_ONE_OPTION_TEMPLATE = '<option value="[[ selectoneValue ]]" selected>[[ displayName ]]</option>';

  QorSelectOne.plugin = function(options) {
    return this.each(function() {
      var $this = $(this);
      var data = $this.data(NAMESPACE);
      var fn;

      if (!data) {
        if (/destroy/.test(options)) {
          return;
        }

        $this.data(NAMESPACE, (data = new QorSelectOne(this, options)));
      }

      if (typeof options === 'string' && $.isFunction((fn = data[options]))) {
        fn.apply(data);
      }
    });
  };

  $(function() {
    var selector = '[data-toggle="qor.selectone"]';
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
        CLASS_BODY_LOADING = '.qor-body__loading';

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
        $ele.each(function() {
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

    function QorSlideout(element, options) {
        this.$element = $(element);
        this.options = $.extend({}, QorSlideout.DEFAULTS, $.isPlainObject(options) && options);
        this.slided = false;
        this.disabled = false;
        this.slideoutType = false;
        this.init();
    }

    QorSlideout.prototype = {
        constructor: QorSlideout,

        init: function() {
            this.build();
            this.bind();
        },

        build: function() {
            var $slideout;

            this.$slideout = $slideout = $(QorSlideout.TEMPLATE).appendTo('body');
            this.$slideoutTemplate = $slideout.html();
        },

        unbuild: function() {
            this.$slideout.remove();
        },

        bind: function() {
            this.$slideout
                .on(EVENT_SUBMIT, 'form', this.submit.bind(this))
                .on(EVENT_CLICK, '.qor-slideout__fullscreen', this.toggleSlideoutMode.bind(this))
                .on(EVENT_CLICK, '[data-dismiss="slideout"]', this.closeSlideout.bind(this));

            $document.on(EVENT_KEYUP, $.proxy(this.keyup, this));
        },

        unbind: function() {
            this.$slideout.off(EVENT_SUBMIT, this.submit).off(EVENT_CLICK);

            $document.off(EVENT_KEYUP, this.keyup);
        },

        keyup: function(e) {
            if (e.which === 27) {
                if ($('.qor-bottomsheets').is(':visible') || $('.qor-modal').is(':visible') || $('#redactor-modal-box').length || $('#dialog').is(':visible')) {
                    return;
                }

                this.hide();
                this.removeSelectedClass();
            }
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

        removeSelectedClass: function() {
            this.$element.find('[data-url]').removeClass(CLASS_IS_SELECTED);
        },

        addLoading: function() {
            $(CLASS_BODY_LOADING).remove();
            var $loading = $(QorSlideout.TEMPLATE_LOADING);
            $loading.appendTo($('body')).trigger('enable');
        },

        toggleSlideoutMode: function() {
            this.$slideout
                .toggleClass('qor-slideout__fullscreen')
                .find('.qor-slideout__fullscreen i')
                .toggle();
        },

        submit: function(e) {
            var $slideout = this.$slideout;
            var $body = this.$body;
            var form = e.target;
            var $form = $(form);
            var _this = this;
            var $submit = $form.find(':submit');

            $slideout.trigger(EVENT_SLIDEOUT_BEFORESEND);

            if (FormData) {
                e.preventDefault();

                $.ajax($form.prop('action'), {
                    method: $form.prop('method'),
                    data: new FormData(form),
                    dataType: 'html',
                    processData: false,
                    contentType: false,
                    beforeSend: function() {
                        $submit.prop('disabled', true);
                        $.fn.qorSlideoutBeforeHide = null;
                    },
                    success: function(html) {
                        var returnUrl = $form.data('returnUrl');
                        var refreshUrl = $form.data('refreshUrl');

                        $slideout.trigger(EVENT_SLIDEOUT_SUBMIT_COMPLEMENT);

                        if (refreshUrl) {
                            window.location.href = refreshUrl;
                            return;
                        }

                        if (returnUrl == 'refresh') {
                            _this.refresh();
                            return;
                        }

                        if (returnUrl && returnUrl != 'refresh') {
                            _this.load(returnUrl);
                        } else {
                            var prefix = '/' + location.pathname.split('/')[1];
                            var flashStructs = [];
                            $(html)
                                .find('.qor-alert')
                                .each(function(i, e) {
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
                    error: function(xhr, textStatus, errorThrown) {
                        var $error;

                        if (xhr.status === 422) {
                            $body.find('.qor-error').remove();
                            $form
                                .find('.qor-field')
                                .removeClass('is-error')
                                .find('.qor-field__error')
                                .remove();

                            $error = $(xhr.responseText).find('.qor-error');
                            $form.before($error);

                            $error.find('> li > label').each(function() {
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
                            window.alert([textStatus, errorThrown].join(': '));
                        }
                    },
                    complete: function() {
                        $submit.prop('disabled', false);
                    }
                });
            }
        },

        load: function(url, data) {
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

            method = data.method ? data.method : 'GET';
            dataType = data.datatype ? data.datatype : 'html';

            load = $.proxy(function() {
                $.ajax(url, {
                    method: method,
                    dataType: dataType,
                    cache: true,
                    ifModified: true,
                    success: $.proxy(function(response) {
                        let $response, $content, $qorFormContainer, $scripts, $links, bodyClass;

                        $(CLASS_BODY_LOADING).remove();

                        if (method === 'GET') {
                            $response = $(response);
                            $content = $response.find(CLASS_MAIN_CONTENT);
                            $qorFormContainer = $content.find('.qor-form-container');
                            this.slideoutType = $qorFormContainer.length && $qorFormContainer.data().slideoutType;

                            if (!$content.length) {
                                return;
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

                                loadScripts($scripts, data, function() {});
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
                                .one(EVENT_SHOWN, function() {
                                    $(this).trigger('enable');
                                })
                                .one(EVENT_HIDDEN, function() {
                                    $(this).trigger('disable');
                                });

                            $slideout.find('.qor-slideout__opennew').attr('href', url);
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
                            $slideout.trigger(EVENT_SLIDEOUT_LOADED, [url, response]);
                        } else {
                            if (data.returnUrl) {
                                this.load(data.returnUrl);
                            } else {
                                this.refresh();
                            }
                        }
                    }, this),

                    error: $.proxy(function() {
                        var errors;
                        $(CLASS_BODY_LOADING).remove();
                        if ($('.qor-error span').length > 0) {
                            errors = $('.qor-error span')
                                .map(function() {
                                    return $(this).text();
                                })
                                .get()
                                .join(', ');
                        } else {
                            errors = 'Server error, please try again later!';
                        }
                        window.alert(errors);
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

        open: function(options) {
            this.addLoading();
            this.load(options.url, options.data);
        },

        reload: function(url) {
            this.hide(true);
            this.load(url);
        },

        show: function() {
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

        shown: function() {
            this.slided = true;
            // Disable to scroll body element
            $('body').addClass(CLASS_OPEN);
            this.$slideout
                .trigger('beforeEnable.qor.slideout')
                .trigger(EVENT_SHOWN)
                .trigger('afterEnable.qor.slideout');
        },

        closeSlideout: function() {
            this.hide();
        },

        hide: function(isReload) {
            let _this = this,
                message = {
                    confirm:
                        'You have unsaved changes on this slideout. If you close this slideout, you will lose all unsaved changes. Are you sure you want to close the slideout?'
                };

            if ($.fn.qorSlideoutBeforeHide) {
                window.QOR.qorConfirm(message, function(confirm) {
                    if (confirm) {
                        _this.hideSlideout(isReload);
                    }
                });
            } else {
                this.hideSlideout(isReload);
            }

            this.removeSelectedClass();
        },

        hideSlideout: function(isReload) {
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

        hidden: function() {
            this.slided = false;

            // Enable to scroll body element
            $('body').removeClass(CLASS_OPEN);

            this.$slideout.removeClass(CLASS_IS_SHOWN).trigger(EVENT_HIDDEN);
        },

        refresh: function() {
            this.hide();

            setTimeout(function() {
                window.location.reload();
            }, 350);
        },

        destroy: function() {
            this.unbind();
            this.unbuild();
            this.$element.removeData(NAMESPACE);
        }
    };

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

    QorSlideout.plugin = function(options) {
        return this.each(function() {
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
      this.$element.on(EVENT_CLICK, '> thead > tr > th', $.proxy(this.sort, this));
    },

    unbind: function () {
      this.$element.off(EVENT_CLICK, this.sort);
    },

    sort: function (e) {
      var $target = $(e.currentTarget);
      var orderBy = $target.data('orderBy');
      var search = location.search;
      var param = 'order_by=' + orderBy;

      // Stop when it is not sortable
      if (!orderBy) {
        return;
      }

      if (/order_by/.test(search)) {
        search = search.replace(/order_by(=\w+)?/, function () {
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
