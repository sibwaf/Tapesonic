-- playlists

CREATE VIEW all_playlists (
    id,
    name,
    artwork_id,
    track_count,
    duration_ms,
    created_at,
    updated_at,
    user_id
) AS

WITH
    tapes_aggregate (user_id, id, track_count, total_duration_ms) AS (
        SELECT
            all_tracks.user_id AS user_id,
            tape_tracks.tape_id AS id,
            count(*) AS track_count,
            sum(all_tracks.duration_ms) AS total_duration_ms
        FROM tape_tracks
        JOIN all_tracks ON tape_tracks.track_id = all_tracks.id
        GROUP BY all_tracks.user_id, tape_tracks.tape_id
    ),
    recommended_playlists_aggregate (user_id, id, track_count, total_duration_ms) AS (
        SELECT
            all_tracks.user_id AS user_id,
            recommended_playlist_tracks.recommended_playlist_id AS id,
            count(*) AS track_count,
            sum(all_tracks.duration_ms) AS total_duration_ms
        FROM recommended_playlist_tracks
        JOIN all_tracks ON recommended_playlist_tracks.track_id = all_tracks.id
        GROUP BY recommended_playlist_tracks.recommended_playlist_id, all_tracks.user_id
    )

SELECT
    tapes.id AS id,
    tapes.name AS name,
    tapes.artwork_id AS artwork_id,
    tapes_aggregate.track_count AS track_count,
    tapes_aggregate.total_duration_ms AS duration_ms,
    tapes.created_at AS created_at,
    tapes.updated_at AS updated_at,
    users.id AS user_id
FROM tapes
JOIN users ON 1 = 1
LEFT JOIN tapes_aggregate ON tapes.id = tapes_aggregate.id AND users.id = tapes_aggregate.user_id
WHERE tapes.type = 'playlist'

UNION ALL

SELECT
    recommended_playlists.id AS id,
    recommended_playlists.name AS name,
    recommended_playlists.artwork_id AS artwork_id,
    recommended_playlists_aggregate.track_count AS track_count,
    recommended_playlists_aggregate.total_duration_ms AS duration_ms,
    recommended_playlists.created_at AS created_at,
    recommended_playlists.updated_at AS updated_at,
    recommended_playlists.user_id AS user_id
FROM recommended_playlists
LEFT JOIN recommended_playlists_aggregate ON recommended_playlists.id = recommended_playlists_aggregate.id AND recommended_playlists.user_id = recommended_playlists_aggregate.user_id
WHERE recommended_playlists_aggregate.track_count > 0;

-- tracks

CREATE VIEW all_playlist_tracks (
    playlist_id,
    track_id,
    track_index
) AS

SELECT
    tape_tracks.tape_id AS playlist_id,
    tape_tracks.track_id AS track_id,
    tape_tracks.list_index AS track_index
FROM tape_tracks
JOIN tapes ON tape_tracks.tape_id = tapes.id
WHERE tapes.type = 'playlist'

UNION ALL

SELECT
    recommended_playlist_tracks.recommended_playlist_id AS playlist_id,
    recommended_playlist_tracks.track_id AS track_id,
    recommended_playlist_tracks.track_index AS track_index
FROM recommended_playlist_tracks;
