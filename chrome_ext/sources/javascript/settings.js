window.onload = function () {
    if (typeof(Storage) !== "undefined") {
        // Code for localStorage/sessionStorage.
        let passphrase = window.localStorage['cash_passphrase'];
        if (passphrase == null || passphrase == '') {
            window.location.href = "../../index.html";
        }
    } else {
        // Sorry! No Web Storage support..
        alert('Sorry! No Web Storage support')
        return
    }

    loadList();

    document.getElementById("bt_add").onclick = function () {
        addNode();
    };

    showLoading(false)
};

function loadList() {
    var cashNodeUrls = JSON.parse(window.localStorage['cash_node_urls']);

    removeChilds('list_account')
    for (var i = 0; i < cashNodeUrls.length; i++) {
        let nodeUrl = cashNodeUrls[i];
        let li = document.createElement('li');
        li.innerHTML = '<a>' + nodeUrl + '</a>' + '<button id="bt_select_' + i + '" style="margin-left: 20px" class="btn btn-primary">Select</button>';
        li.classList = "list-group-item"
        document.getElementById("list_account").appendChild(li);

        document.getElementById("bt_select_" + i).onclick = function (ev) {
            selectNode(nodeUrl);
            return false;
        }
    }
}

function showLoading(show) {
    if (show) {
        document.getElementById("loader").style.display = "block";
        document.getElementById("myDiv").style.display = "none";
    } else {
        document.getElementById("loader").style.display = "none";
        document.getElementById("myDiv").style.display = "block";
    }
}

function addNode() {
    let nodeUrl = document.getElementById("txt_nodeUrl").value;
    let cashNodeUrls = JSON.parse(window.localStorage['cash_node_urls']);
    cashNodeUrls.push(nodeUrl);
    window.localStorage['cash_node_urls'] = JSON.stringify(cashNodeUrls);

    loadList();
}

function selectNode(nodeUrl) {
    window.localStorage['cash_passphrase'] = '';
    window.localStorage['cash_node_url'] = nodeUrl;
    window.location.href = '../../index.html';

    return false;
}
