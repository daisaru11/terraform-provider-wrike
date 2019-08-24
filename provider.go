package main

import (
	"log"

	wrike "github.com/daisaru11/wrike-go"
	"github.com/hashicorp/terraform/helper/schema"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"oauth2_token": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("WRIKE_OAUTH2_TOKEN", nil),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"wrike_task": resourceTask(),
		},
		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	options := &wrike.ClientOptions{
		OAuth2Token: d.Get("oauth2_token").(string),
	}
	client := wrike.NewClient(options)

	log.Printf("[INFO] Wrike client successfully initiated.")

	return client, nil
}
