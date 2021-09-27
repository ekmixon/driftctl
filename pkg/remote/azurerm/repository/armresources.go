package repository

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/resources/armresources"
	"github.com/cloudskiff/driftctl/pkg/remote/azurerm/common"
	"github.com/cloudskiff/driftctl/pkg/remote/cache"
)

type ArmResourcesRespository interface {
	ListAllResourceGroups() ([]*armresources.ResourceGroup, error)
}

type armResourcesRepository struct {
	client *armresources.ResourceGroupsClient
	cache  cache.Cache
}

func NewArmResourcesRepository(con *arm.Connection, config common.AzureProviderConfig, cache cache.Cache) *armResourcesRepository {
	return &armResourcesRepository{
		armresources.NewResourceGroupsClient(con, config.SubscriptionID),
		cache,
	}
}

func (s *armResourcesRepository) ListAllResourceGroups() ([]*armresources.ResourceGroup, error) {
	cacheKey := "armResourcesListAllResourceGroups"
	if v := s.cache.Get(cacheKey); v != nil {
		return v.([]*armresources.ResourceGroup), nil
	}

	pager := s.client.List(nil)
	results := make([]*armresources.ResourceGroup, 0)
	for pager.NextPage(context.Background()) {
		resp := pager.PageResponse()
		if err := pager.Err(); err != nil {
			return nil, err
		}
		results = append(results, resp.ResourceGroupsListResult.Value...)
	}
	if err := pager.Err(); err != nil {
		return nil, err
	}

	s.cache.Put(cacheKey, results)

	return results, nil
}
