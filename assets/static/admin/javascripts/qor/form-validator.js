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
