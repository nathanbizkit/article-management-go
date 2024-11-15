CREATE TABLE IF NOT EXISTS article_management.follows (
	from_user_id INTEGER REFERENCES article_management.users (id) ON DELETE CASCADE,
	to_user_id INTEGER REFERENCES article_management.users (id) ON DELETE CASCADE
);
