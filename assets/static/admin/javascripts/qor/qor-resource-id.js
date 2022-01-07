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