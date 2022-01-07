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