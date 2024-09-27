CREATE TABLE IF NOT EXISTS article_management.users (
	id SERIAL PRIMARY KEY,
	username VARCHAR(100) UNIQUE NOT NULL,
	email VARCHAR(100) UNIQUE NOT NULL,
	password VARCHAR(100) NOT NULL,
	name VARCHAR(100) NOT NULL,
	bio VARCHAR(255) NOT NULL,
	image VARCHAR(255) NOT NULL,
	created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMPTZ
);

DO $$
BEGIN
    CREATE TRIGGER update_users_trigger BEFORE UPDATE ON article_management.users
    FOR EACH ROW EXECUTE PROCEDURE article_management.update_updated_at_column();

    EXCEPTION
        WHEN duplicate_object THEN
        NULL;
END; $$;