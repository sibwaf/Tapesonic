CREATE TABLE sources (
    id TEXT NOT NULL PRIMARY KEY,

    extractor_key TEXT NOT NULL,
    extracted_id TEXT NOT NULL,
    url TEXT NOT NULL,

    title TEXT NOT NULL,
    duration_ms INTEGER NOT NULL,

    uploader TEXT NOT NULL,
    uploader_id TEXT NOT NULL,
    uploaded_at INTEGER NOT NULL,

    album_artist TEXT,
    album_title TEXT,
    album_index INTEGER,
    track_artist TEXT,
    track_title TEXT,
    release_date INTEGER,

    artwork_id TEXT,

    management_policy TEXT NOT NULL,

    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,

    CONSTRAINT sources_artwork_id_fk FOREIGN KEY (artwork_id) REFERENCES artworks (id) ON DELETE SET NULL
) STRICT;

CREATE UNIQUE INDEX sources_url_uniq ON sources (url);

CREATE TABLE source_hierarchy (
    parent_id TEXT NOT NULL,
    child_id TEXT NOT NULL,
    list_index INTEGER NOT NULL,

    CONSTRAINT source_hierarchy_parent_id_fk FOREIGN KEY (parent_id) REFERENCES sources (id) ON DELETE CASCADE,
    CONSTRAINT source_hierarchy_child_id_fk FOREIGN KEY (child_id) REFERENCES sources (id) ON DELETE CASCADE
) STRICT;

CREATE UNIQUE INDEX source_hierarchy_uniq ON source_hierarchy (parent_id, child_id);

CREATE TABLE source_files (
    id TEXT NOT NULL PRIMARY KEY,

    source_id TEXT NOT NULL,
    format TEXT NOT NULL,
    codec TEXT NOT NULL,
    media_path TEXT NOT NULL,

    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,

    CONSTRAINT source_files_source_id_fk FOREIGN KEY (source_id) REFERENCES sources (id) ON DELETE RESTRICT
) STRICT;

CREATE UNIQUE INDEX source_files_source_id_uniq ON source_files (source_id);

CREATE TABLE source_tracks (
    id TEXT NOT NULL PRIMARY KEY,

    source_id TEXT NOT NULL,
    title TEXT NOT NULL,
    start_offset_ms INTEGER NOT NULL,
    end_offset_ms INTEGER NOT NULL,
    artist_id TEXT,

    search_title TEXT NOT NULL,

    CONSTRAINT source_tracks_source_id_fk FOREIGN KEY (source_id) REFERENCES sources (id) ON DELETE CASCADE,
    CONSTRAINT source_tracks_artist_id_fk FOREIGN KEY (artist_id) REFERENCES artists (id) ON DELETE SET NULL
) STRICT;
