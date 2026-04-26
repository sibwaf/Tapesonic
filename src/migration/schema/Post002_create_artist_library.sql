CREATE VIEW all_artists (
    id,
    name,
    artwork_id,
    album_count,
    search_name,
    user_id
) AS

WITH
    artist_raw_aggregate (user_id, artist_id, album_count) AS (
        SELECT
            users.id AS user_id,
            tapes.artist_id AS artist_id,
            count(*) AS album_count
        FROM tapes
        JOIN users ON 1 = 1
        WHERE tapes."type" = 'album'
        GROUP BY users.id, tapes.artist_id

        UNION ALL

        SELECT
            remote_albums_to_users.user_id AS user_id,
            remote_artists.tapesonic_artist_id AS artist_id,
            count(*) AS album_count
        FROM remote_albums
        JOIN remote_albums_to_users ON remote_albums.id = remote_albums_to_users.remote_album_id
        JOIN remote_artists ON remote_albums.remote_id = remote_artists.remote_id AND remote_albums.artist_id = remote_artists.artist_id
        GROUP BY remote_albums_to_users.user_id, remote_artists.tapesonic_artist_id
    ),
    artist_aggregate (user_id, artist_id, album_count) AS (
        SELECT
            user_id AS user_id,
            artist_id AS artist_id,
            sum(album_count) AS album_count
        FROM artist_raw_aggregate
        GROUP BY user_id, artist_id
    )

SELECT
    artists.id AS id,
    artists.name AS name,
    NULL AS artwork_id,
    artist_aggregate.album_count AS album_count,
    artists.search_name AS search_name,
    users.id AS user_id
FROM artists
JOIN users ON 1 = 1
LEFT JOIN artist_aggregate ON users.id = artist_aggregate.user_id AND artists.id = artist_aggregate.artist_id;
