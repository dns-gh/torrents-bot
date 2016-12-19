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
	planningFetchFreq = 10 * time.Minute
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

func (t *torrentManager) DownloadWithQuality(v *bs.Episode, quality, date string) error {
	tmpFile, err := t.t411Client.DownloadTorrentByTerms(v.Show.Title, v.Season, v.Episode, "VOSTFR", quality, date)
	if err != nil {
		return err
	}
	if t.moveToTorrentsPath(tmpFile) {
		_, err := t.bsClient.EpisodeDownloaded(v.ID)
		if err != nil {
			return err
		}
		log.Printf("%s - S%02dE%02d downloaded\n", v.Show.Title, v.Season, v.Episode)
	}
	return nil
}

func (t *torrentManager) Run() {
	ticker := time.NewTicker(planningFetchFreq)
	defer ticker.Stop()
	for _ = range ticker.C {
		shows, err := t.bsClient.EpisodesList(-1, -1)
		if err != nil {
			log.Println(err.Error())
			continue
		}
		log.Printf("checking for episode(s) to download in %d shows...\n", len(shows))
		for _, s := range shows {
			for _, v := range s.Unseen {
				log.Printf("trying HD %s - S%02dE%02d\n", v.Show.Title, v.Season, v.Episode)
				if !v.User.Downloaded {
					err := t.DownloadWithQuality(&v, "TVripHD 720 [Rip HD depuis Source Tv HD]", v.Date)
					if err != nil && err == t411.ErrTorrentNotFound {
						log.Printf("trying SD %s - S%02dE%02d\n", v.Show.Title, v.Season, v.Episode)
						err = t.DownloadWithQuality(&v, "TVrip [Rip SD (non HD) depuis Source Tv HD/SD]", v.Date)
						if err != nil && err == t411.ErrTorrentNotFound {
							log.Printf("trying (no quality filter) %s - S%02dE%02d\n", v.Show.Title, v.Season, v.Episode)
							err = t.DownloadWithQuality(&v, "", v.Date)
							if err != nil && err != t411.ErrTorrentNotFound {
								log.Println(err.Error())
							}
						}
					}
				}
			}
		}
	}
}
