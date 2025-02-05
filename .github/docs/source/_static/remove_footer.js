$(document).ready(function(){
  // Wait a little for the theme_switcher.js to finish
  setTimeout(function(){
    // Remove the link(s) by matching their hrefs
    $('footer a[href="https://github.com/MrDogeBro/sphinx_rtd_dark_mode"]').remove();
    $('footer a[href="http://mrdogebro.com"]').remove();
  }, 500);  // Adjust the delay (in milliseconds) if needed
});
