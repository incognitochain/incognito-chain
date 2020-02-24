###Docker file for testing

###Build image
```
$ docker build -t hoangincognito/incognito-chain -f Dockerfilelocal .
$ docker build -t hoangincognito/incognitochain-highway  -f Dockerfile .
```
Note: Some images was push to docker register at 
https://hub.docker.com/repository/docker/hoangincognito/incognitochain-highway
https://hub.docker.com/repository/docker/hoangincognito/incognito-chain

###Run docker compose
```$ docker-compose -f docker-compose.local.yml up```