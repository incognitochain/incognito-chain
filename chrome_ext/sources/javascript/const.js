var api_url = 'http://127.0.0.1:9334'
// var api_url = 'http://dev.autonomousadmin.com/'
// var api_url = 'http://localhost:5000/'

function numberWithCommas(x) {
    return x.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ",");
}

function removeChilds(id) {
    var elm = document.getElementById(id);
    while (elm != null && elm.hasChildNodes()) {
        elm.removeChild(elm.lastChild);
    }
}