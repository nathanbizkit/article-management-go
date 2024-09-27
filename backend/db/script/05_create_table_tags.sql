CREATE TABLE IF NOT EXISTS article_management.tags(
    id serial PRIMARY KEY,
    name varchar(100) NOT NULL,
    created_at timestamptz DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz
);

DO $$
BEGIN
    CREATE TRIGGER update_tags_trigger
        BEFORE UPDATE ON article_management.tags
        FOR EACH ROW
        EXECUTE PROCEDURE article_management.update_updated_at_column( );
EXCEPTION
    WHEN duplicate_object THEN
        NULL;
END;

$$
