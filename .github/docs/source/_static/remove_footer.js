document.addEventListener('DOMContentLoaded', function () {
  const footerLinks = document.querySelectorAll('footer a');
  footerLinks.forEach(link => {
    if (link.href.includes('sphinx_rtd_dark_mode') || link.href.includes('mrdogebro')) {
      link.parentNode.removeChild(link);
    }
  });
});
