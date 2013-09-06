// main.js contains all application javascript other than bootstrap, jquery, etc.
(function () {
    "use strict";

    // toggle Bootstrap tooltips based on data-toggle="tooltip"
    $(function () {
        $("[data-toggle='tooltip']").tooltip();
    });
    // toggle Bootstrap popover based on data-toggle="popover"
    $(function () {
        $("[data-toggle='popover']").popover();
    });

    // show/hide class descriptions checkbox
    var hideDescriptionSwitch = false;
    $('#toggle-class-description').click(function () {
        hideDescriptionSwitch = !hideDescriptionSwitch;
        if (hideDescriptionSwitch) {
            $('.toggle-description-target').hide();
        } else {
            $('.toggle-description-target').show();
        }
    });

    // filter sections with checkboxes 
    // TODO (kvu787): Use metaprogramming to shorten the following
    var hideClosedSwitch = false;
    var hideFreshmenSwitch = false;
    var hideWithdrawalSwitch = false;
    var sectToggles = [];
    $('#toggle-closed').click(function () {
        hideClosedSwitch = !hideClosedSwitch;
        if (hideClosedSwitch) {
            sectToggles.push('sect-closed');
        } else {
           sectToggles = removeElement(sectToggles, 'sect-closed'); 
        }
        toggleSects($('.sect-target'), sectToggles);
    });
    $('#toggle-freshmen').click(function () {
        hideFreshmenSwitch = !hideFreshmenSwitch;
        if (hideFreshmenSwitch) {
            sectToggles.push('sect-freshmen');
        } else {
           sectToggles = removeElement(sectToggles, 'sect-freshmen'); 
        }
        toggleSects($('.sect-target'), sectToggles);
    });
    $('#toggle-withdrawal').click(function () {
        hideWithdrawalSwitch = !hideWithdrawalSwitch;
        if (hideWithdrawalSwitch) {
            sectToggles.push('sect-withdrawal');
        } else {
           sectToggles = removeElement(sectToggles, 'sect-withdrawal'); 
        }
        toggleSects($('.sect-target'), sectToggles);
    });

    // Loops through elements. Hides element if it has any class 
    // in classList, else shows it. 
    var toggleSects = function (elements, classList) {
        for (var i = 0; i < elements.length; i++) {
            var element = elements[i];
            if (hasAnyClass(element, classList)) {
                element.style['display'] = 'none';
            } else {
                element.style['display'] = 'block';
            }
        }
    }

    // Returns a copy of the input array with all instances of
    // element removed.
    var removeElement = function (array, element) {
        return array = jQuery.grep(array, function(n, i) {
            return n != element;
        });
    }

    // Returns true if the element has any class in classList. Else, 
    // returns false.
    var hasAnyClass = function (element, classList) {
        for (var i = 0; i < classList.length; i++) {
            var classes = element.getAttribute('class').split(' ');
            for (var j = 0; j < classes.length; j++) {
                if (classes[j] == classList[i]) {
                    return true;
                }
            }
        }
        return false;
    }
}());