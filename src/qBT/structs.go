package qBT

// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-Documentation

type JsonMap map[string]interface{}

type TorrentsList struct {
	Hash           string  //	Torrent hash
	Name           string  //	Torrent name
	Size           int     //	Total size (bytes) of files selected for download
	Progress       float64 //	Torrent progress (percentage/100)
	Dlspeed        int     //	Torrent download speed (bytes/s)
	Upspeed        int     //	Torrent upload speed (bytes/s)
	Priority       int     //	Torrent priority. Returns -1 if queuing is disabled or torrent is in seed mode
	Num_seeds      int     //	Number of seeds connected to
	Num_complete   int     //	Number of seeds in the swarm
	Num_leechs     int     //	Number of leechers connected to
	Num_incomplete int     //	Number of leechers in the swarm
	Ratio          float64 //	Torrent share ratio. Max ratio value: 9999.
	Eta            int     //	Torrent ETA (seconds)
	State          string  //	Torrent state. See table here below for the possible values
	Seq_dl         bool    //	True if sequential download is enabled
	F_l_piece_prio bool    //	True if first last piece are prioritized
	Label          string  //	Label of the torrent
	Super_seeding  bool    //	True if super seeding is enabled
	Force_start    bool    //	True if force start is enabled for this torrent
}

type PropertiesGeneral struct {
	Save_path                string  //	Torrent save path
	Creation_date            int     //	Torrent creation date (Unix timestamp)
	Piece_size               int     //	Torrent piece size (bytes)
	Comment                  string  //	Torrent comment
	Total_wasted             int     //	Total data wasted for torrent (bytes)
	Total_uploaded           int     //	Total data uploaded for torrent (bytes)
	Total_uploaded_session   int     //	Total data uploaded this session (bytes)
	Total_downloaded         int     //	Total data uploaded for torrent (bytes)
	Total_downloaded_session int     //	Total data downloaded this session (bytes)
	Up_limit                 int     //	Torrent upload limit (bytes/s)
	Dl_limit                 int     //	Torrent download limit (bytes/s)
	Time_elapsed             int     //	Torrent elapsed time (seconds)
	Seeding_time             int     //	Torrent elapsed time while complete (seconds)
	Nb_connections           int     //	Torrent connection count
	Nb_connections_limit     int     //	Torrent connection count limit
	Share_ratio              float64 //	Torrent share ratio
	Addition_date            int     //	When this torrent was added (unix timestamp)
	Completion_date          int     //	Torrent completion date (unix timestamp)
	Created_by               string  //	Torrent creator
	Dl_speed_avg             int     //	Torrent average download speed (bytes/second)
	Dl_speed                 int     //	Torrent download speed (bytes/second)
	Eta                      int     //	Torrent ETA (seconds)
	Last_seen                int     //	Last seen complete date (unix timestamp)
	Peers                    int     //	Number of peers connected to
	Peers_total              int     //	Number of peers in the swarm
	Pieces_have              int     //	Number of pieces owned
	Pieces_num               int     //	Number of pieces of the torrent
	Reannounce               int     //	Number of seconds until the next announce
	Seeds                    int     //	Number of seeds connected to
	Seeds_total              int     //	Number of seeds in the swarm
	Total_size               int     //	Torrent total size (bytes)
	Up_speed_avg             int     //	Torrent average upload speed (bytes/second)
	Up_speed                 int     //	Torrent upload speed (bytes/second)
}

type PropertiesTrackers struct {
	Url       string //	Tracker url
	Status    string //	Tracker status (translated string). See the table here below for the possible values
	Num_peers int    //	Number of peers for current torrent reported by the tracker
	Msg       string // Tracker message (there is no way of knowing what this message is - it's up to tracker admins)
}

type PropertiesFiles struct {
	Name     string  //	File name (including relative path)
	Size     int     //	File size (bytes)
	Progress float64 //	File progress (percentage/100)
	Priority int     //	File priority. See possible values here below
	Is_seed  bool    //	True if file is seeding/complete
}