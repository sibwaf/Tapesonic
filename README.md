# Tapesonic

A Subsonic-compatible music streaming server initially designed for playing YouTube mixtapes through your favorite Subsonic client. Can import, "cache" and stream anything that can be downloaded via [yt-dlp](https://github.com/yt-dlp/yt-dlp/). No external metadata (last.fm/MusicBrainz/...) required.

Tapesonic is able to fetch recommendation playlists from external services:
- ListenBrainz; matching only tracks you already have in your library for now
- last.fm; including auto-import of tracks you **don't** have in your library

## Warnings

### Everyday usage

***THIS PROJECT IS IN PROTOTYPING/PROOF-OF-CONCEPT STAGE AND IS NOT READY FOR EVERYDAY USE***

Expect:
- Awful UI/UX
- No user support
- Bugs
- No versioning, only bleeding-edge builds
- Barebones Subsonic API implementation - not all Subsonic clients may work
- Being forced to start over - no effort is directed towards maintaining compatibility between versions

### Copyright

This repository and all of its official artifacts (e.g. Docker images) do not contain any copyrighted content. This application is intended only for personal non-commercial use. All responsibility for any possible copyright infringement that could occur during the usage of this application lies solely on its users.

The author of this application does not condone piracy of copyrighted content.

## Quick start

### Run the container

```shell
docker run --rm -p 8080:8080 ghcr.io/sibwaf/tapesonic
```

### Access the webapp and setup your first user

Visit http://localhost:8080/setup (or whatever the host you're using for the docker daemon) to set the credentials for an admin user. This is a one-time setup.

### Add something to your library

1. Click `New Tape` at the top
2. Paste a YouTube/Bandcamp/... URL
3. Click `Search or import`
4. Click `Add all`
5. Click `Next` multiple times, adjusting properties as needed
6. Click `Create`

### Connect your favorite Subsonic client

Address: http://localhost:8080 (or whatever the host you're using for the docker daemon)

Authentication options:
- Username + password; only plaintext credentials supported (usually named "legacy auth" in clients)
- Username + API key; the key can be managed in the `Settings` page of the webapp

Compatibility is (kinda) tested with the following clients:
- [Feishin](https://github.com/jeffvli/feishin) - Windows, Linux, MacOS
- [Sonixd](https://github.com/jeffvli/sonixd) - Windows, Linux, MacOS
- [Tempus](https://github.com/eddyizm/tempus) - Android

## Hosting

### Configuration

All configuration options are passed through environment variables. Not all of the configuration options are listed here, see `src/config/config.go` for the full list.

#### General

- `TAPESONIC_PORT` - HTTP port to listen for requests; `8080` by default
- `TAPESONIC_STREAM_CACHE_SIZE` - target size for the `/cache` directory; `512m` (= 512 MiB) by default
- `TAPESONIC_RECOMMENDATION_PLAYLIST_SYNC_CRON` - cron schedule for recommendation playlist refreshing; `0 0 4 * * *` (each day at 4 AM) by default

#### last.fm

See [last.fm's documentation](https://www.last.fm/api/authentication) on how to create an API account and obtain API key/secret.

- `TAPESONIC_LASTFM_API_KEY` - your last.fm API key
- `TAPESONIC_LASTFM_API_SECRET` - your last.fm API secret

### Persistence

Tapesonic uses multiple directories inside the container to store its data:
- `/data` - the SQLite database with all the metadata; **keep this safe at all costs**
- `/media` - "cached" audio and thumbnails; Tapesonic will be able to auto-recover those in the future, but **keep this safe for now**
- `/cache` - cache for transcoded audio (and maybe more in the future); can be completely lost without any consequences

You can use Docker mounts to keep those directories persisted so you don't lose your data each time container gets restarted.

## Remotes

Tapesonic can act as a reverse proxy for other Subsonic-compatible servers ("remotes") combining all libraries into one so you don't have to switch between them in your Subsonic client of choice. All tracks from the combined library are eligible for matching in recommendation playlists.

You can manage remotes in the `Settings` page of the webapp. Only admin users are able to add/edit/delete remotes.

Each user must provide their own credentials to access the remote's library. Only legacy plaintext auth is supported for now.

Be careful if you have scrobbling to last.fm/ListenBrainz enabled both in Tapesonic and the remote and configure options in the webapp accordingly.

## External services

### ListenBrainz

See [ListenBrainz's documentation](https://listenbrainz.readthedocs.io/en/latest/users/api/index.html) on how to obtain an API token.

Go to the `Settings` page in the webapp to provide Tapesonic with your ListenBrainz token.

Following features will be available if a valid API token is provided:
- Scrobbling
- "Created for you" playlist fetching

### last.fm

Obtain and configure last.fm API key/secret for the server.

Go to the `Settings` page in the webapp and complete the authorization process to allow Tapesonic access to your last.fm account.

Following features will be available if last.fm is authorized:
- Scrobbling
- "Your library"/"Your mix"/"Your recommendations" radio auto-import as playlists

## What to expect in the future

- Better UI/UX
- Streams as radio for all of your "lofi hip hop beats to relax/study to 24/7" needs
- YouTube channels/playlists as podcasts
- Metadata enrichment from last.fm/MusicBrainz/...
- Automatic media search - just use the built-in search in your favorite Subsonic client and let Tapesonic do everything else
- (maybe) Non-Subsonic client/remote support
- (maybe) Lidarr integration - "wanted album" auto-download, media hand-off

## What not to expect

Tapesonic **will not** support streaming user-provided files - it is designed as URL-centered. For streaming any files you already downloaded from somewhere seek other options like [gonic](https://github.com/sentriz/gonic), [Navidrome](https://github.com/navidrome/navidrome), [Jellyfin](https://github.com/jellyfin/jellyfin) or others that already fullfil this use-case.

## Contributing

**No contributions will be accepted until the first real release (1.0.0)**. This includes bugs, incompatibilities, feature requests, pull requests.
