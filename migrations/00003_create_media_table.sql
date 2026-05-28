-- +goose Up
CREATE SCHEMA media;

CREATE TABLE media.originals
(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    film_id UUID NOT NULL,
    key TEXT UNIQUE NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('uploading', 'uploaded', 'transcoding', 'ready', 'failed')),
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ
);

-- +goose StatementBegin
CREATE FUNCTION media.set_updated_at()
    RETURNS TRIGGER AS
$$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER set_updated_at
    BEFORE UPDATE ON media.originals
    FOR EACH ROW
EXECUTE FUNCTION media.set_updated_at();

-- +goose Down
DROP SCHEMA media CASCADE;
