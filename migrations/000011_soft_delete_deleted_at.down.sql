ALTER TABLE users
    ADD COLUMN is_active BOOLEAN NOT NULL DEFAULT true;

UPDATE users SET is_active = false WHERE deleted_at IS NOT NULL;

ALTER TABLE users DROP COLUMN deleted_at;

DROP INDEX IF EXISTS idx_users_deleted_at;
CREATE INDEX idx_users_is_active ON users(is_active);


ALTER TABLE products
    ADD COLUMN is_active BOOLEAN NOT NULL DEFAULT true;

UPDATE products SET is_active = false WHERE deleted_at IS NOT NULL;

ALTER TABLE products DROP COLUMN deleted_at;

DROP INDEX IF EXISTS idx_products_deleted_at;
CREATE INDEX idx_products_is_active ON products(is_active);
