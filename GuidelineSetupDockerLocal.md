###Docker file for testing

###Login Docker Hub
 Create account and login Docker hub
 
###Build image
```
$ docker build -t hoangincognito/incognito-chain -f Dockerfilelocal .
$ docker build -t hoangincognito/incognito-highway  -f Dockerfile .
```
Note: Some images was push to docker register at 
https://hub.docker.com/repository/docker/hoangincognito/incognito-highway
https://hub.docker.com/repository/docker/hoangincognito/incognito-chain

###Run docker compose
```$ docker-compose -f docker-compose.local.yml up```