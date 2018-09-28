function numberWithCommas(x) {
    return x.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ",");
}

window.onload = function () {
    var url = new URL(window.location.href);
    var accountName = url.searchParams.get("accountName");

    document.getElementById("lnk_back").href = 'account_detail.html?account=' + encodeURIComponent(accountName);

    var xhr = new XMLHttpRequest();   // new HttpRequest instance
    xhr.open("POST", api_url);
    xhr.setRequestHeader("Content-Type", "application/json;charset=UTF-8");
    xhr.onreadystatechange = function (oEvent) {
        if (xhr.status == 200 && xhr.readyState == XMLHttpRequest.DONE) {
            var response = JSON.parse(this.responseText.toString());
            if (response.Result != null) {
                document.getElementById("lb_SealerKeySet").innerText = response.Result.SealerKeySet;
                document.getElementById("lb_SealerPublicKey").innerText = response.Result.SealerPublicKey;
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

    var priKey = url.searchParams.get("priKey");
    xhr.send(JSON.stringify({
        jsonrpc: "1.0",
        method: "createsealerkeyset",
        params: priKey,
        id: 1
    }));
};