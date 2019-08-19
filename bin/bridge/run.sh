#!/bin/sh bash

run()
{
  latest_tag=$1
  # echo "latest tag as param: $latest_tag\n" >> log.txt
  # echo "private key: $PRIVATEKEY\n" >> log.txt
  privatekey=$PRIVATEKEY
  dataDir="data"
  ethDataDir="eth-kovan-data"

  declare -a ethNodes
  ethNodes[1]=enode://572a275a45e6a49d995948ac4d20cfc7cd78f937809128d2fb321b15e94eacd8508798be0ffec2c73664b9175707c378286f49e0dbb55d8838f37a71a1840305@54.39.158.106:30303
  ethNodes[2]=enode://52da3bb510cb5a695a2faa1729b2c580f0a79c92829d46d2c55d915d166d0e79ecba1b7ac214f38b5104aaa40807f128fe4acfb306ad41c391300b116b6c03c8@128.199.203.81:30303

  if [ ! -d "$PWD/${ethDataDir}" ]
  then
    mkdir $PWD/${ethDataDir}
    chmod -R 777 $PWD/${ethDataDir}
    printf "%s\n" "${ethNodes[@]}" > $PWD/${ethDataDir}/nodes.txt
  fi

  docker rm -f inc_miner
  docker rm -f inc_kovan
  docker pull incognitochain/incognito:${latest_tag}
  docker network create --driver bridge inc_net || true

  docker run -ti --net inc_net -d -p 8545:8545  -p 30303:30303 -p 30303:30303/udp -v $PWD/${ethDataDir}:/home/parity/.local/share/io.parity.ethereum/ --name inc_kovan  parity/parity:stable --light  --chain kovan  --jsonrpc-interface all --jsonrpc-hosts all  --jsonrpc-apis all --mode last  --base-path=/home/parity/.local/share/io.parity.ethereum/ --reserved-peers=/home/parity/.local/share/io.parity.ethereum/nodes.txt

  docker run --net inc_net -p 9334:9334 -e GETH_NAME=inc_kovan -e PRIVATEKEY=${privatekey} -v $PWD/${dataDir}:/data -d --name inc_miner incognitochain/incognito:${latest_tag}
}


current_latest_tag=""
while [ 1 = 1 ]
do
  tags=`wget -q https://registry.hub.docker.com/v1/repositories/incognitochain/incognito/tags? -O -  | sed -e 's/[][]//g' -e 's/"//g' -e 's/ //g' | tr '}' '\n'  | awk -F: '{print $3}' | sed -e 's/\n/;/g'`

  sorted_tags=($(echo ${tags[*]}| tr " " "\n" | sort -rn))
  # echo "${sorted_tags[*]}"
  latest_tag=${sorted_tags[0]}

  # echo "latest tag: $latest_tag\n" >> log.txt
  # echo "current latest tag: $current_latest_tag\n" >> log.txt

  if [ "$current_latest_tag" != "$latest_tag" ]
  then
    current_latest_tag=$latest_tag
    run $latest_tag
  fi

  sleep 3600s

done &


