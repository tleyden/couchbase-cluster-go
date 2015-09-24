
FROM ubuntu:14.04

MAINTAINER Traun Leyden <tleyden@couchbase.com>

ENV GOPATH /opt/go
ENV GOROOT /usr/local/go

ENV PATH $PATH:$GOPATH/bin:$GOROOT/bin

# Get dependencies
RUN apt-get update && apt-get install -y \
    bc \
    curl \
    emacs \
    git \
    mercurial \
    wget && \
    apt-get clean

# Download and install jq
RUN wget http://stedolan.github.io/jq/download/linux64/jq && \
    chmod +x jq && \
    mv jq /usr/local/bin

# Download and install Go 1.4
RUN wget http://golang.org/dl/go1.4.2.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go1.4.2.linux-amd64.tar.gz && \
    rm go1.4.2.linux-amd64.tar.gz

# install go packages
RUN go get github.com/tools/godep && \
    go get github.com/nsf/gocode && \
    go get code.google.com/p/go.tools/cmd/goimports && \
    go get github.com/golang/lint/golint

# clone emacs conf
RUN git clone https://github.com/fgimenez/.emacs.d.git /root/.emacs.d && \
    cd /root/.emacs.d && \
    git checkout origin/go

# Install couchbase-cluster-go
RUN go get github.com/ClanStyles/couchbase-cluster-go/...

