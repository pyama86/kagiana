FROM golang:latest AS build-env

ENV GO111MODULE=on
WORKDIR /go/src/kagiana
COPY . .

RUN mkdir -p /build
RUN go build -a  -ldflags="-s -w -extldflags \"-static\"" -o=/build/kagiana main.go

FROM alpine:3
# Timezone = Tokyo
RUN apk --no-cache add tzdata zlib && \
    apk add --upgrade --no-cache && \
    cp /usr/share/zoneinfo/Asia/Tokyo /etc/localtime

COPY --from=build-env /build/kagiana /build/kagiana
RUN chmod u+x /build/kagiana

ENTRYPOINT ["/build/kagiana", "server"]
