package requestid

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
)

var validRequestID = regexp.MustCompile(`([a-f\d]{8}(-[a-f\d]{4}){3}-[a-f\d]{12}?)`)

func TestRequestID(t *testing.T) {
	rids := []string{"", "remoteID_12345"}

	fn := func(w http.ResponseWriter, r *http.Request) {
		rid, _ := FromContext(r.Context())
		if rid != r.Header.Get(HeaderName) {
			t.Errorf("requestid in headers and context do not match: %q, %q", rid, r.Header.Get(HeaderName))
		}
		fmt.Fprintf(w, "%s", rid)
	}

	h := RequestIDHandler(http.HandlerFunc(fn))

	for _, rid := range rids {
		w := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "http://example.com/foo", nil)
		if err != nil {
			t.Fatal(err)
		}

		if rid != "" {
			// Pre-set request ID
			req.Header.Set(HeaderName, rid)
		}

		h.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("request '%s': %d != %d", rid, w.Code, http.StatusOK)
		}

		body := strings.TrimSpace(w.Body.String())

		if rid != "" {
			if body != rid {
				t.Errorf("request '%s': %s != %[1]s", rid, body)
			}
			if w.Header().Get(HeaderName) != rid {
				t.Errorf("request header is not set in response headers")
			}
		} else {
			if !validRequestID.MatchString(body) {
				t.Errorf("request '%s': %s is not valid format", rid, body)
			}
			if w.Header().Get(HeaderName) != body {
				t.Errorf("request header is not set in response headers")
			}
		}
	}
}

func TestRequestHandlerMiddleware(t *testing.T) {
	expected := "testrequest"

	fn := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%s,%s", r.Header.Get(HeaderName), expected)
	}

	mux := http.NewServeMux()
	mux.Handle("/", RequestIDHandler(http.HandlerFunc(fn)))

	server := httptest.NewServer(mux)
	defer server.Close()

	res, err := http.Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}

	body, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()

	if err != nil {
		t.Fatal(err)
	}

	results := strings.Split(string(body), ",")
	if len(results) != 2 {
		t.Fatalf("Invalid results: %v", results)
	}
	if !validRequestID.MatchString(results[0]) {
		t.Errorf("%s is not valid request id format", results[0])
	}
	if results[1] != expected {
		t.Errorf("%v != %v", results[1], expected)
	}
	if res.Header.Get(HeaderName) != results[0] {
		t.Errorf("request header is not set in response headers")
	}
}
