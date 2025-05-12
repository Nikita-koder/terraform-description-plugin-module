package terraformdescription

import "github.com/hashicorp/terraform-plugin-framework/resource/schema"

var _ = schema.Schema{
	Description:         "Test Bad",
	MarkdownDescription: "Test Bad",
	Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Optional: true,
			Computed: true,
		}, // want `attribute "id" is missing Description or MarkdownDescription`
		"creator": schema.SingleNestedAttribute{
			Attributes: map[string]schema.Attribute{
				"id":    schema.StringAttribute{Optional: true, Computed: true}, // want `attribute "id" is missing Description or MarkdownDescription`
				"email": schema.StringAttribute{Optional: true, Computed: true}, // want `attribute "email" is missing Description or MarkdownDescription`
				"realm": schema.StringAttribute{Optional: true, Computed: true}, // want `attribute "realm" is missing Description or MarkdownDescription`
			},
			Optional: true,
		}, // want `attribute "creator" is missing Description or MarkdownDescription`
	},
}
