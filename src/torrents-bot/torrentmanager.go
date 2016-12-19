package main

import (
	"log"
	"time"

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
	dst := filepath.Join(t.torrentsPath, filepath.Base(tmp))
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
		log.Printf("checking for %d episode(s)...\n", len(episodes))
		for _, v := range episodes {
			log.Printf("trying %s - S%02dE%02d\n", v.Show.Title, v.Season, v.Episode)
			if !v.User.Downloaded {
				tmpFile, err := t.t411Client.DownloadTorrentByTerms(v.Show.Title, v.Season, v.Episode, "VOSTFR", "TVripHD 720 [Rip HD depuis Source Tv HD]")
				if err != nil {
					if err != t411.ErrTorrentNotFound {
						log.Println(err.Error())
					}
					continue
				}
				if t.moveToTorrentsPath(tmpFile) {
					_, err := t.bsClient.EpisodeDownloaded(v.ID)
					if err != nil {
						log.Println(err.Error())
					}
					log.Printf("%s - S%02dE%02d downloaded\n", v.Show.Title, v.Season, v.Episode)
				}
			}
		}
	}
}
