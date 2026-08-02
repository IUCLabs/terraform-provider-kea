package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestStringOrNull(t *testing.T) {
	if !stringOrNull("").IsNull() {
		t.Error("empty string should map to null")
	}
	if v := stringOrNull("x"); v.ValueString() != "x" {
		t.Errorf("got %q, want x", v.ValueString())
	}
}

func TestInt64OrNull(t *testing.T) {
	if !int64OrNull(0).IsNull() {
		t.Error("zero should map to null")
	}
	if v := int64OrNull(42); v.ValueInt64() != 42 {
		t.Errorf("got %d, want 42", v.ValueInt64())
	}
}

func TestStringsListRoundTrip(t *testing.T) {
	if !stringsToList(nil).IsNull() {
		t.Error("nil slice should map to null list")
	}
	if !stringsToList([]string{}).IsNull() {
		t.Error("empty slice should map to null list")
	}
	l := stringsToList([]string{"a", "b"})
	var diags diag.Diagnostics
	got := listToStrings(context.Background(), l, &diags)
	if diags.HasError() {
		t.Fatalf("diags: %v", diags)
	}
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Errorf("round trip = %v", got)
	}
	// a null list reads back as a nil slice, matching the API's omitempty fields
	if listToStrings(context.Background(), types.ListNull(types.StringType), &diags) != nil {
		t.Error("null list should read back as nil slice")
	}
}

func TestOptionDataRoundTrip(t *testing.T) {
	in := []optionDataAPI{
		{Name: "tftp-server-name", Code: 66, Space: "dhcp4", Data: "10.0.0.1"},
		{Data: "just-data"},
	}
	list := optionDataToList(in)
	var diags diag.Diagnostics
	got := optionDataToAPI(context.Background(), list, &diags)
	if diags.HasError() {
		t.Fatalf("diags: %v", diags)
	}
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if got[0] != in[0] {
		t.Errorf("elem0 = %+v, want %+v", got[0], in[0])
	}
	// zero code/empty name+space round-trip cleanly (null <-> omitempty zero)
	if got[1] != in[1] {
		t.Errorf("elem1 = %+v, want %+v", got[1], in[1])
	}
	if !optionDataToList(nil).IsNull() {
		t.Error("empty option_data should map to null list")
	}
}

func TestSubnetAPIToModel(t *testing.T) {
	api := subnetAPI{
		Name:          "pxe",
		Subnet:        "10.0.0.0/24",
		ID:            0, // omitted -> null
		Pools:         []string{"10.0.0.100-10.0.0.200"},
		Routers:       []string{"10.0.0.1"},
		ClientClasses: []string{"pxeclients"},
		ValidLifetime: 3600,
		OptionData:    []optionDataAPI{{Code: 67, Data: "pxelinux.0"}},
	}
	m := subnetAPIToModel(api)
	if m.Name.ValueString() != "pxe" || m.Subnet.ValueString() != "10.0.0.0/24" {
		t.Errorf("identity fields wrong: %+v", m)
	}
	if !m.SubnetID.IsNull() {
		t.Error("id 0 should be null")
	}
	if m.ValidLifetime.ValueInt64() != 3600 {
		t.Errorf("valid_lifetime = %d", m.ValidLifetime.ValueInt64())
	}
	if !m.DNSServers.IsNull() {
		t.Error("unset dns_servers should be null")
	}
	if m.Pools.IsNull() {
		t.Error("pools should be populated")
	}
}

func TestClientClassAPIToModel(t *testing.T) {
	api := clientClassAPI{
		Name:         "pxeclients",
		Test:         "substring(option[60].hex,0,9) == 'PXEClient'",
		ArchHex:      []string{"0007", "0009"},
		BootFileName: "ipxe.efi",
	}
	m := clientClassAPIToModel(api)
	if m.Name.ValueString() != "pxeclients" || m.Test.ValueString() == "" {
		t.Errorf("fields wrong: %+v", m)
	}
	if m.ArchHex.IsNull() {
		t.Error("arch_hex should be populated")
	}
	if !m.NextServer.IsNull() {
		t.Error("unset next_server should be null")
	}
}
