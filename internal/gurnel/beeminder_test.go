package gurnel

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"log"
)

func TestBeeminderPostServerErrors(t *testing.T) {
	tests := []struct {
		returnCode int
		returnBody string
	}{
		{
			200,
			"",
		},
		{
			404,
			"server not found",
		},
		{
			500,
			"server on fire",
		},
		{
			503,
			"",
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("server return HTTP code %d",
			tt.returnCode), func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.returnCode)
				w.Write([]byte(tt.returnBody))
			})
			server := httptest.NewTLSServer(handler)
			defer server.Close()

			client := beeminderClient{
				Token:     []byte("test"),
				User:      "test",
				c:         *server.Client(),
				serverURL: server.URL,
			}

			result := make(chan error)
			go func() {
				result <- client.postDatapoint("foo", 1)
			}()
			err := <-result

			if tt.returnCode == 200 {
				if err != nil {
					t.Fatalf("expected no error. got %q", err)
				}
			} else {
				if c := strconv.Itoa(tt.returnCode); !strings.Contains(err.Error(), c) {
					t.Fatalf("wrong error code. expected an error containing %q. got %q",
						c, err)
				}
				if !strings.Contains(err.Error(), tt.returnBody) {
					t.Fatalf("wrong error message. expected an error containing %q. got %q",
						tt.returnBody, err)
				}
				if expectedMsg := "no further info"; tt.returnBody == "" &&
					!strings.Contains(err.Error(), expectedMsg) {
					t.Fatalf("wrong error message. expected an error containing %q. got %q",
						expectedMsg, err)
				}
			}
		})
	}
}

func TestBeeminderPostParameters(t *testing.T) {
	tests := []struct {
		goal  string
		count int
		valid bool
	}{
		{
			"testGoal",
			10,
			true,
		},
		{
			"testGoal",
			0,
			true,
		},
		{
			"testGoal",
			-10,
			false,
		},
		{
			"",
			10,
			false,
		},
		{
			"",
			-10,
			false,
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("with goal %q and count %d",
			tt.goal, tt.count), func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if !tt.valid {
					t.Error("expected no API call with invalid parameters passed")
				}
				if v := r.FormValue("value"); v != strconv.Itoa(tt.count) {
					t.Errorf("wrong value. expected %d. got %s", tt.count, v)
				}
				if !strings.Contains(r.RequestURI, tt.goal) {
					t.Errorf("goal not in URL. expected a string containing %q. got %s",
						tt.goal, r.RequestURI)
				}
				w.Write([]byte("OK"))
			})
			server := httptest.NewTLSServer(handler)
			defer server.Close()

			client := beeminderClient{
				Token:     []byte("test"),
				User:      "test",
				c:         *server.Client(),
				serverURL: server.URL,
			}

			result := make(chan error)
			go func() {
				result <- client.postDatapoint(tt.goal, tt.count)
			}()
			err := <-result
			if !tt.valid {
				if err == nil {
					t.Fatal("expected an error with invalid parameters passed")
				}
			}
		})
	}
}

func TestNewBeeminderClient(t *testing.T) {
	tests := []struct {
		user  string
		token string
		valid bool
	}{
		{
			"testUser",
			"testToken",
			true,
		},
		{
			"testUser",
			" testToken ",
			true,
		},
		{
			"",
			" testToken ",
			false,
		},
		{
			"testUser",
			"",
			false,
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("with user %q and token %q",
			tt.user, tt.token), func(t *testing.T) {
			client, err := newBeeminderClient(tt.user, []byte(tt.token))
			if !tt.valid {
				if err == nil {
					t.Fatal("expected an error with invalid parameters passed")
				}
				if client != nil {
					t.Fatal("expected a nil client with invalid parameters passed")
				}
				return
			}
			if tt.user != client.User {
				t.Fatalf("wrong user. expected %q. got %q", tt.user, client.User)
			}
			if s := strings.TrimSpace(tt.token); s != string(client.Token) {
				t.Fatalf("wrong token. expected %q. got %q", s, client.Token)
			}
		})
	}
}

type testTransport struct {
	err error
}

func (t *testTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(&bytes.Buffer{}),
	}, t.err
}

func TestBeeminderClient(t *testing.T) {
	tests := []struct {
		serverURL   string
		transport   http.RoundTripper
		expectedErr string
	}{
		{
			"http://test.com",
			&testTransport{},
			"",
		},
		{
			"http:// not a url",
			&testTransport{},
			"URL error",
		},
		{
			"http://test.com",
			&testTransport{err: fmt.Errorf("test error")},
			"making request",
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("with expected error %q",
			tt.expectedErr), func(t *testing.T) {
			client := beeminderClient{
				serverURL: tt.serverURL,
				c:         http.Client{Transport: tt.transport},
			}

			err := client.postDatapoint("test", 10)
			if tt.expectedErr == "" {
				if err != nil {
					t.Fatalf("expected no error. got %q", err)
				}
			} else {
				if err == nil || !strings.Contains(err.Error(), tt.expectedErr) {
					t.Fatalf("wrong error. expected %q. got %v", tt.expectedErr, err)
				}
			}
		})
	}
}
