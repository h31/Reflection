package qBT

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func checkAndLog(e error, payload []byte) {
	if e != nil {
		tmpfile, _ := ioutil.TempFile("", "reflection")
		tmpfile.Write(payload)
		log.WithField("filename", tmpfile.Name()).Error("Saved payload in file")
		tmpfile.Close()

		panic(e)
	}
}

type Auth struct {
	LoggedIn bool
	Cookie   http.Cookie
}

type Hash string

type ID int

const RECENTLY_ACTIVE ID = -1

type Connection struct {
	addr         *url.URL
	client       *http.Client
	auth         Auth
	TorrentsList TorrentsList
}

type TorrentsList struct {
	useSync       bool
	items         map[Hash]*TorrentInfo
	activity      map[Hash]*time.Time
	rid           int
	mainDataCache MainData
	hashIds       map[ID]Hash
	hashIdMap     map[Hash]ID
	mutex sync.RWMutex
}

func (list *TorrentsList) AllItems() map[Hash]*TorrentInfo {
	return list.items
}

func (list *TorrentsList) GetActive() (resp map[Hash]*TorrentInfo) {
	const timeout = 60 * time.Second
	resp = make(map[Hash]*TorrentInfo)

	for _, item := range list.items {
		activity := list.activity[item.Hash]
		if activity != nil && time.Since(*activity) < timeout {
			resp[item.Hash] = item
		}
	}
	return
}

func (list *TorrentsList) ByID(id ID) *TorrentInfo {
	list.mutex.RLock()
	defer list.mutex.RUnlock()
	if hash, ok := list.hashIds[id]; ok {
		return list.items[hash]
	} else {
		return nil
	}
}

func (list *TorrentsList) ByHash(hash Hash) *TorrentInfo {
	list.mutex.RLock()
	defer list.mutex.RUnlock()
	if item, ok := list.items[hash]; ok {
		return item
	} else {
		return nil
	}
}

func (list *TorrentsList) IDByHash(hash Hash) *ID {
	list.mutex.RLock()
	defer list.mutex.RUnlock()
	if item, ok := list.hashIdMap[hash]; ok {
		return &item
	} else {
		return nil
	}
}

func (list *TorrentsList) ItemsNum() int {
	list.mutex.RLock()
	defer list.mutex.RUnlock()
	return len(list.hashIds)
}

func (q *Connection) Init(baseUrl string, client *http.Client, useSync bool) {
	q.TorrentsList.items = make(map[Hash]*TorrentInfo, 0)
	q.TorrentsList.activity = make(map[Hash]*time.Time)
	q.TorrentsList.hashIds = make(map[ID]Hash)
	q.TorrentsList.useSync = useSync

	apiAddr, _ := url.Parse("api/v2/")
	parsedBaseAddr, _ := url.Parse(baseUrl)
	q.addr = parsedBaseAddr.ResolveReference(apiAddr)
	q.client = client
}

func (q *Connection) IsLoggedIn() bool {
	return q.auth.LoggedIn
}

func (q *Connection) MakeRequestURLWithParam(path string, params map[string]string) string {
	if strings.HasPrefix(path, "/") {
		panic("Invalid API path: " + path)
	}
	parsedPath, err := url.Parse(path)
	check(err)
	u := q.addr.ResolveReference(parsedPath)
	if len(params) > 0 {
		query := u.Query()
		for key, value := range params {
			query.Set(key, value)
		}
		u.RawQuery = query.Encode()
	}

	return u.String()
}

func (q *Connection) MakeRequestURL(path string) string {
	return q.MakeRequestURLWithParam(path, map[string]string{})
}

func (q *Connection) UpdateTorrentListDirectly() (resp []TorrentsList) {
	params := map[string]string{}
	url := q.MakeRequestURLWithParam("torrents/info", params)
	torrents := q.DoGET(url)

	err := json.Unmarshal(torrents, &resp)
	checkAndLog(err, torrents)
	return
}

func (q *Connection) UpdateCachedTorrentsList() {
	torrentsList := &q.TorrentsList
	url := q.MakeRequestURLWithParam("sync/maindata", map[string]string{"rid": string(torrentsList.rid)})
	mainData := q.DoGET(url)

	err := json.Unmarshal(mainData, &torrentsList.mainDataCache)
	checkAndLog(err, mainData)
	torrentsList.rid = torrentsList.mainDataCache.Rid
	now := time.Now()
	for hash, rawTorrentData := range *torrentsList.mainDataCache.Torrents {
		torrentListItem, ok := torrentsList.items[hash]
		if !ok {
			torrentListItem = &TorrentInfo{}
			torrentsList.items[hash] = torrentListItem
		}
		err := json.Unmarshal(rawTorrentData, torrentListItem)
		checkAndLog(err, mainData)
		torrentListItem.Hash = hash
		torrentsList.activity[hash] = &now
	}
}

func (q *Connection) UpdateTorrentsList() {
	if q.TorrentsList.useSync {
		q.UpdateCachedTorrentsList()
	} else {
		q.UpdateTorrentListDirectly()
	}
	q.TorrentsList.UpdateIDs()
}

func (q *Connection) AddNewCategory(category string) {
	url := q.MakeRequestURLWithParam("torrents/createCategory", map[string]string{"category": category})
	q.DoGET(url)
}

func (q *Connection) GetPropsGeneral(hash string) (propGeneral PropertiesGeneral) {
	propGeneralURL := q.MakeRequestURLWithParam("torrents/properties", map[string]string{"hash": hash})
	propGeneralRaw := q.DoGET(propGeneralURL)

	err := json.Unmarshal(propGeneralRaw, &propGeneral)
	checkAndLog(err, propGeneralRaw)
	return
}

func (q *Connection) GetPropsTrackers(hash string) (trackers []PropertiesTrackers) {
	trackersURL := q.MakeRequestURLWithParam("torrents/trackers", map[string]string{"hash": hash})
	trackersRaw := q.DoGET(trackersURL)

	err := json.Unmarshal(trackersRaw, &trackers)

	checkAndLog(err, trackersRaw)
	return
}

func (q *Connection) GetPiecesStates(hash string) (pieces []byte) {
	piecesURL := q.MakeRequestURLWithParam("torrents/pieceStates", map[string]string{"hash": hash})
	piecesRaw := q.DoGET(piecesURL)

	err := json.Unmarshal(piecesRaw, &pieces)

	checkAndLog(err, piecesRaw)
	return
}

func (q *Connection) GetPreferences() (pref Preferences) {
	prefURL := q.MakeRequestURL("app/preferences")
	prefRaw := q.DoGET(prefURL)

	err := json.Unmarshal(prefRaw, &pref)
	checkAndLog(err, prefRaw)
	return
}

func (q *Connection) GetTransferInfo() (info TransferInfo) {
	infoURL := q.MakeRequestURL("transfer/info")
	infoRaw := q.DoGET(infoURL)

	err := json.Unmarshal(infoRaw, &info)
	checkAndLog(err, infoRaw)
	return
}

func (q *Connection) GetMainData() (info TransferInfo) {
	mainDataURL := q.MakeRequestURL("sync/maindata")
	mainDataRaw := q.DoGET(mainDataURL)

	err := json.Unmarshal(mainDataRaw, &info)
	checkAndLog(err, mainDataRaw)
	return
}

func (q *Connection) GetVersion() string {
	versionURL := q.MakeRequestURL("app/version")
	return string(q.DoGET(versionURL))
}

func (q *Connection) GetPropsFiles(hash string) (files []PropertiesFiles) {
	filesURL := q.MakeRequestURLWithParam("torrents/files", map[string]string{"hash": hash})
	filesRaw := q.DoGET(filesURL)

	err := json.Unmarshal(filesRaw, &files)
	checkAndLog(err, filesRaw)
	return
}

func (q *Connection) DoGET(url string) []byte {
	req, err := http.NewRequest("GET", url, nil)
	check(err)
	req.AddCookie(&q.auth.Cookie)

	resp, err := q.client.Do(req)
	check(err)
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	return data
}

func (q *Connection) DoPOST(url string, contentType string, body io.Reader) []byte {
	req, err := http.NewRequest("POST", url, body)
	check(err)
	req.Header.Set("Content-Type", contentType)
	req.AddCookie(&q.auth.Cookie)

	resp, err := q.client.Do(req)
	check(err)
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	return data
}

func (q *Connection) PostForm(url string, data url.Values) []byte {
	return q.DoPOST(url, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
}

func (q *Connection) Login(username, password string) bool {
	resp, err := http.PostForm(q.MakeRequestURL("auth/login"),
		url.Values{"username": {username}, "password": {password}})
	check(err)
	for _, value := range resp.Cookies() {
		if value != nil {
			cookie := *value
			if cookie.Name == "SID" {
				q.auth.LoggedIn = true
				q.auth.Cookie = cookie
				break
			}
		}
	}
	return q.auth.LoggedIn
}

func (q *Connection) SetToggleFlag(path string, hash Hash, newState bool) {
	item := q.TorrentsList.ByHash(hash)
	if item.Seq_dl != newState {
		q.PostForm(q.MakeRequestURL(path),
			url.Values{"hashes": {string(hash)}})
	}
	return
}

func (q *Connection) SetSequentialDownload(hash Hash, newState bool) {
	q.SetToggleFlag("torrents/toggleSequentialDownload", hash, newState)
}

func (q *Connection) SetFirstLastPieceFirst(hash Hash, newState bool) {
	q.SetToggleFlag("torrents/toggleFirstLastPiecePrio", hash, newState)
}

func (list *TorrentsList) UpdateIDs() {
	list.mutex.Lock()
	defer list.mutex.Unlock()

	keepHashes := make(map[ID]bool)
	addedCount := 0

	for hash, _ := range list.items {
		if index, exists := list.hashIdMap[hash]; exists {
			keepHashes[index] = true
		} else {
			var lastIndex = ID(len(list.hashIds))
			list.hashIdMap[hash] = lastIndex
			keepHashes[lastIndex] = true
			list.hashIds[lastIndex] = hash
			addedCount++
		}
	}

	if addedCount > 0 {
		log.WithField("num", addedCount).Info("Added new hashes to IDs table")
	}

	for id, hash := range list.hashIds {
		if _, exists := keepHashes[id]; !exists {
			log.WithField("hash", hash).Info("Hash disappeared from the torrent list")
			delete(list.hashIdMap, hash)
			delete(list.hashIds, id)
		}
	}
}
