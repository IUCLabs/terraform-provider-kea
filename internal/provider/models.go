package provider

// API wire types, mirroring the KeaAPI JSON schema.
// They are duplicated here to keep the provider's dependency surface small.

type optionDataAPI struct {
	Name  string `json:"name,omitempty"`
	Code  int    `json:"code,omitempty"`
	Space string `json:"space,omitempty"`
	Data  string `json:"data"`
}

type subnetAPI struct {
	Name          string          `json:"name"`
	ClusterRef    string          `json:"cluster_ref,omitempty"`
	Subnet        string          `json:"subnet"`
	ID            int             `json:"id,omitempty"`
	Pools         []string        `json:"pools,omitempty"`
	Routers       []string        `json:"routers,omitempty"`
	DNSServers    []string        `json:"dns_servers,omitempty"`
	DomainName    string          `json:"domain_name,omitempty"`
	NextServer    string          `json:"next_server,omitempty"`
	BootFileName  string          `json:"boot_file_name,omitempty"`
	ClientClasses []string        `json:"client_classes,omitempty"`
	ValidLifetime int             `json:"valid_lifetime,omitempty"`
	OptionData    []optionDataAPI `json:"option_data,omitempty"`
}

type clientClassAPI struct {
	Name           string          `json:"name"`
	ClusterRef     string          `json:"cluster_ref,omitempty"`
	Test           string          `json:"test,omitempty"`
	ArchHex        []string        `json:"arch_hex,omitempty"`
	BootFileName   string          `json:"boot_file_name,omitempty"`
	NextServer     string          `json:"next_server,omitempty"`
	ServerHostname string          `json:"server_hostname,omitempty"`
	OptionData     []optionDataAPI `json:"option_data,omitempty"`
}
