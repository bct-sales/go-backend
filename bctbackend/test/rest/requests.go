package rest

import (
	"bctbackend/security"
	"net/http"
	"strings"
	"time"
)

func CreateGetRequest(url string, options ...func(*http.Request)) *http.Request {
	request, err := http.NewRequest("GET", url, nil)

	if err != nil {
		panic(err)
	}

	for _, option := range options {
		option(request)
	}

	return request
}

func CreatePostRequest[T any](url string, payload *T, options ...func(*http.Request)) *http.Request {
	payloadJson := ToJson(payload)
	request, err := http.NewRequest("POST", url, strings.NewReader(payloadJson))

	if err != nil {
		panic(err)
	}

	request.Header.Set("Content-Type", "application/json")

	for _, option := range options {
		option(request)
	}

	return request
}

func CreatePutRequest[T any](url string, payload *T, options ...func(*http.Request)) *http.Request {
	payloadJson := ToJson(payload)
	request, err := http.NewRequest("PUT", url, strings.NewReader(payloadJson))

	if err != nil {
		panic(err)
	}

	request.Header.Set("Content-Type", "application/json")

	for _, option := range options {
		option(request)
	}

	return request
}

func createCookie(sessionId string) *http.Cookie {
	return &http.Cookie{
		Name:     security.SessionCookieName,
		Value:    sessionId,
		Expires:  time.Now().Add(time.Hour),
		Path:     "/",
		Domain:   "localhost",
		HttpOnly: true,
		SameSite: http.SameSiteNoneMode,
		Secure:   false,
	}
}

func WithCookie(sessionId string) func(*http.Request) {
	return func(request *http.Request) {
		cookie := createCookie(sessionId)
		request.AddCookie(cookie)
	}
}
