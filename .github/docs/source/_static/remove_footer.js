$(document).ready(function(){
    setTimeout(function(){
        // Remove the unwanted links
        $('footer a[href="https://github.com/MrDogeBro/sphinx_rtd_dark_mode"]').remove();
        $('footer a[href="http://mrdogebro.com"]').remove();
        
        // Clean up leftover "provided by ." text
        $("footer").contents().filter(function() {
            return this.nodeType === 3 && this.nodeValue.trim() === "provided by .";
        }).remove();
    }, 500);
});
