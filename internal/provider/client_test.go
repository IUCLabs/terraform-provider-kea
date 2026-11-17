package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientSendsBearerToken(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		_, _ = w.Write([]byte(`{"name":"pxe","subnet":"192.0.2.0/24"}`))
	}))
	defer srv.Close()

	c := newAPIClient(srv.URL, "tok")
	var out subnetAPI
	if err := c.put(context.Background(), "/api/v1/subnets/pxe", subnetAPI{Name: "pxe", Subnet: "192.0.2.0/24"}, &out); err != nil {
		t.Fatal(err)
	}
	if gotAuth != "Bearer tok" {
		t.Errorf("auth = %q, want Bearer tok", gotAuth)
	}
}

func TestClientNoTokenNoHeader(t *testing.T) {
	var hadAuth bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, hadAuth = r.Header["Authorization"]
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c := newAPIClient(srv.URL, "")
	if err := c.get(context.Background(), "/api/v1/subnets/x", &subnetAPI{}); err != nil {
		t.Fatal(err)
	}
	if hadAuth {
		t.Error("no Authorization header expected when token empty")
	}
}

func TestClientNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"not found"}`))
	}))
	defer srv.Close()

	err := newAPIClient(srv.URL, "").get(context.Background(), "/api/v1/subnets/ghost", &subnetAPI{})
	if !isNotFound(err) {
		t.Errorf("err = %v, want notFound", err)
	}
}

func TestClientAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"subnet is required"}`))
	}))
	defer srv.Close()

	err := newAPIClient(srv.URL, "t").put(context.Background(), "/api/v1/subnets/h", subnetAPI{}, &subnetAPI{})
	if err == nil {
		t.Fatal("expected api error")
	}
	if isNotFound(err) {
		t.Errorf("400 should not be a notFound error, got %v", err)
	}
}

func TestClientDeleteNoContent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %s, want DELETE", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	if err := newAPIClient(srv.URL, "t").del(context.Background(), "/api/v1/clientclasses/pxe"); err != nil {
		t.Fatal(err)
	}
}
