FROM golang:1.15.2-alpine3.12 AS build-env

ENV GO111MODULE=on

RUN apk --no-cache add git make build-base

WORKDIR /go/src/kagiana

COPY . .

RUN mkdir -p /build
RUN go build -a  -ldflags="-s -w -extldflags \"-static\"" -o=/build/kagiana main.go

FROM alpine:3.12
# Timezone = Tokyo
RUN apk --no-cache add tzdata && \
    cp /usr/share/zoneinfo/Asia/Tokyo /etc/localtime

COPY --from=build-env /build/kagiana /build/kagiana
RUN chmod u+x /build/kagiana

ENTRYPOINT ["/build/kagiana", "server"]
