#!/usr/bin/env bash
# Default params
#host="127.0.0.1:9334"
host="172.105.115.134:20000"

if [[ -n "$2" ]]; then
	host="$2"
fi

getbeaconbeststate=`cat << EOS
{
    "jsonrpc": "2.0",
    "method": "getbeaconbeststate",
    "params":{},
    "id": "0"
}
EOS`

check_error () {
    data=`echo $@ | cut -d ' ' -f 2-`
    name=`echo $@ | cut -d ' ' -f 1`
    error=`echo $data| python -c "import sys, json; print json.load(sys.stdin)['Error']"`
    if [[ "$error" == "None" ]]; then
	    return
    else
        echo $name
        echo $error
    fi
}

status_report (){
    echo "Checking node status ${host} ..."
    head_request="curl -s  -X POST --data"

    # Check beacon chain


    data=`${head_request} "${getbeaconbeststate}" http://${host}`
    if check_error "getbeaconbeststate" "$data"; then

        echo "===================== Beacon Status ====================="
        
        beaconHeight=`echo $data| python -c "import sys, json; print json.load(sys.stdin)['Result']['BeaconHeight']"`
        echo Beacon Height $beaconHeight
        echo Shard Height
        echo $data| python -c "import sys, json; print json.dumps( json.load(sys.stdin)['Result']['BestShardHeight'], sort_keys=True, indent=4); "

        echo "===================== Shard Status ====================="
    fi



}

status_report

