window.onload = function () {
    loadList();

    document.getElementById("bt_add").onclick = function () {
        addNode();
    };

    showLoading(false)
    $('#rpcUsername').val(window.localStorage['rpcUserName'])
    $('#rpcPassword').val(window.localStorage['rpcPassword'])
    $('#saveRpcAuth').click(function () {
        var rpcUserName = $('#rpcUsername').val();
        var rpcPassword = $('#rpcPassword').val();

        window.localStorage['rpcUserName'] = rpcUserName;
        window.localStorage['rpcPassword'] = rpcPassword;
    });
};

function loadList() {
    var cashNodeUrls = JSON.parse(window.localStorage['cash_node_urls']);
    var cashNodeUrl = window.localStorage['cash_node_url']
    removeChilds('list_account')
    for (var i = 0; i < cashNodeUrls.length; i++) {
        var nodeUrl = cashNodeUrls[i];
        var li = document.createElement('li');
        if (nodeUrl != cashNodeUrl) {
            li.innerHTML = '<a style="font-weight: 300;">' + nodeUrl + '</a>' + '<button data="' + nodeUrl + '" id="bt_select_' + i + '" style="margin-left: 20px" class="btn btn-primary" onclick="selectNode(nodeUrl)">Select</button>';
        } else {
            li.innerHTML = '<a style="font-weight: 500; font-style: italic;">' + nodeUrl + '</a>' + '<button id="bt_select_' + i + '" style="margin-left: 20px" class="btn btn-primary">Select</button>';
        }
        li.classList = "list-group-item"
        document.getElementById("list_account").appendChild(li);

        $("#bt_select_" + i).click(function () {
            selectNode($(this).attr('data'));
            return false;
        });
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
    var nodeUrl = document.getElementById("txt_nodeUrl").value;
    var cashNodeUrls = JSON.parse(window.localStorage['cash_node_urls']);
    cashNodeUrls.push(nodeUrl);
    window.localStorage['cash_node_urls'] = JSON.stringify(cashNodeUrls);

    loadList();
}

function selectNode(nodeUrl) {
    alert(nodeUrl)
    window.localStorage['cash_passphrase'] = '';
    window.localStorage['cash_node_url'] = nodeUrl;
    window.location.href = '../../index.html';

    return false;
}
