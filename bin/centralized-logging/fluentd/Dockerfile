FROM fluent/fluentd:v1.10-debian-1

USER root

COPY fluent.conf /fluentd/etc/
COPY entrypoint.sh /bin/
RUN chmod +x /bin/entrypoint.sh
RUN mkdir -p /tailpos && chmod 777 -R /tailpos
RUN mkdir -p /data /log
RUN chmod 777 -R /data
RUN chmod 777 -R /log
COPY server /bin/

USER fluent