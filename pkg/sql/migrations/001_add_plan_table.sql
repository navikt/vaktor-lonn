-- +goose Up
CREATE TABLE beredskapsvakt (
                      id uuid NOT NULL,
                      ident text NOT NULL,
                      plan json NOT NULL,
                      PRIMARY KEY(id)
);

comment on column beredskapsvakt.id is 'Created by Vaktor Plan';

-- +goose Down
DROP TABLE beredskapsvakt;
