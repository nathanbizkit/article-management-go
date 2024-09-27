CREATE TABLE IF NOT EXISTS article_management.article_tags(
	article_id INTEGER REFERENCES article_management.articles(id) ON DELETE CASCADE,
	tag_id INTEGER REFERENCES article_management.tags(id) ON DELETE CASCADE
);
