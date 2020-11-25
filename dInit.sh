go build -o incognito
pushd ../incognito-highway; pm2 --name highway start ./run.sh -- newlc1; git diff popd
pm2 --name n00 start ./run_node.sh -- shard0-0
pm2 --name n01 start ./run_node.sh -- shard0-1
pm2 --name n02 start ./run_node.sh -- shard0-2
pm2 --name n03 start ./run_node.sh -- shard0-3
pm2 --name b0 start ./run_node.sh -- beacon-0
pm2 --name b1 start ./run_node.sh -- beacon-1
pm2 --name b2 start ./run_node.sh -- beacon-2
pm2 --name b3 start ./run_node.sh -- beacon-3
pm2 --name relay start ./run_node.sh -- relaynode
