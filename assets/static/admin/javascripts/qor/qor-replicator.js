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
