$(function() {
    light_up_username();
});

function light_up_username() {
  var username = $("#username");
  username.focus(function() {
    username.prev().css("text-shadow", "0 0 2px white");
  });
  username.focusout(function() {
    username.prev().css("text-shadow", "none");
  });
}
