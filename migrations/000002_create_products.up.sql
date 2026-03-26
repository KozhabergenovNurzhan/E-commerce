CREATE TABLE IF NOT EXISTS categories (
    id         UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    name       VARCHAR(100) NOT NULL,
    slug       VARCHAR(100) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS products (
    id          UUID           PRIMARY KEY DEFAULT uuid_generate_v4(),
    category_id UUID           NOT NULL REFERENCES categories(id),
    name        VARCHAR(255)   NOT NULL,
    description TEXT           NOT NULL DEFAULT '',
    price       NUMERIC(10, 2) NOT NULL CHECK (price > 0),
    stock       INT            NOT NULL DEFAULT 0 CHECK (stock >= 0),
    image_url   TEXT           NOT NULL DEFAULT '',
    is_active   BOOLEAN        NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_products_category_id ON products(category_id);
CREATE INDEX idx_products_is_active   ON products(is_active);
CREATE INDEX idx_products_price       ON products(price);

-- Seed categories
INSERT INTO categories (name, slug) VALUES
    ('Electronics',   'electronics'),
    ('Clothing',      'clothing'),
    ('Books',         'books'),
    ('Home & Garden', 'home-garden'),
    ('Sports',        'sports');
