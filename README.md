# Article Management System

## TODOs

- [ ] Users and Authentication
  - [ ] `POST /login`: Existing user login
  - [ ] `POST /register`: Register a new user
  - [ ] `POST /refresh_token`: Refresh user token with refresh token
  - [ ] `GET /me`: Get current user
  - [ ] `PUT /me`: Update current user
- [ ] Profiles
  - [ ] `GET /profiles/{username}`: Get a profile
  - [ ] `POST /profiles/{username}/follow`: Follow a user
  - [ ] `DELETE /profiles/{username}/follow`: Unfollow a user
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
  - [ ] `DELETE /articles/{slug}/favorite`: Unfavorite an article
- [ ] Default
  - [ ] `GET /tags`: Get tages
