package fcm

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/valyala/fasthttp"
)

func TestSend(t *testing.T) {
	t.Run("send=success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			if req.Header.Get("Authorization") != "key=test" {
				t.Fatalf("expected: key=test\ngot: %s", req.Header.Get("Authorization"))
			}
			rw.WriteHeader(http.StatusOK)
			rw.Header().Set("Content-Type", "application/json")
			fmt.Fprint(rw, `{
				"success": 1,
				"failure":0,
				"results": [{
					"message_id":"q1w2e3r4",
					"registration_id": "t5y6u7i8o9",
					"error": ""
				}]
			}`)
		}))
		defer server.Close()

		client, err := NewClient("test", WithEndpoint(server.URL), WithTimeout(10*time.Second))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		resp, _, err := client.Send(&Message{
			To: "test",
			Data: map[string]interface{}{
				"foo": "bar",
			},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Success != 1 {
			t.Fatalf("expected 1 successes\ngot: %d successes", resp.Success)
		}
		if resp.Failure != 0 {
			t.Fatalf("expected 0 failures\ngot: %d failures", resp.Failure)
		}
	})

	t.Run("send=failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			if req.Header.Get("Authorization") != "key=test" {
				t.Fatalf("expected: key=test\ngot: %s", req.Header.Get("Authorization"))
			}
			rw.WriteHeader(http.StatusBadRequest)
		}))
		defer server.Close()

		client, err := NewClient("test", WithEndpoint(server.URL))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		resp, _, err := client.Send(&Message{
			To: "test",
			Data: map[string]interface{}{
				"foo": "bar",
			},
		})
		if err == nil {
			t.Fatal("expected error but got nil")
		}
		if resp != nil {
			t.Fatalf("expected nil response\ngot: %v response", resp)
		}
	})

	t.Run("send=invalid_token", func(t *testing.T) {
		_, err := NewClient("test", WithEndpoint(""))
		if err == nil {
			t.Fatal("expected error but got nil")
		}
	})

	t.Run("send=invalid_message", func(t *testing.T) {
		c, err := NewClient("test", WithEndpoint("test"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		_, _, err = c.Send(&Message{})
		if err == nil {
			t.Fatal("expected error but go nil")
		}
	})

	t.Run("send=invalid-response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			if req.Header.Get("Authorization") != "key=test" {
				t.Fatalf("expected: key=test\ngot: %s", req.Header.Get("Authorization"))
			}
			rw.WriteHeader(http.StatusOK)
			rw.Header().Set("Content-Type", "application/json")
			fmt.Fprint(rw, `{
				"success": 1,
				"failure":0,
				"results": {
					"message_id":"q1w2e3r4",
					"registration_id": "t5y6u7i8o9",
					"error": ""
				}
			}`)
		}))
		defer server.Close()

		client, err := NewClient("test",
			WithEndpoint(server.URL),
			WithHTTPClient(&fasthttp.Client{}),
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		resp, _, err := client.Send(&Message{
			To: "test",
			Data: map[string]interface{}{
				"foo": "bar",
			},
		})
		if err == nil {
			t.Fatal("expected error but go nil")
		}

		if resp != nil {
			t.Fatalf("expected nil\ngot response: %v", resp)
		}
	})
}

func TestSendWithRetry(t *testing.T) {
	t.Run("send_with_retry=success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			if req.Header.Get("Authorization") != "key=test" {
				t.Fatalf("expected: key=test\ngot: %s", req.Header.Get("Authorization"))
			}
			rw.WriteHeader(http.StatusOK)
			rw.Header().Set("Content-Type", "application/json")
			fmt.Fprint(rw, `{
				"success": 1,
				"failure":0,
				"results": [{
					"message_id":"q1w2e3r4",
					"registration_id": "t5y6u7i8o9",
					"error": ""
				}]
			}`)
		}))
		defer server.Close()

		client, err := NewClient("test",
			WithEndpoint(server.URL),
			WithHTTPClient(&fasthttp.Client{}),
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		resp, _, err := client.SendWithRetry(&Message{
			To: "test",
			Data: map[string]interface{}{
				"foo": "bar",
			},
		}, 3)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Success != 1 {
			t.Fatalf("expected 1 successes\ngot: %d successes", resp.Success)
		}
		if resp.Failure != 0 {
			t.Fatalf("expected 0 failures\ngot: %d failures", resp.Failure)
		}
	})

	t.Run("send_with_retry=failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			if req.Header.Get("Authorization") != "key=test" {
				t.Fatalf("expected: key=test\ngot: %s", req.Header.Get("Authorization"))
			}
			rw.WriteHeader(http.StatusBadRequest)
		}))
		defer server.Close()

		client, err := NewClient("test",
			WithEndpoint(server.URL),
			WithHTTPClient(&fasthttp.Client{}),
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		resp, _, err := client.SendWithRetry(&Message{
			To: "test",
			Data: map[string]interface{}{
				"foo": "bar",
			},
		}, 2)

		if err == nil {
			t.Fatal("expected error\ngot nil")
		}
		if resp != nil {
			t.Fatalf("expected nil response\ngot: %v response", resp)
		}
	})

	t.Run("send_with_retry=success_retry", func(t *testing.T) {
		var attempts int
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			attempts++
			if req.Header.Get("Authorization") != "key=test" {
				t.Fatalf("expected: key=test\ngot: %s", req.Header.Get("Authorization"))
			}
			if attempts < 3 {
				rw.WriteHeader(http.StatusInternalServerError)
			} else {
				rw.WriteHeader(http.StatusOK)
			}
			rw.Header().Set("Content-Type", "application/json")

			fmt.Fprint(rw, `{
				"success": 1,
				"failure":0,
				"results": [{
					"message_id":"q1w2e3r4",
					"registration_id": "t5y6u7i8o9",
					"error": ""
				}]
			}`)
		}))
		defer server.Close()

		client, err := NewClient("test",
			WithEndpoint(server.URL),
			WithHTTPClient(&fasthttp.Client{}),
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		resp, _, err := client.SendWithRetry(&Message{
			To: "test",
			Data: map[string]interface{}{
				"foo": "bar",
			},
		}, 4)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if attempts != 3 {
			t.Fatalf("expected 3 attempts\ngot: %d attempts", attempts)
		}
		if resp.Success != 1 {
			t.Fatalf("expected 1 successes\ngot: %d successes", resp.Success)
		}
		if resp.Failure != 0 {
			t.Fatalf("expected 0 failures\ngot: %d failures", resp.Failure)
		}
	})

	t.Run("send_with_retry=failure_retry", func(t *testing.T) {
		client, err := NewClient("test",
			WithEndpoint("127.0.0.1:80"),
			WithHTTPClient(&fasthttp.Client{

				ReadTimeout: time.Nanosecond,
			}),
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		resp, _, err := client.SendWithRetry(&Message{
			To: "test",
			Data: map[string]interface{}{
				"foo": "bar",
			},
		}, 3)

		if err == nil {
			t.Fatal("expected error\ngot nil")
		}
		if resp != nil {
			t.Fatalf("expected nil response\ngot: %v response", resp)
		}
	})
}
