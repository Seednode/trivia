function toggleTheme() {
    const currentTheme = ('; '+document.cookie).split(`; colorTheme=`).pop().split(';')[0];

    if (currentTheme == "darkMode") {
        newTheme = "lightMode"
    } else {
        newTheme = "darkMode"
    }
    
    document.cookie = "colorTheme=" + newTheme + "; expires=31536000; path=/";

    location.reload()
}

document.addEventListener('DOMContentLoaded', function () {
    document.getElementById('toggle-theme')
    .addEventListener('click', toggleTheme);
})