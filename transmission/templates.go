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
	"download-dir-free-space":      float64(100 * (1 << 30)), // 100 GB
	"idle-seeding-limit":           30,
	"idle-seeding-limit-enabled":   false,
	"queue-stalled-enabled":        true,
	"queue-stalled-minutes":        30,
	"rename-partial-files":         true,
	"rpc-version":                  15,
	"rpc-version-minimum":          1,
	"script-torrent-done-enabled":  false,
	"script-torrent-done-filename": "",
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
	"version": "2.84 (14307)",
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
	"manualAnnounceTime":      0,
	// TODO
	"peers":      []string{},
	"magnetLink": "",
}

var TrackerStatsTemplate = JsonMap{
	"announceState":         0,
	"hasScraped":            false,
	"isBackup":              false,
	"lastAnnounceStartTime": 0,
	"lastAnnounceTime":      0,
	"lastAnnounceTimedOut":  false,
	"lastScrapeResult":      "",
	"lastScrapeStartTime":   0,
	"lastScrapeSucceeded":   false,
	"lastScrapeTime":        0,
	"lastScrapeTimedOut":    0,
	"nextAnnounceTime":      0,
	"nextScrapeTime":        0,
	"scrapeState":           2,
}

var SessionStatsTemplate = JsonMap{
	"current-stats": map[string]int64{
		"filesAdded":   13,
		"sessionCount": 1,
	},
}
