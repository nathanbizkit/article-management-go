CREATE TABLE IF NOT EXISTS article_management.favorite_articles(
    article_id integer REFERENCES article_management.articles(id) ON DELETE CASCADE,
    user_id integer REFERENCES article_management.users(id) ON DELETE CASCADE,
);

