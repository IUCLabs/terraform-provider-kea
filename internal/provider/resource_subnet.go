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
	_ resource.Resource                = &subnetResource{}
	_ resource.ResourceWithImportState = &subnetResource{}
)

type subnetResource struct {
	client *apiClient
}

type subnetResourceModel struct {
	Name          types.String `tfsdk:"name"`
	ClusterRef    types.String `tfsdk:"cluster_ref"`
	Subnet        types.String `tfsdk:"subnet"`
	SubnetID      types.Int64  `tfsdk:"subnet_id"`
	Pools         types.List   `tfsdk:"pools"`
	Routers       types.List   `tfsdk:"routers"`
	DNSServers    types.List   `tfsdk:"dns_servers"`
	DomainName    types.String `tfsdk:"domain_name"`
	NextServer    types.String `tfsdk:"next_server"`
	BootFileName  types.String `tfsdk:"boot_file_name"`
	ClientClasses types.List   `tfsdk:"client_classes"`
	ValidLifetime types.Int64  `tfsdk:"valid_lifetime"`
	OptionData    types.List   `tfsdk:"option_data"`
}

func NewSubnetResource() resource.Resource { return &subnetResource{} }

func (r *subnetResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subnet"
}

func (r *subnetResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A Kea DHCP subnet (KeaSubnet CR): its pool(s), routers, DNS, and PXE/boot options.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "Stable resource id (the KeaSubnet CR name).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"cluster_ref": schema.StringAttribute{
				Description: "Name of the KeaCluster this subnet belongs to.",
				Optional:    true,
			},
			"subnet": schema.StringAttribute{
				Description: "Subnet in CIDR notation, e.g. 192.0.2.0/24.",
				Required:    true,
			},
			"subnet_id": schema.Int64Attribute{
				Description: "Numeric Kea subnet id. Omit to let Kea assign one.",
				Optional:    true,
			},
			"pools": schema.ListAttribute{
				Description: "Address pools, e.g. [\"192.0.2.100-192.0.2.200\"].",
				ElementType: types.StringType,
				Optional:    true,
			},
			"routers": schema.ListAttribute{
				Description: "Default gateway addresses (DHCP option routers).",
				ElementType: types.StringType,
				Optional:    true,
			},
			"dns_servers": schema.ListAttribute{
				Description: "DNS server addresses (domain-name-servers option).",
				ElementType: types.StringType,
				Optional:    true,
			},
			"domain_name": schema.StringAttribute{
				Description: "Domain name handed to clients.",
				Optional:    true,
			},
			"next_server": schema.StringAttribute{
				Description: "PXE next-server (siaddr / TFTP server) address.",
				Optional:    true,
			},
			"boot_file_name": schema.StringAttribute{
				Description: "PXE boot file name.",
				Optional:    true,
			},
			"client_classes": schema.ListAttribute{
				Description: "Client classes required to receive a lease from this subnet.",
				ElementType: types.StringType,
				Optional:    true,
			},
			"valid_lifetime": schema.Int64Attribute{
				Description: "Lease valid lifetime in seconds.",
				Optional:    true,
			},
			"option_data": optionDataSchema("Extra DHCP options applied to this subnet."),
		},
	}
}

func (r *subnetResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *subnetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan subnetResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	r.upsert(ctx, plan, &resp.Diagnostics, &resp.State)
}

func (r *subnetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan subnetResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	r.upsert(ctx, plan, &resp.Diagnostics, &resp.State)
}

func (r *subnetResource) upsert(ctx context.Context, plan subnetResourceModel, diags *diag.Diagnostics, state *tfsdk.State) {
	body := subnetAPI{
		Name:          plan.Name.ValueString(),
		ClusterRef:    plan.ClusterRef.ValueString(),
		Subnet:        plan.Subnet.ValueString(),
		ID:            int(plan.SubnetID.ValueInt64()),
		Pools:         listToStrings(ctx, plan.Pools, diags),
		Routers:       listToStrings(ctx, plan.Routers, diags),
		DNSServers:    listToStrings(ctx, plan.DNSServers, diags),
		DomainName:    plan.DomainName.ValueString(),
		NextServer:    plan.NextServer.ValueString(),
		BootFileName:  plan.BootFileName.ValueString(),
		ClientClasses: listToStrings(ctx, plan.ClientClasses, diags),
		ValidLifetime: int(plan.ValidLifetime.ValueInt64()),
		OptionData:    optionDataToAPI(ctx, plan.OptionData, diags),
	}
	if diags.HasError() {
		return
	}
	var out subnetAPI
	if err := r.client.put(ctx, "/api/v1/subnets/"+pathEscape(body.Name), body, &out); err != nil {
		diags.AddError("upsert subnet failed", err.Error())
		return
	}
	diags.Append(state.Set(ctx, subnetAPIToModel(out))...)
}

func (r *subnetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state subnetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var out subnetAPI
	if err := r.client.get(ctx, "/api/v1/subnets/"+pathEscape(state.Name.ValueString()), &out); err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("read subnet failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, subnetAPIToModel(out))...)
}

func (r *subnetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state subnetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.del(ctx, "/api/v1/subnets/"+pathEscape(state.Name.ValueString())); err != nil {
		resp.Diagnostics.AddError("delete subnet failed", err.Error())
		return
	}
}

func (r *subnetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

func subnetAPIToModel(api subnetAPI) subnetResourceModel {
	return subnetResourceModel{
		Name:          types.StringValue(api.Name),
		ClusterRef:    stringOrNull(api.ClusterRef),
		Subnet:        types.StringValue(api.Subnet),
		SubnetID:      int64OrNull(api.ID),
		Pools:         stringsToList(api.Pools),
		Routers:       stringsToList(api.Routers),
		DNSServers:    stringsToList(api.DNSServers),
		DomainName:    stringOrNull(api.DomainName),
		NextServer:    stringOrNull(api.NextServer),
		BootFileName:  stringOrNull(api.BootFileName),
		ClientClasses: stringsToList(api.ClientClasses),
		ValidLifetime: int64OrNull(api.ValidLifetime),
		OptionData:    optionDataToList(api.OptionData),
	}
}
