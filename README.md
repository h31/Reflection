# Reflection

Reflection allows to use Transmission remote controls with qBittorrent.
It acts as a bridge between Transmission RPC and qBittorrent WebUI API.

Currently Reflection is able to:
* Show torrent list
* Show torrent details (files, trackers, statistics)
* Start, stop, delete torrents
* Add new torrents from a file, a magnet link of an HTTP/HTTPS link
* Choose which files should be downloaded
* Change destination directory
* Show actual free space
* Show peer table

Compatibility:
* Reflection emulates the latest version of Transmission (2.84)
* Requires at least qBittorrent 4.1.0. Older versions of Reflection also support previous qBittorrent versions.
* Tested against Transmission Remote GUI, built-in Transmission Web UI, Torrnado client for Android, Transmission-Qt and Transmission Remote by Yury Polek. Please fill an issue if you experience an incompatibility with some client.

What features are not supported yet:
* Set torrent properties (download/upload speed, etc)
* Show and change torrent client settings

Please note that both qBittorrent and Transmission have some unique features.
For example, some torrent properties such as a private flag are not exposed by qBittorrent.
In case Transmission clients request such information, Reflection responds with a predefined template data. Template values are stored in src/transmission/templates.go.

Please set qBitorrent to English language for best experience! You can set this in qBittorrent.conf under Preferences section:

```
[Preferences]
...
General\Locale=en_US
```

## Usage:

```bash
mkdir Reflection
cd Reflection/
export GOPATH=$(pwd)
go get github.com/h31/Reflection/reflection
./bin/reflection
```

Use a `--help` flag to show settings. Default qBittorrent address is `http://localhost:8080/`.
