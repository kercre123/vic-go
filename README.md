# vic-go

A Go wrapper for [poc.vic-hack](https://github.com/torimos/poc.vic-hack).

## Install example program

On a Linux host machine,

```
cd ~
git clone https://github.com/kercre123/vic-go
cd vic-go
sudo ./download-deps.sh
./compile.sh
./send.sh vectorip
```

-   (replace vectorip with Vector's actual IP)
-   This expects the SSH key to be in the user directory (~/ssh_root_key).
-   If you get an error like `scp: Connection closed`, run ./send.sh with -O like `./send.sh vectorip -O`

-   The compile.sh script should give you a good sense of how you would need to compile your own program.
    -   A compatible toolchain. Ideally, we would statically compile so we could use the special CPU features of Vector's processor, but this is CGO which makes it difficult. So, we must use a timely toolchain. Newest one that works seems to be Linaro's 5.5.
        -   Go works perfectly with old toolchains. It just makes it harder to include stuff like GoCV.
    -   librobot.so must be seperate, as that code is written in C++, and CGO only compiles C code.

## Install example remote control program

```
cd ~
git clone https://github.com/kercre123/vic.go
cd vic-go
sudo ./download-deps.sh
COMPILE_WITH_JPEG=true ./compile.sh rc/rc.go
SEND_WEBROOT=true ./send.sh vectorip
```

-   (replace vectorip with Vector's actual IP)
-   This expects the SSH key to be in the user directory (~/ssh_root_key).
-   If you get an error like `scp: Connection closed`, run ./send.sh with -O like `./send.sh vectorip -O`

# Running

SSH into your bot,

```
systemctl stop anki-robot.target
cd /data/vic-go
./main
```

The default example will show the camera on the LCD, then do some body functions (recieving and transmitting).

The remote control example will host a site at http://vectorip:8888/


