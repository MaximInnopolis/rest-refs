-- +goose Up
-- +goose StatementBegin
CREATE TABLE referrals (
                                id SERIAL PRIMARY KEY,
                                email VARCHAR(255) NOT NULL,
                                referral_code_id INT NOT NULL REFERENCES referral_codes(id) ON DELETE CASCADE,
                                referrer_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                                created_at TIMESTAMPTZ DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS referrals;
-- +goose StatementEnd