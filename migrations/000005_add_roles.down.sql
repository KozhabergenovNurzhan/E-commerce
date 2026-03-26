-- PostgreSQL does not support DROP VALUE on an enum.
-- To rollback: recreate the type and migrate existing data back to 'customer'.
ALTER TABLE users
    ALTER COLUMN role TYPE VARCHAR(50);

DROP TYPE user_role;

CREATE TYPE user_role AS ENUM ('customer', 'admin');

ALTER TABLE users
    ALTER COLUMN role TYPE user_role
    USING (
        CASE role
            WHEN 'manager' THEN 'customer'
            WHEN 'seller'  THEN 'customer'
            ELSE role
        END
    )::user_role;
