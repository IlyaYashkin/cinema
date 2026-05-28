-- +goose Up
CREATE SCHEMA showcase;

CREATE TABLE showcase.films
(
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL,
    description TEXT,
    poster_url  TEXT
);

CREATE TABLE showcase.film_images
(
    id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    film_id UUID REFERENCES showcase.films(id) ON DELETE CASCADE,
    url TEXT NOT NULL
);

-- +goose Down
DROP SCHEMA showcase CASCADE;
