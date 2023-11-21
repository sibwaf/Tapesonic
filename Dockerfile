FROM golang:1.21-alpine AS builder

ARG APP_VERSION

WORKDIR /build
COPY src .

RUN apk add --no-cache --no-interactive build-base

RUN sed -i "s/\"dev\"/\"$APP_VERSION\"/" build/version.go && \
    CGO_ENABLED=1 go build

FROM alpine:3.18

RUN apk add \
    --no-cache \
    --no-interactive \
    yt-dlp \
    ffmpeg

WORKDIR /app
COPY --from=builder /build/tapesonic /app/tapesonic
COPY webapp/dist /app/webapp

ENV TAPESONIC_PORT=8080
ENV TAPESONIC_YTDLP_PATH=/usr/bin/yt-dlp
ENV TAPESONIC_FFMPEG_PATH=/usr/bin/ffmpeg
ENV TAPESONIC_WEBAPP_DIR=/app/webapp
ENV TAPESONIC_DATA_STORAGE_DIR=/data
ENV TAPESONIC_MEDIA_STORAGE_DIR=/media

EXPOSE $TAPESONIC_PORT/tcp

ENTRYPOINT ["/app/tapesonic"]
