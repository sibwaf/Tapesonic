CREATE VIEW all_artworks (
    id,
    remote_id,
    remote_artwork_id
) AS

SELECT
    artworks.id AS id,
    NULL AS remote_id,
    NULL AS remote_artwork_id
FROM artworks

UNION ALL

SELECT
    remote_artworks.id AS id,
    remote_artworks.remote_id AS remote_id,
    remote_artworks.artwork_id AS remote_artwork_id
FROM remote_artworks;
