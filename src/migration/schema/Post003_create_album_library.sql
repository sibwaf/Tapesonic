CREATE VIEW all_albums (
    id,
    artist_id,
    artist,
    title,
    artwork_id,
    track_count,
    duration_ms,
    released_at,
    added_at,
    played_at,
    total_listened_ms,
    search_title,
    user_id
) AS

WITH
    raw_all_tracks (user_id, track_id, duration_ms) AS (
        SELECT
            users.id AS user_id,
            source_tracks.id AS track_id,
            (source_tracks.end_offset_ms - source_tracks.start_offset_ms) AS duration_ms
        FROM source_tracks
        JOIN users ON 1 = 1

        UNION ALL

        SELECT
            remote_tracks_to_users.user_id AS user_id,
            remote_tracks.id AS track_id,
            remote_tracks.duration_ms AS duration_ms
        FROM remote_tracks
        JOIN remote_tracks_to_users ON remote_tracks.id = remote_tracks_to_users.remote_track_id
    ),
    tapes_aggregate (user_id, id, track_count, total_duration_ms, last_listened_at, total_listened_ms) AS (
        SELECT
            raw_all_tracks.user_id AS user_id,
            tape_tracks.tape_id AS id,
            count(*) AS track_count,
            sum(raw_all_tracks.duration_ms) AS total_duration_ms,
            max(track_listens.last_listened_at) AS last_listened_at,
            sum(raw_all_tracks.duration_ms * track_listens.listen_count) AS total_listened_ms
        FROM tape_tracks
        JOIN raw_all_tracks ON tape_tracks.track_id = raw_all_tracks.track_id
        LEFT JOIN track_listens ON tape_tracks.track_id = track_listens.track_id AND raw_all_tracks.user_id = track_listens.user_id
        GROUP BY raw_all_tracks.user_id, tape_tracks.tape_id
    ),
    remote_albums_aggregate (user_id, id, track_count, total_duration_ms, last_listened_at, total_listened_ms) AS (
        SELECT
            remote_tracks_to_users.user_id AS user_id,
            remote_albums.id AS id,
            count(*) AS track_count,
            sum(remote_tracks.duration_ms) AS total_duration_ms,
            max(track_listens.last_listened_at) AS last_listened_at,
            sum(remote_tracks.duration_ms * track_listens.listen_count) AS total_listened_ms
        FROM remote_tracks
        JOIN remote_tracks_to_users ON remote_tracks.id = remote_tracks_to_users.remote_track_id
        JOIN remote_albums ON remote_tracks.remote_id = remote_albums.remote_id AND remote_tracks.album_id = remote_albums.album_id
        LEFT JOIN track_listens ON remote_tracks.id = track_listens.track_id AND remote_tracks_to_users.user_id = track_listens.user_id
        GROUP BY remote_tracks_to_users.user_id, remote_albums.id
    )

SELECT
    tapes.id AS id,
    artists.id AS artist_id,
    artists.name AS artist,
    tapes.name AS title,
    tapes.artwork_id AS artwork_id,
    tapes_aggregate.track_count AS track_count,
    tapes_aggregate.total_duration_ms AS duration_ms,
    tapes.released_at AS released_at,
    tapes.created_at AS added_at,
    tapes_aggregate.last_listened_at AS played_at,
    tapes_aggregate.total_listened_ms AS total_listened_ms,
    tapes.search_name AS search_title,
    users.id AS user_id
FROM tapes
JOIN users ON 1 = 1
LEFT JOIN tapes_aggregate ON tapes.id = tapes_aggregate.id AND users.id = tapes_aggregate.user_id
LEFT JOIN artists ON tapes.artist_id = artists.id
WHERE tapes.type = 'album'

UNION ALL

SELECT
    remote_albums.id AS id,
    artists.id AS artist_id,
    artists.name AS artist,
    remote_albums.title AS title,
    remote_artworks.id AS artwork_id,
    remote_albums_aggregate.track_count AS track_count,
    remote_albums_aggregate.total_duration_ms AS duration_ms,
    remote_albums.released_at AS released_at,
    remote_albums.added_at AS added_at,
    remote_albums_aggregate.last_listened_at AS played_at,
    remote_albums_aggregate.total_listened_ms AS total_listened_ms,
    remote_albums.search_title AS search_title,
    remote_albums_to_users.user_id AS user_id
FROM remote_albums
JOIN remote_albums_to_users ON remote_albums.id = remote_albums_to_users.remote_album_id
LEFT JOIN remote_albums_aggregate ON remote_albums.id = remote_albums_aggregate.id AND remote_albums_to_users.user_id = remote_albums_aggregate.user_id
LEFT JOIN remote_artworks ON remote_albums.remote_id = remote_artworks.remote_id AND remote_albums.artwork_id = remote_artworks.artwork_id
LEFT JOIN remote_artists ON remote_albums.remote_id = remote_artists.remote_id AND remote_albums.artist_id = remote_artists.artist_id
LEFT JOIN artists ON remote_artists.tapesonic_artist_id = artists.id
WHERE remote_albums_aggregate.track_count > 0;
