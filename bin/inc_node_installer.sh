#!/bin/bash

# USAGE: ###############################################################
#   1. Edit following config as you wish:
#		BOOTNODE
#		PORT_RPC
#		PORT_NODE
#		VALIDATOR_K
#		GETH_NAME
#		CHECK_INTERVAL
#       FULLNODE
#	2. To install, reinstall, make changes, run:
#		sudo ./{this script}
#	3. To uninstall, run:
#		sudo ./{this script} uninstall
########################################################################

# check super user
if [ $(whoami) != root ]; then
cat << EOF
	!!! Please run with sudo or su, otherwise it won't work
	!!! Script now exits
EOF
	exit 1
fi

# change config here:
BOOTNODE="mainnet-bootnode.incognito.org:9330"  # this should be left as default

VALIDATOR_K="validator key 1,
             validator key 2,
             validator key 3"  
  # mandatory if you want to run a validator node, multiple validator keys must be separate by commas

GETH_NAME="https://mainnet.infura.io/v3/xxxyyy"  
  # infura link (*) (follow step 3 on this thread to setup infura https://we.incognito.org/t/194)

PORT_RPC="8334"  # change this if you prefer other port, 
  # the first node will use the port the next one use this port+1 and so on

PORT_NODE="9433"  # change this if you prefer other port,
  # the first node will use the port the next one use this port+1 and so on

CHECK_INTERVAL="3600" # 1 hour, interval to checking for new node update

FULLNODE=""  # set to 1 to run as a full node, empty to run as normal node

# New parameters since privacy v2
INDEXER_ACCESS_TOKEN="edeaaff3f1774ad2888673770c6d64097e391bc362d7d6fb34982ddf0efd18cb" 
  # can be generate by running: $ echo "bla bla bla" | sha256sum
NUM_INDEXER_WORKERS=100

# Do not edit lines below unless you know what you're doing
USER="incognito"
SCRIPT="/bin/run_node.sh"
SERVICE="/etc/systemd/system/IncognitoUpdater.service"
DATA_DIR="/home/$USER/node_data"
TMP="/home/$USER/inc_node_latest_tag"
systemctl stop $(basename $SERVICE) 2> /dev/null

function uninstall {
	echo " # Remove update service"
	systemctl disable $(basename $SERVICE)
	systemctl daemon-reload
	echo " # Stop and remove docker images + containers"
	docker container stop $(docker container ls -aqf name=inc_mainnet)
  docker container rm $(docker container ls -aqf name=inc_mainnet)
  docker image rm -f $(docker images -q incognitochain/incognito-mainnet)
	echo " # Clearing node's data"
	rm -Rf $TMP ${DATA_DIR}* $SCRIPT $SERVICE
	echo " # Removing user"
	deluser $USER
}

if [[ $1 = "uninstall" ]]; then
cat << EOF 
!!!===============================================!!!
###                    WARNING                    ###
!!!===============================================!!!
!   Uninstalling and cleanup !
!   This action will remove:
!      - The systemd service: $SERVICE
!      - Docker containers and images
!      - User: $USER 
!      - Node's data: $DATA_DIR
!      - Run script: $SCRIPT
!!! Do you really want to do this? (N/y)
EOF
	read consent
	if [ -z $consent ] || [[ ${consent,,} = "n" ]] || [[ ${consent,,} != "y" ]] ; then
		echo "!!! Good choice !!!"
		exit 1
	else
cat << EOF
#####################################################
#          Too bad! So sad! See you again!          #
#####################################################
EOF
		uninstall
	fi
	exit 1
fi

echo " # Creating incognito user to run node"
useradd $USER
usermod -aG docker ${USER} || echo
mkdir -p $DATA_DIR
chown -R $USER:$USER $DATA_DIR
rm -f $TMP && touch $TMP
chown $USER:$USER $TMP

echo " # Creating systemd service to check for new release"
KILL=$(which pkill)
cat << EOF > $SERVICE
[Unit]
Description = IncognitoChain Node updater
After = network.target network-online.target
Wants = network-online.target

[Service]
Type = simple
User = $USER
ExecStart = $SCRIPT
ExecStop = $KILL $(basename $SCRIPT)
Restart = on-failure
RestartSec = $CHECK_INTERVAL
StartLimitInterval = 60
StartLimitBurst = 60
StandardOutput = syslog
StandardError = syslog
SyslogIdentifier = IncNodeUpdt

[Install]
WantedBy = multi-user.target
EOF

echo " # Creating run node script"
cat << EOF > $SCRIPT
#!/bin/bash
TMP=$TMP
run()
{
  validator_key=(${VALIDATOR_K//,/ })
  bootnode=$BOOTNODE
  data_dir=$DATA_DIR
  rpc_port=$PORT_RPC
  node_port=$PORT_NODE
  geth_name=$GETH_NAME
  geth_port=""
  geth_proto=""
  fullnode=$FULLNODE
  coin_index_access_token=$INDEXER_ACCESS_TOKEN
  num_index_worker=$NUM_INDEXER_WORKERS

EOF
cat << 'EOF' >> $SCRIPT

  latest_tag=$1
  current_tag=$2
  backup_log=0
  container_name=inc_mainnet
  count=0

  if [ -z "$node_port" ]; then
    node_port="9433";
  fi
  if [ -z "$rpc_port" ]; then
    rpc_port="9334";
  fi

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
    docker container rm ${container_name}_${count}

    echo "Start the incognito mainnet docker container"
    set -x
    docker run --restart=always --net inc_net \
      -p $node_port:$node_port -p $rpc_port:$rpc_port \
      -e NODE_PORT=$node_port -e RPC_PORT=$rpc_port -e BOOTNODE_IP=$bootnode \
      -e GETH_NAME=$geth_name -e GETH_PROTOCOL=$geth_proto -e GETH_PORT=$geth_port \
      -e FULLNODE=$fullnode -e MININGKEY=${key} -e TESTNET=false -e LIMIT_FEE=1 \
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
tags=$(curl -s -X GET https://hub.docker.com/v1/repositories/incognitochain/incognito-mainnet/tags  \
	| sed -e 's/[][]//g' -e 's/"//g' -e 's/ //g' | tr '}' '\n'  | awk -F: '{print $3}' | sed -e 's/\n/;/g')
sorted_tags=($(echo ${tags[*]}| tr " " "\n" | sort -rn))
latest_tag=${sorted_tags[0]}
echo "Latest tag is ${latest_tag}"

if [ "$current_latest_tag" != "$latest_tag" ]; then
	echo "Found newer tag , run it!"
	run $latest_tag $current_latest_tag
	current_latest_tag=$latest_tag
	echo $current_latest_tag > $TMP
fi
EOF

chmod +x $SCRIPT

echo " # Enabling service"
systemctl daemon-reload
systemctl enable $(basename $SERVICE)

echo " # Starting service"
systemctl start $(basename $SERVICE)
cat << EOF
 # DONE.
 You can check the installing and starting progress or the service by running:
    $ journalctl | grep Inc
    or 
    $ journalctl t IncNodeUpdt

EOF