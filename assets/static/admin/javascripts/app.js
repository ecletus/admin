$(function() {
    'use strict';

    $(document).on('click.qor.alert', '[data-dismiss="alert"]', function() {
        $(this).closest('.qor-alert').remove();
    });

    setTimeout(function() {
        $('.qor-alert[data-dismissible="true"]').remove();
    }, 5000);
});

$(function () {

  'use strict';

  var $form = $('.qor-page__body > .qor-form-container > form');

  $('.qor-error > li > label').each(function () {
    var $label = $(this);
    var id = $label.attr('for');

    if (id) {
      $form.find('#' + id).
        closest('.qor-field').
        addClass('is-error').
        append($label.clone().addClass('qor-field__error'));
    }
  });

});

$(function () {

  'use strict';

  var modal = (
    '<div class="qor-dialog qor-dialog--global-search" tabindex="-1" role="dialog" aria-hidden="true">' +
      '<div class="qor-dialog-content">' +
        '<form action=[[actionUrl]]>' +
          '<div class="mdl-textfield mdl-js-textfield" id="global-search-textfield">' +
            '<input class="mdl-textfield__input ignore-dirtyform" name="keyword" id="globalSearch" value="" type="text" placeholder="" />' +
            '<label class="mdl-textfield__label" for="globalSearch">[[placeholder]]</label>' +
          '</div>' +
        '</form>' +
      '</div>' +
    '</div>'
  );

  $(document).on('click', '.qor-dialog--global-search', function(e){
    e.stopPropagation();
    if (!$(e.target).parents('.qor-dialog-content').length && !$(e.target).is('.qor-dialog-content')){
      $('.qor-dialog--global-search').remove();
    }
  });

  $(document).on('click', '.qor-global-search--show', function(e){
      e.preventDefault();

      var data = $(this).data();
      var modalHTML = window.Mustache.render(modal, data);

      $('body').append(modalHTML);
      window.componentHandler.upgradeElement(document.getElementById('global-search-textfield'));
      $('#globalSearch').focus();

  });
});

$(function() {
  'use strict';

  var menuDatas = [],
    storageName = 'qoradmin_menu_status',
    lastMenuStatus = localStorage.getItem(storageName);

  if (lastMenuStatus && lastMenuStatus.length) {
    menuDatas = lastMenuStatus.split(',');
  }

  $('.qor-menu-container')
    .on('click', '> ul > li > a', function() {
      var $this = $(this),
        $li = $this.parent(),
        $ul = $this.next('ul'),
        menuName = $li.attr('qor-icon-name');

      if (!$ul.length) {
        return;
      }

      if ($ul.hasClass('in')) {
        menuDatas.push(menuName);

        $li.removeClass('is-expanded');
        $ul
          .one('transitionend', function() {
            $ul.removeClass('collapsing in');
          })
          .addClass('collapsing')
          .height(0);
      } else {
        menuDatas = _.without(menuDatas, menuName);

        $li.addClass('is-expanded');
        $ul
          .one('transitionend', function() {
            $ul.removeClass('collapsing');
          })
          .addClass('collapsing in')
          .height($ul.prop('scrollHeight'));
      }
      localStorage.setItem(storageName, menuDatas);
    })
    .find('> ul > li > a')
    .each(function() {
      var $this = $(this),
        $li = $this.parent(),
        $ul = $this.next('ul'),
        menuName = $li.attr('qor-icon-name');

      if (!$ul.length) {
        return;
      }

      $ul.addClass('collapse');
      $li.addClass('has-menu');

      if (menuDatas.indexOf(menuName) != -1) {
        $ul.height(0);
      } else {
        $li.addClass('is-expanded');
        $ul.addClass('in').height($ul.prop('scrollHeight'));
      }
    });

  var $pageHeader = $('.qor-page > .qor-page__header'),
    $pageBody = $('.qor-page > .qor-page__body'),
    triggerHeight = $pageHeader.find('.qor-page-subnav__header').length ? 96 : 48;

  if ($pageHeader.length) {
    if ($pageHeader.height() > triggerHeight) {
      // see qor-head-fixer.js
      //$pageBody.css('padding-top', $pageHeader.height());
    }

    $('.qor-page').addClass('has-header');
    $('header.mdl-layout__header').addClass('has-action');
  }
});

$(function () {
  $('.qor-mobile--show-actions').on('click', function () {
    $('.qor-page__header').toggleClass('actions-show');
  });
});

$(function() {
    'use strict';

    var $body = $('body'),
        Slideout,
        BottomSheets,
        CLASS_IS_SELECTED = 'is-selected',
        isSlideoutOpened = function() {
            return $body.hasClass('qor-slideout-open');
        },
        isBottomsheetsOpened = function() {
            return $body.hasClass('qor-bottomsheets-open');
        };

    $body.qorBottomSheets();
    $body.qorSlideout();

    Slideout = $body.data('qor.slideout');
    BottomSheets = $body.data('qor.bottomsheets');

    function toggleSelectedCss(ele) {
        $('[data-url]').removeClass(CLASS_IS_SELECTED);
        ele && ele.length && ele.addClass(CLASS_IS_SELECTED);
    }

    function collectSelectID() {
        var $checked = $('.qor-js-table tbody').find('.mdl-checkbox__input:checked'),
            IDs = [];

        if (!$checked.length) {
            return false;
        }

        $checked.each(function() {
            IDs.push(
                $(this)
                    .closest('tr')
                    .data('primary-key')
            );
        });

        return IDs;
    }

    $(document).on('click.qor.openUrl', '[data-url]', function(e) {
        var $this = $(this),
            $target = $(e.target),
            isNewButton = $this.hasClass('qor-button--new'),
            isEditButton = $this.hasClass('qor-button--edit'),
            isInTable = ($this.is('.qor-table tr[data-url]') || $this.closest('.qor-js-table').length) && !$this.closest('.qor-slideout').length, // if table is in slideout, will open bottom sheet
            openData = $this.data(),
            actionData,
            openType = openData.openType,
            hasSlideoutTheme = $this.parents('.qor-theme-slideout').length,
            isActionButton = ($this.hasClass('qor-action-button') || $this.hasClass('qor-action--button')) && !openType;

        if (openType !== "bottomsheet") {
            if ($target.is('.mdl-data-table__select,.qor-table__actions') || $target.parents('.mdl-data-table__select,.qor-table__actions,.qor-actions-bulk').length) {
                return
            }
        }

        // if clicking item's menu actions
        if ($target.closest('.qor-button--actions').length || (!$target.data('url') && $target.is('a')) || (isInTable && isBottomsheetsOpened())) {
            return;
        }

        if (isActionButton) {
            actionData = collectSelectID();
            if (actionData) {
                openData = $.extend({}, openData, {
                    actionData: actionData
                });
            }
        }

        openData.$target = $target;

        if (!openData.method || openData.method.toUpperCase() == 'GET') {
            // Open in BottmSheet: is action button, open type is bottom-sheet
            if (isActionButton || openType == 'bottomsheet') {
                // if is bulk action and no item selected
                if (isActionButton && !actionData && $this.closest('[data-toggle="qor.action.bulk"]').length) {
                    window.QOR.qorConfirm(openData.errorNoItem);
                    return false;
                }

                BottomSheets.open(openData);
                return false;
            }

            // Slideout or New Page: table items, new button, edit button
            if (isInTable || (isNewButton && !isBottomsheetsOpened()) || isEditButton || openType == 'slideout') {
                if (hasSlideoutTheme || openType == 'slideout') {
                    if ($this.hasClass(CLASS_IS_SELECTED)) {
                        Slideout.hide();
                        toggleSelectedCss();
                        return false;
                    } else {
                        Slideout.open(openData);
                        toggleSelectedCss($this);
                        return false;
                    }
                } else {
                    window.location = openData.url;
                    return false;
                }
            }

            // Open in BottmSheet: slideout is opened or openType is Bottom Sheet
            if (isSlideoutOpened() || (isNewButton && isBottomsheetsOpened())) {
                BottomSheets.open(openData);
                return false;
            }

            // Other clicks
            if (hasSlideoutTheme) {
                Slideout.open(openData);
                return false;
            } else {
                BottomSheets.open(openData);
                return false;
            }
        }
    });
});

$(function () {

  'use strict';

  var location = window.location;

  $('.qor-search').each(function () {
    var $this = $(this);
    var $input = $this.find('.qor-search__input');
    var $clear = $this.find('.qor-search__clear');
    var isSearched = !!$input.val();

    var emptySearch = function () {
      var search = location.search.replace(new RegExp($input.attr('name') + '\\=?\\w*'), '');
      if (search == '?'){
        location.href = location.href.split('?')[0];
      } else {
        location.search = location.search.replace(new RegExp($input.attr('name') + '\\=?\\w*'), '');
      }
    };

    $this.closest('.qor-page__header').addClass('has-search');
    $('header.mdl-layout__header').addClass('has-search');

    $clear.on('click', function () {
      if ($input.val() || isSearched) {
        emptySearch();
      } else {
        $this.removeClass('is-dirty');
      }
    });
  });
});
