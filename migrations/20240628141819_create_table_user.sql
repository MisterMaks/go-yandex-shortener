-- +goose Up
-- +goose StatementBegin
CREATE TABLE "user" (
    id serial PRIMARY KEY
);

ALTER TABLE url ADD COLUMN user_id integer DEFAULT NULL REFERENCES "user"(id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE url DROP COLUMN user_id;

DROP TABLE "user";
-- +goose StatementEnd
