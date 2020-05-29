package azuredevops

import (
"github.com/hashicorp/terraform-plugin-sdk/terraform"
"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

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
	config := Config{
		Token:      d.Get("token").(string),
		BaseURL:    d.Get("base_url").(string),
	}
	return config.Client()
}

