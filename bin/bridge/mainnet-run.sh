#!/bin/sh bash

run()
{
  validator_key=xxx
  bootnode="mainnet-bootnode.incognito.org:9330"

  latest_tag=$1
  current_tag=$2
  data_dir="data"
  backup_log=0

  if [ -z "$node_port" ]; then
    node_port="9433";
  fi
  if [ -z "$rpc_port" ]; then
    rpc_port="9334";
  fi

  docker -v || bash -c "wget -qO- https://get.docker.com/ | sh"
  # Remove old container 
  docker rm -f inc_mainnet
  
  if [ "$current_tag" != "" ]; then
    docker image rm -f incognitochain/incognito-mainnet:${current_tag}
  fi

  docker pull incognitochain/incognito-mainnet:${latest_tag}
  docker network create --driver bridge inc_net || true
  # Start the incognito mainnet docker container
  docker run --restart=always --net inc_net -p $node_port:$node_port -p $rpc_port:$rpc_port -e NODE_PORT=$node_port -e RPC_PORT=$rpc_port -e BOOTNODE_IP=$bootnode -e GETH_NAME=mainnet.infura.io/v3/YYYZZZ -e GETH_PROTOCOL=https -e GETH_PORT= -e MININGKEY=${validator_key} -e TESTNET=false -e LIMIT_FEE=1 -v $PWD/${data_dir}:/data -d --name inc_mainnet incognitochain/incognito-mainnet:${latest_tag}

  if [ $backup_log -eq 1 ]; then
    mv $data_dir/log.txt $data_dir/log_$(date "+%Y%m%d_%H%M%S").txt
    mv $data_dir/error_log.txt $data_dir/error_log_$(date "+%Y%m%d_%H%M%S").txt
  fi
}

# kill existing run.sh processes
ps aux | grep $(basename $0) | awk '{ print $2}' | grep -v "^$$\$" | xargs kill -9

current_latest_tag=""
while [ 1 = 1 ]; do
  tags=`curl -X GET https://registry.hub.docker.com/v1/repositories/incognitochain/incognito-mainnet/tags  | sed -e 's/[][]//g' -e 's/"//g' -e 's/ //g' | tr '}' '\n'  | awk -F: '{print $3}' | sed -e 's/\n/;/g'`

  sorted_tags=($(echo ${tags[*]}| tr " " "\n" | sort -rn))
  latest_tag=${sorted_tags[0]}

  if [ "$current_latest_tag" != "$latest_tag" ]; then
    run $latest_tag $current_latest_tag
    current_latest_tag=$latest_tag
  fi

  sleep 1h

done &
