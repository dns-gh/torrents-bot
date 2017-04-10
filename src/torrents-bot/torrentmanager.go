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

func throttleNewT411Client(t411URL, t411Username, t411Password, t411Token string, maxRetry, waitInterval int) (*t411.T411, error) {
	if maxRetry <= 0 {
		maxRetry = int(^uint(0) >> 1)
	}
	// trying to connect once
	t411Client, err := t411.NewT411ClientWithToken(t411URL, t411Username, t411Password, t411Token)
	if err != nil {
		log.Println(err.Error())
	} else {
		return t411Client, nil
	}
	for i := 1; i < maxRetry; i++ {
		time.Sleep(time.Duration(waitInterval) * time.Second)
		t411Client, err := t411.NewT411ClientWithToken(t411URL, t411Username, t411Password, t411Token)
		if err != nil {
			log.Println(err.Error())
			continue
		}
		return t411Client, nil
	}
	return nil, fmt.Errorf("throttleNewT411Client, max retry reached after %d tries", maxRetry)
}

func makeTorrentManager(debug, single bool, torrentsPath string, planningFetchFreq int,
	bsKey, bsUsername, bsPassword, t411Username, t411Password, t411Token, t411URL string, t411MaxRetryt, t411waitInterval int) *torrentManager {
	t411Client, err := throttleNewT411Client(t411URL, t411Username, t411Password, t411Token, t411MaxRetryt, t411waitInterval)
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

func (t *torrentManager) DownloadEpisodeWithQuality(v *bs.Episode, alias, quality, date string) error {
	tmpFile, err := t.t411Client.DownloadTorrentByTerms(alias, v.Season, v.Episode, "VOSTFR", quality, date)
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

func (t *torrentManager) DownloadSeriesWithQuality(v *bs.Show, alias string, season int, quality string) error {
	tmpFile, err := t.t411Client.DownloadTorrentByTerms(alias, season, 0, "VOSTFR", quality, "")
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

func (t *torrentManager) DownloadSeriesWithAlias(v *bs.Show, alias string) (err error) {
	t.print(fmt.Sprintf("trying HD %s - %s complete seasons", alias, v.Seasons))
	if err = t.DownloadSeriesWithQuality(v, alias, 0, "TVripHD 720 [Rip HD depuis Source Tv HD]"); err != nil {
		t.print(fmt.Sprintf("trying SD %s - %s complete seasons", alias, v.Seasons))
		if err = t.DownloadSeriesWithQuality(v, alias, 0, "TVrip [Rip SD (non HD) depuis Source Tv HD/SD]"); err != nil {
			t.print(fmt.Sprintf("trying NQ %s - %s complete seasons", alias, v.Seasons))
			return t.DownloadSeriesWithQuality(v, alias, 0, "")
		}
	}
	return nil
}

func (t *torrentManager) DownloadSeries(v *bs.Show) error {
	err := t.DownloadSeriesWithAlias(v, v.Title)
	if err != nil {
		for _, alias := range v.Aliases {
			err := t.DownloadSeriesWithAlias(v, alias)
			if err != nil {
				continue
			}
		}
	}
	return err
}

func (t *torrentManager) DownloadSeasonWithAlias(v *bs.Show, alias string, season int) (err error) {
	t.print(fmt.Sprintf("trying HD %s - season %d complete", alias, season))
	if err = t.DownloadSeriesWithQuality(v, alias, season, "TVripHD 720 [Rip HD depuis Source Tv HD]"); err != nil {
		t.print(fmt.Sprintf("trying SD %s - season %d complete", alias, season))
		if err = t.DownloadSeriesWithQuality(v, alias, season, "TVrip [Rip SD (non HD) depuis Source Tv HD/SD]"); err != nil {
			t.print(fmt.Sprintf("trying NQ %s - season %d complete", alias, season))
			return t.DownloadSeriesWithQuality(v, alias, season, "")
		}
		return nil
	}
	return nil
}

func (t *torrentManager) DownloadSeason(v *bs.Show, season int) error {
	err := t.DownloadSeasonWithAlias(v, v.Title, season)
	if err != nil {
		for _, alias := range v.Aliases {
			err := t.DownloadSeasonWithAlias(v, alias, season)
			if err != nil {
				continue
			}
		}
	}
	return err
}

func (t *torrentManager) DownloadEpisodeWithAlias(v *bs.Episode, alias string) (err error) {
	t.print(fmt.Sprintf("trying HD %s - S%02dE%02d", alias, v.Season, v.Episode))
	if err = t.DownloadEpisodeWithQuality(v, alias, "TVripHD 720 [Rip HD depuis Source Tv HD]", v.Date); err != nil {
		t.print(fmt.Sprintf("trying SD %s - S%02dE%02d", alias, v.Season, v.Episode))
		if err = t.DownloadEpisodeWithQuality(v, alias, "TVrip [Rip SD (non HD) depuis Source Tv HD/SD]", v.Date); err != nil {
			t.print(fmt.Sprintf("trying NQ %s - S%02dE%02d", alias, v.Season, v.Episode))
			return t.DownloadEpisodeWithQuality(v, alias, "", v.Date)
		}
	}
	return err
}

func (t *torrentManager) DownloadEpisode(v *bs.Episode, aliases []string) error {
	err := t.DownloadEpisodeWithAlias(v, v.Show.Title)
	if err != nil {
		for _, alias := range aliases {
			err := t.DownloadEpisodeWithAlias(v, alias)
			if err != nil {
				continue
			}
		}
	}
	return err
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

func checkAliases(v *bs.Show) {
	if strings.Contains(v.Title, " (") {
		for _, alias := range v.Aliases {
			if !strings.ContainsAny(alias, "()") {
				return
			}
		}
		splitted := strings.Split(v.Title, " (")
		if len(splitted) >= 1 {
			v.Aliases = append(v.Aliases, splitted[0])
		}
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
				checkAliases(show)
				// download the unseen episode
				t.t411Client.OnlyVerified(false)
				err = t.DownloadEpisode(&v, show.Aliases)
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
