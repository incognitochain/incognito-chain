#!/bin/bash
# this is a very stupid work-around to get things work with the current pNode software
# Should not be copy to the new pNode software in the future if we decide to build one.

SERVICE="/etc/systemd/system/IncognitoUpdater.service"
TIMER="/etc/systemd/system/IncognitoUpdater.timer"
USER_NAME="nuc"
INC_HOME="/home/$USER_NAME"
DATA_DIR="$INC_HOME/aos/inco-data"
TMP="$INC_HOME/inc_node_latest_tag"
SCRIPT="$INC_HOME/run_node.sh"

if [ -f "$SERVICE" ]; then
  echo "Service is already installed "
  sudo systemctl start $(basename $SERVICE) 2> /dev/null
  exit 0
fi

# ================================================================
echo "Checking for / Installing Docker and jq (JSON processor)"
sudo apt install docker.io jq -y
sudo systemctl start docker.service
sudo systemctl stop $(basename $SERVICE) 2> /dev/null
sudo systemctl stop $(basename $TIMER) 2> /dev/null

# ================================================================
echo " # Adding $USER_NAME user to docker group"
sudo usermod -aG docker $USER_NAME || echo
rm -f $TMP && touch $TMP
chown $USER_NAME:$USER_NAME $TMP

# ================================================================
echo " # Creating systemd service to check for new release"
cat << EOF > tmp.service
[Unit]
Description = IncognitoChain Node updater
After = network.target network-online.target
Wants = network-online.target

[Service]
Type = oneshot
User = $USER_NAME
ExecStart = $SCRIPT
StandardOutput = syslog
StandardError = syslog
SyslogIdentifier = IncNodeUpdt

[Install]
WantedBy = multi-user.target
EOF
sudo mv tmp.service $SERVICE

# ================================================================
echo " # Creating timer to preodically run the update checker"
cat << EOF > tmp.timer
[Unit]
Description=Run IncognitoUpdater hourly

[Timer]
OnCalendar=hourly
Persistent=true

[Install]
WantedBy=timers.target
EOF
sudo mv tmp.timer $TIMER

echo " # Creating run node script"
cat << EOF > $SCRIPT
#!/bin/bash
TMP=$TMP
data_dir=$DATA_DIR
EOF

cat << 'EOF' >> $SCRIPT
run()
{
  rpc_port=9334
  node_port=9433
  bootnode="mainnet-bootnode.incognito.org:9330"
  validator_key=xxx
  latest_tag=$1
  current_tag=$2
  backup_log=0
  container_name=inc_mainnet

  if [ "$current_tag" != "" ]; then
	  echo "Found new docker tag, remove old one"
    docker image rm -f incognitochain/incognito-mainnet:${current_tag}
  fi

  echo "Pulling new tag: ${latest_tag}"
  docker pull incognitochain/incognito-mainnet:${latest_tag}
  echo "Create new docker network"
  docker network create --driver bridge inc_net || true
  echo "Remove old container"
  docker container stop ${container_name}
  docker container rm ${container_name}

  echo "Start the incognito mainnet docker container"
  set -x
  docker run --restart=always --net inc_net \
    -p $node_port:$node_port -p $rpc_port:$rpc_port \
    -e NODE_PORT=$node_port -e RPC_PORT=$rpc_port -e BOOTNODE_IP=$bootnode \
    -e FULLNODE="" -e MININGKEY=${validator_key} -e TESTNET=false -e LIMIT_FEE=1 \
    -e INDEXER_ACCESS_TOKEN=$coin_index_access_token -e NUM_INDEXER_WORKERS=$num_index_worker \
    -v ${data_dir}:/data -d --name ${container_name} incognitochain/incognito-mainnet:${latest_tag}
  set +x
}

current_latest_tag=$(cat $TMP)
echo "Getting Incognito docker tags"

tags=$(curl -s -X GET https://hub.docker.com/v1/repositories/incognitochain/incognito-mainnet/tags | jq ".[].name")
tags=${tags//\"/}
sorted_tags=($(echo ${tags[*]}| tr " " "\n" | sort -rn))
latest_tag=${sorted_tags[0]}
echo "Latest tag is ${latest_tag}"

if [ "$current_latest_tag" != "$latest_tag" ]; then
	echo "Found newer tag , run it!"
	run $latest_tag $current_latest_tag
	current_latest_tag=$latest_tag
	echo $current_latest_tag > $TMP
else
    echo "Runing latest tag already, no need to update"
fi
EOF

# ================================================================
chmod +x $SCRIPT

echo " # Enabling service"
sudo systemctl daemon-reload
sudo systemctl enable $(basename $SERVICE)
sudo systemctl enable $(basename $TIMER)


echo " # Starting service. Please wait..."
sudo systemctl start $(basename $SERVICE)
sudo systemctl start $(basename $TIMER)
