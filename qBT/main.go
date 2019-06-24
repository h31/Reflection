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

const (
	INVALID_ID         ID = -2
	RECENTLY_ACTIVE_ID ID = -1
)

type Connection struct {
	addr         *url.URL
	client       *http.Client
	auth         Auth
	TorrentsList TorrentsList
}

type TorrentsList struct {
	useSync   bool
	items     map[Hash]*TorrentInfo
	activity  map[Hash]*time.Time
	rid       int
	hashIds   map[ID]Hash
	lastIndex ID
	mutex     sync.RWMutex
}

func (list *TorrentsList) AllItems() map[Hash]*TorrentInfo {
	return list.items
}

func (list *TorrentsList) Slice() []*TorrentInfo {
	list.mutex.RLock()
	defer list.mutex.RUnlock()

	result := make([]*TorrentInfo, 0, len(list.items))
	for _, item := range list.items {
		result = append(result, item)
	}
	return result
}

func (list *TorrentsList) AllIDs() []ID {
	result := make([]ID, list.ItemsNum())
	for i := 0; i < list.ItemsNum(); i++ {
		result[i] = ID(i + 1)
	}
	return result
}

func (list *TorrentsList) GetActive() (resp []*TorrentInfo) {
	const timeout = 60 * time.Second

	for _, item := range list.items {
		activity := list.activity[item.Hash]
		if activity != nil && time.Since(*activity) < timeout {
			resp = append(resp, item)
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

func (list *TorrentsList) ItemsNum() int {
	list.mutex.RLock()
	defer list.mutex.RUnlock()
	return len(list.hashIds)
}

func TorrentInfoListHashes(torrents []*TorrentInfo) []Hash {
	hashesStrings := make([]Hash, len(torrents))
	for i, torrent := range torrents {
		hashesStrings[i] = torrent.Hash
	}
	return hashesStrings
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

func (q *Connection) UpdateTorrentListDirectly() {
	torrents := make([]*TorrentInfo, 0)

	params := map[string]string{}
	url := q.MakeRequestURLWithParam("torrents/info", params)
	torrentsJSON := q.DoGET(url)

	err := json.Unmarshal(torrentsJSON, &torrents)
	checkAndLog(err, torrentsJSON)

	q.TorrentsList.items = make(map[Hash]*TorrentInfo)
	for _, torrent := range torrents {
		q.TorrentsList.items[torrent.Hash] = torrent
		torrent.Id = INVALID_ID
	}
}

func (q *Connection) UpdateCachedTorrentsList() (added, deleted []*TorrentInfo) {
	torrentsList := &q.TorrentsList
	url := q.MakeRequestURLWithParam("sync/maindata", map[string]string{"rid": string(torrentsList.rid)})
	mainData := q.DoGET(url)

	mainDataCache := MainData{}

	err := json.Unmarshal(mainData, &mainDataCache)
	checkAndLog(err, mainData)
	torrentsList.rid = mainDataCache.Rid
	now := time.Now()
	for _, deletedHash := range mainDataCache.Torrents_removed {
		deleted = append(deleted, torrentsList.items[deletedHash])
	}
	for hash, rawTorrentData := range *mainDataCache.Torrents {
		torrent, exists := torrentsList.items[hash]
		if !exists {
			torrent = &TorrentInfo{Id: INVALID_ID}
			torrentsList.items[hash] = torrent
			added = append(added, torrent)
		}
		err := json.Unmarshal(rawTorrentData, torrent)
		checkAndLog(err, mainData)
		torrent.Hash = hash
		torrentsList.activity[hash] = &now
	}
	return
}

func (q *Connection) UpdateTorrentsList() {
	q.TorrentsList.mutex.Lock()
	defer q.TorrentsList.mutex.Unlock()

	if q.TorrentsList.useSync {
		added, deleted := q.UpdateCachedTorrentsList()
		q.TorrentsList.UpdateIDsSynced(added, deleted)
	} else {
		q.UpdateTorrentListDirectly()
		q.TorrentsList.UpdateIDsFullRescan()
	}
}

func (q *Connection) AddNewCategory(category string) {
	url := q.MakeRequestURLWithParam("torrents/createCategory", map[string]string{"category": category})
	q.DoGET(url)
}

func (q *Connection) GetPropsGeneral(hash Hash) (propGeneral PropertiesGeneral) {
	propGeneralURL := q.MakeRequestURLWithParam("torrents/properties", map[string]string{"hash": string(hash)})
	propGeneralRaw := q.DoGET(propGeneralURL)

	err := json.Unmarshal(propGeneralRaw, &propGeneral)
	checkAndLog(err, propGeneralRaw)
	return
}

func (q *Connection) GetPropsTrackers(hash Hash) (trackers []PropertiesTrackers) {
	trackersURL := q.MakeRequestURLWithParam("torrents/trackers", map[string]string{"hash": string(hash)})
	trackersRaw := q.DoGET(trackersURL)

	err := json.Unmarshal(trackersRaw, &trackers)

	checkAndLog(err, trackersRaw)
	return
}

func (q *Connection) GetPiecesStates(hash Hash) (pieces []byte) {
	piecesURL := q.MakeRequestURLWithParam("torrents/pieceStates", map[string]string{"hash": string(hash)})
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

func (q *Connection) GetPropsFiles(hash Hash) (files []PropertiesFiles) {
	filesURL := q.MakeRequestURLWithParam("torrents/files", map[string]string{"hash": string(hash)})
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

func (q *Connection) PostWithHashes(path string, torrents []*TorrentInfo) {
	hashes := ConcatenateHashes(torrents)
	q.PostForm(q.MakeRequestURL(path), url.Values{"hashes": {hashes}})
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

func ConcatenateHashes(torrents []*TorrentInfo) string {
	hashesStrings := make([]string, len(torrents))
	for i, torrent := range torrents {
		hashesStrings[i] = string(torrent.Hash)
	}
	return strings.Join(hashesStrings, "|")
}

func (list *TorrentsList) UpdateIDsFullRescan() {
	addedCount := 0

	for id, hash := range list.hashIds {
		if torrent, exists := list.items[hash]; exists {
			torrent.Id = id
		} else {
			log.WithField("hash", hash).WithField("id", id).Info("Hash disappeared from the torrent list")
			delete(list.hashIds, id)
		}
	}

	for hash, torrent := range list.items {
		if torrent.Id == INVALID_ID {
			list.hashIds[list.lastIndex] = hash
			torrent.Id = list.lastIndex
			list.lastIndex++
			addedCount++
		}
	}

	if addedCount > 0 {
		log.WithField("num", addedCount).Info("Added new hashes to IDs table")
	}
}

func (list *TorrentsList) UpdateIDsSynced(added, deleted []*TorrentInfo) {
	addedCount := 0

	for _, torrent := range deleted {
		if _, exists := list.hashIds[torrent.Id]; exists {
			log.WithField("hash", torrent.Hash).WithField("id", torrent.Id).Info("Hash was removed from the torrent list")
			delete(list.hashIds, torrent.Id)
		}
	}

	for _, torrent := range added {
		if torrent.Id == INVALID_ID {
			list.hashIds[list.lastIndex] = torrent.Hash
			torrent.Id = list.lastIndex
			list.lastIndex++
			addedCount++
		}
	}

	if addedCount > 0 {
		log.WithField("num", addedCount).Info("Added new hashes to IDs table")
	}
}
