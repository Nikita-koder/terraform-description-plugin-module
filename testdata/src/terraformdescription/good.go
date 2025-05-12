package terraformdescription

import "github.com/hashicorp/terraform-plugin-framework/resource/schema"

var _ = schema.Schema{
	Description: "valid", // линтер не должен ругаться
}
