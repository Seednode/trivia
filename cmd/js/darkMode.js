if (localStorage.getItem('mode') === "light") {
    var r = document.querySelector(':root');

    r.style.setProperty('--foreground', '#0d1117')
    r.style.setProperty('--background', '#c9d1d9')
}

function toggleDarkMode () {
    var r = document.querySelector(':root');
    var rs = getComputedStyle(r);

    if (rs.getPropertyValue('--foreground') === "#c9d1d9") {
        r.style.setProperty('--foreground', '#0d1117')
        r.style.setProperty('--background', '#c9d1d9')
    } else {
        r.style.setProperty('--foreground', '#c9d1d9')
        r.style.setProperty('--background', '#0d1117')
    }

    if (localStorage.getItem('mode') === "dark") {
        localStorage.setItem('mode', "light");
    } else {
        localStorage.setItem('mode', "dark");
    }
}

document.addEventListener('DOMContentLoaded', function () {
    document.getElementById('dark-mode')
    .addEventListener('click', toggleDarkMode);
})