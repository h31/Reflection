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

var currentLogLevel = log.DebugLevel

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

func TestTorrentListing(t *testing.T) {
	const apiAddr = "http://localhost:8080"
	log.SetLevel(currentLogLevel)

	defer gock.Off()

	gock.New(apiAddr).
		Get("/api/v2/torrents/info").
		Reply(200).
		File("testdata/torrent_list.json")

	gock.New(apiAddr).
		Post("/api/v2/auth/login").
		Reply(200).
		SetHeader("Set-Cookie", "SID=1")

	// 1
	setUpMocks(apiAddr, "cf7da7ab4d4e6125567bd979994f13bb1f23dddd", "1")

	// 2
	setUpMocks(apiAddr, "842783e3005495d5d1637f5364b59343c7844707", "2")

	client := &http.Client{Transport: &http.Transport{}}
	gock.InterceptClient(client)

	qBTConn.Init(apiAddr, client, false)
	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()
	defer server.CloseClientConnections()
	serverAddr := server.Listener.Addr().(*net.TCPAddr)

	transmissionbt, err := transmissionrpc.New(serverAddr.IP.String(), "", "",
		&transmissionrpc.AdvancedConfig{Port: uint16(serverAddr.Port)})
	Check(err)
	torrents, err := transmissionbt.TorrentGetAll()
	Check(err)
	if len(torrents) != 2 {
		t.Error("Number of torrents is not equal to 2")
	}
	if *torrents[0].Name != "ubuntu-18.04.2-desktop-amd64.iso" {
		t.Error("Unexpected torrent 0")
	}
	if *torrents[1].Name != "ubuntu-18.04.2-live-server-amd64.iso" {
		t.Error("Unexpected torrent 1")
	}
}

func TestTorrentListingRepeated(t *testing.T) {
	prevLogLevel := currentLogLevel
	currentLogLevel = log.ErrorLevel
	for i := 0; i < 50; i++ {
		TestTorrentListing(t)
	}
	currentLogLevel = prevLogLevel
}

func TestSyncing(t *testing.T) {
	const apiAddr = "http://localhost:8080"
	log.SetLevel(currentLogLevel)

	defer gock.Off()

	gock.New(apiAddr).
		Post("/api/v2/auth/login").
		Reply(200).
		SetHeader("Set-Cookie", "SID=1")

	gock.New(apiAddr).
		Get("/api/v2/sync/maindata").
		MatchParam("rid", "0").
		HeaderPresent("Cookie").
		Reply(200).
		File("testdata/sync_initial.json")

	gock.New(apiAddr).
		Get("/api/v2/sync/maindata").
		MatchParam("rid", "1").
		HeaderPresent("Cookie").
		Reply(200).
		File("testdata/sync_1.json")

	gock.New(apiAddr).
		Get("/api/v2/sync/maindata").
		MatchParam("rid", "2").
		HeaderPresent("Cookie").
		Reply(200).
		File("testdata/sync_2.json")

	gock.New(apiAddr).
		Get("/api/v2/sync/maindata").
		MatchParam("rid", "3").
		HeaderPresent("Cookie").
		Reply(200).
		File("testdata/sync_3.json")

	gock.New(apiAddr).
		Get("/api/v2/sync/maindata").
		MatchParam("rid", "4").
		HeaderPresent("Cookie").
		Reply(200).
		File("testdata/sync_4.json")

	// 1
	setUpMocks(apiAddr, "cf7da7ab4d4e6125567bd979994f13bb1f23dddd", "1")

	// 2
	setUpMocks(apiAddr, "842783e3005495d5d1637f5364b59343c7844707", "2")

	// 3
	setUpMocks(apiAddr, "7a1448be6d15bcde08ee9915350d0725775b73a3", "3")

	client := &http.Client{Transport: &http.Transport{}}
	gock.InterceptClient(client)

	qBTConn.Init(apiAddr, client, true)
	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()
	defer server.CloseClientConnections()
	serverAddr := server.Listener.Addr().(*net.TCPAddr)

	transmissionbt, err := transmissionrpc.New(serverAddr.IP.String(), "", "",
		&transmissionrpc.AdvancedConfig{Port: uint16(serverAddr.Port)})
	Check(err)
	torrents, err := transmissionbt.TorrentGetAll()
	Check(err)
	if len(torrents) != 2 {
		t.Error("Number of torrents is not equal to 2")
	}
	if *torrents[0].Name != "ubuntu-18.04.2-live-server-amd64.iso" {
		t.Error("Unexpected torrent 0")
	}
	if *torrents[1].Name != "ubuntu-18.04.2-desktop-amd64.iso" {
		t.Error("Unexpected torrent 1")
	}

	torrents, err = transmissionbt.TorrentGetAll()
	Check(err)
	if len(torrents) != 2 {
		t.Error("Number of torrents is not equal to 2")
	}

	torrents, err = transmissionbt.TorrentGetAll()
	Check(err)
	if len(torrents) != 3 {
		t.Error("Number of torrents is not equal to 3")
	}
	if *torrents[2].Name != "xubuntu-18.04.2-desktop-amd64.iso" {
		t.Error("Unexpected torrent name")
	}

	torrents, err = transmissionbt.TorrentGetAll()
	Check(err)
	if len(torrents) != 3 {
		t.Error("Number of torrents is not equal to 3")
	}

	torrents, err = transmissionbt.TorrentGetAll()
	Check(err)
	if len(torrents) != 2 {
		t.Error("Number of torrents is not equal to 2")
	}
}

func TestSyncingRepeated(t *testing.T) {
	prevLogLevel := currentLogLevel
	currentLogLevel = log.ErrorLevel
	for i := 0; i < 50; i++ {
		TestSyncing(t)
	}
	currentLogLevel = prevLogLevel
}

func setUpMocks(apiAddr string, hash string, name string) {
	gock.New(apiAddr).
		Get("/api/v2/torrents/properties").
		MatchParam("hash", hash).
		Persist().
		Reply(200).
		File("testdata/torrent_" + name + "_properties.json")
	gock.New(apiAddr).
		Get("/api/v2/torrents/trackers").
		MatchParam("hash", hash).
		Persist().
		Reply(200).
		File("testdata/torrent_" + name + "_trackers.json")
	gock.New(apiAddr).
		Get("/api/v2/torrents/pieceStates").
		MatchParam("hash", hash).
		Persist().
		Reply(200).
		File("testdata/torrent_" + name + "_piecestates.json")
	gock.New(apiAddr).
		Get("/api/v2/torrents/files").
		MatchParam("hash", hash).
		Persist().
		Reply(200).
		File("testdata/torrent_" + name + "_files.json")
	gock.New(apiAddr).
		Get("/api/v2/sync/torrentPeers").
		MatchParam("hash", hash).
		MatchParam("rid", "0").
		Persist().
		Reply(200).
		File("testdata/torrent_" + name + "_peers.json")
}
