if (localStorage.getItem('mode') === "light") {
    var r = document.querySelector(':root');

    r.style.setProperty('--foreground', '#0d1117')
    r.style.setProperty('--background', '#c9d1d9')
}

function toggleDarkMode () {
    var r = document.querySelector(':root');

    if (localStorage.getItem('mode') === "light") {
        r.style.setProperty('--foreground', '#c9d1d9')
        r.style.setProperty('--background', '#0d1117')
        localStorage.setItem('mode', "dark");
    } else {
        r.style.setProperty('--foreground', '#0d1117')
        r.style.setProperty('--background', '#c9d1d9')
        localStorage.setItem('mode', "light");
    }
}

document.addEventListener('DOMContentLoaded', function () {
    document.getElementById('dark-mode')
    .addEventListener('click', toggleDarkMode);
})