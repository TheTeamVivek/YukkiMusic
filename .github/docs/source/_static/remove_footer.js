document.addEventListener('DOMContentLoaded', function () {
  setTimeout(() => {
    const footerText = document.querySelector('footer a[href="https://github.com/MrDogeBro/sphinx_rtd_dark_mode"]');
    if (footerText) {
      footerText.parentNode.removeChild(footerText);
    }
  }, 300);
});
