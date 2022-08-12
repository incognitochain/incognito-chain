#!/bin/bash

# ============================================================================================
# Do not edit lines below unless you know what you're doing
BOOTNODE="mainnet-bootnode.incognito.org:9330"  # this should be left as default
SERVICE="/etc/systemd/system/IncognitoUpdater.service"
TIMER="/etc/systemd/system/IncognitoUpdater.timer"
USER_NAME="nuc"
INC_HOME="/home/$USER_NAME"
DATA_DIR="$INC_HOME/aos/inco-data"
TMP="$INC_HOME/inc_node_latest_tag"
SCRIPT="$INC_HOME/run_node.sh"


# =========================== check super user
if [ $(whoami) != root ]; then
sudo -i
fi

# ================================================================
echo "Checking for / Installing Docker and jq (JSON processor)"
apt install docker.io jq -y
systemctl start docker.service
systemctl stop $(basename $SERVICE) 2> /dev/null
systemctl stop $(basename $TIMER) 2> /dev/null

# ================================================================
echo " # Adding $USER_NAME user to docker group"
usermod -aG docker $USER_NAME || echo
rm -f $TMP && touch $TMP
chown $USER_NAME:$USER_NAME $TMP

# ================================================================
echo " # Creating systemd service to check for new release"
cat << EOF > $SERVICE
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

# ================================================================
echo " # Creating timer to preodically run the update checker"
cat << EOF > $TIMER
[Unit]
Description=Run IncognitoUpdater hourly

[Timer]
OnCalendar=hourly
Persistent=true

[Install]
WantedBy=timers.target
EOF

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
  for key in ${validator_key[@]}; do
    echo "Remove old container"
    docker container stop ${container_name}_${count}
    docker container rm ${container_name}_${count}

    echo "Start the incognito mainnet docker container"
    set -x
    docker run --restart=always --net inc_net \
      -p $node_port:$node_port -p $rpc_port:$rpc_port \
      -e NODE_PORT=$node_port -e RPC_PORT=$rpc_port -e BOOTNODE_IP=$bootnode \
      -e FULLNODE="" -e MININGKEY=${key} -e TESTNET=false -e LIMIT_FEE=1 \
      -e INDEXER_ACCESS_TOKEN=$coin_index_access_token -e NUM_INDEXER_WORKERS=$num_index_worker \
      -v ${data_dir}_${count}:/data -d --name ${container_name}_${count} incognitochain/incognito-mainnet:${latest_tag}
    set +x
    ((node_port++))
    ((rpc_port++))
    ((count++))
  done
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
systemctl daemon-reload
systemctl enable $(basename $SERVICE)
systemctl enable $(basename $TIMER)


echo " # Starting service. Please wait..."
systemctl start $(basename $SERVICE)
systemctl start $(basename $TIMER)
cat << EOF
 # ============================= DONE ================================
 To check the installing and starting progress or the running service:
    $ journalctl | grep Inc
    or
    $ journalctl -t IncNodeUpdt
EOF
