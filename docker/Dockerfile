FROM alpine:3.3

ENV CGO_ENABLED=0 \
    GOPATH=/tmp

RUN apk --update --no-cache upgrade && \
    apk add --update --no-cache git gcc && \
    apk add --update --no-cache --repository http://alpine.gliderlabs.com/alpine/edge/community go && \

    go get github.com/tleyden/couchbase-cluster-go/... && \
    mv $GOPATH/bin/* /usr/bin && \

    rm -rf /tmp/* && \
    apk del git go gcc && \
    
    addgroup cbclustergo && \
    adduser -D -g "" -s /bin/sh -G cbclustergo cbclustergo

USER cbclustergo
WORKDIR /home/cbclustergo

CMD ash
