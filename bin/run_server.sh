#!/usr/bin/env bash
declare -a ethNodes
ethNodes[1]=enode://572a275a45e6a49d995948ac4d20cfc7cd78f937809128d2fb321b15e94eacd8508798be0ffec2c73664b9175707c378286f49e0dbb55d8838f37a71a1840305@54.39.158.106:30303
ethNodes[2]=enode://52da3bb510cb5a695a2faa1729b2c580f0a79c92829d46d2c55d915d166d0e79ecba1b7ac214f38b5104aaa40807f128fe4acfb306ad41c391300b116b6c03c8@128.199.203.81:30303

if [ ! -d "/data/eth-data" ]
then
  mkdir -p /data/eth-data
  chmod -R 777 /data/eth-data
  printf "%s\n" "${ethNodes[@]}" > /data/eth-data/nodes.txt
fi

docker pull incognitochain/incognito:${TAG}

docker rm -f ${HOSTNAME};

docker network create --driver bridge inc_net || true

if docker ps | grep inc_kovan ; then
    echo "Kovan already here";
else
    docker run -ti --net inc_net -d -p 8545:8545  -p 30303:30303 -p 30303:30303/udp -v /data/eth-data:/home/parity/.local/share/io.parity.ethereum/ --name inc_kovan  parity/parity:stable --light  --chain kovan  --jsonrpc-interface all --jsonrpc-hosts all  --jsonrpc-apis all --mode last  --base-path=/home/parity/.local/share/io.parity.ethereum/ --reserved-peers=/home/parity/.local/share/io.parity.ethereum/nodes.txt
fi

echo docker run --net inc_net -e GETH_NAME=inc_kovan -e NAME=${HOSTNAME} -p ${NODE_PORT}:${NODE_PORT} -p ${RPC_PORT}:${RPC_PORT} -p ${WS_PORT}:${WS_PORT} -e BOOTNODE_IP="${BOOTNODE_IP}" -v /data/$HOSTNAME:/data -e PRIVATEKEY=${PRIVATEKEY} -e PUBLIC_IP="${PUBLIC_IP}"  -e NODE_PORT="${NODE_PORT}" -e WS_PORT="${WS_PORT}" -e RPC_PORT="${RPC_PORT}" -d --name $HOSTNAME incognitochain/incognito:${TAG} /run_incognito.sh ${CLEAR} > ${HOSTNAME}.asb

docker run --net inc_net -e GETH_NAME=inc_kovan -e NAME=${HOSTNAME} -p ${NODE_PORT}:${NODE_PORT} -p ${RPC_PORT}:${RPC_PORT} -p ${WS_PORT}:${WS_PORT} -e BOOTNODE_IP="${BOOTNODE_IP}" -v /data/$HOSTNAME:/data -e PRIVATEKEY=${PRIVATEKEY} -e PUBLIC_IP="${PUBLIC_IP}"  -e NODE_PORT="${NODE_PORT}" -e WS_PORT="${WS_PORT}" -e RPC_PORT="${RPC_PORT}" -d --name $HOSTNAME incognitochain/incognito:${TAG} /run_incognito.sh ${CLEAR}