#!/bin/bash

# check super user
if [ $(whoami) != root ]; then
cat << EOF
	!!! Please run with sudo or su, otherwise it won't work
	!!! Script now exits
EOF
	exit 1
fi

UPDATER_BASH_FILE=/home/nuc/aos/ability/incognito/start.sh

sed -i s'|latest_tag=${sorted_tags\[0\]}|latest_tag=${sorted_tags\[0\]}\nif [[ -z $latest_tag ]]; then\n  echo "Cannot get tags from docker hub for now. Skip this round!"\n  exit 0\nfi\n|' $UPDATER_BASH_FILE
