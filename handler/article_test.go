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
		fooUser := createRandomUser(t, lct.DB())

		randStr := test.RandomString(t, 20)
		article := model.Article{
			Title:       randStr,
			Description: randStr,
			Body:        randStr,
			Author:      *fooUser,
		}

		expected := article.ResponseArticle(false, false)

		r := message.CreateArticleRequest{
			Title:       article.Title,
			Description: article.Description,
			Body:        article.Body,
			Tags: []string{
				test.RandomString(t, 10),
				test.RandomString(t, 10),
			},
		}

		body, err := json.Marshal(r)
		if err != nil {
			t.Fatal(err)
		}

		req := httptest.NewRequest(http.MethodPut, "/api/v1/articles", bytes.NewReader(body))

		w := httptest.NewRecorder()
		ctx, _ := ctxWithToken(t, lct.Environ(), w, req, fooUser.ID, time.Now())

		h.CreateArticle(ctx)

		var actual message.ArticleResponse
		err = json.NewDecoder(w.Result().Body).Decode(&actual)
		if err != nil {
			t.Fatal(err)
		}
		defer w.Result().Body.Close()

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.NotEmpty(t, actual.ID)
		assert.Equal(t, expected.Title, actual.Title)
		assert.Equal(t, expected.Description, actual.Description)
		assert.Equal(t, expected.Body, actual.Body)
		assert.Equal(t, expected.Favorited, actual.Favorited)
		assert.Equal(t, expected.FavoritesCount, actual.FavoritesCount)
		assert.Equal(t, expected.Author, actual.Author)
		assert.NotEmpty(t, actual.CreatedAt)
		assert.NotEmpty(t, actual.UpdatedAt)
	})

	t.Run("GetArticle", func(t *testing.T) {
		fooUser := createRandomUser(t, lct.DB())

		fooArticle := createRandomArticle(t, lct.DB(), fooUser.ID)
		expected := fooArticle.ResponseArticle(false, false)

		apiUrl := fmt.Sprintf("/api/v1/articles/%d", fooArticle.ID)
		req := httptest.NewRequest(http.MethodGet, apiUrl, nil)

		w := httptest.NewRecorder()
		ctx, _ := ctxWithToken(t, lct.Environ(), w, req, fooUser.ID, time.Now())
		ctx.AddParam("slug", strconv.Itoa(int(fooArticle.ID)))

		h.GetArticle(ctx)

		var actual message.ArticleResponse
		err := json.NewDecoder(w.Result().Body).Decode(&actual)
		if err != nil {
			t.Fatal(err)
		}
		defer w.Result().Body.Close()

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Equal(t, expected.ID, actual.ID)
		assert.Equal(t, expected.Title, actual.Title)
		assert.Equal(t, expected.Description, actual.Description)
		assert.Equal(t, expected.Body, actual.Body)
		assert.Equal(t, expected.Favorited, actual.Favorited)
		assert.Equal(t, expected.FavoritesCount, actual.FavoritesCount)
		assert.Equal(t, expected.Author, actual.Author)
		assert.Equal(t, expected.CreatedAt, actual.CreatedAt)
		assert.Equal(t, expected.UpdatedAt, actual.UpdatedAt)
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
			title string
			query struct {
				tag       string
				author    string
				favorited string
				limit     string
				offset    string
			}
			expected []*model.Article
		}{
			{
				"get articles with default queries",
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
				articles,
			},
			{
				"get articles with limit and offset",
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
				articles[5:10],
			},
			{
				"get articles with tag",
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
				articles[5:10],
			},
			{
				"get articles with author",
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
				articles[0:5],
			},
			{
				"get articles with various queries",
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
				articles[6:8],
			},
			{
				"get articles with favorited queries",
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
			ctx, _ := ctxWithToken(t, lct.Environ(), w, req, reqUser.ID, time.Now())

			h.GetArticles(ctx)

			var actual message.ArticlesResponse
			err := json.NewDecoder(w.Result().Body).Decode(&actual)
			if err != nil {
				t.Fatal(err)
			}
			defer w.Result().Body.Close()

			assert.Equal(t, http.StatusOK, w.Result().StatusCode, tt.title)
			assert.Len(t, actual.Articles, len(tt.expected), tt.title)
			assert.EqualValues(t, len(tt.expected), actual.ArticlesCount, tt.title)

			for i := 0; i < len(actual.Articles); i++ {
				got := actual.Articles[i]
				want := tt.expected[i]

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
			expected []*model.Article
		}{
			{
				"get articles with default queries",
				reqUser,
				struct {
					limit  string
					offset string
				}{
					limit:  "0",
					offset: "0",
				},
				articles[0:5],
			},
			{
				"get articles with queries",
				reqUser,
				struct {
					limit  string
					offset string
				}{
					limit:  "2",
					offset: "1",
				},
				articles[1:3],
			},
			{
				"get articles of user who has no followings",
				fooUser,
				struct {
					limit  string
					offset string
				}{
					limit:  "2",
					offset: "1",
				},
				[]*model.Article{},
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

			var actual message.ArticlesResponse
			err := json.NewDecoder(w.Result().Body).Decode(&actual)
			if err != nil {
				t.Fatal(err)
			}
			defer w.Result().Body.Close()

			assert.Equal(t, http.StatusOK, w.Result().StatusCode, tt.title)
			assert.Len(t, actual.Articles, len(tt.expected), tt.title)
			assert.EqualValues(t, len(tt.expected), actual.ArticlesCount, tt.title)

			for i := 0; i < len(actual.Articles); i++ {
				got := actual.Articles[i]
				want := tt.expected[i]

				assert.Equal(t, want.ID, got.ID, tt.title)
				assert.Equal(t, want.Title, got.Title, tt.title)
				assert.Equal(t, want.Description, got.Description, tt.title)
				assert.Equal(t, want.Body, got.Body, tt.title)
				assert.Equal(t, want.Author.Username, got.Author.Username, tt.title)
			}
		}
	})

	t.Run("UpdateArticle", func(t *testing.T) {
		fooUser := createRandomUser(t, lct.DB())
		fooArticle := createRandomArticle(t, lct.DB(), fooUser.ID)

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
		expected := article.ResponseArticle(false, false)

		r := message.UpdateArticleRequest{
			Title:       randStr,
			Description: randStr,
			Body:        randStr,
		}

		body, err := json.Marshal(r)
		if err != nil {
			t.Fatal(err)
		}

		apiUrl := fmt.Sprintf("/api/v1/articles/%d", fooArticle.ID)
		req := httptest.NewRequest(http.MethodPut, apiUrl, bytes.NewReader(body))

		w := httptest.NewRecorder()
		ctx, _ := ctxWithToken(t, lct.Environ(), w, req, fooUser.ID, time.Now())
		ctx.AddParam("slug", strconv.Itoa(int(fooArticle.ID)))

		h.UpdateArticle(ctx)

		var actual message.ArticleResponse
		err = json.NewDecoder(w.Result().Body).Decode(&actual)
		if err != nil {
			t.Fatal(err)
		}
		defer w.Result().Body.Close()

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Equal(t, expected.ID, actual.ID)
		assert.Equal(t, expected.Title, actual.Title)
		assert.Equal(t, expected.Description, actual.Description)
		assert.Equal(t, expected.Body, actual.Body)
		assert.Equal(t, expected.Favorited, actual.Favorited)
		assert.Equal(t, expected.FavoritesCount, actual.FavoritesCount)
		assert.Equal(t, expected.Author, actual.Author)
		assert.Equal(t, expected.CreatedAt, actual.CreatedAt)
		assert.NotEqual(t, expected.UpdatedAt, actual.UpdatedAt)
	})

	t.Run("DeleteArticle", func(t *testing.T) {
		fooUser := createRandomUser(t, lct.DB())
		fooArticle := createRandomArticle(t, lct.DB(), fooUser.ID)

		apiUrl := fmt.Sprintf("/api/v1/articles/%d", fooArticle.ID)
		req := httptest.NewRequest(http.MethodDelete, apiUrl, nil)

		w := httptest.NewRecorder()
		ctx, _ := ctxWithToken(t, lct.Environ(), w, req, fooUser.ID, time.Now())
		ctx.AddParam("slug", strconv.Itoa(int(fooArticle.ID)))

		h.DeleteArticle(ctx)

		actual, err := h.as.GetByID(context.Background(), fooArticle.ID)

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	t.Run("FavoriteArticle", func(t *testing.T) {
		fooUser := createRandomUser(t, lct.DB())
		barUser := createRandomUser(t, lct.DB())

		barArticle := createRandomArticle(t, lct.DB(), barUser.ID)
		expected := barArticle.ResponseArticle(true, false)
		expected.FavoritesCount = 1

		apiUrl := fmt.Sprintf("/api/v1/articles/%d/favorite", barArticle.ID)
		req := httptest.NewRequest(http.MethodPost, apiUrl, nil)

		w := httptest.NewRecorder()
		ctx, _ := ctxWithToken(t, lct.Environ(), w, req, fooUser.ID, time.Now())
		ctx.AddParam("slug", strconv.Itoa(int(barArticle.ID)))

		h.FavoriteArticle(ctx)

		var actual message.ArticleResponse
		err := json.NewDecoder(w.Result().Body).Decode(&actual)
		if err != nil {
			t.Fatal(err)
		}
		defer w.Result().Body.Close()

		actualFavorited, err := h.as.IsFavorited(context.Background(), barArticle, fooUser)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.True(t, actualFavorited)
		assert.Equal(t, expected.ID, actual.ID)
		assert.Equal(t, expected.Title, actual.Title)
		assert.Equal(t, expected.Description, actual.Description)
		assert.Equal(t, expected.Body, actual.Body)
		assert.Equal(t, expected.Favorited, actual.Favorited)
		assert.Equal(t, expected.FavoritesCount, actual.FavoritesCount)
		assert.Equal(t, expected.Author, actual.Author)
		assert.Equal(t, expected.CreatedAt, actual.CreatedAt)
		assert.NotEqual(t, expected.UpdatedAt, actual.UpdatedAt)
	})

	t.Run("UnfavoriteArticle", func(t *testing.T) {
		fooUser := createRandomUser(t, lct.DB())
		barUser := createRandomUser(t, lct.DB())

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

		apiUrl := fmt.Sprintf("/api/v1/articles/%d/favorite", barArticle.ID)
		req := httptest.NewRequest(http.MethodDelete, apiUrl, nil)

		w := httptest.NewRecorder()
		ctx, _ := ctxWithToken(t, lct.Environ(), w, req, fooUser.ID, time.Now())
		ctx.AddParam("slug", strconv.Itoa(int(barArticle.ID)))

		h.UnfavoriteArticle(ctx)

		var actual message.ArticleResponse
		err = json.NewDecoder(w.Result().Body).Decode(&actual)
		if err != nil {
			t.Fatal(err)
		}
		defer w.Result().Body.Close()

		actualFavorited, err := h.as.IsFavorited(context.Background(), barArticle, fooUser)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.False(t, actualFavorited)
		assert.Equal(t, expected.ID, actual.ID)
		assert.Equal(t, expected.Title, actual.Title)
		assert.Equal(t, expected.Description, actual.Description)
		assert.Equal(t, expected.Body, actual.Body)
		assert.Equal(t, expected.Favorited, actual.Favorited)
		assert.Equal(t, expected.FavoritesCount, actual.FavoritesCount)
		assert.Equal(t, expected.Author, actual.Author)
		assert.Equal(t, expected.CreatedAt, actual.CreatedAt)
		assert.NotEqual(t, expected.UpdatedAt, actual.UpdatedAt)
	})
}
