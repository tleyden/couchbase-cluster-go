
FROM ubuntu:14.04

MAINTAINER Traun Leyden <tleyden@couchbase.com>

ENV GOPATH /opt/go

# Get dependencies
RUN apt-get update && apt-get install -y \
  git \
  golang \
  wget \
  curl 


# Build Sync Gateway
RUN mkdir -p $GOPATH && \
    go get -u -v -t github.com/tleyden/couchbase-cluster-go


# Add Sync Gateway launch script
ADD scripts/sync-gw-start /usr/local/bin/

