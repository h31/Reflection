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
)

func check(e error) {
	if e != nil {
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
	HashIds []string
	Tr      *http.Transport
	Client  *http.Client
	Auth    Auth
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

func (q *Connection) GetTorrentList() (resp []TorrentsList) {
	url := q.MakeRequestURLWithParam("/query/torrents", map[string]string{"sort": "hash"})
	torrents := q.DoGET(url)

	err := json.Unmarshal(torrents, &resp)
	check(err)
	return
}

func (q *Connection) GetPropsGeneral(id int) (propGeneral PropertiesGeneral) {
	propGeneralURL := q.MakeRequestURL("/query/propertiesGeneral/" + q.GetHashForId(id))
	propGeneralRaw := q.DoGET(propGeneralURL)

	err := json.Unmarshal(propGeneralRaw, &propGeneral)
	check(err)
	return
}

func (q *Connection) GetPropsTrackers(id int) (trackers []PropertiesTrackers) {
	trackersURL := q.MakeRequestURL("/query/propertiesTrackers/" + q.GetHashForId(id))
	trackersRaw := q.DoGET(trackersURL)

	err := json.Unmarshal(trackersRaw, &trackers)
	check(err)
	return
}

func (q *Connection) GetPreferences() (pref Preferences) {
	prefURL := q.MakeRequestURL("/query/preferences")
	prefRaw := q.DoGET(prefURL)

	err := json.Unmarshal(prefRaw, &pref)
	check(err)
	return
}

func (q *Connection) GetTransferInfo() (info TransferInfo) {
	infoURL := q.MakeRequestURL("/query/transferInfo")
	infoRaw := q.DoGET(infoURL)

	err := json.Unmarshal(infoRaw, &info)
	check(err)
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
	check(err)
	return
}

func (q *Connection) DoGET(url string) []byte {
	req, err := http.NewRequest("GET", url, nil)
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
	return q.HashIds[id-1]
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

func FindInArray(array []string, item string) bool {
	for _, value := range array {
		if value == item {
			return true
		}
	}
	return false
}

func (q *Connection) CheckAuth() {
	url := q.MakeRequestURL("/query/torrents")
	resp, err := q.Client.Get(url)
	check(err)
	q.Auth.Required = (resp.StatusCode == http.StatusForbidden)
	if q.Auth.Required {
		log.Info("Auth is required")
	} else {
		log.Info("Auth is not required")
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
	if len(q.HashIds) > 0 {
		// HashIDs already filled
		for _, torrent := range torrentsList {
			if FindInArray(q.HashIds, torrent.Hash) == false {
				log.Debug("Received new hash ", torrent.Hash)
				newHashes = append(newHashes, torrent.Hash)
				q.HashIds = append(q.HashIds, torrent.Hash)
			}
		}
	} else {
		q.HashIds = make([]string, len(torrentsList))

		for key, value := range torrentsList {
			q.HashIds[key] = value.Hash
		}
	}
	return
}
