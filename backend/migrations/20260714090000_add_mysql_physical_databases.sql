-- +goose Up
-- +goose StatementBegin
CREATE TABLE mysql_physical_databases (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    database_id        UUID,
    version            TEXT NOT NULL,
    host               TEXT NOT NULL,
    port               INT NOT NULL,
    username           TEXT NOT NULL,
    password           TEXT NOT NULL,
    is_https           BOOLEAN NOT NULL DEFAULT FALSE,
    data_dir           TEXT NOT NULL,
    server_uuid        TEXT,
    is_binlog_enabled  BOOLEAN NOT NULL DEFAULT FALSE
);

ALTER TABLE mysql_physical_databases
    ADD CONSTRAINT uk_mysql_physical_databases_database_id
    UNIQUE (database_id);

ALTER TABLE mysql_physical_databases
    ADD CONSTRAINT fk_mysql_physical_databases_database_id
    FOREIGN KEY (database_id)
    REFERENCES databases (id)
    ON DELETE CASCADE;

CREATE INDEX idx_mysql_physical_databases_database_id
    ON mysql_physical_databases (database_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_mysql_physical_databases_database_id;
DROP TABLE IF EXISTS mysql_physical_databases;
-- +goose StatementEnd
