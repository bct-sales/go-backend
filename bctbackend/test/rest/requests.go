package rest

import (
	"bctbackend/security"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	HTTP_VERB_GET    = "GET"
	HTTP_VERB_POST   = "POST"
	HTTP_VERB_PUT    = "PUT"
	HTTP_VERB_DELETE = "DELETE"
)

func createRequest[T any](verb string, url string, payload *T, options ...func(*http.Request)) *http.Request {
	var reader io.Reader
	if payload != nil {
		payloadJson := ToJson(payload)
		reader = strings.NewReader(payloadJson)
	}
	request, err := http.NewRequest(verb, url, reader)

	if err != nil {
		panic(err)
	}

	for _, option := range options {
		option(request)
	}

	return request
}

func CreateGetRequest(url string, options ...func(*http.Request)) *http.Request {
	return createRequest[any](HTTP_VERB_GET, url, nil, options...)
}

func CreatePostRequest[T any](url string, payload *T, options ...func(*http.Request)) *http.Request {
	options = append(options, WithJsonContentType())
	return createRequest(HTTP_VERB_POST, url, payload, options...)
}

func CreatePutRequest[T any](url string, payload *T, options ...func(*http.Request)) *http.Request {
	options = append(options, WithJsonContentType())
	return createRequest(HTTP_VERB_PUT, url, payload, options...)
}

func createCookie(name string, value string) *http.Cookie {
	return &http.Cookie{
		Name:     name,
		Value:    value,
		Expires:  time.Now().Add(time.Hour),
		Path:     "/",
		Domain:   "localhost",
		HttpOnly: true,
		SameSite: http.SameSiteNoneMode,
		Secure:   false,
	}
}

func createSessionCookie(sessionId string) *http.Cookie {
	return createCookie(security.SessionCookieName, sessionId)
}

func WithCookie(name string, value string) func(*http.Request) {
	return func(request *http.Request) {
		cookie := createCookie(name, value)
		request.AddCookie(cookie)
	}
}

func WithSessionCookie(sessionId string) func(*http.Request) {
	return func(request *http.Request) {
		cookie := createSessionCookie(sessionId)
		request.AddCookie(cookie)
	}
}

func WithHeader(key string, value string) func(*http.Request) {
	return func(request *http.Request) {
		request.Header.Set(key, value)
	}
}

func WithContentType(contentType string) func(*http.Request) {
	return WithHeader("Content-Type", contentType)
}

func WithJsonContentType() func(*http.Request) {
	return WithContentType("application/json")
}
