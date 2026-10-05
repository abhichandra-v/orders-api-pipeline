package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealthz(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
}

func TestCreateAndListOrder(t *testing.T) {
	h := routes()

	// Create a valid order.
	body := strings.NewReader(`{"item":"widget","qty":3}`)
	req := httptest.NewRequest(http.MethodPost, "/orders", body)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create: want 201, got %d (%s)", rec.Code, rec.Body.String())
	}

	// List should now contain it.
	req = httptest.NewRequest(http.MethodGet, "/orders", nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list: want 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "widget") {
		t.Fatalf("list did not contain created order: %s", rec.Body.String())
	}
}

func TestCreateOrderRejectsBadInput(t *testing.T) {
	cases := []string{`{"item":"","qty":1}`, `{"item":"x","qty":0}`, `not json`}
	for _, c := range cases {
		req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(c))
		rec := httptest.NewRecorder()
		routes().ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("input %q: want 400, got %d", c, rec.Code)
		}
	}
}
