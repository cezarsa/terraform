package plans

import (
	"github.com/hashicorp/terraform/addrs"
)

type Changes struct {
	Resources   []*ResourceInstanceChange
	RootOutputs map[string]*OutputChange
}

type ResourceInstanceChange struct {
	Addr addrs.AbsResourceInstance
}

type OutputChange struct {
}
