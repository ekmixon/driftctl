package aws

import (
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
)

type ApiGatewayDomainNameEnumerator struct {
	repository repository.ApiGatewayRepository
	factory    resource.ResourceFactory
}

func NewApiGatewayDomainNameEnumerator(repo repository.ApiGatewayRepository, factory resource.ResourceFactory) *ApiGatewayDomainNameEnumerator {
	return &ApiGatewayDomainNameEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *ApiGatewayDomainNameEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsApiGatewayDomainNameResourceType
}

func (e *ApiGatewayDomainNameEnumerator) Enumerate() ([]*resource.Resource, error) {
	domainNames, err := e.repository.ListAllDomainNames()
	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
	}

	results := make([]*resource.Resource, 0, len(domainNames))

	for _, domainName := range domainNames {
		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				*domainName.DomainName,
				map[string]interface{}{},
			),
		)
	}

	return results, err
}
