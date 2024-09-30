CREATE TABLE IF NOT EXISTS article_management.article_tags (
	article_id INTEGER REFERENCES article_management.articles (id) ON DELETE CASCADE,
	tag_name VARCHAR(50) REFERENCES article_management.tags (name) ON DELETE CASCADE
);
