#!/usr/bin/env bash
## Start chain
tmux send-keys -t bootnode C-C ENTER ./bootnode/bootnode ENTER
sleep 1
tmux send-keys -t shard00 C-C ENTER ./run_node.sh\ shard0-0 ENTER
sleep 1
tmux send-keys -t shard10 C-C ENTER ./run_node.sh\ shard1-0 ENTER
sleep 1
tmux send-keys -t shard01 C-C ENTER ./run_node.sh\ shard0-1 ENTER
sleep 1
tmux send-keys -t shard11 C-C ENTER ./run_node.sh\ shard1-1 ENTER
sleep 1

tmux send-keys -t beacon0 C-C ENTER ./run_node.sh\ beacon-0 ENTER
sleep 1
tmux send-keys -t beacon1 C-C ENTER ./run_node.sh\ beacon-1 ENTER
sleep 1
tmux send-keys -t beacon2 C-C ENTER ./run_node.sh\ beacon-2 ENTER
sleep 1
tmux send-keys -t beacon3 C-C ENTER ./run_node.sh\ beacon-3 ENTER
sleep 1

tmux send-keys -t shard02 C-C ENTER ./run_node.sh\ shard0-2 ENTER
sleep 1
tmux send-keys -t shard03 C-C ENTER ./run_node.sh\ shard0-3 ENTER
sleep 1
tmux send-keys -t shard12 C-C ENTER ./run_node.sh\ shard1-2 ENTER
sleep 1
tmux send-keys -t shard13 C-C ENTER ./run_node.sh\ shard1-3 ENTER
sleep 1
tmux send-keys -t fullnode C-C ENTER ./run_node.sh\ full_node ENTER
