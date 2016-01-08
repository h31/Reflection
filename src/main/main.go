package main

import (
	"net/http"
	"encoding/json"
	"io/ioutil"
	log "github.com/Sirupsen/logrus"
	"net/url"
	"strconv"
	"encoding/base64"
	"mime/multipart"
	"bytes"
	"crypto/sha1"
	"fmt"
	"gopkg.in/alecthomas/kingpin.v2"
	"transmission"
	"qBT"
	"time"
	"strings"
)

var (
	verbose = kingpin.Flag("verbose", "Enable debug output").Short('v').Bool()
	apiAddr = kingpin.Flag("api-addr", "qBittorrent API address").Short('r').Default("http://localhost:8080/").String()
	port = kingpin.Flag("port", "Transmission RPC port").Short('p').Default("9091").Int()
)

var deprecatedFields = []string{
	"announceResponse",
	"seeders",
	"leechers",
	"downloadLimitMode",
	"uploadLimitMode",
	"nextAnnounceTime",
}

var qBTConn qBT.Connection

func IsFieldDeprecated(field string) bool {
	for _, value := range deprecatedFields {
		if value == field {
			return true
		}
	}
	return false
}

func parseIDsArgument(args *json.RawMessage) []int {
	if (args == nil) {
		log.Debug("No IDs provided")
		result := make([]int, qBTConn.GetHashNum())
		for i := 0; i < qBTConn.GetHashNum(); i++ {
			result[i] = i + 1
		}
		return result
	}

	var ids interface{}
	err := json.Unmarshal(*args, &ids)
	Check(err)

	if ids_num, ids_num_ok := ids.(float64); ids_num_ok {
		log.Debug("Query a single ID")
		return []int{int(ids_num)}
	} else if ids_list, ids_list_ok := ids.([]interface{}); ids_list_ok {
		log.Debug("Query an ID list of length ", len(ids_list))
		result := make([]int, len(ids_list))
		for i, value := range ids_list {
			id, id_ok := value.(float64)
			if (id_ok) {
				result[i] = int(id)
			} else {
				log.Error("Hashes as IDs are not supported")
			}
		}
		return result
	} else if _, ids_string_ok := ids.(string); ids_string_ok {
		log.Debug("Query recently-active") // TODO
		result := make([]int, qBTConn.GetHashNum())
		for i := 0; i < qBTConn.GetHashNum(); i++ {
			result[i] = i + 1
		}
		return result
	} else {
		log.Error("Unknown action")
		return []int{}
	}
}

func parseActionArgument(args json.RawMessage) ([]string) {
	var req struct {
		Ids json.RawMessage
	}
	err := json.Unmarshal(args, &req)
	Check(err)

	ids := parseIDsArgument(&req.Ids)
	hashes := make([]string, len(ids))
	for i, value := range ids {
		hashes[i] = qBTConn.GetHashForId(value)
	}
	return hashes
}

func MapTorrentList(dst JsonMap, torrentsList []qBT.TorrentsList, id int) {
	var src qBT.TorrentsList
	for _, value := range torrentsList {
		if (value.Hash == qBTConn.GetHashForId(id)) {
			src = value
		}
	}

	for key, value := range transmission.TorrentGetBase {
		dst[key] = value
	}
	dst["hashString"] = src.Hash
	dst["name"] = src.Name
	dst["recheckProgress"] = src.Progress;
	dst["sizeWhenDone"] = src.Size
	dst["rateDownload"] = src.Dlspeed
	dst["rateUpload"] = src.Upspeed
	dst["uploadRatio"] = src.Ratio
	dst["eta"] = src.Eta
	switch src.State {
	case "pausedUP", "pausedDL":
		dst["status"] = 0 // TR_STATUS_STOPPED
	case "checkingUP", "checkingDL":
		dst["status"] = 2 // TR_STATUS_CHECK
	case "queuedDL":
		dst["status"] = 3 // TR_STATUS_DOWNLOAD_WAIT
	case "downloading", "stalledDL":
		dst["status"] = 4 // TR_STATUS_DOWNLOAD
	case "queuedUP":
		dst["status"] = 5 // TR_STATUS_SEED_WAIT
	case "uploading", "stalledUP":
		dst["status"] = 6 // TR_STATUS_SEED
	default:
		dst["status"] = 0 // TR_STATUS_STOPPED
	}
	dst["leftUntilDone"] = float64(src.Size) * (1 - src.Progress)
	dst["desiredAvailable"] = float64(src.Size) * (1 - src.Progress) // TODO
	dst["haveUnchecked"] = 0 // TODO
}

func MapPropsGeneral(dst JsonMap, propGeneral qBT.PropertiesGeneral) {
	dst["downloadDir"] = propGeneral.Save_path
	dst["pieceSize"] = propGeneral.Piece_size
	dst["pieceCount"] = propGeneral.Pieces_num
	dst["addedDate"] = propGeneral.Addition_date
	dst["startDate"] = propGeneral.Addition_date // TODO
	dst["comment"] = propGeneral.Comment
	dst["dateCreated"] = propGeneral.Creation_date
	dst["creator"] = propGeneral.Created_by
	dst["doneDate"] = propGeneral.Completion_date
	dst["downloadLimit"] = propGeneral.Dl_limit; // TODO: Kb/s?
	dst["uploadLimit"] = propGeneral.Up_limit
	dst["totalSize"] = propGeneral.Total_size
	dst["haveValid"] = propGeneral.Piece_size * propGeneral.Pieces_num
	dst["downloadedEver"] = propGeneral.Total_downloaded
	dst["uploadedEver"] = propGeneral.Total_uploaded
}

func MapPropsTrackers(dst JsonMap, trackers []qBT.PropertiesTrackers) {
	trackersList := make([]JsonMap, len(trackers))
	trackerStats := make([]JsonMap, len(trackers))

	for i, value := range trackers {
		trackersList[i] = make(JsonMap)
		trackersList[i]["announce"] = value.Url
		trackersList[i]["id"] = i // TODO
		trackersList[i]["scrape"] = value.Url
		trackersList[i]["tier"] = 0 // TODO

		trackerStats[i] = make(JsonMap)
		for key, value := range transmission.TrackerStatsTemplate {
			trackerStats[i][key] = value
		}
		trackerStats[i]["announce"] = value.Url
		trackerStats[i]["id"] = i // TODO
		trackerStats[i]["scrape"] = ""
		trackerStats[i]["tier"] = 0 // TODO
	}

	dst["trackers"] = trackersList
	dst["trackerStats"] = trackerStats
}

func MapPropsFiles(dst JsonMap, filesInfo []qBT.PropertiesFiles) {
	fileNum := len(filesInfo)
	files := make([]JsonMap, fileNum)
	fileStats := make([]JsonMap, fileNum)
	priorities := make([]int, fileNum)
	wanted := make([]int, fileNum)
	for i, value := range filesInfo {
		files[i] = make(JsonMap)
		fileStats[i] = make(JsonMap)

		files[i]["bytesCompleted"] = float64(value.Size) * value.Progress
		files[i]["length"] = value.Size
		files[i]["name"] = value.Name

		fileStats[i]["bytesCompleted"] = float64(value.Size) * value.Progress
		if value.Priority == 0 {
			fileStats[i]["wanted"] = false
			wanted[i] = 0
		} else {
			fileStats[i]["wanted"] = true
			wanted[i] = 1
		}
		fileStats[i]["priority"] = 0 // TODO
		priorities[i] = 0 // TODO
	}

	dst["files"] = files
	dst["fileStats"] = fileStats
	dst["priorities"] = priorities
	dst["wanted"] = wanted
}

func TorrentGet(args json.RawMessage) (JsonMap, string) {
	var req transmission.GetRequest
	err := json.Unmarshal(args, &req)
	Check(err)

	torrentList := qBTConn.GetTorrentList()

	if qBTConn.GetHashNum() == 0 {
		qBTConn.FillIDs(torrentList)
		log.Debug("Filling IDs table, new size: ", qBTConn.GetHashNum())
	}

	ids := parseIDsArgument(req.Ids)
	fields := req.Fields

	resultList := make([]JsonMap, len(ids))
	for i, id := range ids {
		translated := make(JsonMap)
		propGeneral := qBTConn.GetPropsGeneral(id)
		trackers := qBTConn.GetPropsTrackers(id)
		files := qBTConn.GetPropsFiles(id)

		MapTorrentList(translated, torrentList, id)
		MapPropsGeneral(translated, propGeneral)
		MapPropsTrackers(translated, trackers)
		MapPropsFiles(translated, files)

		translated["id"] = id
		for _, field := range fields {
			if _, ok := translated[field]; !ok {
				if !IsFieldDeprecated(field) {
					log.Error("Unsupported field: ", field)
				}
			}
		}
		for key := range translated {
			if !Any(fields, key) {
				// Remove unneeded fields
				delete(translated, key)
			}
		}
		resultList[i] = translated
	}
	return JsonMap{"torrents": resultList}, "success"
}

func SessionGet() (JsonMap, string) {
	session := make(JsonMap)
	for key, value := range transmission.SessionGetBase {
		session[key] = value
	}

	version := qBTConn.GetVersion()
	session["version"] = "2.84 (really qBT " + string(version) + ")"
	return session, "success"
}

func FreeSpace(args json.RawMessage) (JsonMap, string) {
	var req JsonMap
	err := json.Unmarshal(args, &req)
	Check(err)

	path := req["path"]

	return JsonMap{
		"path": path,
		"size-bytes": float64(100 * (1 << 30)), // 100 GB
	}, "success"
}

func SessionStats() (JsonMap, string) {
	session := make(JsonMap)
	for key, value := range transmission.SessionGetBase {
		session[key] = value
	}
	return session, "success"
}

func TorrentPause(args json.RawMessage) (JsonMap, string) {
	hashes := parseActionArgument(args)
	for _, hash := range hashes {
		log.Debug("Stopping torrent with hash ", hash)

		http.PostForm(qBTConn.MakeRequestURL("/command/pause"),
			url.Values{"hash": {hash}})
	}
	return JsonMap{}, "success"
}

func TorrentResume(args json.RawMessage) (JsonMap, string) {
	hashes := parseActionArgument(args)
	for _, hash := range hashes {
		log.Debug("Starting torrent with hash ", hash)

		http.PostForm(qBTConn.MakeRequestURL("/command/resume"),
			url.Values{"hash": {hash}})
	}
	return JsonMap{}, "success"
}

func TorrentRecheck(args json.RawMessage) (JsonMap, string) {
	hashes := parseActionArgument(args)
	for _, hash := range hashes {
		log.Debug("Verifying torrent with hash ", hash)

		http.PostForm(qBTConn.MakeRequestURL("/command/recheck"),
			url.Values{"hash": {hash}})
	}
	return JsonMap{}, "success"
}

func TorrentAdd(args json.RawMessage) (JsonMap, string) {
	var req transmission.TorrentAddRequest
	err := json.Unmarshal(args, &req)
	Check(err)

	torrentList := qBTConn.GetTorrentList()
	qBTConn.FillIDs(torrentList)
	oldHashNum := qBTConn.GetHashNum()

	var newHash string
	var newName string
	var newId int

	if (req.Metainfo != nil) {
		log.Debug("Upload torrent using metainfo")

		metainfo, err := base64.StdEncoding.DecodeString(*req.Metainfo)
		log.WithFields(log.Fields{
			"len": len(metainfo),
			"sha1": fmt.Sprintf("%x\n", sha1.Sum(metainfo)),
		}).Debug("Decoded metainfo")

		var buffer bytes.Buffer
		mime := multipart.NewWriter(&buffer)
		mimeWriter, err := mime.CreateFormFile("torrents", "example.torrent")
		Check(err)
		mimeWriter.Write(metainfo)
		mime.Close()

		qBTConn.DoPOST("/command/upload", mime.FormDataContentType(), &buffer)
		log.Debug("Torrent uploaded")

		var parsedMetaInfo MetaInfo
		parsedMetaInfo.ReadTorrentMetaInfoFile(bytes.NewBuffer(metainfo))

		newHash = fmt.Sprintf("%x", parsedMetaInfo.InfoHash)
		newName = parsedMetaInfo.Info.Name

	} else if req.Filename != nil {
		path := *req.Filename
		if (strings.HasPrefix(path, "magnet:?")) {
			path = strings.TrimPrefix(path, "magnet:?")
			params, err := url.ParseQuery(path)
			log.WithFields(log.Fields{
				"params": params,
			}).Debug("Params decoded")
			Check(err)
			newHash = params["xt"][0]
			name, nameProvided := params["dn"]
			if nameProvided {
				newName = strings.TrimPrefix(name[0], "urn:btih:")
			} else {
				newName = "Torrent name"
			}
		}

		http.PostForm(qBTConn.MakeRequestURL("/command/download"),
			url.Values{"urls": {*req.Filename}})
	}

	OuterLoop: for retries := 0; retries < 100; retries++ {
		torrentList := qBTConn.GetTorrentList()
		qBTConn.FillIDs(torrentList)

		for i := oldHashNum; i < qBTConn.GetHashNum(); i++ {
			if qBTConn.GetHashForId(i) == newHash {
				newId = i
				break OuterLoop
			}
		}

		log.Debug("Nothing was found, waiting...")

		time.Sleep(10 * time.Millisecond)
	}

	paused := false
	if req.Paused != nil {
		if value, ok := (*req.Paused).(float64); ok {
			paused = value != 0
		}
		if value, ok := (*req.Paused).(bool); ok {
			paused = value
		}
	}
	if paused {
		http.PostForm(qBTConn.MakeRequestURL("/command/pause"),
			url.Values{"hash": {newHash}})
	} else {
		http.PostForm(qBTConn.MakeRequestURL("/command/resume"),
			url.Values{"hash": {newHash}})
	}

	log.WithFields(log.Fields{
		"hash": newHash,
		"id":   newId,
		"name": newName,
	}).Debug("New torrent")

	return JsonMap{
		"torrent-added": JsonMap{
			"id": newId,
			"name": newName,
			"hashString": newHash,
		},
	}, "success"
}

func handler(w http.ResponseWriter, r *http.Request) {
	var req transmission.RPCRequest
	reqBody, err := ioutil.ReadAll(r.Body)
	log.Debug("Got request ", string(reqBody))
	err = json.Unmarshal(reqBody, &req)
	Check(err)

	var resp JsonMap
	var result string
	switch {
	case req.Method == "session-get":
		resp, result = SessionGet()
	case req.Method == "free-space":
		resp, result = FreeSpace(req.Arguments)
	case req.Method == "torrent-get":
		resp, result = TorrentGet(req.Arguments)
	case req.Method == "session-stats":
		resp, result = SessionStats()
	case req.Method == "torrent-stop":
		resp, result = TorrentPause(req.Arguments)
	case req.Method == "torrent-start":
		resp, result = TorrentResume(req.Arguments)
	case req.Method == "torrent-start-now":
		resp, result = TorrentResume(req.Arguments)
	case req.Method == "torrent-verify":
		resp, result = TorrentRecheck(req.Arguments)
	case req.Method == "torrent-add":
		resp, result = TorrentAdd(req.Arguments)
	}
	response := JsonMap{
		"result": result,
		"arguments": resp,
	}
	if req.Tag != nil {
		response["tag"] = req.Tag
	}
	respBody, err := json.Marshal(response)
	Check(err)
	w.Header().Set("Content-Length", strconv.Itoa(len(respBody)))
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Write(respBody)
}

func main() {
	kingpin.Parse()
	if *verbose {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
	qBTConn.Addr = *apiAddr
	qBTConn.Tr = &http.Transport{
		DisableKeepAlives: true,
	}
	qBTConn.Client = &http.Client{Transport: qBTConn.Tr}

	http.HandleFunc("/transmission/rpc", handler)
	http.HandleFunc("/rpc", handler)
	http.Handle("/", http.FileServer(http.Dir("web/")))
	http.ListenAndServe(":9091", nil)
}
