#!/bin/bash



if [[ ! $1 ]]; then
	echo "You must provide an IP"
	exit 1
fi

ssh -i ~/ssh_root_key root@$1 "mount -o rw,remount / && mount -o rw,remount,exec /data"
set -e
ssh -i ~/ssh_root_key root@$1 "mkdir -p /data/vic-go"
scp $2 -i ~/ssh_root_key build/main root@$1:/data/vic-go/
if [[ ! $NO_LIBS ]]; then
scp $2 -i ~/ssh_root_key build/librobot.so root@$1:/lib/
if [[ -f ./build/libjpeg_interface.so ]]; then
	scp $2 -i ~/ssh_root_key build/libjpeg_interface.so root@$1:/lib/
	scp $2 -i ~/ssh_root_key libjpeg-turbo/lib/libturbojpeg.so.0 root@$1:/lib/
        scp $2 -i ~/ssh_root_key toolchain/arm-linux-gnueabi/lib/libgomp.so.1 root@$1:/lib/
fi
if [[ -f ./build/libanki-camera.so ]]; then
	scp $2 -i ~/ssh_root_key build/libanki-camera.so root@$1:/lib/
fi
fi

if [[ $SEND_WEBROOT ]]; then
	scp $2 -i ~/ssh_root_key -r rc/webroot root@$1:/data/vic-go/
fi

echo 'Sent to the bot! Now you can SSH in, disable the Anki apps, and run "/data/vic-go/main"'
