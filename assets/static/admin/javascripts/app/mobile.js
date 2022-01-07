$(function () {
  $('.qor-mobile--show-actions').on('click', function () {
    const $el = $('.qor-page__header').toggleClass('actions-show');
    $el.is(':visible') ? $el.hide() : $el.show();
  });
});
