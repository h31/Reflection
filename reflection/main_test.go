package main

import (
	"github.com/hekmon/transmissionrpc"
	log "github.com/sirupsen/logrus"
	"gopkg.in/h2non/gock.v1"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAdditionalLocationArguments(t *testing.T) {
	tables := []struct {
		input            string
		args             additionalArguments
		strippedLocation string
		err              error
	}{
		{"/home/user", additionalArguments{}, "/home/user", nil},
		{"/home/user/", additionalArguments{}, "/home/user/", nil},
		{"/home/user/+s", additionalArguments{sequentialDownload: ARGUMENT_TRUE}, "/home/user/", nil},
		{"/home/user+s", additionalArguments{sequentialDownload: ARGUMENT_TRUE}, "/home/user", nil},
		{"/home/user+data", additionalArguments{}, "/home/user+data", nil},
		{"/home/user+data+s", additionalArguments{sequentialDownload: ARGUMENT_TRUE}, "/home/user+data", nil},
		{"/home/user+s/", additionalArguments{}, "/home/user+s/", nil},
		{"/home/user/+f", additionalArguments{firstLastPiecesFirst: ARGUMENT_TRUE}, "/home/user/", nil},
		{"/home/user/+sf", additionalArguments{sequentialDownload: ARGUMENT_TRUE, firstLastPiecesFirst: ARGUMENT_TRUE},
			"/home/user/", nil},
		{"/home/user/+h", additionalArguments{skipChecking: ARGUMENT_TRUE}, "/home/user/", nil},
		{"/home/user/-s", additionalArguments{sequentialDownload: ARGUMENT_FALSE}, "/home/user/", nil},
		{"/home/user/-sh", additionalArguments{sequentialDownload: ARGUMENT_FALSE, skipChecking: ARGUMENT_FALSE}, "/home/user/", nil},
		{"C:\\Users\\+s\\", additionalArguments{}, "C:\\Users\\+s\\", nil},
		{"C:\\Users\\+s", additionalArguments{sequentialDownload: ARGUMENT_TRUE}, "C:\\Users\\", nil},
	}

	for _, table := range tables {
		args, location, err := parseAdditionalLocationArguments(table.input)
		if args != table.args || location != table.strippedLocation || err != table.err {
			t.Errorf("Input %s, expected (%+v, %s, %v), got: (%+v, %s, %v)", table.input,
				table.args, table.strippedLocation, table.err,
				args, location, err)
		}
	}
}

func TestWithStubs(t *testing.T) {
	const apiAddr = "http://localhost:8080"
	log.SetLevel(log.DebugLevel)

	defer gock.Off()

	gock.New(apiAddr).
		Get("/api/v2/torrents/info").
		Reply(200).
		File("torrent_list.json")

	gock.New(apiAddr).
		Post("/api/v2/auth/login").
		Reply(200).
		SetHeader("Set-Cookie", "SID=1")

	// 1
	gock.New(apiAddr).
		Get("/api/v2/torrents/properties").
		MatchParam("hash", "cf7da7ab4d4e6125567bd979994f13bb1f23dddd").
		Reply(200).
		File("torrent_properties.json")

	gock.New(apiAddr).
		Get("/api/v2/torrents/trackers").
		MatchParam("hash", "cf7da7ab4d4e6125567bd979994f13bb1f23dddd").
		Reply(200).
		File("torrent_trackers.json")

	gock.New(apiAddr).
		Get("/api/v2/torrents/pieceStates").
		MatchParam("hash", "cf7da7ab4d4e6125567bd979994f13bb1f23dddd").
		Reply(200).
		File("torrent_piecestates.json")

	gock.New(apiAddr).
		Get("/api/v2/torrents/files").
		MatchParam("hash", "cf7da7ab4d4e6125567bd979994f13bb1f23dddd").
		Reply(200).
		File("torrent_files.json")

	gock.New(apiAddr).
		Get("/api/v2/sync/torrentPeers").
		MatchParam("hash", "cf7da7ab4d4e6125567bd979994f13bb1f23dddd").
		MatchParam("rid", "0").
		Reply(200).
		File("torrent_peers.json")

	// 2
	gock.New(apiAddr).
		Get("/api/v2/torrents/properties").
		MatchParam("hash", "842783e3005495d5d1637f5364b59343c7844707").
		Reply(200).
		File("torrent_2_properties.json")

	gock.New(apiAddr).
		Get("/api/v2/torrents/trackers").
		MatchParam("hash", "842783e3005495d5d1637f5364b59343c7844707").
		Reply(200).
		File("torrent_2_trackers.json")

	gock.New(apiAddr).
		Get("/api/v2/torrents/pieceStates").
		MatchParam("hash", "842783e3005495d5d1637f5364b59343c7844707").
		Reply(200).
		File("torrent_2_piecestates.json")

	gock.New(apiAddr).
		Get("/api/v2/torrents/files").
		MatchParam("hash", "842783e3005495d5d1637f5364b59343c7844707").
		Reply(200).
		File("torrent_2_files.json")

	gock.New(apiAddr).
		Get("/api/v2/sync/torrentPeers").
		MatchParam("hash", "842783e3005495d5d1637f5364b59343c7844707").
		MatchParam("rid", "0").
		Reply(200).
		File("torrent_2_peers.json")

	client := &http.Client{Transport: &http.Transport{}}
	gock.InterceptClient(client)

	qBTConn.Init(apiAddr, client, false)
	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()
	defer server.CloseClientConnections()
	serverAddr := server.Listener.Addr().(*net.TCPAddr)
	println(serverAddr.IP.String())

	transmissionbt, err := transmissionrpc.New(serverAddr.IP.String(), "", "",
		&transmissionrpc.AdvancedConfig{Port: uint16(serverAddr.Port)})
	Check(err)
	torrents, err := transmissionbt.TorrentGetAll()
	Check(err)
	if len(torrents) != 2 {
		t.Fail()
	}
	if *torrents[0].Name != "ubuntu-18.04.2-desktop-amd64.iso" {
		t.Fail()
	}
	if *torrents[1].Name != "ubuntu-18.04.2-live-server-amd64.iso" {
		t.Fail()
	}
}
