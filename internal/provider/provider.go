package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &keaProvider{}

type keaProvider struct {
	version string
}

type keaProviderModel struct {
	Endpoint types.String `tfsdk:"endpoint"`
	Token    types.String `tfsdk:"token"`
}

// New returns the provider constructor.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &keaProvider{version: version}
	}
}

func (p *keaProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "kea"
	resp.Version = p.version
}

func (p *keaProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manage Kea DHCP subnets and PXE client classes through a KeaAPI HTTP service.",
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				Description: "The KeaAPI server base URL (e.g. https://keaapi.example.com).",
				Required:    true,
			},
			"token": schema.StringAttribute{
				Description: "Bearer token for the KeaAPI. Both reads and writes require it. Defaults to the KEA_API_TOKEN environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
		},
	}
}

func (p *keaProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config keaProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	token := config.Token.ValueString()
	if token == "" {
		token = os.Getenv("KEA_API_TOKEN")
	}

	client := newAPIClient(config.Endpoint.ValueString(), token)
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *keaProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewSubnetResource,
		NewClientClassResource,
	}
}

func (p *keaProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return nil
}
