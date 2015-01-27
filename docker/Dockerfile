
FROM ubuntu:14.04

MAINTAINER Traun Leyden <tleyden@couchbase.com>

ENV GOPATH /opt/go

# Get dependencies
RUN apt-get update && apt-get install -y \
  git \
  golang \
  wget \
  curl 

# Build
RUN mkdir -p $GOPATH && \
    go get -u -v -t github.com/tleyden/couchbase-cluster-go/...

# Copy binary to directory in path
ADD $GOPATH/bin/couchbase-cluster /usr/local/bin/

