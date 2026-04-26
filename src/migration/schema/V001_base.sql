-- users

CREATE TABLE users (
    id TEXT NOT NULL PRIMARY KEY,

    name TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL,

    api_key TEXT NOT NULL,

    created_at INTEGER NOT NULL
) STRICT;

CREATE UNIQUE INDEX users_name_uniq ON users (name);

-- background tasks

CREATE TABLE scheduled_tasks (
    id TEXT NOT NULL PRIMARY KEY,

    type TEXT NOT NULL,
    parameters TEXT NOT NULL,

    run_at INTEGER NOT NULL,
    run_attempted_at INTEGER,
    run_succeeded_at INTEGER
) STRICT;

CREATE UNIQUE INDEX scheduled_tasks_uniq ON scheduled_tasks (type, parameters);

-- yt-dlp

CREATE TABLE ytdlp_metadata_cache (
    url TEXT NOT NULL,
    metadata TEXT NOT NULL,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
) STRICT;

CREATE UNIQUE INDEX ytdlp_metadata_cache_url_uniq ON ytdlp_metadata_cache (url);

-- listenbrainz

CREATE TABLE listen_brainz_sessions (
    user_id TEXT NOT NULL PRIMARY KEY,

    token TEXT NOT NULL,
    username TEXT NOT NULL,

    is_scrobbling_enabled INTEGER NOT NULL,

    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,

    CONSTRAINT listen_brainz_sessions_user_id_fk FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
) STRICT;

-- last.fm

CREATE TABLE last_fm_sessions (
    user_id TEXT NOT NULL PRIMARY KEY,

    session_key TEXT NOT NULL,
    username TEXT NOT NULL,

    is_scrobbling_enabled INTEGER NOT NULL,

    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,

    CONSTRAINT last_fm_sessions_user_id_fk FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
) STRICT;

-- artworks

CREATE TABLE artworks (
    id TEXT NOT NULL PRIMARY KEY,
    deduplication_id TEXT NOT NULL,

    file_path TEXT NOT NULL,
    format TEXT NOT NULL,

    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
) STRICT;

CREATE UNIQUE INDEX artworks_deduplication_id_uniq ON artworks (deduplication_id);

-- artists

CREATE TABLE artists (
    id TEXT NOT NULL PRIMARY KEY,

    name TEXT NOT NULL,
    aliases TEXT NOT NULL,

    search_name TEXT NOT NULL,

    music_brainz_id TEXT
) STRICT;

CREATE UNIQUE INDEX artists_music_brainz_id_uniq ON artists (music_brainz_id);

-- stream cache

CREATE TABLE stream_cache (
    id TEXT NOT NULL PRIMARY KEY,

    filename TEXT NOT NULL,
    size INTEGER NOT NULL,
    content_type TEXT NOT NULL,

    created_at INTEGER NOT NULL,
    accessed_at INTEGER NOT NULL
) STRICT;
