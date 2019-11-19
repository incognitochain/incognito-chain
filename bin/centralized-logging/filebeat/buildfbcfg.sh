#!/bin/sh bash

declare -a raw_log_paths
raw_log_paths[1]=/data/*.txt

declare -a json_log_paths
json_log_paths[1]=/data/*.json

logstash_addresses_to_replace=34.94.185.164:5000

if [ ! -z "$RAW_LOG_PATHS" ]
then
  IFS=',' read -r -a raw_log_paths <<< "$RAW_LOG_PATHS"
fi

if [ ! -z "$JSON_LOG_PATHS" ]
then
  IFS=',' read -r -a json_log_paths <<< "$JSON_LOG_PATHS"
fi

if [ ! -z "$LOGSTASH_ADDRESSES" ]
then
  IFS=',' read -r -a logstash_addresses_to_replace <<< "$LOGSTASH_ADDRESSES"
fi

# tab=$'\t'
newline=$'\\\n'

declare -a raw_log_paths_to_replace
for element in "${raw_log_paths[@]}"
do
  # raw_log_paths_to_replace="${raw_log_paths_to_replace}${tab}${tab}- ${element}${newline}"
  raw_log_paths_to_replace="${raw_log_paths_to_replace}  - ${element}${newline}"
done

declare -a json_log_paths_to_replace
for element in "${json_log_paths[@]}"
do
  # json_log_paths_to_replace="${json_log_paths_to_replace}${tab}${tab}- ${element}${newline}"
  json_log_paths_to_replace="${json_log_paths_to_replace}  - ${element}${newline}"
done

cp filebeat.template.yml filebeat.yml

sed -i "s,{raw_log_paths},${raw_log_paths_to_replace},g" filebeat.yml
sed -i "s,{json_log_paths},${json_log_paths_to_replace},g" filebeat.yml
sed -i "s,{logstash_addresses},${logstash_addresses_to_replace},g" filebeat.yml

# execute filebeat
./filebeat -c filebeat.yml -e --strict.perms=false
