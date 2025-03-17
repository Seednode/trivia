function setCategories() {
    selected = document.querySelectorAll('input[type="checkbox"]:checked');
    total = document.querySelectorAll('input[type="checkbox"]');

    let json = {
        categories: [],
    };

    iter = selected.entries();

    let result = iter.next();
    while (!result.done) {
        json.categories.push(result.value[1].name);
        result = iter.next();
    }

    let xhr = new XMLHttpRequest();
    xhr.open("POST", window.location.href + "/categories", true);
    xhr.setRequestHeader("Content-Type", "application/json");
    let data = JSON.stringify({ ...json });
    xhr.send(data);

    alert("Selected " + selected.length + " out of " + total.length + " categories.")
}

document.addEventListener('DOMContentLoaded', function () {
    document.getElementById('set-categories')
    .addEventListener('click', setCategories);
});