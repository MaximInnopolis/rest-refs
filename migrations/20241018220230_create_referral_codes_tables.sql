-- +goose Up
-- +goose StatementBegin
CREATE TABLE referral_codes (
                       id SERIAL PRIMARY KEY,
                       code VARCHAR(255) NOT NULL,
                       expires_at VARCHAR(255) NOT NULL,
                       referrer_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                       created_at TIMESTAMPTZ DEFAULT NOW(),
                       updated_at TIMESTAMPTZ DEFAULT NOW()

);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS referral_codes;
-- +goose StatementEnd