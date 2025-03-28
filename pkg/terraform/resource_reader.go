package terraform

import (
	"github.com/cloudskiff/driftctl/pkg/resource"

	"github.com/zclconf/go-cty/cty"
)

type ResourceReader interface {
	ReadResource(args ReadResourceArgs) (*cty.Value, error)
}

type ReadResourceArgs struct {
	Ty         resource.ResourceType
	ID         string
	Attributes map[string]string
}
