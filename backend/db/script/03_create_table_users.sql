CREATE TABLE IF NOT EXISTS article_management.users(
    id serial PRIMARY KEY,
    username varchar(100) UNIQUE NOT NULL,
    email varchat(100) UNIQUE NOT NULL,
    password VARCHAR(100) NOT NULL,
    name varchar(100) NOT NULL,
    bio varchar(255) NOT NULL,
    image varchar(255) NOT NULL,
    created_at timestamptz DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz
);

DO $$
BEGIN
    CREATE TRIGGER update_users_trigger
        BEFORE UPDATE ON article_management.users
        FOR EACH ROW
        EXECUTE PROCEDURE article_management.update_updated_at_column( );
EXCEPTION
    WHEN duplicate_object THEN
        NULL;
END;

$$;

