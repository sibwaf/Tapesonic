CREATE TABLE track_listens (
    user_id TEXT NOT NULL,
    track_id TEXT NOT NULL,
    listen_count INTEGER NOT NULL,
    last_listened_at INTEGER NOT NULL,

    CONSTRAINT track_listens_user_id_fk FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT track_listens_track_id_fk FOREIGN KEY (track_id) REFERENCES all_track_ids (id) ON DELETE CASCADE
) STRICT;

CREATE UNIQUE INDEX track_listens_uniq ON track_listens (user_id, track_id);
