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
