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
    this.selectedRender = null;
    this.init();
  }

  function firstTextKey(obj) {
    var keys = Object.keys(obj);
    if (keys.length > 1 && keys[0] === "ID") {
      return keys[1];
    }
    return keys[0];
  }

  var glock = {
    lock: false,
    id: 0
  }

  function gunlock() {
    glock.lock = false
  }

  QorSelectOne.prototype = {
    constructor: QorSelectOne,

    id: undefined,

    init: function() {
      let selectedRender = this.$element.data().selectedRender;
      if (selectedRender) {
        eval('this.selectedRender = function(data){'+atob(selectedRender)+'};');
      }
      this.id = glock.id++;
      this.$selectOneSelectedTemplate = this.$element.find('[name="select-one-selected-template"]');
      this.$selectOneSelectedIconTemplate = this.$element.find('[name="select-one-selected-icon"]');
      this.lock = {
        $parent: null,
        $select: null,
      };
      this.bind();
    },

    bind: function() {
      $document
        .on(EVENT_RELOAD, `.${CLASS_ONE}`, this.reloadData.bind(this));
      this.$element
        .on(EVENT_CLICK, CLASS_CLEAR_SELECT, this.clearSelect.bind(this))
        .on(EVENT_CLICK, '[data-selectone-url],[data-selectone-url] .material-icons', this.openBottomSheets.bind(this))
        .on(EVENT_CLICK, CLASS_CHANGE_SELECT, this.changeSelect);
    },

    unbind: function() {
      $document.off(EVENT_RELOAD, `.${CLASS_ONE}`);
      this.$element.off(EVENT_CLICK, CLASS_CLEAR_SELECT).off(EVENT_CLICK, '[data-selectone-url]').off(EVENT_CLICK, CLASS_CHANGE_SELECT);
    },

    clearSelect: function(e) {
      var $target = $(e.target),
        $parent = $target.closest(CLASS_PARENT);

      $parent.find(CLASS_SELECT_FIELD).remove();
      $parent.find(CLASS_SELECT_INPUT).html('<option value="" selected></option>');
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
      if (glock.lock) {
        e.preventDefault();
        return false;
      }

      glock.lock = true;
      setTimeout(gunlock, 1000*3);
      let $this = $(e.target);
      if ($this.is('.material-icons')) {
        $this = $this.parent()
      }
      this.lock.currentData = $this.data();

      this.lock.$parent = $this.closest(CLASS_PARENT);
      this.lock.$select = this.lock.$parent.find('select');

      this.lock.currentData.url = this.lock.currentData.selectoneUrl;
      this.lock.primaryField = this.lock.currentData.remoteDataPrimaryKey;
      this.lock.displayField = this.lock.currentData.remoteDataDisplayKey;
      this.lock.iconField = this.lock.currentData.remoteDataIconKey;

      this.lock.SELECT_ONE_SELECTED_ICON = this.$selectOneSelectedIconTemplate.html();
      let data = $.extend({}, this.lock.currentData);
      if (this.lock.$select.length) {
        data.$element = this.lock.$select;
      }
      $('body').qorBottomSheets('open', data, this.handleSelectOne.bind(this));
    },

    initItem: function() {
      var $selectField = this.lock.$parent.find(CLASS_SELECT_FIELD),
          recordeUrl = this.lock.currentData.remoteRecordeUrl,
          selectedID;

      if (recordeUrl) {
        this.lock.$bottomsheets.find('tr[data-primary-key]').each(function () {
          var $this = $(this), data = $this.data();
          data.url = recordeUrl.replace("{ID}", data.primaryKey)
        })
      }

      if (!$selectField.length) {
        return;
      }

      selectedID = $selectField.data().primaryKey;

      if (selectedID) {
        this.lock.$bottomsheets
          .find('tr[data-primary-key="' + selectedID + '"]')
          .addClass(CLASS_SELECTED)
          .find('td:first')
          .append(this.lock.SELECT_ONE_SELECTED_ICON);
      }
    },

    reloadData: function() {
      this.initItem();
    },

    renderSelectOne: function(data) {
      const res = Mustache.render(this.$selectOneSelectedTemplate.html().replace(/\[\[ *&amp;/g, '[[&'), data);
      return res;
    },

    handleSelectOne: function($bottomsheets) {
      var options = {
        onSelect: this.onSelectResults.bind(this), //render selected item after click item lists
        onSubmit: this.onSubmitResults.bind(this) //render new items after new item form submitted
      };

      $bottomsheets.qorSelectCore(options).addClass(CLASS_ONE);
      this.lock.$bottomsheets = $bottomsheets;
      this.initItem();
    },

    onSelectResults: function(data) {
      this.handleResults(data);
    },

    onSubmitResults: function(data) {
      this.handleResults(data, true);
    },

    handleResults: function(data) {
      let template,
          $parent = this.lock.$parent,
          $selectField = $parent.find(CLASS_SELECT_FIELD);

      data.displayName = this.lock.displayField ? data[this.lock.displayField] :
          (data.Text || data.Name || data.Title || data.Value || data.Code || firstTextKey(data));
      data.selectoneValue = this.lock.primaryField ? data[this.lock.primaryField] : (data.primaryKey || data.ID);

      if (this.lock.iconField) {
          data.icon = data[this.lock.iconField];
      }

      if (data.icon && /\.svg/.test(data.icon)) {
          data.iconSVG = true;
      }

      if (!this.lock.$select.length) {
        return;
      }

      if (this.selectedRender) {
        data.displayText = this.selectedRender(data)
      }
      template = this.renderSelectOne(data);

      if ($selectField.length) {
        $selectField.remove();
      }

      $parent.prepend(template);
      $parent.find(CLASS_SELECT_TRIGGER).hide();

      const res = Mustache.render(QorSelectOne.SELECT_ONE_OPTION_TEMPLATE, data);
      this.lock.$select.html(res);
      // this.lock.$select[0].value = data.primaryKey || data.ID;

      $parent.trigger('qor.selectone.selected', [data]);

      this.lock.$bottomsheets.qorSelectCore('destroy').remove();
      if (!$('.qor-bottomsheets').is(':visible')) {
        $('body').removeClass('qor-bottomsheets-open');
      }
    },

    destroy: function() {
      this.unbind();
      this.$element.removeData(NAMESPACE);
      this.lock = undefined;
    }
  };

  QorSelectOne.SELECT_ONE_OPTION_TEMPLATE = '<option value="[[ selectoneValue ]]" selected>[[ displayName ]]</option>';

  QorSelectOne.plugin = function(options) {
    let args = Array.prototype.slice.call(arguments, 1);
    return this.each(function() {
      let $this = $(this),
        data = $this.data(NAMESPACE),
        fn;

      if (!data) {
        if (/destroy/.test(options)) {
          return;
        }

        $this.data(NAMESPACE, (data = new QorSelectOne(this, options)));
      }

      if (typeof options === 'string' && $.isFunction((fn = data[options]))) {
        fn.apply(data, args);
      }
    });
  };

  $(function() {
    const selector = '[data-toggle="qor.selectone"]';
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
