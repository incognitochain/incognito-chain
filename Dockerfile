# FROM golang:1.14-alpine as builder

# RUN apk add --no-cache make gcc musl-dev linux-headers git

# ADD . /incognitochain
# RUN cd /incognitochain && make build

# # Bring Incognito bin file into a second stage deploy alpine container
# FROM ubuntu:16.04

# WORKDIR /incognitochain
# RUN apt-get update
# RUN apt-get install -y ca-certificates cronolog cron
# RUN apt-get install -y dnsutils

# COPY ./removeoldlog /etc/cron.d/removeoldlog
# RUN chmod 0644 /etc/cron.d/removeoldlog
# RUN crontab /etc/cron.d/removeoldlog

# ARG commit=commit
# ENV commit=$commit

# #COPY --from=builder /incognitochain/incognito /usr/local/bin/
# COPY --from=builder /incognitochain/incognito /incognitochain
# COPY --from=builder /incognitochain/priv2.json /incognitochain/
# COPY --from=builder /incognitochain/whitelist.json /incognitochain/
# COPY --from=builder /incognitochain/config/ /incognitochain/config/
# COPY --from=builder /incognitochain/run_incognito.sh /incognitochain/

# RUN chmod +x /incognitochain/run_incognito.sh
# CMD ["/bin/bash","run_incognito.sh"]

FROM ubuntu:16.04


RUN apt-get update
RUN apt-get install -y ca-certificates cronolog cron
RUN apt-get install -y dnsutils


COPY ./removeoldlog /etc/cron.d/removeoldlog
RUN chmod 0644 /etc/cron.d/removeoldlog
RUN crontab /etc/cron.d/removeoldlog

ARG commit=commit
ENV commit=$commit


COPY ./incognito /
RUN chmod +x /incognito

COPY ./priv2.json /
COPY ./whitelist.json /
COPY ./config/ /config
COPY ./run_incognito.sh /

RUN mkdir /utility

CMD ["/bin/bash","run_incognito.sh"]

ARG commit=commit
ENV commit=$commit

COPY ./incognito /
RUN chmod +x /incognito
