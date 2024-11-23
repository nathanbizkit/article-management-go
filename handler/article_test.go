package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nathanbizkit/article-management/message"
	"github.com/nathanbizkit/article-management/model"
	"github.com/nathanbizkit/article-management/test"
	"github.com/stretchr/testify/assert"
)

func TestIntegration_ArticleHandler(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration tests.")
	}

	gin.SetMode("test")
	h, lct := setup(t)

	t.Run("CreateArticle", func(t *testing.T) {
		shortMaxLenString := strings.Repeat("a", 101)
		tagMaxLenString := strings.Repeat("a", 51)

		fooUser := createRandomUser(t, lct.DB())

		randStr := test.RandomString(t, 20)
		article := model.Article{
			Title:       randStr,
			Description: randStr,
			Body:        randStr,
			Author:      *fooUser,
		}

		tests := []struct {
			title              string
			reqUser            *model.User
			reqBody            *message.CreateArticleRequest
			expectedStatusCode int
			expectedBody       message.ArticleResponse
			expectedError      map[string]interface{}
			hasError           bool
		}{
			{
				"create article: success",
				fooUser,
				&message.CreateArticleRequest{
					Title:       article.Title,
					Description: article.Description,
					Body:        article.Body,
					Tags: []string{
						test.RandomString(t, 10),
						test.RandomString(t, 10),
					},
				},
				http.StatusOK,
				article.ResponseArticle(false, false),
				nil,
				false,
			},
			{
				"create article: wrong current user id",
				&model.User{ID: 0},
				&message.CreateArticleRequest{
					Title:       article.Title,
					Description: article.Description,
					Body:        article.Body,
					Tags: []string{
						test.RandomString(t, 10),
						test.RandomString(t, 10),
					},
				},
				http.StatusNotFound,
				message.ArticleResponse{},
				map[string]interface{}{"error": "current user not found"},
				true,
			},
			{
				"create article: no title",
				fooUser,
				&message.CreateArticleRequest{
					Title:       "",
					Description: article.Description,
					Body:        article.Body,
					Tags: []string{
						test.RandomString(t, 10),
						test.RandomString(t, 10),
					},
				},
				http.StatusBadRequest,
				message.ArticleResponse{},
				map[string]interface{}{"error": "validation error: Title: cannot be blank."},
				true,
			},
			{
				"create article: title is too short",
				fooUser,
				&message.CreateArticleRequest{
					Title:       "abc",
					Description: article.Description,
					Body:        article.Body,
					Tags: []string{
						test.RandomString(t, 10),
						test.RandomString(t, 10),
					},
				},
				http.StatusBadRequest,
				message.ArticleResponse{},
				map[string]interface{}{"error": "validation error: Title: the length must be between 5 and 100."},
				true,
			},
			{
				"create article: title is too long",
				fooUser,
				&message.CreateArticleRequest{
					Title:       shortMaxLenString,
					Description: article.Description,
					Body:        article.Body,
					Tags: []string{
						test.RandomString(t, 10),
						test.RandomString(t, 10),
					},
				},
				http.StatusBadRequest,
				message.ArticleResponse{},
				map[string]interface{}{"error": "validation error: Title: the length must be between 5 and 100."},
				true,
			},
			{
				"create article: description is too short",
				fooUser,
				&message.CreateArticleRequest{
					Title:       article.Title,
					Description: "abc",
					Body:        article.Body,
					Tags: []string{
						test.RandomString(t, 10),
						test.RandomString(t, 10),
					},
				},
				http.StatusBadRequest,
				message.ArticleResponse{},
				map[string]interface{}{"error": "validation error: Description: the length must be between 5 and 100."},
				true,
			},
			{
				"create article: description is too long",
				fooUser,
				&message.CreateArticleRequest{
					Title:       article.Title,
					Description: shortMaxLenString,
					Body:        article.Body,
					Tags: []string{
						test.RandomString(t, 10),
						test.RandomString(t, 10),
					},
				},
				http.StatusBadRequest,
				message.ArticleResponse{},
				map[string]interface{}{"error": "validation error: Description: the length must be between 5 and 100."},
				true,
			},
			{
				"create article: no body",
				fooUser,
				&message.CreateArticleRequest{
					Title:       article.Title,
					Description: article.Description,
					Body:        "",
					Tags: []string{
						test.RandomString(t, 10),
						test.RandomString(t, 10),
					},
				},
				http.StatusBadRequest,
				message.ArticleResponse{},
				map[string]interface{}{"error": "validation error: Body: cannot be blank."},
				true,
			},
			{
				"create article: no tags",
				fooUser,
				&message.CreateArticleRequest{
					Title:       article.Title,
					Description: article.Description,
					Body:        article.Body,
					Tags:        []string{},
				},
				http.StatusBadRequest,
				message.ArticleResponse{},
				map[string]interface{}{"error": "validation error: Tags: cannot be blank."},
				true,
			},
			{
				"create article: empty tag name",
				fooUser,
				&message.CreateArticleRequest{
					Title:       article.Title,
					Description: article.Description,
					Body:        article.Body,
					Tags:        []string{""},
				},
				http.StatusBadRequest,
				message.ArticleResponse{},
				map[string]interface{}{"error": "validation error: Tags: (0: tag name cannot be blank.)."},
				true,
			},
			{
				"create article: tag name is too short",
				fooUser,
				&message.CreateArticleRequest{
					Title:       article.Title,
					Description: article.Description,
					Body:        article.Body,
					Tags:        []string{"a"},
				},
				http.StatusBadRequest,
				message.ArticleResponse{},
				map[string]interface{}{"error": "validation error: Tags: (0: tag name length must be between 3 and 50.)."},
				true,
			},
			{
				"create article: tag name is too long",
				fooUser,
				&message.CreateArticleRequest{
					Title:       article.Title,
					Description: article.Description,
					Body:        article.Body,
					Tags:        []string{tagMaxLenString},
				},
				http.StatusBadRequest,
				message.ArticleResponse{},
				map[string]interface{}{"error": "validation error: Tags: (0: tag name length must be between 3 and 50.)."},
				true,
			},
		}

		for _, tt := range tests {
			body, err := json.Marshal(tt.reqBody)
			if err != nil {
				t.Fatal(err)
			}

			req := httptest.NewRequest(http.MethodPut, "/api/v1/articles", bytes.NewReader(body))
			w := httptest.NewRecorder()
			ctx, _ := ctxWithToken(t, lct.Environ(), w, req, tt.reqUser.ID, time.Now())

			h.CreateArticle(ctx)

			assert.Equal(t, tt.expectedStatusCode, w.Result().StatusCode, tt.title)

			if tt.hasError {
				actualBody := test.GetResponseBody[map[string]interface{}](t, w.Result())
				assert.Equal(t, tt.expectedError, actualBody, tt.title)
			} else {
				actualBody := test.GetResponseBody[message.ArticleResponse](t, w.Result())
				assert.NotEmpty(t, actualBody.ID, tt.title)
				assert.Equal(t, tt.expectedBody.Title, actualBody.Title, tt.title)
				assert.Equal(t, tt.expectedBody.Description, actualBody.Description, tt.title)
				assert.Equal(t, tt.expectedBody.Body, actualBody.Body, tt.title)
				assert.Equal(t, tt.expectedBody.Favorited, actualBody.Favorited, tt.title)
				assert.Equal(t, tt.expectedBody.FavoritesCount, actualBody.FavoritesCount, tt.title)
				assert.Equal(t, tt.expectedBody.Author, actualBody.Author, tt.title)
				assert.NotEmpty(t, actualBody.CreatedAt, tt.title)
				assert.NotEmpty(t, actualBody.UpdatedAt, tt.title)
			}
		}
	})

	t.Run("GetArticle", func(t *testing.T) {
		fooUser := createRandomUser(t, lct.DB())
		fooArticle := createRandomArticle(t, lct.DB(), fooUser.ID)

		tests := []struct {
			title              string
			reqUser            *model.User
			reqSlug            string
			expectedStatusCode int
			expectedBody       message.ArticleResponse
			expectedError      map[string]interface{}
			hasError           bool
		}{
			{
				"get article: success",
				fooUser,
				strconv.Itoa(int(fooArticle.ID)),
				http.StatusOK,
				fooArticle.ResponseArticle(false, false),
				nil,
				false,
			},
			{
				"get article: allow public access",
				&model.User{ID: 0},
				strconv.Itoa(int(fooArticle.ID)),
				http.StatusOK,
				fooArticle.ResponseArticle(false, false),
				nil,
				false,
			},
			{
				"get article: wrong current user id",
				&model.User{ID: 99},
				strconv.Itoa(int(fooArticle.ID)),
				http.StatusNotFound,
				message.ArticleResponse{},
				map[string]interface{}{"error": "current user not found"},
				true,
			},
			{
				"get article: invalid slug",
				fooUser,
				"invalid_slug",
				http.StatusBadRequest,
				message.ArticleResponse{},
				map[string]interface{}{"error": "invalid slug"},
				true,
			},
			{
				"get article: wrong slug",
				fooUser,
				"0",
				http.StatusNotFound,
				message.ArticleResponse{},
				map[string]interface{}{"error": "article not found"},
				true,
			},
		}

		for _, tt := range tests {
			apiUrl := fmt.Sprintf("/api/v1/articles/%v", tt.reqSlug)
			req := httptest.NewRequest(http.MethodGet, apiUrl, nil)

			w := httptest.NewRecorder()
			ctx, _ := ctxWithToken(t, lct.Environ(), w, req, tt.reqUser.ID, time.Now())
			ctx.AddParam("slug", tt.reqSlug)

			h.GetArticle(ctx)

			assert.Equal(t, tt.expectedStatusCode, w.Result().StatusCode, tt.title)

			if tt.hasError {
				actualBody := test.GetResponseBody[map[string]interface{}](t, w.Result())
				assert.Equal(t, tt.expectedError, actualBody, tt.title)
			} else {
				actualBody := test.GetResponseBody[message.ArticleResponse](t, w.Result())
				assert.Equal(t, tt.expectedBody, actualBody, tt.title)
			}
		}
	})

	t.Run("GetArticles", func(t *testing.T) {
		fooUser := createRandomUser(t, lct.DB())
		barUser := createRandomUser(t, lct.DB())
		reqUser := createRandomUser(t, lct.DB())

		tag := model.Tag{Name: test.RandomString(t, 10)}

		articles := make([]*model.Article, 0, 10)
		for i := 0; i < 10; i++ {
			randStr := test.RandomString(t, 10)
			article := model.Article{
				Title:       randStr,
				Description: randStr,
				Body:        randStr,
			}

			if i < 5 {
				article.UserID = fooUser.ID
				article.Author = *fooUser
				article.Tags = []model.Tag{tag}
			} else {
				article.UserID = barUser.ID
				article.Author = *barUser
			}

			articles = append(articles, &article)
		}

		for i, article := range articles {
			createdArticle, err := h.as.Create(context.Background(), article)
			if err != nil {
				t.Fatal(err)
			}

			article.ID = createdArticle.ID
			article.CreatedAt = createdArticle.CreatedAt
			article.UpdatedAt = createdArticle.UpdatedAt

			if i < 5 {
				err := h.as.AddFavorite(context.Background(), article, fooUser,
					func(favoritesCount int64, updatedAt time.Time) {
						article.FavoritesCount = favoritesCount
						article.UpdatedAt = updatedAt
					})
				if err != nil {
					t.Fatal(err)
				}
			}

			// delay creating articles, so we can sort them by date
			time.Sleep(1 * time.Second)
		}

		sort.SliceStable(articles, func(i, j int) bool {
			return articles[i].CreatedAt.After(articles[j].CreatedAt)
		})

		tests := []struct {
			title   string
			reqUser *model.User
			query   struct {
				tag       string
				author    string
				favorited string
				limit     string
				offset    string
			}
			expectedStatusCode int
			expectedArticles   []*model.Article
		}{
			{
				"get articles: with default queries",
				reqUser,
				struct {
					tag       string
					author    string
					favorited string
					limit     string
					offset    string
				}{
					tag:       "",
					author:    "",
					favorited: "",
					limit:     "0",
					offset:    "0",
				},
				http.StatusOK,
				articles,
			},
			{
				"get articles: allow public access with default queries",
				&model.User{ID: 0},
				struct {
					tag       string
					author    string
					favorited string
					limit     string
					offset    string
				}{
					tag:       "",
					author:    "",
					favorited: "",
					limit:     "0",
					offset:    "0",
				},
				http.StatusOK,
				articles,
			},
			{
				"get articles: with limit and offset",
				reqUser,
				struct {
					tag       string
					author    string
					favorited string
					limit     string
					offset    string
				}{
					tag:       "",
					author:    "",
					favorited: "",
					limit:     "5",
					offset:    "5",
				},
				http.StatusOK,
				articles[5:10],
			},
			{
				"get articles: allow public access with limit and offset",
				&model.User{ID: 0},
				struct {
					tag       string
					author    string
					favorited string
					limit     string
					offset    string
				}{
					tag:       "",
					author:    "",
					favorited: "",
					limit:     "5",
					offset:    "5",
				},
				http.StatusOK,
				articles[5:10],
			},
			{
				"get articles: with tag",
				reqUser,
				struct {
					tag       string
					author    string
					favorited string
					limit     string
					offset    string
				}{
					tag:       tag.Name,
					author:    "",
					favorited: "",
					limit:     "0",
					offset:    "0",
				},
				http.StatusOK,
				articles[5:10],
			},
			{
				"get articles: allow public access with tag",
				&model.User{ID: 0},
				struct {
					tag       string
					author    string
					favorited string
					limit     string
					offset    string
				}{
					tag:       tag.Name,
					author:    "",
					favorited: "",
					limit:     "0",
					offset:    "0",
				},
				http.StatusOK,
				articles[5:10],
			},
			{
				"get articles: with author",
				reqUser,
				struct {
					tag       string
					author    string
					favorited string
					limit     string
					offset    string
				}{
					tag:       "",
					author:    barUser.Username,
					favorited: "",
					limit:     "0",
					offset:    "0",
				},
				http.StatusOK,
				articles[0:5],
			},
			{
				"get articles: allow public access with author",
				&model.User{ID: 0},
				struct {
					tag       string
					author    string
					favorited string
					limit     string
					offset    string
				}{
					tag:       "",
					author:    barUser.Username,
					favorited: "",
					limit:     "0",
					offset:    "0",
				},
				http.StatusOK,
				articles[0:5],
			},
			{
				"get articles: with various queries",
				reqUser,
				struct {
					tag       string
					author    string
					favorited string
					limit     string
					offset    string
				}{
					tag:       tag.Name,
					author:    fooUser.Username,
					favorited: "",
					limit:     "2",
					offset:    "1",
				},
				http.StatusOK,
				articles[6:8],
			},
			{
				"get articles: allow public access with various queries",
				&model.User{ID: 0},
				struct {
					tag       string
					author    string
					favorited string
					limit     string
					offset    string
				}{
					tag:       tag.Name,
					author:    fooUser.Username,
					favorited: "",
					limit:     "2",
					offset:    "1",
				},
				http.StatusOK,
				articles[6:8],
			},
			{
				"get articles: with favorited queries",
				reqUser,
				struct {
					tag       string
					author    string
					favorited string
					limit     string
					offset    string
				}{
					tag:       "",
					author:    "",
					favorited: fooUser.Username,
					limit:     "0",
					offset:    "0",
				},
				http.StatusOK,
				articles[5:10],
			},
			{
				"get articles: allow public access with favorited queries",
				&model.User{ID: 0},
				struct {
					tag       string
					author    string
					favorited string
					limit     string
					offset    string
				}{
					tag:       "",
					author:    "",
					favorited: fooUser.Username,
					limit:     "0",
					offset:    "0",
				},
				http.StatusOK,
				articles[5:10],
			},
		}

		for _, tt := range tests {
			req := httptest.NewRequest(http.MethodPut, "/api/v1/articles", nil)

			q := req.URL.Query()
			q.Add("tag", tt.query.tag)
			q.Add("username", tt.query.author)
			q.Add("favorited", tt.query.favorited)
			q.Add("limit", tt.query.limit)
			q.Add("offset", tt.query.offset)
			req.URL.RawQuery = q.Encode()

			w := httptest.NewRecorder()
			ctx, _ := ctxWithToken(t, lct.Environ(), w, req, tt.reqUser.ID, time.Now())

			h.GetArticles(ctx)

			actualBody := test.GetResponseBody[message.ArticlesResponse](t, w.Result())

			assert.Equal(t, tt.expectedStatusCode, w.Result().StatusCode, tt.title)
			assert.Len(t, actualBody.Articles, len(tt.expectedArticles), tt.title)
			assert.Len(t, tt.expectedArticles, int(actualBody.ArticlesCount), tt.title)

			for i := 0; i < len(actualBody.Articles); i++ {
				got := actualBody.Articles[i]
				want := tt.expectedArticles[i]

				assert.Equal(t, want.ID, got.ID, tt.title)
				assert.Equal(t, want.Title, got.Title, tt.title)
				assert.Equal(t, want.Description, got.Description, tt.title)
				assert.Equal(t, want.Body, got.Body, tt.title)
				assert.Equal(t, want.Author.Username, got.Author.Username, tt.title)
			}
		}
	})

	t.Run("GetFeedArticles", func(t *testing.T) {
		fooUser := createRandomUser(t, lct.DB())
		barUser := createRandomUser(t, lct.DB())
		reqUser := createRandomUser(t, lct.DB())

		err := h.us.Follow(context.Background(), reqUser, barUser)
		if err != nil {
			t.Fatal(err)
		}

		tag := model.Tag{Name: test.RandomString(t, 10)}

		articles := make([]*model.Article, 0, 10)
		for i := 0; i < 10; i++ {
			randStr := test.RandomString(t, 10)
			article := model.Article{
				Title:       randStr,
				Description: randStr,
				Body:        randStr,
			}

			if i < 5 {
				article.UserID = fooUser.ID
				article.Author = *fooUser
				article.Tags = []model.Tag{tag}
			} else {
				article.UserID = barUser.ID
				article.Author = *barUser
			}

			articles = append(articles, &article)
		}

		for i, article := range articles {
			createdArticle, err := h.as.Create(context.Background(), article)
			if err != nil {
				t.Fatal(err)
			}

			article.ID = createdArticle.ID
			article.CreatedAt = createdArticle.CreatedAt
			article.UpdatedAt = createdArticle.UpdatedAt

			if i < 5 {
				err := h.as.AddFavorite(context.Background(), article, fooUser,
					func(favoritesCount int64, updatedAt time.Time) {
						article.FavoritesCount = favoritesCount
						article.UpdatedAt = updatedAt
					})
				if err != nil {
					t.Fatal(err)
				}
			}

			// delay creating articles, so we can sort them by date
			time.Sleep(1 * time.Second)
		}

		sort.SliceStable(articles, func(i, j int) bool {
			return articles[i].CreatedAt.After(articles[j].CreatedAt)
		})

		tests := []struct {
			title   string
			reqUser *model.User
			query   struct {
				limit  string
				offset string
			}
			expectedStatusCode int
			expectedArticles   []*model.Article
			expectedError      map[string]interface{}
			hasError           bool
		}{
			{
				"get articles: with default queries",
				reqUser,
				struct {
					limit  string
					offset string
				}{
					limit:  "0",
					offset: "0",
				},
				http.StatusOK,
				articles[0:5],
				nil,
				false,
			},
			{
				"get articles: with queries",
				reqUser,
				struct {
					limit  string
					offset string
				}{
					limit:  "2",
					offset: "1",
				},
				http.StatusOK,
				articles[1:3],
				nil,
				false,
			},
			{
				"get articles: with from user who does not follow anyone",
				fooUser,
				struct {
					limit  string
					offset string
				}{
					limit:  "2",
					offset: "1",
				},
				http.StatusOK,
				[]*model.Article{},
				nil,
				false,
			},
			{
				"get articles: wrong current user id",
				&model.User{ID: 0},
				struct {
					limit  string
					offset string
				}{
					limit:  "0",
					offset: "0",
				},
				http.StatusNotFound,
				nil,
				map[string]interface{}{"error": "current user not found"},
				true,
			},
		}

		for _, tt := range tests {
			req := httptest.NewRequest(http.MethodPut, "/api/v1/articles/feed", nil)

			q := req.URL.Query()
			q.Add("limit", tt.query.limit)
			q.Add("offset", tt.query.offset)
			req.URL.RawQuery = q.Encode()

			w := httptest.NewRecorder()
			ctx, _ := ctxWithToken(t, lct.Environ(), w, req, tt.reqUser.ID, time.Now())

			h.GetFeedArticles(ctx)

			assert.Equal(t, tt.expectedStatusCode, w.Result().StatusCode, tt.title)

			if tt.hasError {
				actualBody := test.GetResponseBody[map[string]interface{}](t, w.Result())
				assert.Equal(t, tt.expectedError, actualBody, tt.title)
			} else {
				actualBody := test.GetResponseBody[message.ArticlesResponse](t, w.Result())
				assert.Len(t, actualBody.Articles, len(tt.expectedArticles), tt.title)
				assert.Len(t, tt.expectedArticles, int(actualBody.ArticlesCount), tt.title)

				for i := 0; i < len(actualBody.Articles); i++ {
					got := actualBody.Articles[i]
					want := tt.expectedArticles[i]

					assert.Equal(t, want.ID, got.ID, tt.title)
					assert.Equal(t, want.Title, got.Title, tt.title)
					assert.Equal(t, want.Description, got.Description, tt.title)
					assert.Equal(t, want.Body, got.Body, tt.title)
					assert.Equal(t, want.Author.Username, got.Author.Username, tt.title)
				}
			}
		}
	})

	t.Run("UpdateArticle", func(t *testing.T) {
		shortMaxLenString := strings.Repeat("a", 101)

		fooUser := createRandomUser(t, lct.DB())
		fooArticle := createRandomArticle(t, lct.DB(), fooUser.ID)

		barUser := createRandomUser(t, lct.DB())
		barArticle := createRandomArticle(t, lct.DB(), barUser.ID)

		err := h.as.AddFavorite(context.Background(), fooArticle, fooUser,
			func(favoritesCount int64, updatedAt time.Time) {
				fooArticle.FavoritesCount = favoritesCount
				fooArticle.UpdatedAt = updatedAt
			},
		)
		if err != nil {
			t.Fatal(err)
		}

		randStr := test.RandomString(t, 20)
		article := model.Article{
			ID:             fooArticle.ID,
			Title:          randStr,
			Description:    randStr,
			Body:           randStr,
			Tags:           fooArticle.Tags,
			UserID:         fooArticle.UserID,
			Author:         fooArticle.Author,
			FavoritesCount: fooArticle.FavoritesCount,
			CreatedAt:      fooArticle.CreatedAt,
			UpdatedAt:      fooArticle.UpdatedAt,
		}

		tests := []struct {
			title              string
			reqUser            *model.User
			reqSlug            string
			reqBody            *message.UpdateArticleRequest
			expectedStatusCode int
			expectedBody       message.ArticleResponse
			expectedError      map[string]interface{}
			hasError           bool
		}{
			{
				"update article: success",
				fooUser,
				strconv.Itoa(int(fooArticle.ID)),
				&message.UpdateArticleRequest{
					Title:       article.Title,
					Description: article.Description,
					Body:        article.Body,
				},
				http.StatusOK,
				article.ResponseArticle(true, false),
				nil,
				false,
			},
			{
				"update article: ignore zero value field",
				fooUser,
				strconv.Itoa(int(fooArticle.ID)),
				&message.UpdateArticleRequest{
					Title:       article.Title,
					Description: article.Description,
					Body:        "",
				},
				http.StatusOK,
				article.ResponseArticle(true, false),
				nil,
				false,
			},
			{
				"update article: wrong current user id",
				&model.User{ID: 0},
				strconv.Itoa(int(fooArticle.ID)),
				&message.UpdateArticleRequest{},
				http.StatusNotFound,
				message.ArticleResponse{},
				map[string]interface{}{"error": "current user not found"},
				true,
			},
			{
				"update article: invalid slug",
				fooUser,
				"invalid_slug",
				&message.UpdateArticleRequest{},
				http.StatusBadRequest,
				message.ArticleResponse{},
				map[string]interface{}{"error": "invalid slug"},
				true,
			},
			{
				"update article: wrong slug",
				fooUser,
				"0",
				&message.UpdateArticleRequest{},
				http.StatusNotFound,
				message.ArticleResponse{},
				map[string]interface{}{"error": "article not found"},
				true,
			},
			{
				"update article: forbidden to update other user's article",
				fooUser,
				strconv.Itoa(int(barArticle.ID)),
				&message.UpdateArticleRequest{},
				http.StatusForbidden,
				message.ArticleResponse{},
				map[string]interface{}{"error": "forbidden"},
				true,
			},
			{
				"update article: title is too short",
				fooUser,
				strconv.Itoa(int(fooArticle.ID)),
				&message.UpdateArticleRequest{
					Title:       "abc",
					Description: article.Description,
					Body:        article.Body,
				},
				http.StatusBadRequest,
				message.ArticleResponse{},
				map[string]interface{}{"error": "validation error: Title: the length must be between 5 and 100."},
				true,
			},
			{
				"update article: title is too long",
				fooUser,
				strconv.Itoa(int(fooArticle.ID)),
				&message.UpdateArticleRequest{
					Title:       shortMaxLenString,
					Description: article.Description,
					Body:        article.Body,
				},
				http.StatusBadRequest,
				message.ArticleResponse{},
				map[string]interface{}{"error": "validation error: Title: the length must be between 5 and 100."},
				true,
			},
		}

		for _, tt := range tests {
			body, err := json.Marshal(tt.reqBody)
			if err != nil {
				t.Fatal(err)
			}

			apiUrl := fmt.Sprintf("/api/v1/articles/%v", tt.reqSlug)
			req := httptest.NewRequest(http.MethodPut, apiUrl, bytes.NewReader(body))

			w := httptest.NewRecorder()
			ctx, _ := ctxWithToken(t, lct.Environ(), w, req, tt.reqUser.ID, time.Now())
			ctx.AddParam("slug", tt.reqSlug)

			h.UpdateArticle(ctx)

			assert.Equal(t, tt.expectedStatusCode, w.Result().StatusCode, tt.title)

			if tt.hasError {
				actualBody := test.GetResponseBody[map[string]interface{}](t, w.Result())
				assert.Equal(t, tt.expectedError, actualBody, tt.title)
			} else {
				actualBody := test.GetResponseBody[message.ArticleResponse](t, w.Result())
				assert.Equal(t, tt.expectedBody.ID, actualBody.ID, tt.title)
				assert.Equal(t, tt.expectedBody.Title, actualBody.Title, tt.title)
				assert.Equal(t, tt.expectedBody.Description, actualBody.Description, tt.title)
				assert.Equal(t, tt.expectedBody.Body, actualBody.Body, tt.title)
				assert.Equal(t, tt.expectedBody.Favorited, actualBody.Favorited, tt.title)
				assert.Equal(t, tt.expectedBody.FavoritesCount, actualBody.FavoritesCount, tt.title)
				assert.Equal(t, tt.expectedBody.Author, actualBody.Author, tt.title)
				assert.Equal(t, tt.expectedBody.CreatedAt, actualBody.CreatedAt, tt.title)
				assert.NotEqual(t, tt.expectedBody.UpdatedAt, actualBody.UpdatedAt, tt.title)
			}
		}
	})

	t.Run("DeleteArticle", func(t *testing.T) {
		fooUser := createRandomUser(t, lct.DB())
		fooArticle := createRandomArticle(t, lct.DB(), fooUser.ID)

		barUser := createRandomUser(t, lct.DB())
		barArticle := createRandomArticle(t, lct.DB(), barUser.ID)

		tests := []struct {
			title              string
			reqUser            *model.User
			reqSlug            string
			expectedStatusCode int
			expectedError      map[string]interface{}
			hasError           bool
		}{
			{
				"delete article: success",
				fooUser,
				strconv.Itoa(int(fooArticle.ID)),
				http.StatusNoContent,
				nil,
				false,
			},
			{
				"delete article: wrong current user id",
				&model.User{ID: 0},
				strconv.Itoa(int(fooArticle.ID)),
				http.StatusNotFound,
				map[string]interface{}{"error": "current user not found"},
				true,
			},
			{
				"delete article: invalid slug",
				fooUser,
				"invalid_slug",
				http.StatusBadRequest,
				map[string]interface{}{"error": "invalid slug"},
				true,
			},
			{
				"delete article: wrong slug",
				fooUser,
				"0",
				http.StatusNotFound,
				map[string]interface{}{"error": "article not found"},
				true,
			},
			{
				"delete article: forbidden to delete other user's article",
				fooUser,
				strconv.Itoa(int(barArticle.ID)),
				http.StatusForbidden,
				map[string]interface{}{"error": "forbidden"},
				true,
			},
		}

		for _, tt := range tests {
			apiUrl := fmt.Sprintf("/api/v1/articles/%v", tt.reqSlug)
			req := httptest.NewRequest(http.MethodDelete, apiUrl, nil)

			w := httptest.NewRecorder()
			ctx, _ := ctxWithToken(t, lct.Environ(), w, req, tt.reqUser.ID, time.Now())
			ctx.AddParam("slug", tt.reqSlug)

			h.DeleteArticle(ctx)

			assert.Equal(t, tt.expectedStatusCode, w.Result().StatusCode, tt.title)

			if tt.hasError {
				actualBody := test.GetResponseBody[map[string]interface{}](t, w.Result())
				assert.Equal(t, tt.expectedError, actualBody, tt.title)
			} else {
				reqID, err := strconv.ParseUint(tt.reqSlug, 10, 0)
				if err != nil {
					t.Fatal(err)
				}

				actualArticle, err := h.as.GetByID(context.Background(), uint(reqID))

				assert.Error(t, err, tt.title)
				assert.Nil(t, actualArticle, tt.title)
			}
		}
	})

	t.Run("FavoriteArticle", func(t *testing.T) {
		fooUser := createRandomUser(t, lct.DB())
		barUser := createRandomUser(t, lct.DB())

		fooArticle := createRandomArticle(t, lct.DB(), fooUser.ID)
		barArticle := createRandomArticle(t, lct.DB(), barUser.ID)

		err := h.as.AddFavorite(context.Background(), fooArticle, fooUser,
			func(favoritesCount int64, updatedAt time.Time) {
				fooArticle.FavoritesCount = favoritesCount
				fooArticle.UpdatedAt = updatedAt
			},
		)
		if err != nil {
			t.Fatal(err)
		}

		expected := barArticle.ResponseArticle(true, false)
		expected.FavoritesCount = 1

		tests := []struct {
			title              string
			reqUser            *model.User
			reqSlug            string
			expectedStatusCode int
			expectedBody       message.ArticleResponse
			expectedError      map[string]interface{}
			hasError           bool
		}{
			{
				"favorite article: success",
				fooUser,
				strconv.Itoa(int(barArticle.ID)),
				http.StatusOK,
				expected,
				nil,
				false,
			},
			{
				"favorite article: wrong current user id",
				&model.User{ID: 0},
				strconv.Itoa(int(barArticle.ID)),
				http.StatusNotFound,
				message.ArticleResponse{},
				map[string]interface{}{"error": "current user not found"},
				true,
			},
			{
				"favorite article: invalid slug",
				fooUser,
				"invalid_slug",
				http.StatusBadRequest,
				message.ArticleResponse{},
				map[string]interface{}{"error": "invalid slug"},
				true,
			},
			{
				"favorite article: wrong slug",
				fooUser,
				"0",
				http.StatusNotFound,
				message.ArticleResponse{},
				map[string]interface{}{"error": "article not found"},
				true,
			},
			{
				"favorite article: already favorited this article",
				fooUser,
				strconv.Itoa(int(fooArticle.ID)),
				http.StatusBadRequest,
				message.ArticleResponse{},
				map[string]interface{}{"error": "you already favorited this article"},
				true,
			},
		}

		for _, tt := range tests {
			apiUrl := fmt.Sprintf("/api/v1/articles/%v/favorite", tt.reqSlug)
			req := httptest.NewRequest(http.MethodPost, apiUrl, nil)

			w := httptest.NewRecorder()
			ctx, _ := ctxWithToken(t, lct.Environ(), w, req, tt.reqUser.ID, time.Now())
			ctx.AddParam("slug", tt.reqSlug)

			h.FavoriteArticle(ctx)

			assert.Equal(t, tt.expectedStatusCode, w.Result().StatusCode, tt.title)

			if tt.hasError {
				actualBody := test.GetResponseBody[map[string]interface{}](t, w.Result())
				assert.Equal(t, tt.expectedError, actualBody, tt.title)
			} else {
				actualBody := test.GetResponseBody[message.ArticleResponse](t, w.Result())
				assert.Equal(t, tt.expectedBody.ID, actualBody.ID, tt.title)
				assert.Equal(t, tt.expectedBody.Title, actualBody.Title, tt.title)
				assert.Equal(t, tt.expectedBody.Description, actualBody.Description, tt.title)
				assert.Equal(t, tt.expectedBody.Body, actualBody.Body, tt.title)
				assert.Equal(t, tt.expectedBody.Favorited, actualBody.Favorited, tt.title)
				assert.Equal(t, tt.expectedBody.FavoritesCount, actualBody.FavoritesCount, tt.title)
				assert.Equal(t, tt.expectedBody.Author, actualBody.Author, tt.title)
				assert.Equal(t, tt.expectedBody.CreatedAt, actualBody.CreatedAt, tt.title)
				assert.NotEqual(t, tt.expectedBody.UpdatedAt, actualBody.UpdatedAt, tt.title)

				reqID, err := strconv.ParseUint(tt.reqSlug, 10, 0)
				if err != nil {
					t.Fatal(err)
				}

				actualArticle, err := h.as.GetByID(context.Background(), uint(reqID))
				if err != nil {
					t.Fatal(err)
				}

				actualFavorited, err := h.as.IsFavorited(context.Background(), actualArticle, tt.reqUser)
				if err != nil {
					t.Fatal(err)
				}

				assert.True(t, actualFavorited, tt.title)
			}
		}
	})

	t.Run("UnfavoriteArticle", func(t *testing.T) {
		fooUser := createRandomUser(t, lct.DB())
		barUser := createRandomUser(t, lct.DB())

		fooArticle := createRandomArticle(t, lct.DB(), fooUser.ID)
		barArticle := createRandomArticle(t, lct.DB(), barUser.ID)

		err := h.as.AddFavorite(context.Background(), barArticle, fooUser,
			func(favoritesCount int64, updatedAt time.Time) {
				barArticle.FavoritesCount = favoritesCount
				barArticle.UpdatedAt = updatedAt
			})
		if err != nil {
			t.Fatal(err)
		}

		expected := barArticle.ResponseArticle(false, false)
		expected.FavoritesCount = 0

		tests := []struct {
			title              string
			reqUser            *model.User
			reqSlug            string
			expectedStatusCode int
			expectedBody       message.ArticleResponse
			expectedError      map[string]interface{}
			hasError           bool
		}{
			{
				"unfavorite article: success",
				fooUser,
				strconv.Itoa(int(barArticle.ID)),
				http.StatusOK,
				expected,
				nil,
				false,
			},
			{
				"unfavorite article: wrong current user id",
				&model.User{ID: 0},
				strconv.Itoa(int(barArticle.ID)),
				http.StatusNotFound,
				message.ArticleResponse{},
				map[string]interface{}{"error": "current user not found"},
				true,
			},
			{
				"unfavorite article: invalid slug",
				fooUser,
				"invalid_slug",
				http.StatusBadRequest,
				message.ArticleResponse{},
				map[string]interface{}{"error": "invalid slug"},
				true,
			},
			{
				"unfavorite article: wrong slug",
				fooUser,
				"0",
				http.StatusNotFound,
				message.ArticleResponse{},
				map[string]interface{}{"error": "article not found"},
				true,
			},
			{
				"unfavorite article: already unfavorited this article",
				fooUser,
				strconv.Itoa(int(fooArticle.ID)),
				http.StatusBadRequest,
				message.ArticleResponse{},
				map[string]interface{}{"error": "you did not favorite this article"},
				true,
			},
		}

		for _, tt := range tests {
			apiUrl := fmt.Sprintf("/api/v1/articles/%v/favorite", tt.reqSlug)
			req := httptest.NewRequest(http.MethodDelete, apiUrl, nil)

			w := httptest.NewRecorder()
			ctx, _ := ctxWithToken(t, lct.Environ(), w, req, tt.reqUser.ID, time.Now())
			ctx.AddParam("slug", tt.reqSlug)

			h.UnfavoriteArticle(ctx)

			assert.Equal(t, tt.expectedStatusCode, w.Result().StatusCode, tt.title)

			if tt.hasError {
				actualBody := test.GetResponseBody[map[string]interface{}](t, w.Result())
				assert.Equal(t, tt.expectedError, actualBody, tt.title)
			} else {
				actualBody := test.GetResponseBody[message.ArticleResponse](t, w.Result())
				assert.Equal(t, tt.expectedBody.ID, actualBody.ID, tt.title)
				assert.Equal(t, tt.expectedBody.Title, actualBody.Title, tt.title)
				assert.Equal(t, tt.expectedBody.Description, actualBody.Description, tt.title)
				assert.Equal(t, tt.expectedBody.Body, actualBody.Body, tt.title)
				assert.Equal(t, tt.expectedBody.Favorited, actualBody.Favorited, tt.title)
				assert.Equal(t, tt.expectedBody.FavoritesCount, actualBody.FavoritesCount, tt.title)
				assert.Equal(t, tt.expectedBody.Author, actualBody.Author, tt.title)
				assert.Equal(t, tt.expectedBody.CreatedAt, actualBody.CreatedAt, tt.title)
				assert.NotEqual(t, tt.expectedBody.UpdatedAt, actualBody.UpdatedAt, tt.title)

				reqID, err := strconv.ParseUint(tt.reqSlug, 10, 0)
				if err != nil {
					t.Fatal(err)
				}

				actualArticle, err := h.as.GetByID(context.Background(), uint(reqID))
				if err != nil {
					t.Fatal(err)
				}

				actualFavorited, err := h.as.IsFavorited(context.Background(), actualArticle, tt.reqUser)
				if err != nil {
					t.Fatal(err)
				}

				assert.False(t, actualFavorited, tt.title)
			}
		}
	})
}
