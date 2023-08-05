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
