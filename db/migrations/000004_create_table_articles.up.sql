CREATE TABLE IF NOT EXISTS article_management.articles (
	id SERIAL PRIMARY KEY,
	title VARCHAR(100) NOT NULL,
	description VARCHAR(255) NOT NULL,
	body TEXT NOT NULL,
	user_id INTEGER REFERENCES article_management.users (id) ON DELETE cascade,
	favorites_count INTEGER NOT NULL DEFAULT 0,
	created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMPTZ
);

DO $$
BEGIN
    CREATE OR REPLACE TRIGGER update_articles_trigger BEFORE UPDATE ON article_management.articles
    FOR EACH ROW EXECUTE PROCEDURE article_management.update_updated_at_column();

    EXCEPTION
        WHEN duplicate_object THEN
        NULL;
END; $$;
