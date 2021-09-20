package aws

import (
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
)

type ApiGatewayResourceEnumerator struct {
	repository repository.ApiGatewayRepository
	factory    resource.ResourceFactory
}

func NewApiGatewayResourceEnumerator(repo repository.ApiGatewayRepository, factory resource.ResourceFactory) *ApiGatewayResourceEnumerator {
	return &ApiGatewayResourceEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *ApiGatewayResourceEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsApiGatewayResourceResourceType
}

func (e *ApiGatewayResourceEnumerator) Enumerate() ([]*resource.Resource, error) {
	apis, err := e.repository.ListAllRestApis()
	if err != nil {
		return nil, remoteerror.NewResourceListingErrorWithType(err, string(e.SupportedType()), aws.AwsApiGatewayRestApiResourceType)
	}

	results := make([]*resource.Resource, 0)

	resources, err := e.repository.ListAllRestApiResources(apis)
	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
	}

	for _, resource := range resources {
		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				*resource.Id,
				map[string]interface{}{
					"rest_api_id": *resource.RestApiId,
					"path":        *resource.Path,
				},
			),
		)
	}

	return results, err
}
