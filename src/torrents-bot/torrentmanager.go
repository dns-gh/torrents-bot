package main

import (
	"log"
	"time"

	"strconv"

	"os"
	"path/filepath"

	bs "github.com/dns-gh/bs-client/bsclient"
	t411 "github.com/dns-gh/t411-client/t411client"
)

var (
	planningFetchFreq = 10 * time.Second
)

type torrentManager struct {
	bsClient     *bs.BetaSeries
	t411Client   *t411.T411
	torrentsPath string
	torrents     map[string]bool
}

func makeTorrentPath(path string) string {
	torrentsPath, err := filepath.Abs(path)
	if err != nil {
		log.Fatalln(err.Error())
	}

	if _, err := os.Stat(torrentsPath); err != nil {
		if os.IsNotExist(err) {
			err = os.Mkdir(torrentsPath, os.ModeDir+0666)
			if err != nil {
				log.Fatalln(err.Error())
			}
		} else {
			log.Fatalln(err.Error())
		}
	}
	return torrentsPath
}

func makeTorrentManager(torrentsPath, bsKey, bsUsername, bsPassword, t411Username, t411Password string) *torrentManager {
	t411Client, err := t411.NewT411Client("", t411Username, t411Password)
	if err != nil {
		log.Fatalln(err.Error())
	}

	bsClient, err := bs.NewBetaseriesClient(bsKey, bsUsername, bsPassword)
	if err != nil {
		log.Fatalln(err.Error())
	}

	manager := &torrentManager{
		bsClient:     bsClient,
		t411Client:   t411Client,
		torrentsPath: makeTorrentPath(torrentsPath),
		torrents:     make(map[string]bool),
	}
	return manager
}

func (t *torrentManager) moveToTorrentsPath(tmp string) bool {
	defer func() {
		err := os.Remove(tmp)
		if err != nil {
			log.Println(err.Error())
		}
	}()
	dst := filepath.Join(t.torrentsPath, filepath.Base(tmp)+".torrent")
	err := os.Rename(tmp, dst)
	if err != nil {
		err = copyFile(tmp, dst)
		if err != nil {
			log.Println(err.Error())
			return false
		}
	}
	return true
}

func (t *torrentManager) Run() {
	ticker := time.NewTicker(planningFetchFreq)
	defer ticker.Stop()
	for _ = range ticker.C {
		episodes, err := t.bsClient.PlanningMember(-1, true, "")
		if err != nil {
			log.Println(err.Error())
			continue
		}
		for _, v := range episodes {
			if downloaded, ok := t.torrents[strconv.Itoa(v.ID)]; !ok || !downloaded {
				log.Println("TV Show:", v.Show.Title)
				t.torrents[strconv.Itoa(v.ID)] = false
				tmpFile, err := t.t411Client.DownloadTorrentByTerms(v.Show.Title, v.Season, v.Episode, "VOSTFR", "")
				if err != nil {
					log.Println(err.Error())
					continue
				}
				if t.moveToTorrentsPath(tmpFile) {
					t.torrents[strconv.Itoa(v.ID)] = true
				}
			}
		}
	}
}
