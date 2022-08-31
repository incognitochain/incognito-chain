#!/bin/bash

SERVICE_FILE=/etc/systemd/system/IncognitoUpdater.service
RUN_NODE_FILE=$(cat $SERVICE_FILE | grep ExecStart | cut -d " " -f3)

sed -i s'|echo "Latest tag is ${latest_tag}"|echo "Latest tag is ${latest_tag}\nif [[ -z $latest_tag ]]; then\n  echo "Cannot get tags from docker hub for now. Skip this round!"\n  exit 0\nfi|' $RUN_NODE_FILE