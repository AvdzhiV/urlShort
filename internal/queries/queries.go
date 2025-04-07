package queries

const SelectOriginalURL = `
SELECT original_url
FROM url_records
WHERE short_url = $1
`
const SelectShortlURL = `
SELECT short_url
FROM url_records
WHERE original_url = $1
`

const PutShortURL = `
INSERT INTO url_records (uuid, short_url, original_url)
VALUES (gen_random_uuid(), $1, $2)
ON CONFLICT (original_url)
DO NOTHING RETURNING short_url
`
const InsertBatch = `
INSERT INTO url_records (uuid, short_url, original_url)
VALUES (gen_random_uuid(), $1, $2)
`