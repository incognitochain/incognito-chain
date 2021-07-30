#!/bin/bash

open_shard() {
    counter=0
    while [ $counter -le 0 ]
    do
        if [ $counter -ne 0 ]; then
            echo "tmux new-window \"bash run_node.sh shard${counter}-0\""
            tmux new-window "bash run_node.sh shard${counter}-0"
        fi
        echo "tmux new-window \"bash run_node.sh shard${counter}-1\""
        tmux new-window "bash run_node.sh shard${counter}-1"

        echo "tmux new-window \"bash run_node.sh shard${counter}-2\""
        tmux new-window "bash run_node.sh shard${counter}-2"

        echo "tmux new-window \"bash run_node.sh shard${counter}-3\""
        tmux new-window "bash run_node.sh shard${counter}-3"

        echo "Done shard${counter}"
        ((counter++))
    done
}

open_beacon() {
    echo "tmux new-window \"bash run_node.sh beacon-0\""
    tmux new-window "bash run_node.sh beacon-0"
    echo "tmux new-window \"bash run_node.sh beacon-1\""
    tmux new-window "bash run_node.sh beacon-1"
    echo "tmux new-window \"bash run_node.sh beacon-2\""
    tmux new-window "bash run_node.sh beacon-2"
    echo "tmux new-window \"bash run_node.sh beacon-3\""
    tmux new-window "bash run_node.sh beacon-3"
}

#tmux new -s my_session
#tmux new-window "bash run_node.sh shard${counter}-2"

tmux new-session -d 'bash'
open_shard
#open_beacon