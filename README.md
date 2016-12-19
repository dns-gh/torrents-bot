# torrents-bot

[![Go Report Card](https://goreportcard.com/badge/github.com/dns-gh/torrents-bot)](https://goreportcard.com/report/github.com/dns-gh/torrents-bot)

torrents-bot is a Go bot managing tv show torrents. It uses t411 and betaseries APIs: https://api.t411.li/ and https://www.betaseries.com/api/

## Motivation

For fun, practice and to automatize torrents downloading tasks.

Feel free to join my efforts!

## Installation

- It requires Go language of course. You can set it up by downloading it here: https://golang.org/dl/
- Install it here C:/Go.
- Set your GOPATH, GOROOT and PATH environment variables with:

```
export GOROOT=C:/Go
export GOPATH=WORKING_DIR
export PATH=C:/Go/bin:${PATH}
```

or:

```
@working_dir $ source build/go.sh
```

and then set up your API keys/tokens/secrets in a torrents-bot.config file

```
{
    "BS_API_KEY": "your_betaseries_api_key",
    "bs-password": "your_betaseries_password",
    "bs-username": "your_betaseries_username",
    "debug": "false",
    "t411-password": "your_t411_password",
    "t411-username": "your_t411_username",
    "torrents-path": "./torrents"
}
```

You can get them here: http://www.t411.li/ and https://www.betaseries.com/api/

## Build and usage

```
@working_dir $ go get ./...
@working_dir $ go install torrents-bot
@working_dir $ bin/torrents-bot.exe
```
will :

- download every unseen tv show episode of your betaseries account from the t411 website into 'torrents-path' folder.
- mark them as 'downloaded' in your betaseries account.

Make sure to load torrents from this path with your torrent client in order to download them automatically.

## License

See the included LICENSE file.