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
            common:{
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

    $(function () {
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

        QOR.ajaxError = function(xhr, textStatus, errorThrown) {
            QOR.alert(QOR.ajaxErrorString.apply(this, arguments));
        };

        QOR.ajaxErrorString = function(xhr, textStatus, errorThrown) {
            return "<strong>"+QOR.messages.common.ajaxError + "<strong></strong>:<br/>"+ [textStatus, errorThrown].join(': ')
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
                values = pairs[i+1];
                values.forEach(function (value) {
                    $form.append($(`<input type="hidden" name="${pairs[i]}" value="${value}">`))
                })
            }
            $form.submit();
            return false;
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
    });
});
