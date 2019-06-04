# repoTagger

repoTagger is a GO API to get starred repositories from GitHub users, store,
tag the repositories and search by tag.

## Installation

Download:
```bash
go get github.com/rschio/repoTagger
```

Set ENVs (optional), path of database and API port.
```bash
export REPOTAGGER_DBPATH=$GOPATH/src/github.com/rschio/repoTagger/repoTagger.db
export REPOTAGGER_PORT=8080
```

Install the binary:
```bash
go install github.com/rschio/repoTagger
```

## Test
Tests:
```bash
cd $GOPATH/src/github.com/rschio/repoTagger
go test ./...
```

## Run

Run:
```bash
repoTagger
```

Run on Docker:
```bash
cd $GOPATH/src/github.com/rschio/repoTagger
docker build -t repotagger .
docker run --name [name] -p [PORT]:[PORT] repotagger
```
