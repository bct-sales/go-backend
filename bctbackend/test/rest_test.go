package test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	models "bctbackend/database/models"

	"github.com/stretchr/testify/assert"
)

func TestListItems(t *testing.T) {
	db, router := createRestRouter()
	w := httptest.NewRecorder()
	defer db.Close()

	// Create an example user for testing
	request, err := http.NewRequest("GET", "/api/v1/items", nil)

	if assert.NoError(t, err) {
		router.ServeHTTP(w, request)

		assert.Equal(t, 200, w.Code)

		expected := []models.Item{}
		actual := fromJson[[]models.Item](w.Body.String())
		assert.Equal(t, expected, *actual)
	}
}
