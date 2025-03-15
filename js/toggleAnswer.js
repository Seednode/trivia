function toggleAnswer() {
    var x = document.getElementById("answer");
    if (x.style.display === "block") {
        x.style.display = "none";
    } else {
        x.style.display = "block";
    }
}

document.addEventListener('DOMContentLoaded', function () {
    document.getElementById('toggle-answer')
    .addEventListener('click', toggleAnswer);
});