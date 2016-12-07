# torrents-bot

torrents-bot is a Go bot managing torrents and videos. It uses t411 and betaseries APIs: https://api.t411.li/ and https://www.betaseries.com/api/

## Motivation

For fun, practice and to automatize torrents/video downloading tasks.

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

and then set up your API keys/tokens/secrets:

```
export T411_USERNAME="your_t411_username"
export T411_PASSWORD="your_t411_password"
export BS_API_KEY="your_betaseries_api_key"
```

You can find get them here: http://www.t411.li/ and https://www.betaseries.com/api/

## Build and usage

```
@working_dir $ go install torrents-bot
@working_dir $ bin/torrents-bot.exe
```

## License

See the included LICENSE file.