# Detergent

P2P search, discovery and sharing on the BitTorrent network.

## What is it?

Detergent is a new type of BitTorrent client that includes search and discovery
backed entirely by P2P. There are no central servers. This stands in contrast
to traditional BitTorrent indexes like The Pirate Bay. Users run a Detergent
node on their own machine and interact with the BitTorrent network locally.

## Installing

Detergent requires `go` 1.11+ to build.

```
git clone https://github.com/toby/det.git
cd det
go build
```

## Usage

Deteregent is still very early in the development process. Not all
functionality is currently available (see *Functional Roadmap*). Mainly, the
distributed search and trending features don't exist yet, so `det` requires
some time to build up a local database.

### Listening

Detergent uses a command line tool named `det` to build and query the Torrent
database. The `det` tool has many commands but to get started you have to build
your local Torrent database by listening:

`./det listen`

This will run `det` in listen mode and watch for Torrents that people are
sharing on the network. You should start to see Torrent file names log to the
console. You can stop listening by typing **Ctrl+C**.

### Querying

There are a few ways to query the local Torrent database. Searching looks for
matching Torrents using a full-text search:

`./det search TERM`

You can view the most popular Torrents since your first listen (based on
Announces in your DHT) with:

`./det popular`

There is also a timeline view with most popular Torrents by day:

`./det timeline`

The query commands can include a `limit` argument to specify the number of
desired results:

`./det popular --limit=1000`

Overall system stats can be displayed with:

`./det info`

### Torrenting

While primitive in its current state, `det` does offer basic Torrent functionality:

```
./det download MAGNETURL
./det resolve MAGNET URL
```

## Functional Roadmap

- [x] Command line interface
- [x] Seed Torrents
- [x] Download Torrents
- [x] Store and index announces from BitTorrent DHT
- [x] Resolve and store magnet url Torrent metadata
- [x] Search Torrent metadata stored on Detergent peer
- [x] Show popular and trending Torrents
- [x] Detergent peer discovery
- [ ] Distributed searching
- [ ] Web interface
- [ ] Distributed trending
- [ ] Content publication
- [ ] Content curation and promotion

### Command line interface

The `det` command provides a CLI for accessing Detergent functionality. While
functional, there is work to be done to clean up output and make it easier to
pipe into other commands.

### Seed Torrents

Some Detergent functionality requires the ability to seed files and arbitrary
byte slices. This is currently supported from the CLI and in the code. The
`torrent/storage` interface from [anacrolix/torrent](https://github.com/anacrolix/torrent)
is implemented for read-only `[]byte` slices in [bytes.go](https://github.com/toby/det/blob/master/bytes.go).

### Download Torrents

As you can imagine Torrents can be downloaded using `det`. This is provided for
user convenience and required for some internal Detergent functionality.

### Store and index announces from BitTorrent DHT

At the heart of Detergent is the BitTorrent DHT best described by
[BEP-5](http://www.bittorrent.org/beps/bep_0005.html). At a high level, each
Detergent peer acts as a functional DHT node that also indexes every Torrent
announce that it encounters. Announces are stored individually in a SQLite
database and provide sorting order for the trending functionality.

### Resolve and store magnet url Torrent metadata

Torrent announces contain an infohash for the Torrent but don't provide any
metadata about the files provided by the Torrent (title, file names...). Each
infohash discovered on the DHT must be resolved on the Torrent network to
retrieve user readable information. Detergent has naive parallelization of
resolving but has much room to improve on performance.

### Search Torrent metadata stored on Detergent peer

Torrent metadata is stored locally in the SQLite database with
[FTS4](https://www.sqlite.org/fts3.html) full text indexing. Currently only the
top level filename is indexed. Soon all files in the Torrent should be indexed.

### Show popular and trending Torrents

Both `det popular` and `det timeline` provide *very* basic trending analysis.
The queries themselves are in [db.sql](https://github.com/toby/det/blob/master/sql/db.sql).

### Detergent peer discovery

Detergent uses an experimental Torrent based peer discovery mechanism. The goal
is to link Detergent peers together to provide a network for distributed search
and trending. The process for peer discovery works roughly like this:

1. Detergent peers seed a deterministic `detergent.json` file that contains a protocol string.
2. Each Detergent peer also shares a deterministic `PEER_ID.json` file containing its Torrent peer id.
3. Peer IDs and IP addresses of other Torrent clients sharing `detergent.json` are noted.
4. `PEER_ID.json` files are constructed and seeded for each potential Detergent peer sharing `detergent.json`.
5. Any `PEER_ID.json` file that can be downloaded should represent a peer that implements the Detergent protocol.

Further details are available in [discovery.go](https://github.com/toby/det/blob/master/discovery.go).

### Distributed searching

Once other Detergent peers are known, searches can be propagated outward
through the network with results returned to the querying peer. Detergent peers
should search their own local database first. The exact mechanism for
distributed search is TBD but will probably include multiple hops through the
peer network.

### Web interface

Detergent should have a web interface that provides search and trending.
Ideally media will be streamable in the web view as that's supported by the
underlying Torrent library.

In addition to streaming media, text content including Markdown should be
rendered and displayed in the client. With the addition of content publication
and curation, this should provide a powerful platform for distributed
communication.

### Distributed trending

More challenging than distributed search, trending should include aggregate
information from as many Detergent peers as possible. Challenges include the
oversampling of infohashes "near" a node ID per each Detergent client, how to
dedupe identical announces seen by multiple peers and host of other unforeseen
issues that will surely arise.


### Content publication

Detergent should include a first-class mechanism for authoring and publishing
text content in the form of Markdown. BitTorrent has historically been very
media heavy but can just as easily be used to share text content. By focusing
on discovery and search, Detergent should provide a natural platform on which
to publish Markdown based posts.

### Content curation and promotion

In addition to trending and search, Detergent should allow users to curate,
organize and promote content. Using techniques based on
[key.run](https://github.com/toby/keyrun), infohashes can be organized
into many namespaces and sorted by cryptocurrency transaction amounts. By
leveraging blockchain based namespacing and encoding, there's an opportunity to
build up durable indexes of infohashes that are manually curated.

## Known Limitations

* Top level filenames are the only metadata indexed
* Long wait for indexes to build up (distributed search will help)
* Poor CLI output formatting
* Many other things, this is very early
