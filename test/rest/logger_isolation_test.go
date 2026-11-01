//go:build test

package rest

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"

	"bctbackend/logging"
	path "bctbackend/server/paths"
	server "bctbackend/server"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"

	"github.com/stretchr/testify/require"
)

// recordedEntry represents a single captured log line.
type recordedEntry struct {
	level      string
	message    string
	attributes map[string]any
}

// recordingState is the shared, thread-safe log store of a recordingLogger and
// all loggers derived from it through With.
type recordingState struct {
	mutex   sync.Mutex
	entries []recordedEntry
}

// recordingLogger is a thread-safe logging.Logger that captures every log line
// together with its attributes, so that tests can assert on what gets logged.
type recordingLogger struct {
	state      *recordingState
	attributes map[string]any
}

func newRecordingLogger() *recordingLogger {
	return &recordingLogger{state: &recordingState{}, attributes: make(map[string]any)}
}

func (l *recordingLogger) Debug(message string) { l.record("debug", message) }
func (l *recordingLogger) Info(message string)  { l.record("info", message) }
func (l *recordingLogger) Warn(message string)  { l.record("warn", message) }
func (l *recordingLogger) Error(message string) { l.record("error", message) }

func (l *recordingLogger) With(key string, value any) logging.Logger {
	l.state.mutex.Lock()
	defer l.state.mutex.Unlock()

	attributes := make(map[string]any, len(l.attributes)+1)
	for k, v := range l.attributes {
		attributes[k] = v
	}
	attributes[key] = value

	return &recordingLogger{state: l.state, attributes: attributes}
}

func (l *recordingLogger) record(level string, message string) {
	l.state.mutex.Lock()
	defer l.state.mutex.Unlock()

	l.state.entries = append(l.state.entries, recordedEntry{level: level, message: message, attributes: l.attributes})
}

func (l *recordingLogger) snapshot() []recordedEntry {
	l.state.mutex.Lock()
	defer l.state.mutex.Unlock()

	return append([]recordedEntry(nil), l.state.entries...)
}

// newIsolationTestServer creates a REST server over an in-memory database,
// wired to a recording logger.
func newIsolationTestServer(t *testing.T) (*DatabaseFixture, *recordingLogger, *server.Server) {
	databaseFixture, db := NewDatabaseFixture(t, WithDefaultCategories)
	fixture := &databaseFixture
	logger := newRecordingLogger()
	server := aux.CreateRestServer(db, logger, fixture.Clock)

	return fixture, logger, server
}

func doLoginRequest(restServer *server.Server, username string, password string) *httptest.ResponseRecorder {
	form := url.Values{}
	form.Add("username", username)
	form.Add("password", password)

	request, _ := http.NewRequest("POST", path.Login().String(), bytes.NewBufferString(form.Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	recorder := httptest.NewRecorder()
	restServer.ServeHTTP(recorder, request)

	return recorder
}

func loginRequest(t *testing.T, restServer *server.Server, username string, password string) *httptest.ResponseRecorder {
	t.Helper()

	recorder := doLoginRequest(restServer, username, password)
	require.NotEqual(t, http.StatusInternalServerError, recorder.Code, "unexpected server error: %s", recorder.Body.String())

	return recorder
}

func TestRestLoggerIsolation(t *testing.T) {
	t.Run("Attributes of one login request do not leak into the next request", func(t *testing.T) {
		fixture, logger, restServer := newIsolationTestServer(t)
		defer restServer.Shutdown()
		defer fixture.Close()

		seller := fixture.Seller(t)

		// First request: a successful login, whose handler decorates the logger
		// with request-specific attributes (user ID and submitted password).
		successfulLogin := doLoginRequest(restServer, seller.UserID.String(), seller.Password)
		require.Equal(t, http.StatusOK, successfulLogin.Code)

		entriesAfterFirstRequest := len(logger.snapshot())

		// Second request: an unrelated request, whose log lines must not carry
		// any attribute from the first request.
		unrelatedRequest := loginRequest(t, restServer, "not-a-number", "irrelevant")
		require.Equal(t, http.StatusBadRequest, unrelatedRequest.Code, "invalid user ID should be rejected")

		for _, entry := range logger.snapshot()[entriesAfterFirstRequest:] {
			for key, value := range entry.attributes {
				require.NotEqual(t, seller.UserID, value, "user ID from first request leaked into log attribute %q", key)
				require.NotEqual(t, seller.Password, value, "password from first request leaked into log attribute %q", key)
			}
		}
	})

	t.Run("Concurrent login requests do not race on the logger", func(t *testing.T) {
		fixture, _, restServer := newIsolationTestServer(t)
		defer restServer.Shutdown()
		defer fixture.Close()

		const workers = 32
		const requestsPerWorker = 8

		start := make(chan struct{})
		var wg sync.WaitGroup
		codes := make(chan int, workers*requestsPerWorker)

		for worker := 0; worker < workers; worker++ {
			wg.Add(1)

			go func(worker int) {
				defer wg.Done()
				<-start

				for i := 0; i < requestsPerWorker; i++ {
					username := fmt.Sprintf("not-a-number-%d-%d", worker, i)
					recorder := doLoginRequest(restServer, username, "irrelevant")
					codes <- recorder.Code
				}
			}(worker)
		}

		close(start)
		wg.Wait()
		close(codes)

		for code := range codes {
			require.Equal(t, http.StatusBadRequest, code)
		}
	})
}
