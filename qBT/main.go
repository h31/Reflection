package qBT

import (
	"encoding/json"
	"github.com/iancoleman/orderedmap"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strconv"
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
	INVALID_ID ID = -1
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

type TorrentInfoList []*TorrentInfo

func (torrents TorrentInfoList) Len() int {
	return len(torrents)
}
func (torrents TorrentInfoList) Swap(i, j int) {
	torrents[i], torrents[j] = torrents[j], torrents[i]
}

func (torrents TorrentInfoList) Less(i, j int) bool {
	return torrents[i].Id < torrents[j].Id
}

func (torrents TorrentInfoList) ConcatenateHashes() string {
	hashesStrings := make([]string, len(torrents))
	for i, torrent := range torrents {
		hashesStrings[i] = string(torrent.Hash)
	}
	return strings.Join(hashesStrings, "|")
}

func (torrents TorrentInfoList) Hashes() []Hash {
	hashesStrings := make([]Hash, len(torrents))
	for i, torrent := range torrents {
		hashesStrings[i] = torrent.Hash
	}
	return hashesStrings
}

func (list *TorrentsList) AllItems() map[Hash]*TorrentInfo {
	return list.items
}

func (list *TorrentsList) Slice() TorrentInfoList {
	list.mutex.RLock()
	defer list.mutex.RUnlock()

	result := make(TorrentInfoList, 0, len(list.items))
	for _, item := range list.items {
		result = append(result, item)
	}
	sort.Sort(result)
	return result
}

func (list *TorrentsList) GetActive() (resp TorrentInfoList) {
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

func (q *Connection) Init(baseUrl string, client *http.Client, useSync bool) {
	q.TorrentsList.items = make(map[Hash]*TorrentInfo, 0)
	q.TorrentsList.activity = make(map[Hash]*time.Time)
	q.TorrentsList.hashIds = make(map[ID]Hash)
	q.TorrentsList.useSync = useSync
	q.TorrentsList.rid = 0

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

func (q *Connection) UpdateTorrentListDirectly() TorrentInfoList {
	torrents := make(TorrentInfoList, 0)

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
	return torrents
}

func (q *Connection) UpdateCachedTorrentsList() (added, deleted TorrentInfoList) {
	torrentsList := &q.TorrentsList
	url := q.MakeRequestURLWithParam("sync/maindata", map[string]string{"rid": strconv.Itoa(torrentsList.rid)})
	mainData := q.DoGET(url)

	mainDataCache := MainData{}

	err := json.Unmarshal(mainData, &mainDataCache)
	checkAndLog(err, mainData)

	torrentsList.rid = mainDataCache.Rid
	now := time.Now()
	for _, deletedHash := range mainDataCache.Torrents_removed {
		deleted = append(deleted, torrentsList.items[deletedHash])
		delete(torrentsList.items, deletedHash)
	}

	if mainDataCache.Torrents != nil {
		orderedTorrentsMap := orderedmap.New()

		err = json.Unmarshal(*mainDataCache.Torrents, &orderedTorrentsMap)
		checkAndLog(err, *mainDataCache.Torrents)

		nativeTorrentsMap := make(map[Hash]*json.RawMessage)

		err = json.Unmarshal(*mainDataCache.Torrents, &nativeTorrentsMap)
		checkAndLog(err, *mainDataCache.Torrents)

		for _, hashString := range orderedTorrentsMap.Keys() {
			hash := Hash(hashString)
			torrent, exists := torrentsList.items[hash]
			if !exists {
				torrent = &TorrentInfo{Id: INVALID_ID}
				torrentsList.items[hash] = torrent
				added = append(added, torrent)
			}
			err := json.Unmarshal(*nativeTorrentsMap[hash], torrent)
			checkAndLog(err, mainData)
			torrent.Hash = hash
			torrentsList.activity[hash] = &now
		}
	}

	return
}

func (q *Connection) UpdateTorrentsList() {
	q.TorrentsList.mutex.Lock()
	defer q.TorrentsList.mutex.Unlock()

	if q.TorrentsList.useSync {
		added, deleted := q.UpdateCachedTorrentsList()
		q.TorrentsList.DeleteIDsSync(deleted)
		q.TorrentsList.UpdateIDs(added)
	} else {
		added := q.UpdateTorrentListDirectly()
		q.TorrentsList.DeleteIDsFullRescan()
		q.TorrentsList.UpdateIDs(added)
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

func (q *Connection) PostWithHashes(path string, torrents TorrentInfoList) {
	hashes := torrents.ConcatenateHashes()
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

func (list *TorrentsList) DeleteIDsSync(deleted TorrentInfoList) {
	for _, torrent := range deleted {
		if _, exists := list.hashIds[torrent.Id]; exists {
			log.WithField("hash", torrent.Hash).WithField("id", torrent.Id).Info("Hash was removed from the torrent list")
			delete(list.hashIds, torrent.Id)
		}
	}
}

func (list *TorrentsList) DeleteIDsFullRescan() {
	for id, hash := range list.hashIds {
		if torrent, exists := list.items[hash]; exists {
			torrent.Id = id
		} else {
			log.WithField("hash", hash).WithField("id", id).Info("Hash disappeared from the torrent list")
			delete(list.hashIds, id)
		}
	}
}

func (list *TorrentsList) UpdateIDs(added TorrentInfoList) {
	for _, torrent := range added {
		if torrent.Id == INVALID_ID {
			list.hashIds[list.lastIndex] = torrent.Hash
			torrent.Id = list.lastIndex
			log.WithField("hash", torrent.Hash).WithField("id", list.lastIndex).Info("Torrent got assigned ID")
			list.lastIndex++
		}
	}
}
