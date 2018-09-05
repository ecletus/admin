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
