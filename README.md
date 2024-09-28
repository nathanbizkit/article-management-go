# Article Management System

## TODOs

- [ ] Users and Authentication
  - [ ] `POST /signin`: Existing user login
  - [ ] `POST /signup`: Register a new user
  - [ ] `GET /me`: Get current user
  - [ ] `PUT /me`: Update current user
- [ ] Profiles
  - [ ] `GET /profiles/{username}`: Get a profile
  - [ ] `POST /profiles/{username}/follow`: Follow a user
  - [ ] `POST /profiles/{username}/unfollow`: Unfollow a user
- [ ] Articles
  - [ ] `GET /articles/feed`: Get recent articles from users you follow
  - [ ] `GET /articles`: Get recent articles globally
  - [ ] `POST /articles`: Create an article
  - [ ] `GET /articles/{slug}`: Get an article
  - [ ] `PUT /articles/{slug}`: Update an article
  - [ ] `DELETE /articles/{slug}`: DELETE an article
- [ ] Comments
  - [ ] `GET /articles/{slug}/comments`: Get comments for an article
  - [ ] `POST /articles/{slug}/commends`: Create a comment for an article
  - [ ] `DELETE /articles/{slug}/comments/{id}`: Delete a comment for an article
- [ ] Favorites
  - [ ] `POST /articles/{slug}/favorite`: Favorite an article
  - [ ] `POST /articles/{slug}/unfavorite`: Unfavorite an article
- [ ] Default
  - [ ] `GET /tags`: Get tages
