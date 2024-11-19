# Article Management System

## Getting Started

API endpoints:

- HTTP: [http://localhost:8000/api](http://localhost:8000/api)
- HTTPS: [https://localhost:8443/api](https://localhost:8443/api)

### Running

1. Copy `template.env` and edit it accordingly
2. In `docker-compose.yml`, set env file location for `app` service accordingly
3. TLS is enabled when `TLS_CERT_FILE` and `TLS_KEY_FILE` is set in env file
4. If TLS is enabled, set cert directory location for `app` service in `docker-compose.yml`
5. Start by running `docker-compose up -d`
6. Stop by running `docker-compose down`

### Testing

```bash
# unit
make unittest
make unitcoverage

# integration (docker is required)
make integrationtest
make integrationcoverage

# overall coverage
make coverage

# e2e
# Set --ssl-client-cert in e2e/run-api-tests.sh first
docker-compose up -d
make e2etest
```

## TODOs

- [x] Users and Authentication
  - [x] `POST /login`: Existing user login
  - [x] `POST /register`: Register a new user
  - [x] `POST /refresh_token`: Refresh user token with refresh token
  - [x] `GET /me`: Get current user
  - [x] `PUT /me`: Update current user
- [x] Profiles
  - [x] `GET /profiles/{username}`: Get a profile
  - [x] `POST /profiles/{username}/follow`: Follow a user
  - [x] `DELETE /profiles/{username}/follow`: Unfollow a user
- [x] Articles
  - [x] `GET /articles/feed`: Get recent articles from users you follow
  - [x] `GET /articles`: Get recent articles globally
  - [x] `POST /articles`: Create an article
  - [x] `GET /articles/{slug}`: Get an article
  - [x] `PUT /articles/{slug}`: Update an article
  - [x] `DELETE /articles/{slug}`: DELETE an article
- [x] Comments
  - [x] `GET /articles/{slug}/comments`: Get comments for an article
  - [x] `POST /articles/{slug}/commends`: Create a comment for an article
  - [x] `DELETE /articles/{slug}/comments/{id}`: Delete a comment for an article
- [x] Favorites
  - [x] `POST /articles/{slug}/favorite`: Favorite an article
  - [x] `DELETE /articles/{slug}/favorite`: Unfavorite an article
- [x] Default
  - [x] `GET /tags`: Get tages
