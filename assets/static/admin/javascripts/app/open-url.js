$(function () {
        'use strict';
        const EVENT_ENABLE = 'enable';

        if (window.PRINT_MODE) {
            $('a').attr('href', 'javascript:void(0)');
            $(document)
                .on(EVENT_ENABLE, function (e) {
                    $('a', e.target).each(function () {
                        $(this).attr('href', 'javascript:void(0)');
                    })
                })
            return;
        }

        let $body = $('body'),
            Slideout,
            BottomSheets,
            CLASS_IS_SELECTED = 'is-selected',
            isSlideoutOpened = function () {
                return $body.hasClass('qor-slideout-open');
            },
            isBottomsheetsOpened = function () {
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
            let $checked = $('.qor-js-table tbody').find('.mdl-checkbox__input:checked'),
                IDs = [];

            if (!$checked.length) {
                return false;
            }

            $checked.each(function () {
                IDs.push(
                    $(this)
                        .closest('tr')
                        .data('primary-key')
                );
            });

            return IDs;
        }

        $(document).on('click.qor.openUrl', '[data-url]', function (e) {
            let $this = $(this),
                $target = $(e.target),
                isNewButton = $this.hasClass('qor-button--new'),
                isEditButton = $this.hasClass('qor-button--edit'),
                isInTable = ($this.is('.qor-table tr[data-url]') || $this.closest('.qor-js-table').length) && !$this.closest('.qor-slideout').length, // if table is in slideout, will open bottom sheet
                openData = $this.data(),
                actionData,
                openType = openData.openType,
                hasSlideoutTheme = $this.parents('.qor-theme-slideout').length,
                isActionButton = $this.hasClass('qor-action-button') && !openType,
                isActionBulkButton = $this.hasClass('qor-action-button--form'),
                dataUrl;

            if (openType !== "bottomsheet") {
                if (!isActionBulkButton && ($target.is('.mdl-data-table__select,.qor-table__actions') || $target.parents('.mdl-data-table__select,.qor-table__actions,.qor-actions-bulk').length)) {
                    return
                }
            }

            // if clicking item's menu actions
            if ($target.closest('.qor-button--actions').length || ((dataUrl = $target.data('url')) === "" || (!dataUrl && $target.is('a'))) || (isInTable && isBottomsheetsOpened())) {
                return;
            } else if ((dataUrl = $target.data('url')))

                if (isActionButton) {
                    actionData = collectSelectID();
                    if (actionData) {
                        openData = $.extend({}, openData, {
                            actionData: actionData
                        });
                    }
                }

            if (!openData.url) {
                return;
            }

            openData.$target = $target;

            if (!openData.method || openData.method.toUpperCase() === 'GET') {
                // Open in BottmSheet: is action button or inside in bottomsheet, open type is bottom-sheet
                if (isActionButton || openType === 'bottomsheet' || $target.closest('.qor-bottomsheets__body').length) {
                    // if is bulk action and no item selected
                    if (isActionButton && !actionData && $this.closest('[data-toggle="qor.action.bulk"]').length) {
                        window.QOR.qorConfirm(openData.errorNoItem);
                        return false;
                    }

                    BottomSheets.open(openData);
                    return false;
                }

                // Slideout or New Page: table items, new button, edit button
                if (isInTable || (isNewButton && !isBottomsheetsOpened()) || isEditButton || openType === 'slideout') {
                    if (hasSlideoutTheme || openType === 'slideout') {
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
    }
);
