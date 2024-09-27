CREATE TABLE IF NOT EXISTS article_management.favorite_articles(
	article_id INTEGER REFERENCES article_management.articles(id) ON DELETE CASCADE,
	user_id INTEGER REFERENCES article_management.users(id) ON DELETE CASCADE,
);
