# g15mpris

An G15Daemon LCD face to display what music you're playing, written in go

# Building

## Prerequisites

You need to have `g15daemon`, `libg15`, `libg15render` and their headers installed

ArchLinux: `sudo pacman -S g15daemon libg15 libg15render`

On others you probably need to compile it yourself, because it's **old**

## Building

```bash
git clone https://github.com/ragon000/g15mpris
cd g15mpris
go build
```

To make the binary avaliable to your whole system, run

```bash
sudo cp g15mpris /usr/local/bin
```


# TODO

- [ ] Scrolling text if the title/album is too long
- [ ] displaying how many seconds the music will play until end (maybe with a bar at the bottom)
- [ ] make the buttons do something, maybe repeat, random, spotify favorite (if possible)
- [ ] better design
