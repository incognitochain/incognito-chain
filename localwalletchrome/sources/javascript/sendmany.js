window.onload = function () {
    var xhr = new XMLHttpRequest();   // new HttpRequest instance
    xhr.open("POST", window.localStorage['cash_node_url']);
    var auth = "Basic " + $.base64.encode(window.localStorage['rpcUserName'] + ":" + window.localStorage['rpcPassword']);
    xhr.setRequestHeader("Authorization", auth);
    xhr.setRequestHeader("Content-Type", "application/json;charset=UTF-8");
    xhr.onreadystatechange = function (oEvent) {
        if (xhr.status == 200 && xhr.readyState == XMLHttpRequest.DONE) {
            var response = JSON.parse(this.responseText.toString());
            if (response.Result != null) {
                document.getElementById("lb_publicKey").innerText = response.Result.PublicKey;
                document.getElementById("lb_readonlyKey").innerText = response.Result.ReadonlyKey;
                document.getElementById("loader").style.display = "none";
                document.getElementById("myDiv").style.display = "block";
                dumpprivkey(response.Result.PublicKey)
                getbalance();
            } else {
                if (response.Error != null) {
                    alert(response.Error.message);
                } else {
                    alert('Bad response');
                }
            }
        }
    };
    var url = new URL(window.location.href);
    var account = url.searchParams.get("account");
    xhr.send(JSON.stringify({
        jsonrpc: "1.0",
        method: "getaccountaddress",
        params: account,
        id: 1
    }));

    document.getElementById("bt_send").onclick = function () {
        sendmany();
        return false;
    };

    addNewRow();
};

function addNewRow() {
    let tbSend = document.getElementById("tb_send");
    let index = tbSend.getElementsByTagName("tr").length;
    let row = tbSend.insertRow(index);
    let cell1 = row.insertCell(0);
    let cell2 = row.insertCell(1);
    let cell3 = row.insertCell(2);
    cell1.innerHTML = '<td><input id="txt_address" class="form-control txt_address"></input></td>';
    cell2.innerHTML = '<td><input id="txt_amount" class="form-control txt_amount"></input></td>';
    cell3.innerHTML = '';

    let txtAmount = tbSend.getElementsByTagName("tr")[index].getElementsByClassName('txt_amount')[0];
    txtAmount.addEventListener('keypress', function (e) {
        if (e.keyCode == 13) {
            let length = tbSend.getElementsByTagName("tr").length;
            if (this.parentNode.parentNode.rowIndex == length - 1) {
                tbSend.rows[length - 1].cells[2].innerHTML = '<td><button id="bt_remove" type="submit" class="btn btn-primary">Remove</button></td>';

                let btSend = tbSend.getElementsByTagName("tr")[length - 1].getElementsByTagName('button')[0];
                btSend.onclick = function (ev) {
                    tbSend.deleteRow(this.parentNode.parentNode.rowIndex);
                    return false;
                };

                addNewRow();

                saveData();
            }
            return false;
        }
    });

    tbSend.getElementsByTagName("tr")[index].getElementsByClassName('txt_address')[0].focus();

    saveData();
}

function saveData() {

}

function dumpprivkey(publicKey) {
    var xhr = new XMLHttpRequest();   // new HttpRequest instance
    xhr.open("POST", window.localStorage['cash_node_url']);
    xhr.setRequestHeader("Content-Type", "application/json;charset=UTF-8");
    var auth = "Basic " + $.base64.encode(window.localStorage['rpcUserName'] + ":" + window.localStorage['rpcPassword']);
    xhr.setRequestHeader("Authorization", auth);
    xhr.onreadystatechange = function (oEvent) {
        if (xhr.status == 200 && xhr.readyState == XMLHttpRequest.DONE) {
            var response = JSON.parse(this.responseText.toString());
            if (response.Result != null) {
                document.getElementById("lb_privateKey").innerText = response.Result.PrivateKey;
                var url = new URL(window.location.href);
                var account = url.searchParams.get("account");
                document.getElementById("lnk_createsealerkeyset").href = 'create_sealer_keyset.html?priKey=' + encodeURIComponent(response.Result.PrivateKey) + '&accountName=' + encodeURIComponent(account);
            } else {
                if (response.Error != null) {
                    alert(response.Error.message);
                } else {
                    alert('Bad response');
                }
            }
        }
    };
    xhr.send(JSON.stringify({
        jsonrpc: "1.0",
        method: "dumpprivkey",
        params: publicKey,
        id: 1
    }));
}

function getbalance() {
    var url = new URL(window.location.href);
    var account = url.searchParams.get("account");
    var passphrase = window.localStorage['cash_passphrase'];

    var xhr = new XMLHttpRequest();   // new HttpRequest instance
    xhr.open("POST", window.localStorage['cash_node_url']);
    xhr.setRequestHeader("Content-Type", "application/json;charset=UTF-8");
    var auth = "Basic " + $.base64.encode(window.localStorage['rpcUserName'] + ":" + window.localStorage['rpcPassword']);
    xhr.setRequestHeader("Authorization", auth);
    xhr.onreadystatechange = function (oEvent) {
        if (xhr.status == 200 && xhr.readyState == XMLHttpRequest.DONE) {
            var response = JSON.parse(this.responseText.toString());
            if (response.Result != null) {
                document.getElementById("lb_balance").innerText = numberWithCommas(response.Result);
            } else {
                if (response.Error != null) {
                    alert(response.Error.message);
                } else {
                    alert('Bad response');
                }
            }
        }
    };
    xhr.send(JSON.stringify({
        jsonrpc: "1.0",
        method: "getbalance",
        params: [account, 1, passphrase],
        id: 1
    }));
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

function sendmany() {
    let priKey = document.getElementById("lb_privateKey").innerText;

    showLoading(true);

    var xhr = new XMLHttpRequest();   // new HttpRequest instance
    xhr.open("POST", window.localStorage['cash_node_url']);
    xhr.setRequestHeader("Content-Type", "application/json;charset=UTF-8");
    var auth = "Basic " + $.base64.encode(window.localStorage['rpcUserName'] + ":" + window.localStorage['rpcPassword']);
    xhr.setRequestHeader("Authorization", auth);
    xhr.onreadystatechange = function (oEvent) {
        showLoading(false);

        if (xhr.status == 200 && xhr.readyState == XMLHttpRequest.DONE) {
            var response = JSON.parse(this.responseText.toString());
            if (response.Result != null && response.Result != '') {
            } else {
                if (response.Error != null) {
                    alert(response.Error.message)
                } else {
                    alert('Bad response');
                }
            }
        }
    };
    var dest = {};

    let txtAddresses = document.getElementsByClassName('txt_address');
    let txtAmounts = document.getElementsByClassName('txt_amount');
    for (let i = 0; i < txtAddresses.length; i++) {
        let pubKey = txtAddresses[i].value;
        let amount = txtAmounts[i].value;

        dest[pubKey] = parseInt(amount);
    }

    xhr.send(JSON.stringify({
        jsonrpc: "1.0",
        method: "sendmany",
        params: [priKey, dest, -1, 1],
        id: 1
    }));
}