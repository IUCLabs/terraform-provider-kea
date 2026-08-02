package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// optionDataSchema is the shared option_data nested-list attribute used by both
// resources. Each element is a DHCP option (name/code/space + required data).
func optionDataSchema(desc string) schema.ListNestedAttribute {
	return schema.ListNestedAttribute{
		Description: desc,
		Optional:    true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"name":  schema.StringAttribute{Description: "Option name.", Optional: true},
				"code":  schema.Int64Attribute{Description: "Numeric option code.", Optional: true},
				"space": schema.StringAttribute{Description: "Option space.", Optional: true},
				"data":  schema.StringAttribute{Description: "Option value.", Required: true},
			},
		},
	}
}

// optionDataModel is the nested-attribute form of a DHCP option value.
type optionDataModel struct {
	Name  types.String `tfsdk:"name"`
	Code  types.Int64  `tfsdk:"code"`
	Space types.String `tfsdk:"space"`
	Data  types.String `tfsdk:"data"`
}

func optionDataObjectType() types.ObjectType {
	return types.ObjectType{AttrTypes: map[string]attr.Type{
		"name":  types.StringType,
		"code":  types.Int64Type,
		"space": types.StringType,
		"data":  types.StringType,
	}}
}

// stringOrNull maps an omitempty API string to a Terraform value: empty -> null.
func stringOrNull(s string) types.String {
	if s == "" {
		return types.StringNull()
	}
	return types.StringValue(s)
}

// int64OrNull maps an omitempty API int to a Terraform value: zero -> null.
func int64OrNull(i int) types.Int64 {
	if i == 0 {
		return types.Int64Null()
	}
	return types.Int64Value(int64(i))
}

// listToStrings reads a Terraform list attribute into a Go slice. A null or
// unknown list yields nil, so it round-trips against the API's omitempty fields.
func listToStrings(ctx context.Context, l types.List, diags *diag.Diagnostics) []string {
	if l.IsNull() || l.IsUnknown() {
		return nil
	}
	out := make([]string, 0, len(l.Elements()))
	diags.Append(l.ElementsAs(ctx, &out, false)...)
	return out
}

// stringsToList renders an API slice as a Terraform list. An empty/nil slice
// becomes a null list so it compares equal to an omitted config attribute.
func stringsToList(s []string) types.List {
	if len(s) == 0 {
		return types.ListNull(types.StringType)
	}
	elems := make([]attr.Value, 0, len(s))
	for _, v := range s {
		elems = append(elems, types.StringValue(v))
	}
	return types.ListValueMust(types.StringType, elems)
}

// optionDataToAPI reads the option_data nested list into wire structs.
func optionDataToAPI(ctx context.Context, l types.List, diags *diag.Diagnostics) []optionDataAPI {
	if l.IsNull() || l.IsUnknown() {
		return nil
	}
	var models []optionDataModel
	diags.Append(l.ElementsAs(ctx, &models, false)...)
	if diags.HasError() {
		return nil
	}
	out := make([]optionDataAPI, 0, len(models))
	for _, m := range models {
		out = append(out, optionDataAPI{
			Name:  m.Name.ValueString(),
			Code:  int(m.Code.ValueInt64()),
			Space: m.Space.ValueString(),
			Data:  m.Data.ValueString(),
		})
	}
	return out
}

// optionDataToList renders wire structs back into a nested list. Empty -> null.
func optionDataToList(in []optionDataAPI) types.List {
	t := optionDataObjectType()
	if len(in) == 0 {
		return types.ListNull(t)
	}
	elems := make([]attr.Value, 0, len(in))
	for _, o := range in {
		obj := types.ObjectValueMust(t.AttrTypes, map[string]attr.Value{
			"name":  stringOrNull(o.Name),
			"code":  int64OrNull(o.Code),
			"space": stringOrNull(o.Space),
			"data":  types.StringValue(o.Data),
		})
		elems = append(elems, obj)
	}
	return types.ListValueMust(t, elems)
}
