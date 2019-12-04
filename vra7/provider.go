package vra7

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"github.com/terraform-providers/terraform-provider-vra7/sdk"
)

//Provider - This function initializes the provider schema
//also the config function and resource mapping
func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema:        providerSchema(),
		ConfigureFunc: providerConfig,
		ResourcesMap: map[string]*schema.Resource{
			"vra7_deployment": resourceVra7Deployment(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"vra7_deployment": dataSourceVra7Deployment(),
		},
	}
}

//providerSchema - To set provider fields
func providerSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"username": {
			Type:        schema.TypeString,
			Required:    true,
			DefaultFunc: schema.EnvDefaultFunc("VRA7_USERNAME", nil),
			Description: "Tenant administrator username.",
		},
		"password": {
			Type:        schema.TypeString,
			Required:    true,
			DefaultFunc: schema.EnvDefaultFunc("VRA7_PASSWORD", nil),
			Description: "Tenant administrator password.",
		},
		"tenant": {
			Type:        schema.TypeString,
			Required:    true,
			DefaultFunc: schema.EnvDefaultFunc("VRA7_TENANT", nil),
			Description: "Specifies the tenant URL token determined by the system administrator" +
				"when creating the tenant, for example, support.",
		},
		"host": {
			Type:        schema.TypeString,
			Required:    true,
			DefaultFunc: schema.EnvDefaultFunc("VRA7_HOST", nil),
			Description: "host name.domain name of the vRealize Automation server, " +
				"for example, mycompany.mktg.mydomain.com.",
		},
		"insecure": {
			Type:        schema.TypeBool,
			DefaultFunc: schema.EnvDefaultFunc("VRA7_INSECURE", nil),
			Optional:    true,
			Description: "Specify whether to validate TLS certificates.",
		},
	}
}

//Function use - To authenticate terraform provider
func providerConfig(r *schema.ResourceData) (interface{}, error) {
	//Create a client handle to perform REST calls for various operations upon the resource

	user := r.Get("username").(string)
	password := r.Get("password").(string)
	tenant := r.Get("tenant").(string)
	baseURL := r.Get("host").(string)
	insecure := r.Get("insecure").(bool)
	vraClient := sdk.NewClient(user, password, tenant, baseURL, insecure)

	//Authenticate user
	err := vraClient.Authenticate()

	//Raise an error on authentication fail
	if err != nil {
		return nil, fmt.Errorf("Error: Unable to get auth token: %v", err)
	}

	//Return client handle on success
	return &vraClient, nil
}
