CREATE TABLE IF NOT EXISTS article_management.articles(
    id serial PRIMARY KEY,
    title varchar(100) NOT NULL,
    description varchar(255) NOT NULL,
    body text NOT NULL,
    user_id integer REFERENCES article_management.users(id) ON DELETE CASCADE,
    favorites_count integer NOT NULL DEFAULT 0,
    created_at timestamptz DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz
);

DO $$
BEGIN
    CREATE TRIGGER update_articles_trigger
        BEFORE UPDATE ON article_management.articles
        FOR EACH ROW
        EXECUTE PROCEDURE article_management.update_updated_at_column( );
EXCEPTION
    WHEN duplicate_object THEN
        NULL;
END;

$$
