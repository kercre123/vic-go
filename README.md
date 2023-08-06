# vic-go

A Go wrapper for [poc.vic-hack](https://github.com/torimos/poc.vic-hack).

Vector installation instructions:

On a Linux host machine,

```
cd ~
git clone https://github.com/kercre123/vic-go
cd vic-go
sudo ./download-deps.sh
./compile.sh
./send.sh vectorip
```

(replace vectorip with Vector's actual IP)

That expects the SSH key to be in the user directory (~/ssh_root_key).

If you get an error like `scp: Connection closed`, run ./send.sh with -O like `./send.sh vectorip -O`

Then, to run it, SSH in and do:

```
systemctl stop anki-robot.target
/data/vic-go/main
```

The default example takes the touch sensor input. If it's being touched, the bot will raise the lift. If not touched, it will lower the lift.

Full spine communication is implemented, except for the proximity sensor.

Camera, screen, IMU, and speaker are in the works, but not functional yet.

