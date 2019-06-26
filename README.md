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

## Compatibility
* Reflection emulates the latest version of Transmission (2.94)
* Requires at least qBittorrent 4.1.0.
* Tested against Transmission Remote GUI, built-in Transmission Web UI, Torrnado client for Android, Transmission-Qt and Transmission Remote by Yury Polek. Please fill an issue if you experience an incompatibility with any client.

What features are not supported yet:
* Setting torrent properties (download/upload speed, etc)
* Showing and changing torrent client settings

## qBittorrent and Transmission-specific options

Please note that both qBittorrent and Transmission have some unique features.
For example, some torrent properties such as a private flag are not exposed by qBittorrent.
In case Transmission clients request such information, Reflection responds with predefined template data. Template values are stored in [transmission/templates.go](https://github.com/h31/Reflection/blob/master/transmission/templates.go).

To enable some qBittorrent-specific options, use a "Specify torrent location" command in your Transmission GUI
and append a special flag to the path. You can also specify such paths when adding a new torrent.
For example, if download directory is `/home/user/`, those paths can be used:
* `/home/user/+s` to enable sequential download
* `/home/user/+f` to download first and last pieces first 
* `/home/user/+h` to skip hash checking when adding torrent.

It is possible to combine several commands, i.e. `/home/user/+sf`. Use `-` sign instead of `+` to disable an option.
 f your want to disable command processing and treat a path just as a path, end it with `/`, i.e. `/home/user/my+path+s/`.

## Usage:

```bash
mkdir Reflection
cd Reflection/
export GOPATH=$(pwd)
go get github.com/h31/Reflection/reflection
./bin/reflection
```

Use a `--help` flag to show settings. Default qBittorrent address is `http://localhost:8080/`.
