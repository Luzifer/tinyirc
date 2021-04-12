[![Go Report Card](https://goreportcard.com/badge/github.com/Luzifer/tinyirc)](https://goreportcard.com/report/github.com/Luzifer/tinyirc)
![](https://badges.fyi/github/license/Luzifer/tinyirc)
![](https://badges.fyi/github/downloads/Luzifer/tinyirc)
![](https://badges.fyi/github/latest-release/Luzifer/tinyirc)
![](https://knut.in/project-status/tinyirc)

# Luzifer / tinyirc

This is a very tiny CLI IRC client which takes raw IRC commands and sends them to the established connection.

## Usage

```console
# go install github.com/Luzifer/tinyirc@latest

# cat .env
JOIN=#tezrian
NICK=tezrian
PORT=6697
SEND_LIMIT=1s
SERVER=irc.chat.twitch.tv
SERVER_PASS=oauth:***
TLS=true
USER=tezrian

# echo "PRIVMSG #tezrian :Test" | envrun tinyirc
:tezrian!tezrian@tezrian.tmi.twitch.tv JOIN #tezrian
```
