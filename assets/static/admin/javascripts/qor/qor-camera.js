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