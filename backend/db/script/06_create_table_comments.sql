CREATE TABLE IF NOT EXISTS article_management.comments(
	id SERIAL PRIMARY KEY,
	body TEXT NOT NULL,
	user_id INTEGER REFERENCES article_management.users(id) ON DELETE CASCADE,
	article_id INTEGER REFERENCES article_management.articles(id) ON DELETE CASCADE,
	created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMPTZ
);

DO $$
BEGIN
	CREATE TRIGGER update_comments_trigger
		BEFORE UPDATE ON article_management.comments
		FOR EACH ROW
		EXECUTE PROCEDURE article_management.update_updated_at_column( );
EXCEPTION
	WHEN duplicate_object THEN
		NULL;
END;

$$
