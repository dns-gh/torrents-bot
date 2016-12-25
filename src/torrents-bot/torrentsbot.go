package main

import (
	"flag"
	"log"
	"path/filepath"

	"github.com/dns-gh/betterlog"
	conf "github.com/dns-gh/flagsconfig"
)

const (
	torrentsPathFlag      = "torrents-path"
	planningFetchFreqFlag = "freq"
	t411UsernameFlag      = "t411-username"
	t411PasswordFlag      = "t411-password"
	bsUsernameFlag        = "bs-username"
	bsPasswordFlag        = "bs-password"
	bsKeyFlag             = "BS_API_KEY"
	configFilename        = "torrents-bot.config"
	debugFlag             = "debug"
	singleFlag            = "single"
)

func main() {
	torrentsPath := flag.String(torrentsPathFlag, "./torrents", "[bot / t411] torrents folder")
	planningFetchFreq := flag.Int(planningFetchFreqFlag, 10, "[bot] planning fetch frequency in minutes")
	t411Username := flag.String(t411UsernameFlag, "", "[bot / t411] username")
	t411Password := flag.String(t411PasswordFlag, "", "[bot / t411] password")
	bsUsername := flag.String(bsUsernameFlag, "", "[bot / bs] username")
	bsPassword := flag.String(bsPasswordFlag, "", "[bot / bs] password")
	bsKey := flag.String(bsKeyFlag, "", "[bot / bs] api key")
	single := flag.Bool(singleFlag, false, "[bot] single shot mode")
	debug := flag.Bool(debugFlag, false, "[bot] debug mode")
	_, err := conf.NewConfig(configFilename)
	f, err := betterlog.MakeDateLogger(filepath.Join("debug", "tbot.log"))
	if err != nil {
		log.Fatalln(err.Error())
	}
	defer f.Close()
	log.Printf("[bot / t411] %s: %s\n", torrentsPathFlag, *torrentsPath)
	log.Printf("[bot] %s: %d\n", planningFetchFreqFlag, *planningFetchFreq)
	log.Printf("[bot / t411] %s: %s\n", t411UsernameFlag, *t411Username)
	log.Printf("[bot / t411] %s: %s\n", t411PasswordFlag, *t411Password)
	log.Printf("[bot / bs] %s: %s\n", bsUsernameFlag, *bsUsername)
	log.Printf("[bot / bs] %s: %s\n", bsPasswordFlag, *bsPassword)
	log.Printf("[bot / bs] %s: %s\n", bsKeyFlag, *bsKey)
	log.Printf("[bot] %s: %t\n", debugFlag, *debug)

	manager := makeTorrentManager(*debug, *single, *torrentsPath, *planningFetchFreq, *bsKey, *bsUsername, *bsPassword, *t411Username, *t411Password)
	token, err := manager.t411Client.GetToken()
	if err != nil {
		token = err.Error()
	}
	log.Printf("[bot / t411] token: %s\n", token)
	manager.Run()
}
