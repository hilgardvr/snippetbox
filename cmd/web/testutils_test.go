package main

import (
	"bytes"
	"html"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
	"hilgardvr.com/snippetbox/internal/models/mocks"
)

var csrfTokenRx = regexp.MustCompile(`<input type="hidden" name="csrf_token" value="(.+)">`)

func extractCSRFToken(t *testing.T, body string) string {
	matches := csrfTokenRx.FindStringSubmatch(body)
	if len(matches) < 2 {
		t.Fatal("no CSRF token found in body")
	}
	return html.UnescapeString(matches[1])
}

func newTestApplication(t *testing.T) *application {
	tempalteCache, err := newTemplateCache()
	if err != nil {
		t.Fatal(err)
	}
	formDecoder := form.NewDecoder()
	sessionManager := scs.New()
	sessionManager.Lifetime = 12 * time.Hour
	sessionManager.Cookie.Secure = true
	return &application{
		logger:         slog.New(slog.DiscardHandler),
		snippets:       &mocks.SnippetModel{},
		users:          &mocks.UserModel{},
		templateCache:  tempalteCache,
		formDecoder:    formDecoder,
		sessionManager: sessionManager,
	}
}

type testServer struct {
	*httptest.Server
}

func newTestServer(t *testing.T, h http.Handler) *testServer {
	ts := httptest.NewTLSServer(h)
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	ts.Client().Jar = jar
	ts.Client().CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	return &testServer{ts}
}

func (ts *testServer) get(t *testing.T, urlPath string) (int, http.Header, string) {
	rs, err := ts.Client().Get(ts.URL + urlPath)
	if err != nil {
		t.Fatal(err)
	}
	defer rs.Body.Close()
	body, err := io.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}
	body = bytes.TrimSpace(body)
	return rs.StatusCode, rs.Header, string(body)
}

func (ts *testServer) postForm(t *testing.T, urlPath string, form url.Values) (int, http.Header, string) {
	req, err := http.NewRequest(http.MethodPost, ts.URL+urlPath, strings.NewReader(form.Encode()))
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Sec-Fetch-Site", "same-origin")

	res, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	body = bytes.TrimSpace(body)
	return res.StatusCode, res.Header, string(body)
}
