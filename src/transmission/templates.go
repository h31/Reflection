package transmission

type JsonMap map[string]interface{}

var SessionGetBase = JsonMap{
	"alt-speed-down":               50,
	"alt-speed-enabled":            false,
	"alt-speed-time-begin":         540,
	"alt-speed-time-day":           127,
	"alt-speed-time-enabled":       false,
	"alt-speed-time-end":           1020,
	"alt-speed-up":                 50,
	"blocklist-enabled":            false,
	"blocklist-size":               393006,
	"blocklist-url":                "http://www.example.com/blocklist",
	"cache-size-mb":                4,
	"config-dir":                   "/var/lib/transmission-daemon",
	"dht-enabled":                  true,
	"download-dir":                 "/var/lib/transmission-daemon/downloads",
	"download-dir-free-space":      float64(100 * (1 << 30)), // 100 GB
	"download-queue-enabled":       true,
	"download-queue-size":          5,
	"encryption":                   "preferred",
	"idle-seeding-limit":           30,
	"idle-seeding-limit-enabled":   false,
	"incomplete-dir":               "/var/lib/transmission-daemon/downloads",
	"incomplete-dir-enabled":       false,
	"lpd-enabled":                  false,
	"peer-limit-global":            200,
	"peer-limit-per-torrent":       50,
	"peer-port":                    44444,
	"peer-port-random-on-start":    false,
	"pex-enabled":                  true,
	"port-forwarding-enabled":      true,
	"queue-stalled-enabled":        true,
	"queue-stalled-minutes":        30,
	"rename-partial-files":         true,
	"rpc-version":                  15,
	"rpc-version-minimum":          1,
	"script-torrent-done-enabled":  false,
	"script-torrent-done-filename": "",
	"seed-queue-enabled":           false,
	"seed-queue-size":              10,
	"seedRatioLimit":               2,
	"seedRatioLimited":             false,
	"speed-limit-down":             100,
	"speed-limit-down-enabled":     false,
	"speed-limit-up":               100,
	"speed-limit-up-enabled":       false,
	"start-added-torrents":         true,
	"trash-original-torrent-files": false,
	"units": map[string]interface{}{
		"memory-bytes": 1024,
		"memory-units": []string{
			"KiB",
			"MiB",
			"GiB",
			"TiB",
		},
		"size-bytes": 1000,
		"size-units": []string{
			"kB",
			"MB",
			"GB",
			"TB",
		},
		"speed-bytes": 1000,
		"speed-units": []string{
			"kB/s",
			"MB/s",
			"GB/s",
			"TB/s",
		},
	},
	"utp-enabled": false,
	"version":     "2.84 (14307)",
}

var TorrentGetBase = JsonMap{
	"errorString":             "",
	"metadataPercentComplete": 1,
	"isFinished":              false,
	"queuePosition":           0, // Looks like not supported by qBittorent
	"seedRatioLimit":          2,
	"seedRatioMode":           0, // No local limits in qBittorrent
	"activityDate":            1443977197,
	"secondsDownloading":      500,
	"secondsSeeding":          80000,
	"isPrivate":               false, // Not exposed by qBittorrent
	"honorsSessionLimits":     true,
	"webseedsSendingToUs":     0,
	"bandwidthPriority":       0,
	"seedIdleLimit":           10,
	"seedIdleMode":            0, // TR_IDLELIMIT_GLOBAL
	// TODO
	"peers": []string{},
}

var TrackerStatsTemplate = JsonMap{
	"announceState":         0,
	"downloadCount":         -1,
	"hasAnnounced":          false,
	"hasScraped":            false,
	"host":                  "http://example.com:80",
	"isBackup":              false,
	"lastAnnouncePeerCount": 0,
	"lastAnnounceResult":    "",
	"lastAnnounceStartTime": 0,
	"lastAnnounceSucceeded": false,
	"lastAnnounceTime":      0,
	"lastAnnounceTimedOut":  false,
	"lastScrapeResult":      "",
	"lastScrapeStartTime":   0,
	"lastScrapeSucceeded":   false,
	"lastScrapeTime":        0,
	"lastScrapeTimedOut":    0,
	"leecherCount":          -1,
	"nextAnnounceTime":      0,
	"nextScrapeTime":        0,
	"scrapeState":           2,
	"seederCount":           -1,
}

var SessionStatsTemplate = JsonMap{
	"activeTorrentCount": 0,
	"current-stats": map[string]int64{
		"filesAdded":    13,
		"secondsActive": 99633,
		"sessionCount":  1,
		"uploadedBytes": 26478335758,
	},
	"pausedTorrentCount": 127,
	"torrentCount":       127,
}
