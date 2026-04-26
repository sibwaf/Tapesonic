-- base

CREATE TABLE remotes (
    id TEXT NOT NULL PRIMARY KEY,

    name TEXT NOT NULL,
    type TEXT NOT NULL,
    url TEXT NOT NULL,

    is_scrobble_replication_enabled INTEGER NOT NULL,
    is_external_scrobbling_enabled INTEGER NOT NULL,

    created_at INTEGER NOT NULL
) STRICT;

CREATE TABLE remotes_to_users (
    remote_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    credentials TEXT,

    CONSTRAINT remotes_to_users_remote_id_fk FOREIGN KEY (remote_id) REFERENCES remotes (id) ON DELETE CASCADE,
    CONSTRAINT remotes_to_users_user_id_fk FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
) STRICT;

CREATE UNIQUE INDEX remotes_to_users_uniq ON remotes_to_users (remote_id, user_id);

-- artworks

CREATE TABLE remote_artworks (
    id TEXT NOT NULL PRIMARY KEY,

    remote_id TEXT NOT NULL,
    artwork_id TEXT NOT NULL,

    CONSTRAINT remote_artworks_remote_id_fk FOREIGN KEY (remote_id) REFERENCES remotes (id) ON DELETE CASCADE
) STRICT;

CREATE UNIQUE INDEX remote_artworks_uniq ON remote_artworks (remote_id, artwork_id);

-- artists

CREATE TABLE remote_artists (
    id TEXT NOT NULL PRIMARY KEY,

    remote_id TEXT NOT NULL,
    artist_id TEXT NOT NULL,

    name TEXT NOT NULL,
    artwork_id TEXT,
    music_brainz_id TEXT,
    tapesonic_artist_id TEXT,

    search_name TEXT NOT NULL,

    CONSTRAINT remote_artists_remote_id_fk FOREIGN KEY (remote_id) REFERENCES remotes (id) ON DELETE CASCADE,
    CONSTRAINT remote_artists_tapesonic_artist_id_fk FOREIGN KEY (tapesonic_artist_id) REFERENCES artists (id) ON DELETE RESTRICT
) STRICT;

CREATE UNIQUE INDEX remote_artists_uniq ON remote_artists (remote_id, artist_id);

CREATE TABLE remote_artists_to_users (
    remote_artist_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    sync_tag TEXT NOT NULL,

    CONSTRAINT remote_artists_to_users_remote_artist_id_fk FOREIGN KEY (remote_artist_id) REFERENCES remote_artists (id) ON DELETE CASCADE,
    CONSTRAINT remote_artists_to_users_user_id_fk FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
) STRICT;

CREATE UNIQUE INDEX remote_artists_to_users_uniq ON remote_artists_to_users (remote_artist_id, user_id);

-- albums

CREATE TABLE remote_albums (
    id TEXT NOT NULL PRIMARY KEY,

    remote_id TEXT NOT NULL,
    album_id TEXT NOT NULL,

    title TEXT NOT NULL,
    artwork_id TEXT,
    artist_id TEXT,
    added_at INTEGER NOT NULL,
    released_at INTEGER,

    search_title TEXT NOT NULL,

    CONSTRAINT remote_albums_remote_id_fk FOREIGN KEY (remote_id) REFERENCES remotes (id) ON DELETE CASCADE
) STRICT;

CREATE UNIQUE INDEX remote_albums_uniq ON remote_albums (remote_id, album_id);

CREATE TABLE remote_albums_to_users (
    remote_album_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    sync_tag TEXT NOT NULL,

    CONSTRAINT remote_albums_to_users_remote_album_id_fk FOREIGN KEY (remote_album_id) REFERENCES remote_albums (id) ON DELETE CASCADE,
    CONSTRAINT remote_albums_to_users_user_id_fk FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
) STRICT;

CREATE UNIQUE INDEX remote_albums_to_users_uniq ON remote_albums_to_users (remote_album_id, user_id);

-- tracks

CREATE TABLE remote_tracks (
    id TEXT NOT NULL PRIMARY KEY,

    remote_id TEXT NOT NULL,
    track_id TEXT NOT NULL,

    title TEXT NOT NULL,
    artwork_id TEXT,
    artist TEXT,
    artist_id TEXT,
    album TEXT,
    album_id TEXT,
    album_index INTEGER,
    duration_ms INTEGER NOT NULL,

    search_title TEXT NOT NULL,

    CONSTRAINT remote_tracks_remote_id_fk FOREIGN KEY (remote_id) REFERENCES remotes (id) ON DELETE CASCADE
) STRICT;

CREATE UNIQUE INDEX remote_tracks_uniq ON remote_tracks (remote_id, track_id);

CREATE TABLE remote_tracks_to_users (
    remote_track_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    sync_tag TEXT NOT NULL,

    CONSTRAINT remote_tracks_to_users_remote_track_id_fk FOREIGN KEY (remote_track_id) REFERENCES remote_tracks (id) ON DELETE CASCADE,
    CONSTRAINT remote_tracks_to_users_user_id_fk FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
) STRICT;

CREATE UNIQUE INDEX remote_tracks_to_users_uniq ON remote_tracks_to_users (remote_track_id, user_id);
