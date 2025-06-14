[![license](https://img.shields.io/github/license/revilo732/rss2webex-hook.svg)](https://github.com/revilo732/rss2webex-hook/blob/master/LICENSE)

# RSS2Webex-Hook

This project is a self-hosted utility which will make HTTP POST
requests to Webex chanels when new items appear in an RSS feed.

## Credit

This is just a slightly modified version of the [rss2hook](https://github.com/skx/rss2hook) repo, tailored for Webex.


## Installation


## Build with Go Modules (Go 1.11 or higher) - golang.org  

    git clone https://github.com/revilo732/rss2webex-hook ;# make sure to clone outside of GOPATH
    cd rss2webex-hook
    go install



## Setup

There are three parts to the setup:

* [Create webhook url for room](https://apphub.webex.com/applications/incoming-webhooks-cisco-systems-38054-23307-75252)
* Configure the list of feeds and the corresponding hooks to POST to.
* Ensure the program is running.

For the first create a configuration-file like so:

    http://example.com/feed.rss = https://webhook.example.com/notify/me

(There is a sample configuration file [sample.cfg](sample.cfg) which
will demonstrate this more verbosely.)

You can use your favourite supervision tool to launch the deamon, but you
can test interactively like so:

     $ rss2hook -config ./sample.cfg



### Sample Webhook Receiver

There is a simple webserver located beneath [webhook/](webhook/) which
will listen upon http://localhost:8080, and dump any POST submission to the
console.

You can launch it like so:

     cd webhook/
     go run webhook.go

Testing it via `curl` would look like this:

      $ curl --header "Content-Type: application/json"  \
      --request POST \
      --data '{"username":"blah","password":"blah"}' \
      http://localhost:8080/

The [sample.cfg](sample.cfg) file will POST to this end-point so you can
see how things work:

    $ rss2hook --config=sample.cfg



## Implementation Notes

* By default the server will poll all configured feeds immediately
upon startup.
   * It will look for changes every five minutes.
* To ensure items are only announced once state is kept on the filesystem.
   * Beneath the directory `~/.rss2hook/seen/`.
* Feed items are submitted to the webhook as JSON.

