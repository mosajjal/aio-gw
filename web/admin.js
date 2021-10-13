// auto switch dark mode if the browser is on dark mode
if (window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches) {
    halfmoon.toggleDarkMode();
    document.getElementById('darkmode').checked = true;
}

// fetch data inside the forms
function getSettings() {
    var data = {};
    var inputs = $.querySelectorAll('input, select, textarea');
    for (var i = 0; i < inputs.length; i++) {
        var input = inputs[i];
        if (input.name) {
            data[input.name] = input.value;
        }
    }
    return data;
}


// toggle a form's editablility 

// post data of a form to the server