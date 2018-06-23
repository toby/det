# detergent

Distributed torrents.

## Install

The `anacrolix/torrent` and `anacrolix/dht` libs have recently changed
interface and break compatibility with `det`. Please use the following
installation procedure as a temporary fix.

```bash
go get git.playgrub.com/toby/det

cd $GOPATH/src/github.com/anacrolix/dht
git remote add det https://git.playgrub.com/toby/dht.git
git fetch det
git checkout det

cd $GOPATH/src/github.com/anacrolix/torrent
git remote add det https://git.playgrub.com/toby/torrent.git
git fetch det
git checkout det

cd $GOPATH/src/git.playgrub.com/toby/det
go clean; go build
```

## Usage

Detergent uses a cli called `det`. The longer you run this command in
`listen` mode, the better off you will be.

```
det command [command options] [arguments...]

COMMANDS:
     listen, l    Build torrent database from network
     search, s    Search resolved torrents
     resolve, r   Resolve a magnet URI
     popular, p   List top torrents of all time
     timeline, t  List most popular torrents found each day
     download, d  Download magnet URL
     seed         Seed file on torrent network
     peer         Respond to queries but don't listen for announces
     info, i      Show Detergent info
     help, h      Shows a list of commands or help for one command
```
