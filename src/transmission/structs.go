package transmission

import "encoding/json"

type RPCRequest struct {
	Method    string
	Tag       *int
	Arguments json.RawMessage
}

type GetRequest struct {
	Ids    *json.RawMessage
	Fields []string
}

type TorrentAddRequest struct {
	Cookies           *string      //  pointer to a string of one or more cookies.
	Download_dir      *string      `json:"download-dir"` //    path to download the torrent to
	Filename          *string      //   filename or URL of the .torrent file
	Metainfo          *string      //   base64-encoded .torrent content
	Paused            *interface{} //    if true, don't start the torrent
	Peer_limit        *int         //   maximum number of peers
	BandwidthPriority *int         //   torrent's bandwidth tr_priority_t
	Files_wanted      *[]JsonMap   //   indices of file(s) to download
	Files_unwanted    *[]JsonMap   //    indices of file(s) to not download
	Priority_high     *[]JsonMap   //    indices of high-priority file(s)
	Priority_low      *[]JsonMap   //    indices of low-priority file(s)
	Priority_normal   *[]JsonMap   //    indices of normal-priority file(s)
}
