#!/bin/bash
# ver: 22.10
# ============================================================================================
# 1. Interactive mode: just run following command then follow the steps.
#       sudo ./{this script}
#	2. Preconfig mode: change the below configs, then run:
#    		sudo ./{this script} -y
#	3. To uninstall, run:
#		    sudo ./{this script} -u
# 4. If you want to add more node after running this script:
#      4.1. Open /home/incognito/validator_keys, append more keys to it.
#          (separate by commas, no spaces)
#      4.2. Delete /home/incognito/inc_node_latest_tag
#      4.3. Start IncognitoUpdater service again:
#          sudo systemctl start IncognitoUpdater.service
#        or just run the run_node.sh script:
#          sudo /home/incognito/run_node.sh
# ============================================================================================

# ============================= CHANGE CONFIG HERE ===========================================
declare -A CONFIGS HINTS
CONFIGS[VALIDATOR_K]="validator_key_1,validator_key_2,validator_key_3"
CONFIGS[GETH_NAME]="https://mainnet.infura.io/v3/xxxyyy"
CONFIGS[RPC_PORT]="8334"
CONFIGS[NODE_PORT]="9433"
CONFIGS[NUM_INDEXER_WORKERS]="0"
CONFIGS[INDEXER_ACCESS_TOKEN]="edeaaff3f1774ad2888673770c6d64097e391bc362d7d6fb34982ddf0efd18cb"

# =================================== CONFIGS GUIDE ============================================
HINTS[VALIDATOR_K]="! Input validator keys here, multiple validator keys must be separated by commas (no spaces):"
HINTS[GETH_NAME]='! Infura link. Example: "https://mainnet.infura.io/v3/xxxyyy".
   (follow step 3 on this thread to setup infura https://we.incognito.org/t/194):'
HINTS[RPC_PORT]="! RPC port, should be left as default.
   The first node uses this port, the next one uses port+1 and so on: "
HINTS[NODE_PORT]="! Node port, should be left as default.
   The first node uses this port, the next one uses port+1 and so on: "
HINTS[NUM_INDEXER_WORKERS]='! Number of coin indexer worker. To disable this, set it to 0.
   only needed when using your node to create transaction, query balance. Can be set to 0 for normal validator node'
HINTS[INDEXER_ACCESS_TOKEN]='! Indexer access token, can be generated by running: $ echo "bla bla bla" | sha256sum
   only needed when using your node to create transaction, query balance. Can ignore this setting for validator node'
# =================================== END CONFIG ============================================

# ============================================================================================
# Do not edit lines below unless you know what you're doing
BOOTNODE="mainnet-bootnode.incognito.org:9330"  # this should be left as default
FULLNODE=""  # set to 1 to run as a full node, empty to run as normal node
SERVICE="/etc/systemd/system/IncognitoUpdater.service"
TIMER="/etc/systemd/system/IncognitoUpdater.timer"
USER_NAME="incognito"
INC_HOME="/home/$USER_NAME"
DATA_DIR="$INC_HOME/node_data"
KEY_FILE="$INC_HOME/validator_keys"
SCRIPT="$INC_HOME/run_node.sh"
CONTAINER_NAME="inc_mainnet"
CONTAINER_INDEX_START=0

# check super user
if [ $(whoami) != root ]; then
cat << EOF
	!!! Please run with sudo or su, otherwise it won't work
	!!! Script now exits
EOF
	exit 1
fi

function uninstall {
cat << EOF
!!!===============================================!!!
###                    WARNING                    ###
!!!===============================================!!!
!   Uninstalling and cleanup !
!   This action will remove:
!      - The systemd service and timer: $SERVICE $TIMER
!      - Docker containers and images
!      - User: $USER_NAME
!      - $INC_HOME (including all node's data, logs and everything else inside the folder)
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
    echo " # Remove update service"
    systemctl stop $(basename $SERVICE) 2> /dev/null
    systemctl stop $(basename $TIMER) 2> /dev/null
    systemctl disable $(basename $SERVICE)
    systemctl disable $(basename $TIMER)
    systemctl daemon-reload
    echo " # Stop and remove docker images + containers"
    docker container stop $(docker container ls -aqf name=$CONTAINER_NAME)
    docker container rm $(docker container ls -aqf name=$CONTAINER_NAME)
    docker image rm -f $(docker images -q incognitochain/incognito-mainnet)
    docker network rm inc_net
    echo " # Removing user"
    deluser $USER_NAME
    echo " # Removing $INC_HOME (including all node's data, logs and everything else inside the folder)"
    rm -Rf $INC_HOME $SERVICE $TIMER
    exit 1
  fi
}

# ============================================================================================
function show_usage {
cat << EOF
   $(basename $0) [-u|h] [-y 1|2]
      -h: print this help message then exit.
      -y: Run setup with default settings, non-interactive mode.
          + [-y 1] to install on a new server or reinstall/update node configs current server.
                   This option will interrupt your node operation (if there's any)..
          + [-y 2] to update the Updater service, not the node configs such as port, access token, validator key ...
                   This option won't affect your node operation.
                   If you want to update the node configs, please use option 1
      -u: Uninstall auto update service

   Example:
    To uninstall the auto update service
      $(basename $0) -u

   or
    To install the auto update service with non-interactive mode
      $(basename $0) -y 1|2

EOF
exit 0
}

function get_conf_from_existing_container {
    # read current config from docker
  container_id=$(docker container ls -f name=${CONTAINER_NAME}_${CONTAINER_INDEX_START} -q)
  if [[ -z $container_id ]]; then
    echo "!! There's no existing Incognito container. Please use option 1 next time"
    exit 0
  fi
  CURRENT_CONF=$(docker inspect $container_id | jq '.[0].Config.Env' | tr \[\] {})
  CURRENT_CONF=${CURRENT_CONF//=/\":\"}
  CONFIGS[VALIDATOR_K]=$(cat $KEY_FILE)
  CONFIGS[GETH_NAME]=$(echo $CURRENT_CONF | jq .GETH_NAME)
  CONFIGS[RPC_PORT]=$(echo $CURRENT_CONF | jq .RPC_PORT)
  CONFIGS[NODE_PORT]=$(echo $CURRENT_CONF | jq .NODE_PORT)
  CONFIGS[NUM_INDEXER_WORKERS]=$(echo $CURRENT_CONF | jq .NUM_INDEXER_WORKERS)
  CONFIGS[INDEXER_ACCESS_TOKEN]=$(echo $CURRENT_CONF | jq .INDEXER_ACCESS_TOKEN)
}

function get_conf_from_input {
  for conf in ${!HINTS[@]}; do
    echo "${HINTS[$conf]}"
    printf "  default: ${CONFIGS[$conf]}\n\t> "; read input
    if [[ ! -z $input ]]; then CONFIGS[$conf]=$input; fi
  done
}

function prepare_new_setup {
  # try cleaning up old service and docker if there's any
  systemctl stop $(basename $SERVICE) 2> /dev/null
  systemctl stop $(basename $TIMER) 2> /dev/null
  docker container rm ${CONTAINER_NAME}_${CONTAINER_INDEX_START}
  echo " # Creating $USER_NAME user to run node"
  useradd $USER_NAME
  usermod -aG docker $USER_NAME || echo
  mkdir -p $INC_HOME
  chown -R $USER_NAME:$USER_NAME $INC_HOME
  echo ${CONFIGS[VALIDATOR_K]} > $KEY_FILE
}

while getopts "uy:ht" option; do
   case "$option" in
      "h")
         show_usage
         ;;
      "u")
         uninstall
         ;;
      "y")
         interactive_mode=${OPTARG}
         if [[ $interactive_mode != 1 && $interactive_mode != 2 ]]; then show_usage; fi
         ;;
      "t")
         test_mode=1
         ;;
      ?)
         show_usage
         ;;
   esac
done

echo "Checking for / Installing Docker and jq (JSON processor)"
apt install docker.io jq -y
systemctl start docker.service
echo "Finished. Launching setup script..."
sleep 3
clear
cat << 'EOF'
      ██╗███╗░░██╗░█████╗░░█████╗░░██████╗░███╗░░██╗██╗████████╗░█████╗░░░░░█████╗░██████╗░░██████╗░
      ██║████╗░██║██╔══██╗██╔══██╗██╔════╝░████╗░██║██║╚══██╔══╝██╔══██╗░░░██╔══██╗██╔══██╗██╔════╝░
      ██║██╔██╗██║██║░░╚═╝██║░░██║██║░░██╗░██╔██╗██║██║░░░██║░░░██║░░██║░░░██║░░██║██████╔╝██║░░██╗░
      ██║██║╚████║██║░░██╗██║░░██║██║░░╚██╗██║╚████║██║░░░██║░░░██║░░██║░░░██║░░██║██╔══██╗██║░░╚██╗
      ██║██║░╚███║╚█████╔╝╚█████╔╝╚██████╔╝██║░╚███║██║░░░██║░░░╚█████╔╝██╗╚█████╔╝██║░░██║╚██████╔╝
      ╚═╝╚═╝░░╚══╝░╚════╝░░╚════╝░░╚═════╝░╚═╝░░╚══╝╚═╝░░░╚═╝░░░░╚════╝░╚═╝░╚════╝░╚═╝░░╚═╝░╚═════╝░
            __    __                  __                   ______                       __              __
           /  \  /  |                /  |                 /      \                     /  |            /  |
 __     __ $$  \ $$ |  ______    ____$$ |  ______        /$$$$$$  |  _______   ______  $$/   ______   _$$ |_
/  \   /  |$$$  \$$ | /      \  /    $$ | /      \       $$ \__$$/  /       | /      \ /  | /      \ / $$   |
$$  \ /$$/ $$$$  $$ |/$$$$$$  |/$$$$$$$ |/$$$$$$  |      $$      \ /$$$$$$$/ /$$$$$$  |$$ |/$$$$$$  |$$$$$$/
 $$  /$$/  $$ $$ $$ |$$ |  $$ |$$ |  $$ |$$    $$ |       $$$$$$  |$$ |      $$ |  $$/ $$ |$$ |  $$ |  $$ | __
  $$ $$/   $$ |$$$$ |$$ \__$$ |$$ \__$$ |$$$$$$$$/       /  \__$$ |$$ \_____ $$ |      $$ |$$ |__$$ |  $$ |/  |
   $$$/    $$ | $$$ |$$    $$/ $$    $$ |$$       |      $$    $$/ $$       |$$ |      $$ |$$    $$/   $$  $$/
    $/     $$/   $$/  $$$$$$/   $$$$$$$/  $$$$$$$/        $$$$$$/   $$$$$$$/ $$/       $$/ $$$$$$$/     $$$$/
                                                                                           $$ |
                                                                                           $$ |
                                                                                           $$/
EOF
if [[ -z $interactive_mode ]]; then # interactive mode, taking user input
cat << EOF
=======================================================
     Start setup with interactive mode
=======================================================

!! To select default values, just hit ENTER !!
EOF

  while cat << EOF
! 1: Install on a new server or reinstall/update node configs current server.
      This option will interrupt your node operation (if there's any)..
! 2: Update current Updater service, not the node config such as port, access token, validator key ....
      This option won't affect your node operation.
      If you want to update the node configs, please use option 1
EOF
  do
    read interactive_select
    if [[ $interactive_select == 1 ]]; then
      echo "  !!  INSTALLING NEW NODE UPDATER SERVICE"
      # now get new config from user input
      get_conf_from_input
      prepare_new_setup
      break
    elif [[ $interactive_select == 2 ]]; then
      echo "  !!  UPDATING CURRENT NODE UPDATER SERVICE"
      # read current config from docker
      get_conf_from_existing_container
      break
    fi
  done
else
cat << EOF
=======================================================
     Start setup with non-interactive mode
=======================================================
EOF
  case $interactive_mode in
    1) # -y1 use declared CONFIGS above, nothing to do
      prepare_new_setup
      ;;
    2) # -y2 read current docker configs
      get_conf_from_existing_container
      ;;
    ?)
      show_usage
      ;;
  esac
fi
cat << EOF
Setup with following configurations:
      Validator keys: ${CONFIGS[VALIDATOR_K]}
      Infura: ${CONFIGS[GETH_NAME]}
      RPC port: ${CONFIGS[RPC_PORT]}
      Node port: ${CONFIGS[NODE_PORT]}
      Number of indexer worker: ${CONFIGS[NUM_INDEXER_WORKERS]}
      Coin indexer access token: ${CONFIGS[INDEXER_ACCESS_TOKEN]}
EOF

if [[ -z $test_mode ]]; then echo ""; else exit 0; fi

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

echo " # Creating timer to preodically run the update checker"
cat << EOF > $TIMER
[Unit]
Description=Run IncognitoUpdater hourly

[Timer]
OnCalendar=hourly
RandomizedDelaySec=1000
Persistent=true

[Install]
WantedBy=timers.target
EOF

echo " # Creating run node script"
cat << EOF > $SCRIPT
#!/bin/bash
key_file=$KEY_FILE
function do_update {
  bootnode=$BOOTNODE
  data_dir=$DATA_DIR
  rpc_port=${CONFIGS[RPC_PORT]}
  node_port=${CONFIGS[NODE_PORT]}
  geth_name=${CONFIGS[GETH_NAME]}
  geth_port=""
  geth_proto=""
  fullnode=$FULLNODE
  coin_index_access_token=${CONFIGS[INDEXER_ACCESS_TOKEN]}
  num_index_worker=${CONFIGS[NUM_INDEXER_WORKERS]}
  container_name=$CONTAINER_NAME
  count=$CONTAINER_INDEX_START

EOF

cat << 'EOF' >> $SCRIPT
  validator_key=$(cat $key_file)
  validator_key=(${validator_key//,/ })
  new_tag=$1
  old_tag=$2

  if [[ ! -z $old_tag ]]; then
	  echo "Found new docker tag, remove old one"
    docker image rm -f incognitochain/incognito-mainnet:${old_tag}
  fi

  echo "Pulling new tag: ${new_tag}"
  docker pull incognitochain/incognito-mainnet:${new_tag}
  echo "Create new docker network"
  docker network create --driver bridge inc_net || true
  for key in ${validator_key[@]}; do
    echo "Remove old container if there's any"
    docker container stop ${container_name}_${count}
    docker container rm ${container_name}_${count}

    echo "Start the incognito mainnet docker container"
    set -x
    docker run --restart=always --net inc_net \
      -p $node_port:$node_port -p $rpc_port:$rpc_port \
      -e NODE_PORT=$node_port -e RPC_PORT=$rpc_port -e BOOTNODE_IP=$bootnode \
      -e GETH_NAME=$geth_name -e GETH_PROTOCOL=$geth_proto -e GETH_PORT=$geth_port \
      -e FULLNODE=$fullnode -e MININGKEY=${key} -e TESTNET=false \
      -e INDEXER_ACCESS_TOKEN=$coin_index_access_token -e NUM_INDEXER_WORKERS=$num_index_worker \
      -v ${data_dir}_${count}:/data -d --name ${container_name}_${count} incognitochain/incognito-mainnet:${new_tag}
    set +x
    ((node_port++))
    ((rpc_port++))
    ((count++))
  done
}

container_id=($(docker container ls -f name=$container_name -q))
if [[ $container_id != "" ]]; then
  current_tag=$(docker inspect ${container_id[0]} | jq '.[].Config.Image' | tr -d \" | cut -d ":" -f2)
fi
echo "Getting Incognito docker tags"
tags=$(curl -s -X GET https://hub.docker.com/v2/namespaces/incognitochain/repositories/incognito-mainnet/tags?page_size=100 | jq '.results[].name' | tr -d "\"")
IFS=$'\n'
sorted_tags=($(sort -nr <<< "${tags[*]}"))
latest_tag=${sorted_tags[0]}
unset IFS
echo "Current tag |${current_tag}| - Latest tag |${latest_tag}|"
if [[ -z $latest_tag ]]; then
  echo "Cannot get tags from docker hub for now. Skip this round!"
  exit 0
fi

if [ "$current_tag" != "$latest_tag" ]; then
	do_update $latest_tag $current_tag
fi
EOF

chmod +x $SCRIPT

echo " # Enabling service"
systemctl daemon-reload
systemctl enable $(basename $SERVICE)
systemctl enable $(basename $TIMER)


echo " # Starting service. Please wait..."
systemctl start $(basename $SERVICE)
systemctl start $(basename $TIMER)
cat << EOF
 # DONE.
 To check the installing and starting progress or the running service:
    $ journalctl | grep Inc
    or
    $ journalctl -t IncNodeUpdt
EOF
