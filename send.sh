#!/bin/bash

set -e

if [[ ! $1 ]]; then
	echo "You must provide an IP"
	exit 1
fi

ssh -i ~/ssh_root_key root@$1 "mount -o rw,remount / && mount -o rw,remount,exec /data && mkdir -p /data/vic-go"
scp $2 -i ~/ssh_root_key build/main root@$1:/data/vic-go/
scp $2 -i ~/ssh_root_key build/*.so root@$1:/lib/

echo 'Sent to the bot! Now you can SSH in, disable the Anki apps, and run "/data/vic-go/main"'
