package qBT

import (
	"encoding/json"
)

// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-Documentation

type JsonMap map[string]interface{}

type TorrentInfo struct {
	Id             ID      //   Transmission's ID
	Hash           Hash    //	Torrent hash
	Name           string  //	Torrent name
	Size           int64   //	Total size (bytes) of files selected for download
	Total_size     int64   //	Torrent total size (bytes)
	Progress       float64 //	Torrent progress (percentage/100)
	Dlspeed        int     //	Torrent download speed (bytes/s)
	Upspeed        int     //	Torrent upload speed (bytes/s)
	Priority       int     //	Torrent priority. Returns -1 if queuing is disabled or torrent is in seed mode
	Num_seeds      int     //	Number of seeds connected to
	Num_complete   int     //	Number of seeds in the swarm
	Num_leechs     int     //	Number of leechers connected to
	Num_incomplete int     //	Number of leechers in the swarm
	Ratio          float64 //	Torrent share ratio. Max ratio value: 9999.
	Eta            int64   //	Torrent ETA (seconds)
	State          string  //	Torrent state. See table here below for the possible values
	Seq_dl         bool    //	True if sequential download is enabled
	F_l_piece_prio bool    //	True if first last piece are prioritized
	Label          string  //	Label of the torrent
	Super_seeding  bool    //	True if super seeding is enabled
	Force_start    bool    //	True if force start is enabled for this torrent
	Save_path      string  //	Torrent save path
	Added_on       int64
	Completion_on  int64 //   Torrent completion time
}

type PeerInfo struct {
	Up_speed   int
	Uploaded   int64
	Dl_speed   int
	Port       int
	Downloaded int64
	Client     string
	Country    string
	Flags      string
	IP         string
	Progress   float64 //	Torrent progress (percentage/100)
}

type PropertiesGeneral struct {
	Save_path                string  //	Torrent save path
	Creation_date            int64   //	Torrent creation date (Unix timestamp)
	Piece_size               int     //	Torrent piece size (bytes)
	Comment                  string  //	Torrent comment
	Total_wasted             int64   //	Total data wasted for torrent (bytes)
	Total_uploaded           int64   //	Total data uploaded for torrent (bytes)
	Total_uploaded_session   int64   //	Total data uploaded this session (bytes)
	Total_downloaded         int64   //	Total data uploaded for torrent (bytes)
	Total_downloaded_session int64   //	Total data downloaded this session (bytes)
	Up_limit                 int     //	Torrent upload limit (bytes/s)
	Dl_limit                 int     //	Torrent download limit (bytes/s)
	Time_elapsed             int     //	Torrent elapsed time (seconds)
	Seeding_time             int     //	Torrent elapsed time while complete (seconds)
	Nb_connections           int     //	Torrent connection count
	Nb_connections_limit     int     //	Torrent connection count limit
	Share_ratio              float64 //	Torrent share ratio
	Addition_date            int64   //	When this torrent was added (unix timestamp)
	Completion_date          int64   //	Torrent completion date (unix timestamp)
	Created_by               string  //	Torrent creator
	Dl_speed_avg             int     //	Torrent average download speed (bytes/second)
	Dl_speed                 int     //	Torrent download speed (bytes/second)
	Eta                      int64   //	Torrent ETA (seconds)
	Last_seen                int64   //	Last seen complete date (unix timestamp)
	Peers                    int     //	Number of peers connected to
	Peers_total              int     //	Number of peers in the swarm
	Pieces_have              int     //	Number of pieces owned
	Pieces_num               int     //	Number of pieces of the torrent
	Reannounce               int64   //	Number of seconds until the next announce
	Seeds                    int     //	Number of seeds connected to
	Seeds_total              int     //	Number of seeds in the swarm
	Total_size               int64   //	Torrent total size (bytes)
	Up_speed_avg             int     //	Torrent average upload speed (bytes/second)
	Up_speed                 int     //	Torrent upload speed (bytes/second)
}

type PropertiesTrackers struct {
	Url       string //	Tracker url
	Status    int    //	Tracker status. See the table below for possible values
	Num_peers int    //	Number of peers for current torrent reported by the tracker
	Msg       string // Tracker message (there is no way of knowing what this message is - it's up to tracker admins)
}

type PropertiesFiles struct {
	Name     string  //	File name (including relative path)
	Size     int64   //	File size (bytes)
	Progress float64 //	File progress (percentage/100)
	Priority int     //	File priority. See possible values here below
	Is_seed  bool    //	True if file is seeding/complete
}

type TransferInfo struct {
	Dl_info_speed     int    //	Global download rate (bytes/s)
	Dl_info_data      int64  //	Data downloaded this session (bytes)
	Up_info_speed     int    //	Global upload rate (bytes/s)
	Up_info_data      int64  //	Data uploaded this session (bytes)
	Dl_rate_limit     int    //	Download rate limit (bytes/s)
	Up_rate_limit     int    //	Upload rate limit (bytes/s)
	Dht_nodes         int    //	DHT nodes connected to
	Connection_status string //	Connection status. See possible values here below
}

type MainData struct {
	Rid                int
	Full_update        bool
	Torrents           *json.RawMessage
	Torrents_removed   []Hash
	Categories         *json.RawMessage
	Categories_removed *json.RawMessage
	Queueing           bool
	Server_state       *TransferInfo
}

type Preferences struct {
	Locale                         string      //	Currently selected language (e.g. en_GB for english)
	Save_path                      string      //	Default save path for torrents, separated by slashes
	Temp_path_enabled              bool        //	True if folder for incomplete torrents is enabled
	Temp_path                      string      //	Path for incomplete torrents, separated by slashes
	Scan_dirs                      interface{} //	List of watch folders to add torrent automatically; slashes are used as path separators; list entries are separated by commas
	Download_in_scan_dirs          []bool      //	True if torrents should be downloaded to watch folder; list entries are separated by commas
	Export_dir_enabled             bool        //	True if .torrent file should be copied to export directory upon adding
	Export_dir                     string      //	Path to directory to copy .torrent files ifexport_dir_enabled is enabled; path is separated by slashes
	Mail_notification_enabled      bool        //	True if e-mail notification should be enabled
	Mail_notification_email        string      //	e-mail to send notifications to
	Mail_notification_smtp         string      //	smtp server for e-mail notifications
	Mail_notification_ssl_enabled  bool        //	True if smtp server requires SSL connection
	Mail_notification_auth_enabled bool        //	True if smtp server requires authentication
	Mail_notification_username     string      //	Username for smtp authentication
	Mail_notification_password     string      //	Password for smtp authentication
	Autorun_enabled                bool        //	True if external program should be run after torrent has finished downloading
	Autorun_program                string      //	Program path/name/arguments to run ifautorun_enabled is enabled; path is separated by slashes; you can use %f and%n arguments, which will be expanded by qBittorent as path_to_torrent_file and torrent_name (from the GUI; not the .torrent file name) respectively
	Preallocate_all                bool        //	True if file preallocation should take place, otherwise sparse files are used
	Queueing_enabled               bool        //	True if torrent queuing is enabled
	Max_active_downloads           int         //	Maximum number of active simultaneous downloads
	Max_active_torrents            int         //	Maximum number of active simultaneous downloads and uploads
	Max_active_uploads             int         //	Maximum number of active simultaneous uploads
	Dont_count_slow_torrents       bool        //	If true torrents w/o any activity (stalled ones) will not be counted towards max_active_*limits; see dont_count_slow_torrents for more information
	Max_ratio_enabled              bool        //	True if share ratio limit is enabled
	Max_ratio                      float64     //	Get the global share ratio limit
	Max_ratio_act                  int         //	Action performed when a torrent reaches the maximum share ratio. See list of possible values here below.
	Incomplete_files_ext           bool        //	If true .!qB extension will be appended to incomplete files
	Listen_port                    int         //	Port for incoming connections
	Upnp                           bool        //	True if UPnP/NAT-PMP is enabled
	Random_port                    bool        //	True if the port is randomly selected
	Dl_limit                       int         //	Global download speed limit in B/s; -1means no limit is applied
	Up_limit                       int         //	Global upload speed limit in B/s; -1means no limit is applied
	Max_connec                     int         //	Maximum global number of simultaneous connections
	Max_connec_per_torrent         int         //	Maximum number of simultaneous connections per torrent
	Max_uploads                    int         //	Maximum number of upload slots
	Max_uploads_per_torrent        int         //	Maximum number of upload slots per torrent
	Enable_utp                     bool        //	True if uTP protocol should be enabled; this option is only available in qBittorent built against libtorrent version 0.16.X and higher
	Limit_utp_rate                 bool        //	True if [du]l_limit should be applied to uTP connections; this option is only available in qBittorent built against libtorrent version 0.16.X and higher
	Limit_tcp_overhead             bool        //	True if [du]l_limit should be applied to estimated TCP overhead (service data: e.g. packet headers)
	Alt_dl_limit                   int         //	Alternative global download speed limit in KiB/s
	Alt_up_limit                   int         //	Alternative global upload speed limit in KiB/s
	Scheduler_enabled              bool        //	True if alternative limits should be applied according to schedule
	Schedule_from_hour             int         //	Scheduler starting hour
	Schedule_from_min              int         //	Scheduler starting minute
	Schedule_to_hour               int         //	Scheduler ending hour
	Schedule_to_min                int         //	Scheduler ending minute
	Scheduler_days                 int         //	Scheduler days. See possible values here below
	Dht                            bool        //	True if DHT is enabled
	DhtSameAsBT                    bool        //	True if DHT port should match TCP port
	Dht_port                       int         //	DHT port if dhtSameAsBT is false
	Pex                            bool        //	True if PeX is enabled
	Lsd                            bool        //	True if LSD is eanbled
	Encryption                     int         //	See list of possible values here below
	Anonymous_mode                 bool        //	If true anonymous mode will be enabled; read more here; this option is only available in qBittorent built against libtorrent version 0.16.X and higher
	Proxy_type                     int         //	See list of possible values here below
	Proxy_ip                       string      //	Proxy Address address or domain name
	Proxy_port                     int         //	Proxy port
	Proxy_peer_connections         bool        //	True if peer and web seed connections should be proxified; this option will have any effect only in qBittorent built against libtorrent version 0.16.X and higher
	Force_proxy                    bool        //	True if the connections not supported by the proxy are disabled
	Proxy_auth_enabled             bool        //	True proxy requires authentication; doesn't apply to SOCKS4 proxies
	Proxy_username                 string      //	Username for proxy authentication
	Proxy_password                 string      //	Password for proxy authentication
	Ip_filter_enabled              bool        //	True if external Address filter should be enabled
	Ip_filter_path                 string      //	Path to Address filter file (.dat, .p2p, .p2b files are supported); path is separated by slashes
	Ip_filter_trackers             bool        //	True if Address filters are applied to trackers
	Web_ui_port                    int         //	WebUI port
	Web_ui_upnp                    bool        //	True if UPnP is used for the WebUI port
	Web_ui_username                string      //	WebUI username
	Web_ui_password                string      //	MD5 hash of WebUI password; hash is generated from the following string:username:Web UI Access:plain_text_web_ui_password
	Bypass_local_auth              bool        //	True if auithetication challenge for loopback address (127.0.0.1) should be disabled
	Use_https                      bool        //	True if WebUI HTTPS access is eanbled
	Ssl_key                        string      //	SSL keyfile contents (this is a not a path)
	Ssl_cert                       string      //	SSL certificate contents (this is a not a path)
	Dyndns_enabled                 bool        //	True if server DNS should be updated dynamically
	Dyndns_service                 int         //	See list of possible values here below
	Dyndns_username                string      //	Username for DDNS service
	Dyndns_password                string      //	Password for DDNS service
	Dyndns_domain                  string      //	Your DDNS domain name
}
