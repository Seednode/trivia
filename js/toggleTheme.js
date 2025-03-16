function toggleTheme() {
    const currentTheme = ('; '+document.cookie).split(`; colorTheme=`).pop().split(';')[0];

    if (currentTheme == "lightMode") {
        newTheme = "darkMode"
    } else {
        newTheme = "lightMode"
    }
    
    document.cookie = "colorTheme=" + newTheme + "; expires=31536000; path=/";

    location.reload()
}

document.addEventListener('DOMContentLoaded', function () {
    document.getElementById('toggle-theme')
    .addEventListener('click', toggleTheme);
})