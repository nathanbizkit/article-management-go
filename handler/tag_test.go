package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/nathanbizkit/article-management/message"
	"github.com/nathanbizkit/article-management/model"
	"github.com/nathanbizkit/article-management/test"
	"github.com/stretchr/testify/assert"
)

func TestIntegration_TagHandler(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration tests.")
	}

	gin.SetMode("test")
	h, lct := setup(t)

	t.Run("GetTags", func(t *testing.T) {
		fooUser := createRandomUser(t, lct.DB())
		barUser := createRandomUser(t, lct.DB())

		tags := make([]string, 0, 20)
		for i := 0; i < 10; i++ {
			randStr := test.RandomString(t, 10)
			a := model.Article{
				Title:       randStr,
				Description: randStr,
				Body:        randStr,
			}

			if i < 5 {
				a.UserID = fooUser.ID
				a.Author = *fooUser
			} else {
				a.UserID = barUser.ID
				a.Author = *barUser
			}

			tag1 := test.RandomString(t, 10)
			tag2 := test.RandomString(t, 10)
			tags = append(tags, tag1)
			tags = append(tags, tag2)

			a.Tags = []model.Tag{{Name: tag1}, {Name: tag2}}

			_, err := h.as.Create(context.Background(), &a)
			if err != nil {
				t.Fatal(err)
			}
		}

		expected := message.TagsResponse{Tags: tags}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/tags", nil)

		h.GetTags(c)

		var actual message.TagsResponse
		err := json.NewDecoder(w.Result().Body).Decode(&actual)
		if err != nil {
			t.Fatal(err)
		}
		defer w.Result().Body.Close()

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Equal(t, expected, actual)
	})
}
