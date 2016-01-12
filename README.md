# Reflection

Reflection allows to use Transmission remote controls with qBittorrent.
It acts as a bridge between Transmission RPC and qBittorrent WebUI API.

Currently Reflection is able to:
* Show torrent list
* Show torrent details (files, trackers, statistics)
* Start, stop, delete torrents
* Add new torrents from a file, a magnet link of an HTTP/HTTPS link
* Choose which files should be downloaded

Compatibility:
* Reflection emulates a latest version of Transmission (2.84)
* Tested against qBittorrent 3.3.1, should work with 3.2.x too.
* Tested against Transmission Remote GUI, built-in Transmission Web UI, Torrnado client for Android

What features are not supported yet:
* Authentication
* Set torrent properties (download/upload speed, etc)
* Show and change torrent client settings
* Show peer table

Please note that some features are not implemented in qBittorrent WebUI API, so it's not possible to use them.

## Usage:

```base
go get main
go build main
./main
```

Binaries for some popular platforms (Windows, Linux, OS X) can be downloaded from a release page.

Use a `--help` flag to show settings. Default qBittorrent address is `http://localhost:8080/`.

Reflection works best with this qBittorrent settings (can be changed in ~/.config/qBittorrent/qBittorrent.conf):
* `WebUI\LocalHostAuth=false` - otherwise authentication error will happen
* `Downloads\StartInPause=true` - highly recommended for a stable work