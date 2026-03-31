ALTER TABLE users
    ADD COLUMN deleted_at TIMESTAMPTZ;

UPDATE users SET deleted_at = updated_at WHERE is_active = false;

ALTER TABLE users DROP COLUMN is_active;

DROP INDEX IF EXISTS idx_users_is_active;
CREATE INDEX idx_users_deleted_at ON users(deleted_at) WHERE deleted_at IS NOT NULL;


ALTER TABLE products
    ADD COLUMN deleted_at TIMESTAMPTZ;

UPDATE products SET deleted_at = updated_at WHERE is_active = false;

ALTER TABLE products DROP COLUMN is_active;

DROP INDEX IF EXISTS idx_products_is_active;
CREATE INDEX idx_products_deleted_at ON products(deleted_at) WHERE deleted_at IS NOT NULL;
