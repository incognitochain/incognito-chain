window.onload = function () {
    document.getElementById("bt_save").onclick = function () {
        var passphrase = document.getElementById('txt_passphrase').value
        if (typeof(Storage) !== "undefined") {
            // Code for localStorage/sessionStorage.
            window.localStorage['cash_passphrase'] = passphrase;
            window.location.href = 'index.html'
        } else {
            // Sorry! No Web Storage support..
            alert('Sorry! No Web Storage support')
        }
    };
}