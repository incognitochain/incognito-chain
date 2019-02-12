FROM alpine:3.9

ADD ./constant /
ADD ./bootnode /
RUN chmod +x /constant
RUN chmod +x /bootnode
ADD ./run_bootnode.sh /
ADD ./run_constant.sh /
RUN chmod +x /run_bootnode.sh
RUN chmod +x /run_constant.sh

CMD ["/bin/sh"]
