package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &clientClassResource{}
	_ resource.ResourceWithImportState = &clientClassResource{}
)

type clientClassResource struct {
	client *apiClient
}

type clientClassResourceModel struct {
	Name           types.String `tfsdk:"name"`
	ClusterRef     types.String `tfsdk:"cluster_ref"`
	Test           types.String `tfsdk:"test"`
	ArchHex        types.List   `tfsdk:"arch_hex"`
	BootFileName   types.String `tfsdk:"boot_file_name"`
	NextServer     types.String `tfsdk:"next_server"`
	ServerHostname types.String `tfsdk:"server_hostname"`
	OptionData     types.List   `tfsdk:"option_data"`
}

func NewClientClassResource() resource.Resource { return &clientClassResource{} }

func (r *clientClassResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_clientclass"
}

func (r *clientClassResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A Kea PXE client class (KeaClientClass CR) matching clients by test expression or architecture, with boot options.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "Stable resource id (the KeaClientClass CR name).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"cluster_ref": schema.StringAttribute{
				Description: "Name of the KeaCluster this class belongs to.",
				Optional:    true,
			},
			"test": schema.StringAttribute{
				Description: "Kea test expression selecting matching clients. One of test or arch_hex is required.",
				Optional:    true,
			},
			"arch_hex": schema.ListAttribute{
				Description: "Client architecture types (hex) to match. One of test or arch_hex is required.",
				ElementType: types.StringType,
				Optional:    true,
			},
			"boot_file_name": schema.StringAttribute{
				Description: "PXE boot file name handed to matching clients.",
				Optional:    true,
			},
			"next_server": schema.StringAttribute{
				Description: "PXE next-server (siaddr / TFTP server) address.",
				Optional:    true,
			},
			"server_hostname": schema.StringAttribute{
				Description: "PXE server hostname (sname).",
				Optional:    true,
			},
			"option_data": optionDataSchema("Extra DHCP options applied to matching clients."),
		},
	}
}

func (r *clientClassResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*apiClient)
	if !ok {
		resp.Diagnostics.AddError("unexpected provider data type", fmt.Sprintf("got %T", req.ProviderData))
		return
	}
	r.client = client
}

func (r *clientClassResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan clientClassResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	r.upsert(ctx, plan, &resp.Diagnostics, &resp.State)
}

func (r *clientClassResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan clientClassResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	r.upsert(ctx, plan, &resp.Diagnostics, &resp.State)
}

func (r *clientClassResource) upsert(ctx context.Context, plan clientClassResourceModel, diags *diag.Diagnostics, state *tfsdk.State) {
	body := clientClassAPI{
		Name:           plan.Name.ValueString(),
		ClusterRef:     plan.ClusterRef.ValueString(),
		Test:           plan.Test.ValueString(),
		ArchHex:        listToStrings(ctx, plan.ArchHex, diags),
		BootFileName:   plan.BootFileName.ValueString(),
		NextServer:     plan.NextServer.ValueString(),
		ServerHostname: plan.ServerHostname.ValueString(),
		OptionData:     optionDataToAPI(ctx, plan.OptionData, diags),
	}
	if diags.HasError() {
		return
	}
	var out clientClassAPI
	if err := r.client.put(ctx, "/api/v1/clientclasses/"+pathEscape(body.Name), body, &out); err != nil {
		diags.AddError("upsert clientclass failed", err.Error())
		return
	}
	diags.Append(state.Set(ctx, clientClassAPIToModel(out))...)
}

func (r *clientClassResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state clientClassResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var out clientClassAPI
	if err := r.client.get(ctx, "/api/v1/clientclasses/"+pathEscape(state.Name.ValueString()), &out); err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("read clientclass failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, clientClassAPIToModel(out))...)
}

func (r *clientClassResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state clientClassResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.del(ctx, "/api/v1/clientclasses/"+pathEscape(state.Name.ValueString())); err != nil {
		resp.Diagnostics.AddError("delete clientclass failed", err.Error())
		return
	}
}

func (r *clientClassResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

func clientClassAPIToModel(api clientClassAPI) clientClassResourceModel {
	return clientClassResourceModel{
		Name:           types.StringValue(api.Name),
		ClusterRef:     stringOrNull(api.ClusterRef),
		Test:           stringOrNull(api.Test),
		ArchHex:        stringsToList(api.ArchHex),
		BootFileName:   stringOrNull(api.BootFileName),
		NextServer:     stringOrNull(api.NextServer),
		ServerHostname: stringOrNull(api.ServerHostname),
		OptionData:     optionDataToList(api.OptionData),
	}
}
