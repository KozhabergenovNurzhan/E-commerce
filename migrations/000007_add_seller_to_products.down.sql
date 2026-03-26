DROP INDEX IF EXISTS idx_products_seller_id;

ALTER TABLE products DROP COLUMN IF EXISTS seller_id;
