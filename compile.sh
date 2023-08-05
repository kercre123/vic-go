
if [[ ! -f main.go ]]; then
	echo "This must be run in the vic-go directory"
	exit 1
fi

if [[ ! -f ./toolchain/bin/arm-linux-gnueabi-g++ ]]; then
	echo "Run the ./download-deps.sh (toolchain not found)"
fi

mkdir -p build

/home/kerigan/toolchain/vicgcc/bin/arm-linux-gnueabi-g++ \
-w -shared \
-o build/librobot.so \
hacksrc/libs/spine.cpp \
hacksrc/spine_demo.cpp \
hacksrc/libs/utils.cpp \
-Iinclude -fPIC

CC="/home/kerigan/toolchain/vicgcc/bin/arm-linux-gnueabi-gcc -w -Lbuild" \
CFLAGS="-Iinclude -fPIC -Lbuild" \
GOARM=7 \
GOARCH=arm \
CGO_ENABLED=1 \
go build \
-ldflags '-w -s' \
-o build/main \
main.go

echo "Compiled successfully! Now you can send to the robot with ./send.sh <robotip> (expects ssh_root_key in user directory)"
