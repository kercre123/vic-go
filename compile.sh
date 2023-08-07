#!/bin/bash

set -e

if [[ ! -f main.go ]]; then
	echo "This must be run in the vic-go directory"
	exit 1
fi

if [[ ! -f ./toolchain/bin/arm-linux-gnueabi-g++ ]]; then
	echo "Run the ./download-deps.sh (toolchain not found)"
	exit 1
fi

mkdir -p build

ABSPATH="${PWD}"

export LD_LIBRARY_PATH=$ABSPATH/toolchain/lib:$ABSPATH/toolchain/arm-linux-gnueabi/libc/usr/lib:$ABSPATH/toolchain/arm-linux-gnueabi/libc/lib/

$ABSPATH/toolchain/bin/arm-linux-gnueabi-g++ \
-w -shared \
-o build/librobot.so \
hacksrc/libs/spine.cpp \
hacksrc/spine_demo.cpp \
hacksrc/libs/utils.cpp \
hacksrc/libs/lcd.cpp \
hacksrc/lcd_demo.cpp \
hacksrc/libs/cam.cpp \
hacksrc/cam_demo.cpp \
-Iinclude -fPIC

CC="$ABSPATH/toolchain/bin/arm-linux-gnueabi-gcc -w -Lbuild" \
CFLAGS="-Iinclude" \
CGO_LDFLAGS="-ldl" \
GOARM=7 \
GOARCH=arm \
CGO_ENABLED=1 \
go build \
-ldflags '-w -s' \
-o build/main \
main.go

echo "Compiled successfully! Now you can send to the robot with ./send.sh <robotip> (expects ssh_root_key in user directory)"
