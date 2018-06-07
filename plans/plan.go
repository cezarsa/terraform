package plans

import (
	"github.com/hashicorp/terraform/configs"
	"github.com/hashicorp/terraform/states"
	"github.com/zclconf/go-cty/cty"
)

type Plan struct {
	VariableValues   map[string]cty.Value
	Config           *configs.Config
	PriorState       *states.State
	Changes          *Changes
	TerraformVersion string
	ProviderSHA256s  map[string][]byte
}
