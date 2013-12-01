// main.js contains all application javascript other than bootstrap, jquery, etc.

// Implement a string.startsWith function.
// CREDIT: http://stackoverflow.com/questions/646628/javascript-startswith
if (typeof String.prototype.startsWith != 'function') {
  String.prototype.startsWith = function (str){
    return this.slice(0, str.length) == str;
  };
}

// Handle changing the <body> padding-top and showing 'hide search' when resized.
(function () {
    "use strict";

    var searchBoxHidden = true;
    $('#search-box-form').hide();
    $('#show-search-div').show();
    $('body').css('padding-top', '130px');

    var changeTopMargin = function () {
        if ($(window).width() < 768) {
            if (searchBoxHidden) {
                $('body').css('padding-top', '130px');
            } else {
                $('body').css('padding-top', '180px');
            }
            $('#search-box-form').hide();
            $('#show-search-div').show();
            $('body').css('padding-top', '130px');
        } else {
            searchBoxHidden = false;
            $('body').css('padding-top', '70px');
            $('#search-box-form').show();
            $('#hide-search-div').hide();
            $('#show-search-div').hide();
        }
    };

    $('#hide-search-link').click(function() {
        searchBoxHidden = true;
        $('#search-box-form').hide();
        $('#show-search-div').show();
        $('body').css('padding-top', '130px');
    });
    $('#show-search-link').click(function() {
        searchBoxHidden = false;
        $('#show-search-div').hide();
        $('#search-box-form').show();
        $('body').css('padding-top', '180px');
    });

    changeTopMargin();
    $(window).resize(function () {
        changeTopMargin();
    });
}());

// Handle selector button for search box.
(function () {
    "use strict";

    var categories = ['All', 'Classes', 'Departments', 'Colleges'];
    $('#category-selector').click(function () {
        categories.unshift(categories.pop());
        $('#category-selector').text(categories[0]);
        $('#category-input').val(categories[0]);
    });
}());

// Handle department page quick links. 
(function () {
    "use strict";
    var offset = 65;

    $('.department-quick-link').click(function(event) {
        event.preventDefault();
        $($(this).attr('href'))[0].scrollIntoView();
        scrollBy(0, -offset);
    });
}());

// Handle navbar search box.
(function () {
    "use strict";

    var processCategory = function(searchSelector, categorySelectorText, categorySelectorInput) {
        var search = $(searchSelector).val();
        search = $.trim(search);
        var mapping = [
            ['.a', 'All'],
            ['.g', 'Colleges'],
            ['.d', 'Departments'],
            ['.c', 'Classes']
        ];
        console.log(search);
        $.each(mapping, function(index, value) {
            if (search.startsWith(value[0])) {
                var category = value[1];
                $(searchSelector).val(search.substr(2));
                $(categorySelectorText).text(category);
                $(categorySelectorInput).val(category);
            } 
        });
    };

    $(document).on('click', '#magic-search-box', function () {
        $('#magic-search-box-div').addClass('open');
            var search = $(this).val();
            var category = $('#category-input').val();
            $.ajax({
                url: '/search',
                type: 'POST',
                dataType: 'script',
                data : { 
                    search : search,
                    category : category
                }
            });
    });

    var timeoutYo = null;
    $(document).on('keyup', '#magic-search-box', function () {
        window.clearTimeout(timeoutYo);
        processCategory(this, '#category-selector', '#category-input');
        var search = $(this).val();
        var category = $('#category-input').val();
        timeoutYo = window.setTimeout(function () {
            $.ajax({
                url: '/search',
                type: 'POST',
                dataType: 'script',
                data : { 
                    search : search,
                    category : category
                }
            });
        }, 150);
    });
}());

// Enable bootstrap tooltips and popovers.
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
}());

// Show/hide class descriptions checkbox
(function () {
    "use strict";
    
    var hideDescriptionSwitch = false;
    $('#toggle-class-description').click(function () {
        hideDescriptionSwitch = !hideDescriptionSwitch;
        if (hideDescriptionSwitch) {
            $('.toggle-description-target').hide();
        } else {
            $('.toggle-description-target').show();
        }
    });
}());

// Handle filters for section index.
(function () {
    "use strict";

    // Loops through elements. Hides element if it has any class 
    // in classList, else shows it. 
    var toggleSects = function (elements, classList) {
        for (var i = 0; i < elements.length; i++) {
            var element = elements[i];
            if (hasAnyClass(element, classList)) {
                element.style.display = 'none';
            } else {
                element.style.display = 'block';
            }
        }
    };

    // Returns a copy of the input array with all instances of
    // element removed.
    var removeElement = function (array, element) {
        return array === jQuery.grep(array, function(n, i) {
            return n !== element;
        });
    };

    // Returns true if the element has any class in classList. Else, 
    // returns false.
    var hasAnyClass = function (element, classList) {
        for (var i = 0; i < classList.length; i++) {
            var classes = element.getAttribute('class').split(' ');
            for (var j = 0; j < classes.length; j++) {
                if (classes[j] === classList[i]) {
                    return true;
                }
            }
        }
        return false;
    };


    // filter sections with checkboxes 
    // TODO (kvu787): Use AngularJS to replace the following
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
}());
