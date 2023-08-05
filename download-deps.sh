#!/bin/bash

set -e

UNAME=`uname -a`

if [[ ! -f main.go ]]; then
	echo "This script must be run in the vic-go directory"
	exit 1
fi

if [[ "${UNAME}" == *"x86_64"* ]]; then
    ARCH="x86_64"
    echo "amd64 architecture confirmed."
else
    echo "Your CPU architecture not supported. This script currently supports x86_64."
    exit 1
fi

if [[ $EUID -ne 0 ]]; then
    echo "This script must be run as root. sudo ./download-deps.sh"
    exit 1
fi

        if [[ ! -f /usr/local/go/bin/go ]]; then
	    echo "Downloading go"
	    mkdir golang
	    cd golang
            if [[ ${ARCH} == "x86_64" ]]; then
                wget -q --show-progress --no-check-certificate https://go.dev/dl/go1.19.4.linux-amd64.tar.gz
                rm -rf /usr/local/go && tar -C /usr/local -xzf go1.19.4.linux-amd64.tar.gz
            elif [[ ${ARCH} == "aarch64" ]]; then
                wget -q --show-progress --no-check-certificate https://go.dev/dl/go1.19.4.linux-arm64.tar.gz
                rm -rf /usr/local/go && tar -C /usr/local -xzf go1.19.4.linux-arm64.tar.gz
            elif [[ ${ARCH} == "armv7l" ]]; then
                wget -q --show-progress --no-check-certificate https://go.dev/dl/go1.19.4.linux-armv6l.tar.gz
                rm -rf /usr/local/go && tar -C /usr/local -xzf go1.19.4.linux-armv6l.tar.gz
            fi
            ln -s /usr/local/go/bin/go /usr/bin/go
	    cd ..
	    rm -rf golang
        fi


if [[ ! -f ./toolchain/bin/arm-linux-gnueabi-gcc ]]; then
	echo "Downloading arm toolchain"
	mkdir -p toolchain-down
	cd toolchain-down
	wget https://releases.linaro.org/components/toolchain/binaries/5.5-2017.10/arm-linux-gnueabi/gcc-linaro-5.5.0-2017.10-x86_64_arm-linux-gnueabi.tar.xz
	echo "Decompressing toolchain..."
	tar -xf gcc-linaro-5.5.0-2017.10-x86_64_arm-linux-gnueabi.tar.xz
	mv gcc-linaro-5.5.0-2017.10-x86_64_arm-linux-gnueabi ../toolchain
	cd ..
	rm -rf toolchain-down
	chmod -R +rwx ./toolchain/
fi

echo "Done! Now you can run ./compile.sh."
exit 0
