FROM docker.elastic.co/beats/filebeat:7.4.0

USER root

WORKDIR /usr/share/filebeat

COPY ./buildfbcfg.sh .
COPY ./filebeat.template.yml .

RUN chmod -R 777 .

CMD ["/bin/sh", "buildfbcfg.sh"]

RUN chown root:filebeat /usr/share/filebeat/filebeat.yml
USER filebeat
