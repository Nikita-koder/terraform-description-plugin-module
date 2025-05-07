package terraformdescription

import "github.com/hashicorp/terraform-plugin-framework/resource/schema"

func resource() schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional: true,
				Computed: true,
			}, // want `attribute "id" is missing Description or MarkdownDescription`
			"name": schema.StringAttribute{
				Optional:    true,
				Description: "Name of the resource",
			},
		},
	}
}
