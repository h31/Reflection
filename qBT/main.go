package qBT

import (
	"encoding/json"
	"errors"
	log "github.com/Sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
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
	Required bool
	LoggedIn bool
	Cookie   http.Cookie

	Username string
	Password string
}

type Connection struct {
	Addr string
	// Hash to ID map. Array index is an ID
	lastId   int
	HashIds  []string
	Tr       *http.Transport
	Client   *http.Client
	Auth     Auth
	MainData MainData
}

func (q *Connection) MakeRequestURLWithParam(path string, params map[string]string) string {
	u, err := url.Parse(q.Addr)
	check(err)
	u.Path = path
	query := u.Query()
	for key, value := range params {
		query.Set(key, value)
	}
	u.RawQuery = query.Encode()

	return u.String()
}

func (q *Connection) MakeRequestURL(path string) string {
	return q.MakeRequestURLWithParam(path, map[string]string{})
}

func (q *Connection) getTorrentListDirect() (resp []TorrentsList) {
	url := q.MakeRequestURLWithParam("/query/torrents", map[string]string{"sort": "hash"})
	torrents := q.DoGET(url)

	err := json.Unmarshal(torrents, &resp)
	checkAndLog(err, torrents)
	return
}

func (q *Connection) getTorrentListCached() (resp []TorrentsList) {
	url := q.MakeRequestURL("/sync/maindata")
	mainData := q.DoGET(url)

	err := json.Unmarshal(mainData, &q.MainData)
	checkAndLog(err, mainData)
	for hash, torrentData := range *q.MainData.Torrents {
		torrentData.Hash = hash
		resp = append(resp, torrentData)
	}
	return
}

func (q *Connection) GetTorrentList() (resp []TorrentsList, newHashes []string) {
	resp = q.getTorrentListDirect()

	if q.GetHashNum() == 0 || q.GetHashNum() != len(resp) {
		newHashes = q.FillIDs(resp)
		log.Debug("Filling IDs table, new size: ", q.GetHashNum())
	}
	return
}

func (q *Connection) GetPropsGeneral(id int) (propGeneral PropertiesGeneral) {
	propGeneralURL := q.MakeRequestURL("/query/propertiesGeneral/" + q.GetHashForId(id))
	propGeneralRaw := q.DoGET(propGeneralURL)

	err := json.Unmarshal(propGeneralRaw, &propGeneral)
	checkAndLog(err, propGeneralRaw)
	return
}

func (q *Connection) GetPropsTrackers(id int) (trackers []PropertiesTrackers) {
	trackersURL := q.MakeRequestURL("/query/propertiesTrackers/" + q.GetHashForId(id))
	trackersRaw := q.DoGET(trackersURL)

	err := json.Unmarshal(trackersRaw, &trackers)

	checkAndLog(err, trackersRaw)
	return
}

func (q *Connection) GetPreferences() (pref Preferences) {
	prefURL := q.MakeRequestURL("/query/preferences")
	prefRaw := q.DoGET(prefURL)

	err := json.Unmarshal(prefRaw, &pref)
	checkAndLog(err, prefRaw)
	return
}

func (q *Connection) GetTransferInfo() (info TransferInfo) {
	infoURL := q.MakeRequestURL("/query/transferInfo")
	infoRaw := q.DoGET(infoURL)

	err := json.Unmarshal(infoRaw, &info)
	checkAndLog(err, infoRaw)
	return
}

func (q *Connection) GetMainData() (info TransferInfo) {
	mainDataURL := q.MakeRequestURL("/sync/maindata")
	mainDataRaw := q.DoGET(mainDataURL)

	err := json.Unmarshal(mainDataRaw, &info)
	checkAndLog(err, mainDataRaw)
	return
}

func (q *Connection) GetVersion() string {
	versionURL := q.MakeRequestURL("/version/qbittorrent")
	return string(q.DoGET(versionURL))
}

func (q *Connection) GetPropsFiles(id int) (files []PropertiesFiles) {
	filesURL := q.MakeRequestURL("/query/propertiesFiles/" + q.GetHashForId(id))
	filesRaw := q.DoGET(filesURL)

	err := json.Unmarshal(filesRaw, &files)
	checkAndLog(err, filesRaw)
	return
}

func (q *Connection) DoGET(lurl string) []byte {
	req, err := http.NewRequest("GET", lurl, nil)
	if q.Auth.Required {
		req.AddCookie(&q.Auth.Cookie)
	}
	check(err)
	resp, err := q.Client.Do(req)
	check(err)
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	return data
}

func (q *Connection) DoPOST(url string, contentType string, body io.Reader) []byte {
	req, err := http.NewRequest("POST", url, body)
	check(err)
	req.Header.Set("Content-Type", contentType)
	if q.Auth.Required {
		req.AddCookie(&q.Auth.Cookie)
	}

	resp, err := q.Client.Do(req)
	check(err)
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	return data
}

func (q *Connection) PostForm(url string, data url.Values) []byte {
	return q.DoPOST(url, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
}

func (q *Connection) GetHashForId(id int) string {
	if len(q.HashIds) >= id {
		return q.HashIds[id-1]
	} else {
		return "a"
	}
}

func (q *Connection) GetHashNum() int {
	return len(q.HashIds)
}

func (q *Connection) GetIdOfHash(hash string) (int, error) {
	for index, value := range q.HashIds {
		if value == hash {
			return index + 1, nil
		}
	}
	return 0, errors.New("No such hash")
}

func (q *Connection) CheckAuth() error {
	requestURL := q.MakeRequestURL("/query/torrents")
	resp, err := q.Client.Get(requestURL)
	if err != nil {
		return err
	}
	q.Auth.Required = (resp.StatusCode == http.StatusForbidden)
	if q.Auth.Required {
		log.Info("Auth is required")
	} else {
		log.Info("Auth is not required")
	}
	return err
}

func (q *Connection) TryToCheckAuth(num_of_retries int) {
	for i := 0; i < num_of_retries; i++ {
		err := q.CheckAuth()
		if err == nil {
			return
		} else {
			log.Error("qBittorrent RPC is not available, error is ", err)
			if i == num_of_retries-1 {
				panic(err)
			} else {
				time.Sleep(5 * time.Second)
			}
		}
	}
}

func (q *Connection) Login(username, password string) bool {
	resp, err := http.PostForm(q.MakeRequestURL("/login"),
		url.Values{"username": {username}, "password": {password}})
	check(err)
	authOK := false
	for _, value := range resp.Cookies() {
		if value != nil {
			cookie := *value
			if cookie.Name == "SID" {
				authOK = true
				q.Auth.Cookie = cookie
				break
			}
		}
	}
	return authOK
}

func (q *Connection) FillIDs(torrentsList []TorrentsList) (newHashes []string) {
	// Refill the table completely to handle removed hashes
	q.HashIds = make([]string, len(torrentsList))

	for key, value := range torrentsList {
		q.HashIds[key] = value.Hash
	}
	return
}
