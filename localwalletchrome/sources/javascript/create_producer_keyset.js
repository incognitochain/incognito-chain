window.onload = function () {
    var url = new URL(window.location.href);
    var accountName = url.searchParams.get("accountName");

    document.getElementById("lnk_back").href = 'account_detail.html?account=' + encodeURIComponent(accountName);

    showLoading(true);

    var xhr = new XMLHttpRequest();   // new HttpRequest instance
    xhr.open("POST", window.localStorage['cash_node_url']);
    var auth = "Basic " + $.base64.encode(window.localStorage['rpcUserName'] + ":" + window.localStorage['rpcPassword']);
    xhr.setRequestHeader("Authorization", auth);
    xhr.setRequestHeader("Content-Type", "application/json;charset=UTF-8");
    xhr.onreadystatechange = function (oEvent) {
        showLoading(false);

        if (xhr.status == 200 && xhr.readyState == XMLHttpRequest.DONE) {
            var response = JSON.parse(this.responseText.toString());
            if (response.Result != null) {
                document.getElementById("lb_ProducerKeySet").innerText = response.Result.ProducerKeySet;
                document.getElementById("lb_ProducerPublicKey").innerText = response.Result.ProducerPublicKey;
                document.getElementById("loader").style.display = "none";
                document.getElementById("myDiv").style.display = "block";
            } else {
                if (response.Error != null) {
                    alert(response.Error.message);
                } else {
                    alert('Bad response');
                }
                document.getElementById("lnk_back").click();
            }
        }
    };

    var priKey = url.searchParams.get("priKey");
    xhr.send(JSON.stringify({
        jsonrpc: "1.0",
        method: "createproducerkeyset",
        params: priKey,
        id: 1
    }));
};

function showLoading(show) {
    if (show) {
        document.getElementById("loader").style.display = "block";
        document.getElementById("myDiv").style.display = "none";
    } else {
        document.getElementById("loader").style.display = "none";
        document.getElementById("myDiv").style.display = "block";
    }
}