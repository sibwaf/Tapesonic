FROM golang:1.21-alpine AS builder

ARG APP_VERSION

WORKDIR /build
COPY src .

RUN sed -i "s/\"dev\"/\"$APP_VERSION\"/" build/version.go && \
    go build

FROM alpine:3.18

RUN apk add \
    --no-cache \
    --no-interactive \
    yt-dlp=2023.07.06-r1 \
    ffmpeg=6.0-r15

WORKDIR /app
COPY --from=builder /build/tapesonic /app/tapesonic

ENV TAPESONIC_PORT=8080
ENV TAPESONIC_YTDLP_PATH=/usr/bin/yt-dlp
ENV TAPESONIC_FFMPEG_PATH=/usr/bin/ffmpeg
ENV TAPESONIC_MEDIA_STORAGE_DIR=/media

EXPOSE $TAPESONIC_PORT/tcp

ENTRYPOINT ["/app/tapesonic"]
