-- name: create-completed-table
CREATE TABLE IF NOT EXISTS completed(
  infoHash,
  "index",
  unique(infoHash, "index") ON CONFLICT IGNORE
);

-- name: create-torrent-table
CREATE TABLE IF NOT EXISTS torrent(
  infoHash TEXT UNIQUE,
  name TEXT DEFAULT NULL,
  resolved_at DATE DEFAULT NULL,
  created_at DATE DEFAULT (strftime('%s', 'now')),
  announce_count INTEGER DEFAULT 0
);

-- name: create-announce-table
CREATE TABLE IF NOT EXISTS announce(
  infoHash TEXT,
  peerID TEXT,
  created_at DATE DEFAULT (strftime('%s', 'now')),
  unique(infoHash, peerID) ON CONFLICT IGNORE
);

-- name: create-search-table
CREATE VIRTUAL TABLE IF NOT EXISTS search_torrent
USING FTS4(name, infoHash);

-- name: create-torrent
INSERT INTO torrent (infoHash) VALUES (?)

-- name: create-torrent-search
INSERT INTO search_torrent (infoHash, name) VALUES (?,?)

-- name: create-announce
INSERT INTO announce (infoHash, peerID) VALUES (?,?)

-- name: update-announce-count
UPDATE torrent SET announce_count = announce_count + 1 WHERE infoHash = ?

-- name: set-torrent-meta
UPDATE torrent
SET name = ?, resolved_at = (strftime('%s', 'now'))
WHERE infohash = ?;

-- name: get-torrent
SELECT infoHash, name, created_at, resolved_at FROM torrent WHERE infoHash = ?

-- name: search-torrents
SELECT t.announce_count, t.infoHash, t.name, t.resolved_at, t.created_at
FROM search_torrent AS s
INNER JOIN torrent AS t ON s.infoHash = t.infoHash
WHERE s.name MATCH ?
ORDER BY t.announce_count DESC LIMIT ?

-- name: completed-exists
SELECT EXISTS(
  SELECT * FROM completed WHERE infoHash = ? AND "index" = ?
);

-- name: insert-completed
INSERT INTO completed (infoHash, "index") VALUES (?, ?)

-- name: delete-completed
DELETE FROM completed WHERE infoHash = ? AND "index" = ?

-- name: popular-torrents
SELECT announce_count, infoHash, name, created_at, resolved_at
FROM torrent
ORDER BY announce_count DESC LIMIT ?;

-- name: popular-torrents-day
SELECT announce_count, infoHash, name, created_at, resolved_at
FROM torrent
WHERE datetime(created_at, 'unixepoch') <= datetime('now', ?)
AND datetime(created_at, 'unixepoch') > datetime('now', ?)
ORDER BY announce_count DESC LIMIT ?;

-- name: total-torrents
SELECT count(*) FROM torrent

-- name: total-resolved
SELECT count(*) FROM torrent WHERE resolved_at IS NOT NULL

-- name: total-announces
SELECT count(*) FROM announce
