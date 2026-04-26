CREATE TABLE all_artwork_ids (
    id TEXT NOT NULL PRIMARY KEY,

    artwork_id TEXT,
    remote_artwork_id TEXT,

    CONSTRAINT all_artwork_ids_artwork_id_fk FOREIGN KEY (artwork_id) REFERENCES artworks (id) ON DELETE CASCADE,
    CONSTRAINT all_artwork_ids_remote_artwork_id_fk FOREIGN KEY (remote_artwork_id) REFERENCES remote_artworks (id) ON DELETE CASCADE
) STRICT;

CREATE TABLE all_track_ids (
    id TEXT NOT NULL PRIMARY KEY,

    source_track_id TEXT,
    remote_track_id TEXT,

    CONSTRAINT all_track_ids_source_track_id_fk FOREIGN KEY (source_track_id) REFERENCES source_tracks (id) ON DELETE CASCADE,
    CONSTRAINT all_track_ids_remote_track_id_fk FOREIGN KEY (remote_track_id) REFERENCES remote_tracks (id) ON DELETE CASCADE
) STRICT;

-- tapes

CREATE TABLE tapes (
    id TEXT NOT NULL PRIMARY KEY,

    name TEXT NOT NULL,
    type TEXT NOT NULL,
    artist_id TEXT,
    artwork_id TEXT,
    released_at INTEGER,

    created_by TEXT NOT NULL,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,

    search_name TEXT NOT NULL,

    CONSTRAINT tapes_artist_id_fk FOREIGN KEY (artist_id) REFERENCES artists (id) ON DELETE SET NULL,
    CONSTRAINT tapes_artwork_id_fk FOREIGN KEY (artwork_id) REFERENCES all_artwork_ids (id) ON DELETE SET NULL,
    CONSTRAINT tapes_created_by_fk FOREIGN KEY (created_by) REFERENCES users (id) ON DELETE RESTRICT
) STRICT;

CREATE TABLE tape_tracks (
    tape_id TEXT NOT NULL,
    track_id TEXT NOT NULL,
    list_index INTEGER NOT NULL,

    CONSTRAINT tape_tracks_tape_id_fk FOREIGN KEY (tape_id) REFERENCES tapes (id) ON DELETE CASCADE,
    CONSTRAINT tape_tracks_track_id_fk FOREIGN KEY (track_id) REFERENCES all_track_ids (id) ON DELETE CASCADE
) STRICT;

CREATE UNIQUE INDEX tape_tracks_uniq ON tape_tracks (tape_id, track_id);

-- recommendations

CREATE TABLE recommended_playlists (
    id TEXT NOT NULL PRIMARY KEY,

    provider TEXT NOT NULL,
    provider_playlist_id TEXT NOT NULL,

    user_id TEXT NOT NULL,
    name TEXT NOT NULL,
    artwork_id TEXT,

    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,

    sync_tag TEXT NOT NULL,

    CONSTRAINT recommended_playlists_user_id_fk FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT recommended_playlists_artwork_id_fk FOREIGN KEY (artwork_id) REFERENCES artworks (id) ON DELETE SET NULL
) STRICT;

CREATE UNIQUE INDEX recommended_playlists_uniq ON recommended_playlists (provider, provider_playlist_id);

CREATE TABLE recommended_playlist_tracks (
    recommended_playlist_id TEXT NOT NULL,

    artist TEXT NOT NULL,
    title TEXT NOT NULL,
    track_id TEXT NOT NULL,
    track_index INTEGER NOT NULL,

    CONSTRAINT recommended_playlist_tracks_recommended_playlist_id_fk FOREIGN KEY (recommended_playlist_id) REFERENCES recommended_playlists (id) ON DELETE CASCADE,
    CONSTRAINT recommended_playlist_tracks_track_id_fk FOREIGN KEY (track_id) REFERENCES all_track_ids (id) ON DELETE CASCADE
) STRICT;
