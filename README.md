# prototype

Pre: install glide https://github.com/Masterminds/glide

- Clone Project
- to install dependencies Run command-line below 
    ```
    $glide install
    $go build ./
    ```
- get first node up 
    ```
    $./cash-prototype
    ```
- get node `n` up
    ```
    $./cash-prototype -listen “127.0.0.1:9555” --norpc --datadir “data1" --connect “/ip4/127.0.0.1/tcp/9333/ipfs/QmawrS2w63oXTq9dS8sFYk6ebttLPpdKm7eosTUPx4YGu8” --generate --miningaddr “mgnUx4Ah4VBvtaL7U1VXkmRjKUk3h8pbst”
    ```

## run with docker-compose
* To start dev container
    ```
    $docker-compose -p cashdev -f dev-env/docker-compose.yaml up -d
    ``` 
* To stop dev container
    ```
    $docker-compose -f dev-env/docker-compose.yaml down
    ``` 
* To start developing
    ```
    $docker exec -it cash-prototype-dev sh
    ``` 
    ```
    $glide install
    ```
    ```
    $go build
    ```
    ```
    $./cash-prototype
    ```
* To start other nodes (these nodes will start will config file in dev-env/nodes-data/node-<NODE_NUMBER>)
    ```
    $docker run -i -t --net cashdev_cash-net --mount type=bind,src=$PWD/cash-prototype,dst=/cash-prototype --mount type=bind,src=$PWD/dev-env/nodes-data/node<REPLACE THIS WILL NODE_NUMBER>,dst=/nodedata --expose 9333 alpine:3.7 /cash-prototype --configfile /nodedata/config.conf
    ```