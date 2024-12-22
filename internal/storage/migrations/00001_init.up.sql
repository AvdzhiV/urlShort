CREATE TABLE IF NOT EXISTS url_records (
    uuid UUID NOT NULL,
    short_url VARCHAR(255) UNIQUE NOT NULL,
    original_url TEXT UNIQUE NOT NULL
);
