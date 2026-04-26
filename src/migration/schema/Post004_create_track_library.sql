CREATE VIEW all_tracks (
    id,
    source_id,
    remote_id,
    remote_track_id,
    artist,
    artist_id,
    album,
    album_id,
    album_artist_id,
    album_released_at,
    album_track_index,
    title,
    artwork_id,
    duration_ms,
    played_at,
    search_artist,
    search_album,
    search_title,
    user_id
) AS

WITH track_albums (track_id, album_id, album_name, album_search_name, album_artist_id, album_artwork_id, album_released_at, list_index) AS (
    SELECT
        tape_tracks.track_id AS track_id,
        tapes.id AS album_id,
        tapes.name AS album_name,
        tapes.search_name AS album_search_name,
        tapes.artist_id AS album_artist_id,
        tapes.artwork_id AS album_artwork_id,
        tapes.released_at AS album_released_at,
        tape_tracks.list_index AS list_index
    FROM tape_tracks
    JOIN tapes ON tape_tracks.tape_id = tapes.id
    WHERE tapes.type = 'album'

    UNION ALL

    SELECT
        remote_tracks.id AS track_id,
        remote_albums.id AS album_id,
        remote_albums.title AS album_name,
        remote_albums.search_title AS search_album,
        album_artists.id AS album_artist_id,
        remote_artworks.id AS album_artwork_id,
        remote_albums.released_at AS album_released_at,
        remote_tracks.album_index AS list_index
    FROM remote_tracks
    JOIN remote_albums ON remote_tracks.remote_id = remote_albums.remote_id AND remote_tracks.album_id = remote_albums.album_id
    LEFT JOIN remote_artists remote_album_artists ON remote_albums.remote_id = remote_album_artists.remote_id AND remote_albums.artist_id = remote_album_artists.artist_id
    LEFT JOIN artists album_artists ON remote_album_artists.tapesonic_artist_id = album_artists.id
    JOIN remote_artworks ON remote_albums.remote_id = remote_artworks.remote_id AND remote_albums.artwork_id = remote_artworks.artwork_id
)

SELECT
    source_tracks.id AS id,
    sources.id AS source_id,
    NULL AS remote_id,
    NULL AS remote_track_id,
    artists.name AS artist,
    artists.id AS artist_id,
    track_albums.album_name AS album,
    track_albums.album_id AS album_id,
    track_albums.album_artist_id AS album_artist_id,
    track_albums.album_released_at AS album_released_at,
    track_albums.list_index AS album_track_index,
    source_tracks.title AS title,
    coalesce(track_albums.album_artwork_id, sources.artwork_id) AS artwork_id,
    source_tracks.end_offset_ms - source_tracks.start_offset_ms AS duration_ms,
    track_listens.last_listened_at AS played_at,
    artists.search_name AS search_artist,
    track_albums.album_search_name AS search_album,
    source_tracks.search_title AS search_title,
    users.id AS user_id
FROM source_tracks
JOIN users ON 1 = 1
LEFT JOIN artists ON source_tracks.artist_id = artists.id
LEFT JOIN track_albums ON source_tracks.id = track_albums.track_id
LEFT JOIN sources ON source_tracks.source_id = sources.id
LEFT JOIN track_listens ON source_tracks.id = track_listens.track_id AND users.id = track_listens.user_id

UNION ALL

SELECT
    remote_tracks.id AS id,
    NULL AS source_id,
    remote_tracks.remote_id AS remote_id,
    remote_tracks.track_id AS remote_track_id,
    artists.name AS artist,
    artists.id AS artist_id,
    track_albums.album_name AS album,
    track_albums.album_id AS album_id,
    track_albums.album_artist_id AS album_artist_id,
    track_albums.album_released_at AS album_released_at,
    track_albums.list_index AS album_track_index,
    remote_tracks.title AS title,
    remote_artworks.id AS artwork_id,
    remote_tracks.duration_ms AS duration_ms,
    track_listens.last_listened_at AS played_at,
    artists.search_name AS search_artist,
    track_albums.album_search_name AS search_album,
    remote_tracks.search_title AS search_title,
    remote_tracks_to_users.user_id AS user_id
FROM remote_tracks
JOIN remote_tracks_to_users ON remote_tracks.id = remote_tracks_to_users.remote_track_id
LEFT JOIN remote_artists ON remote_tracks.remote_id = remote_artists.remote_id AND remote_tracks.artist_id = remote_artists.artist_id
LEFT JOIN artists ON remote_artists.tapesonic_artist_id = artists.id
LEFT JOIN track_albums ON remote_tracks.id = track_albums.track_id
LEFT JOIN remote_artworks ON remote_tracks.remote_id = remote_artworks.remote_id AND remote_tracks.artwork_id = remote_artworks.artwork_id
LEFT JOIN track_listens ON remote_tracks.id = track_listens.track_id AND remote_tracks_to_users.user_id = track_listens.user_id;
