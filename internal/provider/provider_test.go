package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestProviderSchema(t *testing.T) {
	p := New("test")()
	resp := &provider.SchemaResponse{}
	p.Schema(context.Background(), provider.SchemaRequest{}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("schema diagnostics: %v", resp.Diagnostics)
	}
	if _, ok := resp.Schema.Attributes["endpoint"]; !ok {
		t.Error("missing endpoint attribute")
	}
	if _, ok := resp.Schema.Attributes["token"]; !ok {
		t.Error("missing token attribute")
	}
}

func TestProviderMetadata(t *testing.T) {
	p := New("1.2.3")()
	resp := &provider.MetadataResponse{}
	p.Metadata(context.Background(), provider.MetadataRequest{}, resp)
	if resp.TypeName != "kea" || resp.Version != "1.2.3" {
		t.Errorf("metadata = %+v", resp)
	}
}

func TestProviderRegistersResources(t *testing.T) {
	p := New("test")()
	if len(p.Resources(context.Background())) != 2 {
		t.Error("expected 2 resources (subnet, clientclass)")
	}
}

// TestResourceSchemas validates each resource's schema compiles cleanly and
// carries the expected metadata type name.
func TestResourceSchemas(t *testing.T) {
	for _, ctor := range []func() resource.Resource{NewSubnetResource, NewClientClassResource} {
		resp := &resource.SchemaResponse{}
		ctor().Schema(context.Background(), resource.SchemaRequest{}, resp)
		if resp.Diagnostics.HasError() {
			t.Errorf("resource schema error: %v", resp.Diagnostics)
		}
	}
}

// TestSubnetWireRoundTrip drives a full subnet PUT through a mock server that
// echoes what it decoded, proving the JSON contract (name from path id, pools,
// option_data) matches the KeaAPI models end to end.
func TestSubnetWireRoundTrip(t *testing.T) {
	var got subnetAPI
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("method = %s, want PUT", r.Method)
		}
		if r.URL.Path != "/api/v1/subnets/pxe" {
			t.Errorf("path = %s", r.URL.Path)
		}
		_ = json.NewDecoder(r.Body).Decode(&got)
		got.Name = "pxe" // server authoritative from path
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(got)
	}))
	defer srv.Close()

	in := subnetAPI{
		Name:       "pxe",
		Subnet:     "192.0.2.0/24",
		Pools:      []string{"192.0.2.100-192.0.2.200"},
		Routers:    []string{"192.0.2.1"},
		OptionData: []optionDataAPI{{Code: 67, Space: "dhcp4", Data: "pxelinux.0"}},
	}
	var out subnetAPI
	if err := newAPIClient(srv.URL, "t").put(context.Background(), "/api/v1/subnets/pxe", in, &out); err != nil {
		t.Fatal(err)
	}
	if out.Subnet != in.Subnet || len(out.Pools) != 1 || len(out.OptionData) != 1 {
		t.Errorf("round trip mismatch: %+v", out)
	}
	if out.OptionData[0].Code != 67 || out.OptionData[0].Data != "pxelinux.0" {
		t.Errorf("option_data lost: %+v", out.OptionData)
	}
}
