CREATE TABLE IF NOT EXISTS article_management.tags (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ
);

DO $$
BEGIN
    CREATE TRIGGER update_tags_trigger BEFORE UPDATE ON article_management.tags
    FOR EACH ROW EXECUTE PROCEDURE article_management.update_updated_at_column();

    EXCEPTION
        WHEN duplicate_object THEN
        NULL;
END; $$;
