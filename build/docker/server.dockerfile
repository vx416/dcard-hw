FROM golang:1.15.2 AS builder
WORKDIR /decard-work
ENV GO111MODULE=on 

COPY . .

RUN apt-get -qq update && apt-get -yqq install upx
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o service

RUN strip ./service
RUN upx -q -9 ./service

FROM alpine:latest
ARG BUILD_TIME
ARG SHA1_VER


RUN apk update && \
    apk upgrade && \
    apk add --no-cache curl tzdata && \
    apk add ca-certificates && \
    rm -rf /var/cache/apk/*

WORKDIR /decard-work
COPY --from=builder /decard-work/service /decard-work/service

RUN ls
ENV SHA1_VER=${SHA1_VER}
ENV BUILD_TIME=${BUILD_TIME}
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser && \
    chown -R appuser:appuser /decard-work
USER appuser

CMD ["./service", "server"]
