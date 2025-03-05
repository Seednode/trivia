if (localStorage.getItem('mode') === "light") {
    var r = document.querySelector(':root');

    r.style.setProperty('--background', '#fdf6e3')
    r.style.setProperty('--highlight', '#eee8d5')
    r.style.setProperty('--comment', '#93a1a1')
    r.style.setProperty('--content', '#657b83')
    r.style.setProperty('--emphasis', '#586e75')
}

function toggleDarkMode () {
    var r = document.querySelector(':root');

    if (localStorage.getItem('mode') === "light") {
        r.style.setProperty('--background', '#002b36')
        r.style.setProperty('--highlight', '#073642')
        r.style.setProperty('--comment', '#586e75')
        r.style.setProperty('--content', '#839496')
        r.style.setProperty('--emphasis', '#93a1a1')

        localStorage.setItem('mode', "dark");
    } else {
        r.style.setProperty('--background', '#fdf6e3')
        r.style.setProperty('--highlight', '#eee8d5')
        r.style.setProperty('--comment', '#93a1a1')
        r.style.setProperty('--content', '#657b83')
        r.style.setProperty('--emphasis', '#586e75')

        localStorage.setItem('mode', "light");
    }
}

document.addEventListener('DOMContentLoaded', function () {
    document.getElementById('dark-mode')
    .addEventListener('click', toggleDarkMode);
})