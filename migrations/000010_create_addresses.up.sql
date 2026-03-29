CREATE TABLE IF NOT EXISTS addresses (
    id          BIGSERIAL     PRIMARY KEY,
    user_id     BIGINT        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    full_name   VARCHAR(200)  NOT NULL,
    phone       VARCHAR(20)   NOT NULL,
    country     VARCHAR(100)  NOT NULL,
    city        VARCHAR(100)  NOT NULL,
    street      VARCHAR(255)  NOT NULL,
    postal_code VARCHAR(20)   NOT NULL,
    is_default  BOOLEAN       NOT NULL DEFAULT false,
    created_at  TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_addresses_user_id ON addresses(user_id);

ALTER TABLE orders ADD COLUMN IF NOT EXISTS address_id BIGINT REFERENCES addresses(id) ON DELETE SET NULL;
