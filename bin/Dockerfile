FROM alpine:3.9
COPY ./constant /
COPY ./bootnode /
RUN chmod +x /constant
RUN chmod +x /bootnode
COPY ./run_bootnode.sh /
COPY ./run_constant.sh /
RUN chmod +x /run_bootnode.sh
RUN chmod +x /run_constant.sh

CMD ["/bin/sh"]
