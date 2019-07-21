package server

const (
	sqlCreateTorrentTable = `CREATE TABLE IF NOT EXISTS torrent(
  infoHash TEXT UNIQUE,
  name TEXT DEFAULT NULL,
  resolved_at DATE DEFAULT NULL,
  created_at DATE DEFAULT (strftime('%s', 'now')),
  announce_count INTEGER DEFAULT 0,
  unique(infoHash) ON CONFLICT IGNORE)`

	sqlCreateAnnounceTable = `CREATE TABLE IF NOT EXISTS announce(
  infoHash TEXT,
  peerID TEXT,
  created_at DATE DEFAULT (strftime('%s', 'now')),
  unique(infoHash, peerID) ON CONFLICT IGNORE)`

	sqlCreateSearchTable = `CREATE VIRTUAL TABLE IF NOT EXISTS search_torrent
USING FTS4(infoHash PRIMARY KEY, name TEXT)`

	sqlCreateTorrent = `INSERT INTO torrent (infoHash) VALUES (?)`

	sqlCreateTorrentSearch = `INSERT INTO search_torrent (infoHash, name) VALUES (?,?)`

	sqlCreateAnnounce = `INSERT INTO announce (infoHash, peerID) VALUES (?,?)`

	sqlUpdateAnnounceCount = `UPDATE torrent SET announce_count = announce_count + 1 WHERE infoHash = ?`

	sqlSetTorrentMeta = `UPDATE torrent
SET name = ?, resolved_at = (strftime('%s', 'now'))
WHERE infohash = ?`

	sqlGetTorrent = `SELECT announce_count, infoHash, name, created_at, resolved_at
FROM torrent WHERE infoHash = ?`

	sqlSearchTorrents = `SELECT t.announce_count, t.infoHash, t.name, t.created_at, t.resolved_at
FROM search_torrent AS s
INNER JOIN torrent AS t ON s.infoHash = t.infoHash
WHERE s.name MATCH ?
ORDER BY t.announce_count DESC LIMIT ?`

	sqlPopularTorrents = `SELECT announce_count, infoHash, name, created_at, resolved_at
FROM torrent
ORDER BY announce_count DESC LIMIT ?;`

	sqlPopularTorrentsDay = `SELECT announce_count, infoHash, name, created_at, resolved_at
FROM torrent
WHERE datetime(created_at, 'unixepoch') <= datetime('now', ?)
AND datetime(created_at, 'unixepoch') > datetime('now', ?)
ORDER BY announce_count DESC LIMIT ?`

	sqlTotalTorrents = `SELECT count(*) FROM torrent`

	sqlTotalResolved = `SELECT count(*) FROM torrent WHERE resolved_at IS NOT NULL`

	sqlTotalAnnounces = `SELECT count(*) FROM announce`
)
