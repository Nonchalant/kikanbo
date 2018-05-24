# kikanbo

kikanbo is command line tool for managing device connection status.
And [kikanbo](http://kikanbo.co.jp/english) is noodle I love 🍜

You can know current device connection status by mention on slack.

<img src="https://github.com/Nonchalant/kikanbo/blob/master/docs/kikanbo.png">

## Requirements

- [Go](https://golang.org/)
- [Slack Bot](https://my.slack.com/services/new/bot)


## Installation

```
$ brew install go // If not installed
$ export PATH=${HOME}/go/bin:${PATH} or export PATH=${GOPATH}/bin:${PATH}

$ go get -u github.com/Nonchalant/kikanbo
```


## Setup

Prepare `.env` file.

```
KIKANBO_TOKEN=xoxb-xxxxxxxxxx-xxxxxxxxxxxx-xxxxxxxxxxxxxxxxxxxxxxxx
BOT_MEMBER_ID=<SLACK_BOT_MEMBER_ID>
DEVICES_FILE_DIR=$HOME/.kikanbo // Option
DEVICES_FILE_PATH=$HOME/.kikanbo/devices.json // Option
```

### KIKANBO_TOKEN

Slack Bot Token

### BOT_MEMBER_ID

Slack Bot Member Id. You can get on slack app.

<img src="https://github.com/Nonchalant/kikanbo/blob/master/docs/slack_member_id.png" width="300">

### DEVICES_FILE_DIR

Device List File Dir. Default is `$HOME/.kikanbo`.

### DEVICES_FILE_PATH

Device List File Path. Default is `$HOME/.kikanbo/devices.json`.


## Usage

### Command

```
$ kikanbo run
```

### Slack

#### Defaults

e.g. `@kikanbo`

Show all devices.

#### Keyword

e.g. `@kikanbo iOS 10`

Show devices contains keyword (e.g. `iOS 10`)


## Notice

kikanbo is depends on macOS commands `instruments`. Please run on macOS.
