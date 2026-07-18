-- +goose Up
-- +goose StatementBegin

ALTER TABLE users_settings
    ALTER COLUMN is_allow_external_registrations SET DEFAULT FALSE,
    ALTER COLUMN is_allow_member_invitations SET DEFAULT FALSE;

UPDATE users_settings
SET
    is_allow_external_registrations = FALSE,
    is_allow_member_invitations = FALSE;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE users_settings
    ALTER COLUMN is_allow_external_registrations SET DEFAULT TRUE,
    ALTER COLUMN is_allow_member_invitations SET DEFAULT TRUE;

-- +goose StatementEnd
