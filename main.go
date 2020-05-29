package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/plugin"
	"github.com/mathieugravil/terraform-provider-azuredevops/azuredevops"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: azuredevops.Provider})
}
