FROM golang:1.21-alpine3.18 AS builder

ARG APP_VERSION

WORKDIR /build
COPY src .

RUN sed -i "s/\"dev\"/\"$APP_VERSION\"/" build/version.go
RUN apk add --no-cache --no-interactive build-base icu-dev
RUN CGO_ENABLED=1 go build --tags icu

FROM alpine:3.18

RUN apk add \
    --no-cache \
    --no-interactive \
    yt-dlp \
    ffmpeg \
    icu

WORKDIR /app
COPY --from=builder /build/tapesonic /app/tapesonic
COPY webapp/dist /app/webapp

ENV TAPESONIC_PORT=8080
ENV TAPESONIC_YTDLP_PATH=/usr/bin/yt-dlp
ENV TAPESONIC_FFMPEG_PATH=/usr/bin/ffmpeg
ENV TAPESONIC_WEBAPP_DIR=/app/webapp
ENV TAPESONIC_DATA_STORAGE_DIR=/data
ENV TAPESONIC_MEDIA_STORAGE_DIR=/media
ENV TAPESONIC_CACHE_DIR=/cache

EXPOSE $TAPESONIC_PORT/tcp

ENTRYPOINT ["/app/tapesonic"]
