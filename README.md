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
* run docker build
    ```
    $docker-compose build
    ```
* then run docker up
    ```
    $docker-compose up
    ``` 
    