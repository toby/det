# Detergent

Detergent is a distributed protocol for searching and indexing the BitTorrent
network. It uses the `det` command line app to query information stored from
the Torrent DHT (magnet links).

## Install

Building requires `go` with module support (`go1.11beta2` at the time of this
writing).

```
git clone https://git.playgrub.com/toby/det.git
cd det
go1.11beta2 build
```

## Usage

To start you'll need to build up a database of Torrent metainfo by running
`det listen`. You can run other commands while det is listening.

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
