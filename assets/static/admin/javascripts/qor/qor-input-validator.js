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