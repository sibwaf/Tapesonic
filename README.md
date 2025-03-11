# Tapesonic

A Subsonic-compatible music streaming server initially designed for playing YouTube mixtapes through your favorite Subsonic client. Can import and stream anything that can be downloaded via [yt-dlp](https://github.com/yt-dlp/yt-dlp/). No external metadata (last.fm/MusicBrainz/...) required.

Tapesonic is also able to fetch recommendation playlists from external services:
- ListenBrainz; only tracks you already have in your library
- last.fm; with auto-import of tracks you don't already have in your library - **get actual recommendations in your self-hosted streaming**!

Tapesonic can act as a proxy to a different Subsonic-compatible server combining both libraries so you don't have to switch between multiple servers in your Subsonic client of choice.

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

```shell
docker run --rm -p 8080:8080 -e TAPESONIC_USERNAME=user -e TAPESONIC_PASSWORD=pass ghcr.io/sibwaf/tapesonic
```

### Access the UI

http://localhost:8080 (or whatever the host you're using for the docker daemon)

Credentials: `user`/`pass`

Import something:
1. Click `New Tape` at the top
2. Paste a YouTube/Bandcamp URL
3. Click `Import`
4. Click `Add all`
5. Click `Next` multiple times, adjusting data as needed
6. Click `Create`

### Connect your favorite Subsonic client

http://localhost:8080 (or whatever the host you're using for the docker daemon)

Credentials: `user`/`pass`

Compatibility is (kinda) tested with the following clients:
- [Feishin](https://github.com/jeffvli/feishin) - Windows, Linux, MacOS
- [Sonixd](https://github.com/jeffvli/sonixd) - Windows, Linux, MacOS
- [Tempo](https://github.com/CappielloAntonio/tempo) - Android

Compatibility was (kinda) tested with the following clients in the past:
- [Supersonic](https://github.com/dweymouth/supersonic) - Windows, Linux, MacOS

## Hosting

### Configuration

All configuration options are passed through environment variables. Not all of the configuration options are listed here, see `src/config/config.go` for the full list.

#### General

- `TAPESONIC_PORT` - HTTP port to listen for requests; 8080 by default
- `TAPESONIC_USERNAME` - username for accessing the server from web UI and Subsonic clients
- `TAPESONIC_PASSWORD` - password for accessing the server from web UI and Subsonic clients
- `TAPESONIC_SCROBBLE_MODE` - controls which tracks will be scrobbled to external services like last.fm/ListenBrainz; scrobbling to non-configured external services will be silently skipped
  - `none` - nothing will be scrobbled to external services; this is the default value
  - `tapesonic` - only tracks hosted by this Tapesonic instance will be scrobbled to external services
  - `all` - everything played through this Tapesonic instance (both Tapesonic's own library and proxied library) will be scrobbled to external services

#### Proxying

- `TAPESONIC_SUBSONIC_PROXY_URL` - URL for the server you want Tapesonic to proxy including the protocol; ex. http://gonic.myserver.local
- `TAPESONIC_SUBSONIC_PROXY_USERNAME` - username Tapesonic will use when accessing the proxied server
- `TAPESONIC_SUBSONIC_PROXY_PASSWORD` - password Tapesonic will use when accessing the proxied server

If proxying is configured, Tapesonic will serve both it's own library as well as the library from the proxied server. This means that if you already have a Subsonic-compatible server running you can point Tapesonic to it and configure your clients to only access Tapesonic without the need to switch between servers. This also allows Tapesonic to use the proxied library for matching tracks of external playlists.

Be careful if you have scrobbling to last.fm/ListenBrainz enabled both in Tapesonic and the proxied server and configure the `TAPESONIC_SCROBBLE_MODE` accordingly so you don't get duplicated scrobbles.

#### ListenBrainz

- `TAPESONIC_LISTENBRAINZ_TOKEN` - your ListenBrainz API token

See [ListenBrainz's documentation](https://listenbrainz.readthedocs.io/en/latest/users/api/index.html) on how to obtain an API token.

Following features will be enabled if a valid API token is configured:
- Scrobbling (if scrobbling is enabled in general configuration)
- "Created for you" playlist auto-import using tracks from your library; import happens each day at 04:00 by default

#### last.fm

- `TAPESONIC_LASTFM_API_KEY` - your last.fm API key
- `TAPESONIC_LASTFM_API_SECRET` - your last.fm API secret

See [last.fm's documentation](https://www.last.fm/api/authentication) on how to create an API account and obtain API key/secret.

To complete last.fm configuration you'll have to go to `Settings` in the web UI and complete the authorization process to allow Tapesonic to access your account.

Following features will be enabled if last.fm is configured:
- Scrobbling (if scrobbling is enabled in general configuration)
- "Your library"/"Your mix"/"Your recommendations" radio auto-import as playlists using both tracks from your library and auto-importing the missing ones; import happens each day at 04:00 by default

### Persistence

Tapesonic uses multiple directories inside the container to store its data:
- `/data` - the SQLite database with all the metadata; **keep this safe at all costs**
- `/media` - cached audio and thumbnails; Tapesonic will be able to auto-recover those in the future, but **keep this safe for now**
- `/cache` - cache for transcoded audio (and maybe more in the future); can be completely lost without any consequences

You can use Docker mounts to keep those directories persisted so you don't lose your data each time container gets restarted.

## What to expect in the future

- Better UI/UX
- Streams as radio for all of your "lofi hip hop beats to relax/study to 24/7" needs
- YouTube channels/playlists as podcasts
- Metadata enrichment from last.fm/MusicBrainz/...
- Automatic media search - just use the built-in search in your favorite Subsonic client and let Tapesonic do everything else
- Support for multi-user usage
- (maybe) Non-Subsonic client/proxying support (most likely as a separate project)
- (maybe) Lidarr integration - "wanted album" auto-download, media hand-off

## What not to expect

Tapesonic **will not** support streaming user-provided files - it is designed as URL-centered. For streaming any files you already downloaded from somewhere seek other options like [gonic](https://github.com/sentriz/gonic), [Navidrome](https://github.com/navidrome/navidrome), [Jellyfin](https://github.com/jellyfin/jellyfin) or others that already fullfil this use-case.

## Contributing

**No contributions will be accepted until the first real release (1.0.0)**. This includes bugs, incompatibilities, feature requests, pull requests.
