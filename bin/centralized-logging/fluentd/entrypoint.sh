#!/bin/sh





sed -i "s/%SERVER%/${server}/g" /fluentd/etc/fluent.conf
#source vars if file exists
DEFAULT=/etc/default/fluentd

if [ -r $DEFAULT ]; then
    set -o allexport
    . $DEFAULT
    set +o allexport
fi

# If the user has supplied only arguments append them to `fluentd` command
if [ "${1#-}" != "$1" ]; then
    set -- fluentd "$@"
fi

# If user does not supply config file or plugins, use the default
if [ "$1" = "fluentd" ]; then
    if [ -z $server ];then
        echo "ERROR: Must include env param 'server'"
        exit -1
    fi

    if ! echo $@ | grep ' \-c' ; then
       set -- "$@" -c /fluentd/etc/${FLUENTD_CONF}
    fi

    if ! echo $@ | grep ' \-p' ; then
       set -- "$@" -p /fluentd/plugins
    fi
fi

echo "$@"
exec "$@"