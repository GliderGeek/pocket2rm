### Update 15/09/2020: The code has been updated to work with the new software version of reMarkable which uses a trash instead of immediate deletion.

# pocket2rm
`pocket2rm` is a tool to get articles from read-later platform [pocket](https://app.getpocket.com/) on the [reMarkable paper tablet](https://remarkable.com/). 

- retrieve URLs for articles from pocket (last 10)
- PDFs are downloaded directly, webpages are converted to a [readable format](https://github.com/go-shiori/go-readability) and converted to epub
- runs on reMarkable directly, does not use reMarkable cloud.
- sync is user-triggered (removing synchronization file)

## Prerequisites
- SSH connection with remarkable: [https://remarkablewiki.com/tech/ssh](https://remarkablewiki.com/tech/ssh)
- golang + dependencies
- scp

## Installation
- create a pocket application: [https://getpocket.com/developer/apps/new](https://getpocket.com/developer/apps/new) to obtain a `consumerKey`. The application only needs the 'Retrieve' permission.

- Inside cmd/pocket2rm-setup folder: `go build main.go`
- Inside cmd/pocket2rm folder: `GOOS=linux GOARCH=arm GOARM=7 go build -o pocket2rm.arm`
- Inside cmd/pocket2rm-reload folder: `GOOS=linux GOARCH=arm GOARM=7 go build -o pocket2rm-reload.arm`

- execute `cmd/pocket2rm-setup/main`

- `scp $HOME/.pocket2rm root@10.11.99.1:/home/root/.`
- `scp cmd/pocket2rm/pocket2rm.arm root@10.11.99.1:/home/root/.`
- `scp cmd/pocket2rm/pocket2rm.service root@10.11.99.1:/etc/systemd/system/.`
- `scp cmd/pocket2rm-reload/pocket2rm-reload.arm root@10.11.99.1:/home/root/.`
- `scp cmd/pocket2rm-reload/pocket2rm-reload.service root@10.11.99.1:/etc/systemd/system/.`

- `ssh root@10.11.99.1 systemctl start pocket2rm-reload`

## Improvements
- input consumerKey in popup (removes commandline run)
- move scp commands to pocket2rm-setup
- provide binaries
- images in epubs
- improve repo structure (duplicate utils, dependencies)

## Non-goals
- support other read-later services / e-reader targets

## Alternatives:
- there is [google-chrome plugin](https://chrome.google.com/webstore/detail/send-to-remarkable/mcfkooagiaelmfpkgegmbobdcpcbdbgh) which sends articles to reMarkable
- reMarkable is planning to release an [offical chrome plugin](https://support.remarkable.com/hc/en-us/articles/360006830977-Read-on-reMarkable-Google-Chrome-plug-in)

## Credit
There were a few projects, apart from the dependencies, which were very helpful:
- https://github.com/Evidlo/remarkable_news
- https://github.com/koreader/koreader
- https://github.com/nick8325/remarkable-fs
- https://github.com/jessfraz/morningpaper2remarkable
