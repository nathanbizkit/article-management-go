# Article Management System

## Getting Started

API endpoints: checking out `swagger.json` file in `doc/swagger/v1` with [swagger-editor](https://editor-next.swagger.io/) to explore more about api routes.

- HTTP: [http://localhost:8000/api/v1](http://localhost:8000/api/v1)
- HTTPS: [https://localhost:8443/api/v1](https://localhost:8443/api/v1)

### Running

1. Create env file `env/local.env` using `env/template.env` as a template.
2. Set env values accordingly:
   1. TLS is enabled when `TLS_CERT_FILE` and `TLS_KEY_FILE` is set.
   2. For database, you can use the settings from `docker-compose.yml` if you want to use `db` service.
3. Set `docker-compose.yml`:
   1. Update `env_file` in `app` service to point to the env file you just created. (`env/local.env`)
   2. Update `ports` in `app` service to reflect the ports in env file.
   3. If TLS is enabled, update `app` volumes to bind to your certificate folder location.
4. If you change app ports from anything other than `8000` and `8443`, you need to update `Dockerfile` to expose those ports correctly.
5. Start by running `docker-compose up -d` and stop by running `docker-compose down`.

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
  - [x] `DELETE /articles/{slug}`: Delete an article
- [x] Comments
  - [x] `GET /articles/{slug}/comments`: Get comments for an article
  - [x] `POST /articles/{slug}/commends`: Create a comment for an article
  - [x] `DELETE /articles/{slug}/comments/{id}`: Delete a comment for an article
- [x] Favorites
  - [x] `POST /articles/{slug}/favorite`: Favorite an article
  - [x] `DELETE /articles/{slug}/favorite`: Unfavorite an article
- [x] Default
  - [x] `GET /tags`: Get tages
