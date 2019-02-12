BUILD
    docker build . -t constant
RUN BOOTNODE
    docker run -p 9330:9330 \
        -e PORT='9330' \
        -it constant /run_bootnode.sh
RUN CONSTANT
    docker run \
        -p 9331:9331 \
        -e DISCOVERPEERSADDRESS='127.0.0.1:9330' \
        -e SPENDINGKEY='112t8rqGc71CqjrDCuReGkphJ4uWHJmiaV7rVczqNhc33pzChmJRvikZNc3Dt5V7quhdzjWW9Z4BrB2BxdK5VtHzsG9JZdZ5M7yYYGidKKZV' \
        -e EXTERNALADDRESS='127.0.0.1' \
        -e PORT='9331' \
        -it constant /run_constant.sh
