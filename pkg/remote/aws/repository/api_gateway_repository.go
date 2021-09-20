package repository

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/aws/aws-sdk-go/service/apigateway/apigatewayiface"
	"github.com/cloudskiff/driftctl/pkg/remote/cache"
)

type ApiGatewayRepository interface {
	ListAllRestApis() ([]*apigateway.RestApi, error)
	ListAllRestApiResources([]*apigateway.RestApi) ([]*RestApiResource, error)
}

type apigatewayRepository struct {
	client apigatewayiface.APIGatewayAPI
	cache  cache.Cache
}

func NewApiGatewayRepository(session *session.Session, c cache.Cache) *apigatewayRepository {
	return &apigatewayRepository{
		apigateway.New(session),
		c,
	}
}

func (r *apigatewayRepository) ListAllRestApis() ([]*apigateway.RestApi, error) {
	if v := r.cache.Get("apigatewayListAllRestApis"); v != nil {
		return v.([]*apigateway.RestApi), nil
	}

	var restApis []*apigateway.RestApi
	input := apigateway.GetRestApisInput{}
	err := r.client.GetRestApisPages(&input,
		func(resp *apigateway.GetRestApisOutput, lastPage bool) bool {
			restApis = append(restApis, resp.Items...)
			return !lastPage
		},
	)
	if err != nil {
		return nil, err
	}

	r.cache.Put("apigatewayListAllRestApis", restApis)
	return restApis, nil
}

func (r *apigatewayRepository) ListAllRestApiResources(apis []*apigateway.RestApi) ([]*RestApiResource, error) {
	var apiResources []*RestApiResource
	for _, api := range apis {
		a := *api
		cacheKey := fmt.Sprintf("apigatewayListAllRestApiResources_api_%s", *a.Id)
		if v := r.cache.Get(cacheKey); v != nil {
			apiResources = append(apiResources, v.([]*RestApiResource)...)
			continue
		}

		var resources []*RestApiResource
		input := &apigateway.GetResourcesInput{
			RestApiId: a.Id,
		}
		err := r.client.GetResourcesPages(input, func(res *apigateway.GetResourcesOutput, lastPage bool) bool {
			for _, item := range res.Items {
				i := *item
				resources = append(resources, &RestApiResource{
					Resource:  i,
					RestApiId: a.Id,
				})
			}
			return !lastPage
		})
		if err != nil {
			return nil, err
		}

		r.cache.Put(cacheKey, resources)
		apiResources = append(apiResources, resources...)
	}
	return apiResources, nil
}

type RestApiResource struct {
	apigateway.Resource
	RestApiId *string
}
