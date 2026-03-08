-- +goose Up
CREATE SCHEMA showcase;

CREATE TABLE showcase.genres
(
    id TEXT PRIMARY KEY
);

CREATE TABLE showcase.films
(
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL,
    description TEXT,
    genre_id    TEXT REFERENCES showcase.genres (id),
    rating      NUMERIC(3, 1) CHECK ( rating >= 0 AND rating <= 10 ),
    poster_url  TEXT
);

CREATE TABLE showcase.film_images
(
    id SERIAL PRIMARY KEY,
    film_id UUID REFERENCES showcase.films(id) ON DELETE CASCADE,
    url TEXT NOT NULL
);

-- +goose Down
DROP SCHEMA showcase CASCADE;


