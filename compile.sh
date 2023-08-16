#!/bin/bash

set -e

COMPILEFILE="./main.go"

if [[ $1 ]]; then
	COMPILEFILE=$@
fi

ABSPATH="${PWD}"

TOOLCHAIN="$ABSPATH/toolchain/bin/arm-linux-gnueabi-"
#TOOLCHAIN="$ABSPATH/../vic-tool/arm-unknown-linux-gnueabi/bin/arm-unknown-linux-gnueabi-"

if [[ ! -f main.go ]]; then
	echo "This must be run in the vic-go directory"
	exit 1
fi

if [[ ! -f ./toolchain/bin/arm-linux-gnueabi-g++ ]]; then
	echo "Run the ./download-deps.sh (toolchain not found)"
	exit 1
fi

mkdir -p build

${TOOLCHAIN}g++ \
-w -shared \
-o build/librobot.so \
hacksrc/libs/spine.cpp \
hacksrc/spine_demo.cpp \
hacksrc/libs/utils.cpp \
hacksrc/libs/lcd.cpp \
hacksrc/lcd_demo.cpp \
hacksrc/libs/cam.cpp \
hacksrc/cam_demo.cpp \
-Iinclude -fPIC \
-O3 -mfpu=neon-vfpv4 -mfloat-abi=softfp \
-mcpu=cortex-a7 -flto -ffast-math \
-Ilibjpeg-turbo/include \
-std=c++11

${TOOLCHAIN}g++ \
-w -shared \
-o build/libjpeg_interface.so \
hacksrc/jpeg.cpp \
-Iinclude -fPIC \
-O3 -mfpu=neon-vfpv4 -mfloat-abi=softfp \
-mcpu=cortex-a7 -flto -ffast-math \
-std=c++11 -Ilibjpeg-turbo/include -fopenmp

${TOOLCHAIN}gcc \
-shared \
-o build/libanki-camera.so \
anki/platform/camera/vicos/camera_client/camera_client.c \
anki/platform/gpio/gpio.c \
anki/platform/camera/vicos/camera_client/log.c \
-I anki/platform/camera/vicos/camera_client/ \
-I anki/platform/camera/vicos/camera_client/inc \
-I anki \
-I anki/platform/gpio/inc \
-fPIC -lpthread \
-O3 -flto -ffast-math -mfpu=neon-vfpv4 -mfloat-abi=softfp

CC="${TOOLCHAIN}gcc -w -Lbuild" \
CGO_CFLAGS="-Iinclude -O3 -mfpu=neon-vfpv4 -mfloat-abi=softfp -mcpu=cortex-a7 -ffast-math -flto -Ilibjpeg-turbo/include" \
CGO_LDFLAGS="-ldl" \
GOARM=7 \
GOARCH=arm \
CGO_ENABLED=1 \
go build \
-ldflags '-w -s' \
-o build/main \
$COMPILEFILE

echo "Compiled successfully! Now you can send to the robot with ./send.sh <robotip> (expects ssh_root_key in user directory)"
