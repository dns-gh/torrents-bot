package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"os"
	"path/filepath"

	bs "github.com/dns-gh/bs-client/bsclient"
	t411 "github.com/dns-gh/t411-client/t411client"
)

type torrentManager struct {
	bsClient          *bs.BetaSeries
	t411Client        *t411.T411
	torrentsPath      string
	planningFetchFreq time.Duration
	singleShot        bool
	debug             bool
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

func makeTorrentManager(debug, single bool, torrentsPath string, planningFetchFreq int,
	bsKey, bsUsername, bsPassword, t411Username, t411Password, t411Token string) *torrentManager {
	t411Client, err := t411.NewT411ClientWithToken("", t411Username, t411Password, t411Token)
	if err != nil {
		log.Fatalln(err.Error())
	}

	bsClient, err := bs.NewBetaseriesClient(bsKey, bsUsername, bsPassword)
	if err != nil {
		log.Fatalln(err.Error())
	}

	manager := &torrentManager{
		bsClient:          bsClient,
		t411Client:        t411Client,
		torrentsPath:      makeTorrentPath(torrentsPath),
		planningFetchFreq: time.Duration(planningFetchFreq) * time.Minute,
		singleShot:        single,
		debug:             debug,
	}
	return manager
}

func (t *torrentManager) moveToTorrentsPath(tmp string) bool {
	defer func() {
		err := os.Remove(tmp)
		if err != nil && !strings.Contains(err.Error(), "The system cannot find the file specified") {
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

func (t *torrentManager) DownloadEpisodeWithQuality(v *bs.Episode, quality, date string) error {
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
		return nil
	}
	return fmt.Errorf("could not move torrent to output path")
}

func (t *torrentManager) DownloadSeriesWithQuality(v *bs.Show, season int, quality string) error {
	tmpFile, err := t.t411Client.DownloadTorrentByTerms(v.Title, season, 0, "VOSTFR", quality, "")
	if err != nil {
		return err
	}
	if t.moveToTorrentsPath(tmpFile) {
		episodes, err := t.bsClient.ShowsEpisodes(v.ID, season, 0)
		if err != nil {
			log.Println(err.Error())
		}
		for _, episode := range episodes {
			_, err := t.bsClient.EpisodeDownloaded(episode.ID)
			if err != nil {
				log.Println(err.Error())
			}
		}
		if season == 0 {
			log.Printf("%s - %s seasons / complete series downloaded\n", v.Title, v.Seasons)
		} else {
			log.Printf("%s - season %d complete downloaded\n", v.Title, season)
		}
		return nil
	}
	return fmt.Errorf("could not move torrent to output path")
}

func (t *torrentManager) DownloadSeries(v *bs.Show) error {
	t.print(fmt.Sprintf("trying HD %s - %s complete seasons", v.Title, v.Seasons))
	err := t.DownloadSeriesWithQuality(v, 0, "TVripHD 720 [Rip HD depuis Source Tv HD]")
	if isTorrentNotFound(err) {
		t.print(fmt.Sprintf("trying SD %s - %s complete seasons", v.Title, v.Seasons))
		err = t.DownloadSeriesWithQuality(v, 0, "TVrip [Rip SD (non HD) depuis Source Tv HD/SD]")
		if isTorrentNotFound(err) {
			t.print(fmt.Sprintf("trying NQ %s - %s complete seasons", v.Title, v.Seasons))
			err = t.DownloadSeriesWithQuality(v, 0, "")
		}
	}
	return err
}

func (t *torrentManager) DownloadSeason(v *bs.Show, season int) error {
	t.print(fmt.Sprintf("trying HD %s - season %d complete", v.Title, season))
	err := t.DownloadSeriesWithQuality(v, season, "TVripHD 720 [Rip HD depuis Source Tv HD]")
	if isTorrentNotFound(err) {
		t.print(fmt.Sprintf("trying SD %s - season %d complete", v.Title, season))
		err = t.DownloadSeriesWithQuality(v, season, "TVrip [Rip SD (non HD) depuis Source Tv HD/SD]")
		if isTorrentNotFound(err) {
			t.print(fmt.Sprintf("trying NQ %s - season %d complete", v.Title, season))
			err = t.DownloadSeriesWithQuality(v, season, "")
		}
	}
	return err
}

func (t *torrentManager) DownloadEpisode(v *bs.Episode) error {
	t.print(fmt.Sprintf("trying HD %s - S%02dE%02d", v.Show.Title, v.Season, v.Episode))
	err := t.DownloadEpisodeWithQuality(v, "TVripHD 720 [Rip HD depuis Source Tv HD]", v.Date)
	if isTorrentNotFound(err) {
		t.print(fmt.Sprintf("trying SD %s - S%02dE%02d", v.Show.Title, v.Season, v.Episode))
		err = t.DownloadEpisodeWithQuality(v, "TVrip [Rip SD (non HD) depuis Source Tv HD/SD]", v.Date)
		if isTorrentNotFound(err) {
			t.print(fmt.Sprintf("trying NQ %s - S%02dE%02d", v.Show.Title, v.Season, v.Episode))
			err = t.DownloadEpisodeWithQuality(v, "", v.Date)
		}
	}
	return err
}

func isTorrentNotFound(err error) bool {
	return err != nil && err == t411.ErrTorrentNotFound
}

func (t *torrentManager) print(text string) {
	if t.debug {
		log.Printf("%s\n", text)
	}
}

func logIfNotTorrentNotFound(err error) {
	// if the error is not of type "not Found", log it
	if err != nil && err != t411.ErrTorrentNotFound {
		log.Println(err.Error())
	}
}

func (t *torrentManager) download() {
	shows, err := t.bsClient.EpisodesList(-1, -1)
	if err != nil {
		log.Println(err.Error())
		return
	}
	log.Printf("checking for episode(s) to download in %d shows...\n", len(shows))
	for _, s := range shows {
		seasonsToSkip := make(map[int]struct{})
		for _, v := range s.Unseen {
			_, ok := seasonsToSkip[v.Season]
			if !v.User.Downloaded && !ok {
				show, err := t.bsClient.ShowDisplay(v.Show.ID)
				if err != nil {
					log.Println(err.Error())
					break
				}
				// if the episode is not special
				// and the episode is not the first
				// then try to download the complete series or season
				if v.Special != 1 && v.Episode <= 1 {
					t.t411Client.OnlyVerified(true)
					if show.Status == "Ended" {
						err := t.DownloadSeries(show)
						if err == nil {
							break
						}
						logIfNotTorrentNotFound(err)
						// try to download season by season if complete series is not found
						err = t.DownloadSeason(show, v.Season)
						if err == nil {
							seasonsToSkip[v.Season] = struct{}{}
							continue
						}
						logIfNotTorrentNotFound(err)
					}

					if show.Status == "Continuing" {
						err := t.DownloadSeason(show, v.Season)
						if err == nil {
							seasonsToSkip[v.Season] = struct{}{}
							continue
						}
						logIfNotTorrentNotFound(err)
					}
				}

				// download the unseen episode
				t.t411Client.OnlyVerified(false)
				err = t.DownloadEpisode(&v)
				logIfNotTorrentNotFound(err)
			}
		}
	}
}

// TODO: add webrip quality filter download just after SD quality ?
// it may be useful for shows displayed on websites first.
func (t *torrentManager) Run() {
	if t.singleShot {
		t.download()
		return
	}
	ticker := time.NewTicker(t.planningFetchFreq)
	defer ticker.Stop()
	for range ticker.C {
		t.download()
	}
}
