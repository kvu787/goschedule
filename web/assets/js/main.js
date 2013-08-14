var toggleDescriptionSwitch = true;
$('#toggle-description').click(function() {
    if (toggleDescriptionSwitch) {
        $('.toggle-description-target').hide();
        $('#toggle-description').text('Show descriptions');
    } else {
        $('.toggle-description-target').show();
        $('#toggle-description').text('Hide descriptions');
    }
    toggleDescriptionSwitch = !toggleDescriptionSwitch
});