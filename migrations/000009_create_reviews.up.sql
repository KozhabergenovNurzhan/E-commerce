CREATE TABLE IF NOT EXISTS reviews (
    id         BIGSERIAL      PRIMARY KEY,
    product_id BIGINT         NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    user_id    BIGINT         NOT NULL REFERENCES users(id)    ON DELETE CASCADE,
    rating     SMALLINT       NOT NULL CHECK (rating BETWEEN 1 AND 5),
    comment    TEXT           NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    UNIQUE (product_id, user_id)
);

CREATE INDEX idx_reviews_product_id ON reviews(product_id);
CREATE INDEX idx_reviews_user_id    ON reviews(user_id);
