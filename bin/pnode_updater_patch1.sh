#!/bin/bash

# check super user
if [ $(whoami) != root ]; then
cat << EOF
	!!! Please run with sudo or su, otherwise it won't work
	!!! Script now exits
EOF
	exit 1
fi
apt install jq -y

UPDATER_BASH_FILE=/home/nuc/aos/ability/incognito/start.sh

if [[ -z $(grep "Cannot get tags from docker hub for now" $UPDATER_BASH_FILE) ]]; then
	sed -i s'|latest_tag=${sorted_tags\[0\]}|latest_tag=${sorted_tags\[0\]}\nif [[ -z $latest_tag ]]; then\n  echo "Cannot get tags from docker hub for now. Skip this round!"\n  exit 0\nfi\n|' $UPDATER_BASH_FILE
fi

if [[ -z $(grep "hub.docker.com/v2/" $UPDATER_BASH_FILE) ]]; then
	sed -i s'!tags=`curl.*!tags=$(curl -sX GET https://hub.docker.com/v2/namespaces/incognitochain/repositories/incognito-mainnet/tags?page_size=100 | jq ".results[].name" | tr -d "\\"")!' $UPDATER_BASH_FILE
fi
