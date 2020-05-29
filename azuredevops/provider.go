package azuredevops

import (
//	"github.com/hashicorp/terraform/helper/schema"
//"github.com/hashicorp/terraform/terraform"
"github.com/hashicorp/terraform-plugin-sdk/terraform"
"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
"context"
	"github.com/microsoft/azure-devops-go-api/azuredevops"
	"github.com/microsoft/azure-devops-go-api/azuredevops/core"
)

func Provider() terraform.ResourceProvider {
    return &schema.Provider{
       Schema: map[string]*schema.Schema{
			"token": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("AZUREDEVOPS_TOKEN", nil),
				Description: descriptions["token"],
			},
			"base_url": {
				Type:         schema.TypeString,
				Optional:     true,
				DefaultFunc:  schema.EnvDefaultFunc("AZUREDEVOPS_BASE_URL", ""),
			},
		},

	//	DataSourcesMap: map[string]*schema.Resource{
//			"azuredevops_project": dataSourceAzuredevopsProject(),
//		},
		ResourcesMap: map[string]*schema.Resource{
			"azuredevops_project":                    resourceAzuredevopsProject(),
			
		},

		ConfigureFunc: providerConfigure,
    }
}

var descriptions map[string]string

func init() {
	descriptions = map[string]string{
		"token": "The personnal access  token used to connect to AzureDevops.",
		"base_url": "The AzureDevops Base API URL",
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	address := d.Get("base_url").(string)
	token := d.Get("token").(string)
	connection := azuredevops.NewPatConnection(address, token)
	ctx := context.Background()
	myclient, err := core.NewClient(ctx, connection)
	return myclient, err
}

