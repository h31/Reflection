package main

import (
	"bytes"
	"github.com/jackpal/bencode-go"
	"crypto/sha1"
	"io"
	"fmt"
	"time"
)

// Structs into which torrent metafile is
// parsed and stored into.
type FileDict struct {
	Length int64    "length"
	Path   []string "path"
	Md5sum string   "md5sum"
}

type InfoDict struct {
	FileDuration []int64 "file-duration"
	FileMedia    []int64 "file-media"
	// Single file
	Name   string "name"
	Length int64  "length"
	Md5sum string "md5sum"
	// Multiple files
	Files       []FileDict "files"
	PieceLength int64      "piece length"
	Pieces      string     "pieces"
	Private     int64      "private"
}

type MetaInfo struct {
	Info         InfoDict   "info"
	InfoHash     string     "info hash"
	Announce     string     "announce"
	AnnounceList [][]string "announce-list"
	CreationDate int64      "creation date"
	Comment      string     "comment"
	CreatedBy    string     "created by"
	Encoding     string     "encoding"
}

// Open .torrent file, un-bencode it and load them into MetaInfo struct.
func (metaInfo *MetaInfo) ReadTorrentMetaInfoFile(src io.Reader) bool {
	// Decode bencoded metainfo file.
	fileMetaData, er := bencode.Decode(src)
	if er != nil {
		return false
	}

	// fileMetaData is map of maps of... maps. Get top level map.
	metaInfoMap, ok := fileMetaData.(map[string]interface{})
	if !ok {
		return false
	}

	// Enumerate through child maps.
	var bytesBuf bytes.Buffer
	for mapKey, mapVal := range metaInfoMap {
		switch mapKey {
		case "info":
			if er = bencode.Marshal(&bytesBuf, mapVal); er != nil {
				return false
			}

			infoHash := sha1.New()
			infoHash.Write(bytesBuf.Bytes())
			metaInfo.InfoHash = string(infoHash.Sum(nil))

			if er = bencode.Unmarshal(&bytesBuf, &metaInfo.Info); er != nil {
				return false
			}

		case "announce-list":
			if er = bencode.Marshal(&bytesBuf, mapVal); er != nil {
				return false
			}
			if er = bencode.Unmarshal(&bytesBuf, &metaInfo.AnnounceList); er != nil {
				return false
			}

		case "announce":
			metaInfo.Announce = mapVal.(string)

		case "creation date":
			metaInfo.CreationDate = mapVal.(int64)

		case "comment":
			metaInfo.Comment = mapVal.(string)

		case "created by":
			metaInfo.CreatedBy = mapVal.(string)

		case "encoding":
			metaInfo.Encoding = mapVal.(string)
		}
	}

	return true
}

// Print torrent meta info struct data.
func (metaInfo *MetaInfo) DumpTorrentMetaInfo() {
	fmt.Println("Announce:", metaInfo.Announce)
	fmt.Println("Announce List:")
	for _, anncListEntry := range metaInfo.AnnounceList {
		for _, elem := range anncListEntry {
			fmt.Println("    ", elem)
		}
	}
	strCreationDate := time.Unix(metaInfo.CreationDate, 0)
	fmt.Println("Creation Date:", strCreationDate)
	fmt.Println("Comment:", metaInfo.Comment)
	fmt.Println("Created By:", metaInfo.CreatedBy)
	fmt.Println("Encoding:", metaInfo.Encoding)
	fmt.Printf("InfoHash: %X\n", metaInfo.InfoHash)
	fmt.Println("Info:")
	fmt.Println("    Piece Length:", metaInfo.Info.PieceLength)
	piecesList := metaInfo.getPiecesList()
	fmt.Printf("    Pieces:%X -- %X\n", len(piecesList), len(metaInfo.Info.Pieces)/20)
	fmt.Println("    File Duration:", metaInfo.Info.FileDuration)
	fmt.Println("    File Media:", metaInfo.Info.FileMedia)
	fmt.Println("    Private:", metaInfo.Info.Private)
	fmt.Println("    Name:", metaInfo.Info.Name)
	fmt.Println("    Length:", metaInfo.Info.Length)
	fmt.Println("    Md5sum:", metaInfo.Info.Md5sum)
	fmt.Println("    Files:")
	for _, fileDict := range metaInfo.Info.Files {
		fmt.Println("        Length:", fileDict.Length)
		fmt.Println("        Path:", fileDict.Path)
		fmt.Println("        Md5sum:", fileDict.Md5sum)
	}
}

// Splits pieces string into an array of 20 byte SHA1 hashes.
func (metaInfo *MetaInfo) getPiecesList() []string {
	var piecesList []string
	piecesLen := len(metaInfo.Info.Pieces)
	for i, j := 0, 0; i < piecesLen; i, j = i+20, j+1 {
		piecesList = append(piecesList, metaInfo.Info.Pieces[i:i+19])
	}
	return piecesList
}