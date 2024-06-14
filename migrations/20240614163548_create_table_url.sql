-- +goose Up
-- +goose StatementBegin
CREATE TABLE url (
    id serial PRIMARY KEY,
    url text UNIQUE NOT NULL CHECK (url <> ''),
    url_id text UNIQUE NOT NULL CHECK (url_id <> '')
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE url;
-- +goose StatementEnd
