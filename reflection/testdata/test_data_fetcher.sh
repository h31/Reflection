#!/bin/sh

hash=842783e3005495d5d1637f5364b59343c7844707
num=2

curl -o torrent_${num}_list.json        http://localhost:8080/api/v2/torrents/info
curl -o torrent_${num}_properties.json  http://localhost:8080/api/v2/torrents/properties?hash=$hash
curl -o torrent_${num}_trackers.json    http://localhost:8080/api/v2/torrents/trackers?hash=$hash
curl -o torrent_${num}_piecestates.json http://localhost:8080/api/v2/torrents/pieceStates?hash=$hash
curl -o torrent_${num}_files.json       http://localhost:8080/api/v2/torrents/files?hash=$hash
curl -o torrent_${num}_peers.json       http://localhost:8080/api/v2/sync/torrentPeers?hash=$hash\&rid=0
