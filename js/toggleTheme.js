function toggleTheme() {
    let xhr = new XMLHttpRequest();
    xhr.open("POST", window.location.href + "/theme/" + document.querySelector('input[name="theme"]:checked').value, true);
    xhr.setRequestHeader("Content-Type", "application/json");
    xhr.send("");

    handleHardReload(window.location.href);
}

async function handleHardReload(url) {
    await fetch(url, {
        headers: {
            Pragma: 'no-cache',
            Expires: '-1',
            'Cache-Control': 'no-cache',
        },
    });
    window.location.href = url;
    
    window.location.reload();
}

document.addEventListener('DOMContentLoaded', function () {
    document.getElementById('set-theme')
    .addEventListener('click', toggleTheme);
})

function getCookie(name) {
    const value = `; ${document.cookie}`;
    const parts = value.split(`; ${name}=`);
    if (parts.length === 2) return parts.pop().split(';').shift();
}

var colorTheme = getCookie("colorTheme");
if (colorTheme == "lightMode") {
    document.getElementById('light-mode').checked = true;
} else {
    document.getElementById('dark-mode').checked = true;
}