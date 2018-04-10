# detergent

Distributed torrents.

## Install

```bash
$ go get git.playgrub.com/toby/det
```

## Usage

Detergent uses a cli called `det`. The longer you run this command in
`listen` mode, the better off you will be.

```
det [global options] command [command options] [arguments...]

COMMANDS:
     listen, l   Build torrent database from network
     search, s   Search resolved torrents
     resolve, r  Resolve a magnet hash
     popular, p  List popular torrents
     info, i     Show Detergent info
     help, h     Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h           show help
   --print-version, -V  print only the version
```
