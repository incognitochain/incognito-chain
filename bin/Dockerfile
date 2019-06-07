FROM alpine:3.9

ARG commit=commit
ENV commit=$commit

COPY ./constant /
COPY ./bootnode /

RUN chmod +x /constant
RUN chmod +x /bootnode

COPY ./run_bootnode.sh /
COPY ./run_constant.sh /
COPY ./run_fullnode.sh /

RUN chmod +x /run_bootnode.sh
RUN chmod +x /run_constant.sh
RUN chmod +x /run_fullnode.sh

RUN mkdir /utility

COPY ./txs-shard0-noprivacy-5000.json /utility/
COPY ./txs-shard0-noprivacy-9000.json /utility/
COPY ./txs-shard0-privacy-5000.json /utility/
COPY ./txs-shard0-privacy-3000.1.json /utility/
COPY ./txs-shard0-privacy-3000.2.json /utility/
COPY ./txs-shard0-privacy-3000.3.json /utility/

COPY ./txs-shard1-noprivacy-5000.json /utility/
COPY ./txs-shard1-noprivacy-9000.json /utility/
COPY ./txs-shard1-privacy-5000.json /utility/
COPY ./txs-shard1-privacy-3000.1.json /utility/
COPY ./txs-shard1-privacy-3000.2.json /utility/
COPY ./txs-shard1-privacy-3000.3.json /utility/
COPY ./txs-shard1-cstoken-5000.json /utility/
COPY ./txs-shard1-cstokenprivacy-5000.json /utility/

CMD ["/bin/sh"]
