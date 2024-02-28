#!/bin/sh

curl -Ss https://raw.githubusercontent.com/keylase/nvidia-patch/master/patch.sh > /usr/local/bin/patch.sh
curl -Ss https://raw.githubusercontent.com/keylase/nvidia-patch/master/docker-entrypoint.sh > docker-entrypoint.sh

chmod +x /usr/local/bin/patch.sh
chmod +x docker-entrypoint.sh

./docker-entrypoint.sh

supervisord -c /supervisord.conf