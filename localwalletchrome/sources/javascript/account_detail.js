window.onload = function () {
    var xhr = new XMLHttpRequest();   // new HttpRequest instance
    xhr.open("POST", window.localStorage['cash_node_url']);
    xhr.setRequestHeader("Content-Type", "application/json;charset=UTF-8");
    var auth = "Basic " + $.base64.encode(window.localStorage['rpcUserName'] + ":" + window.localStorage['rpcPassword']);
    xhr.setRequestHeader("Authorization", auth);
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
    let url = new URL(window.location.href);
    let account = url.searchParams.get("account");
    xhr.send(JSON.stringify({
        jsonrpc: "1.0",
        method: "getaccountaddress",
        params: account,
        id: 1
    }));

    document.getElementById("bt_send").onclick = function () {
        sendmany()
        return false;
    };
    document.getElementById("bt_send_many").onclick = function () {
        window.location.href = 'sendmany.html?account=' + encodeURIComponent(account);
        return false;
    };

    $('#bt_register').click(function () {
        var r = confirm();
        if (r) {
            sendRegistrationCandidate();
        }
        return false;
    });

    showNetworkInfo();
};

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
    var priKey = document.getElementById("lb_privateKey").innerText;
    var pubKey = document.getElementById("txt_address").value;
    var amount = document.getElementById("txt_amount").value;

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
                alert("Success");
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
    dest[pubKey] = parseInt(amount);
    xhr.send(JSON.stringify({
        jsonrpc: "1.0",
        method: "sendmany",
        params: [priKey, dest, -1, 1],
        id: 1
    }));
}

function sendRegistrationCandidate() {
    var priKey = $('#lb_privateKey').text();
    var candidateFee = parseInt($('#candidateFee').val());
    var candidatePeerInfo = $('#listAddresses input:checked').val();
    var auth = "Basic " + $.base64.encode(window.localStorage['rpcUserName'] + ":" + window.localStorage['rpcPassword']);
    $.ajax({
        url: window.localStorage['cash_node_url'],
        type: 'POST',
        data: JSON.stringify({
            jsonrpc: "1.0",
            method: "sendregistration",
            params: [priKey, candidateFee, -1, 1, candidatePeerInfo],
            id: 1
        }),
        beforeSend: function (xhr) {
            xhr.setRequestHeader('Authorization', auth);
            xhr.setRequestHeader("Content-Type", "application/json;charset=UTF-8");
        },
        success: function (response) {
            if (response.Result != null && response.Result != '') {
                alert("Success");
            } else {
                if (response.Error != null) {
                    alert(response.Error.message)
                } else {
                    alert('Bad response');
                }
            }
        }
    });
}

function showNetworkInfo() {
    var auth = "Basic " + $.base64.encode(window.localStorage['rpcUserName'] + ":" + window.localStorage['rpcPassword']);
    $.ajax({
        url: window.localStorage['cash_node_url'],
        type: 'POST',
        data: JSON.stringify({
            jsonrpc: "1.0",
            method: "getnetworkinfo",
            params: [],
            id: 1
        }),
        beforeSend: function (xhr) {
            xhr.setRequestHeader('Authorization', auth);
            xhr.setRequestHeader("Content-Type", "application/json;charset=UTF-8");
        },
        success: function (response) {
            if (response.Result != null) {
                $('#listAddresses').html('');
                response.Result.LocalAddresses.forEach(function (value) {
                    $('#listAddresses').append('<li style="margin-right: 5px;"><input value="' + value + '" type="radio" name="lcAddresses"/>' + value + '</li>')
                });
            }
            else {
                if (response.Error != null) {
                    alert(response.Error.message)
                } else {
                    alert('Bad response');
                }
            }
        }
    });
}