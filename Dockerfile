FROM golang

ADD . /go/src/github.com/rschio/repoTagger

RUN go get github.com/mattn/go-sqlite3
RUN go install github.com/rschio/repoTagger

ENTRYPOINT /go/bin/repoTagger

ENV REPOTAGGER_DBPATH=/go/src/github.com/rschio/repoTagger/repoTagger.db
ENV REPOTAGGER_PORT=8080

EXPOSE 8080
