package repository

import (
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/cloudskiff/driftctl/pkg/remote/cache"
	awstest "github.com/cloudskiff/driftctl/test/aws"

	"github.com/stretchr/testify/mock"

	"github.com/r3labs/diff/v2"
	"github.com/stretchr/testify/assert"
)

func Test_apigatewayRepository_ListAllRestApis(t *testing.T) {
	apis := []*apigateway.RestApi{
		{Id: aws.String("restapi1")},
		{Id: aws.String("restapi2")},
		{Id: aws.String("restapi3")},
		{Id: aws.String("restapi4")},
		{Id: aws.String("restapi5")},
		{Id: aws.String("restapi6")},
	}

	tests := []struct {
		name    string
		mocks   func(client *awstest.MockFakeApiGateway, store *cache.MockCache)
		want    []*apigateway.RestApi
		wantErr error
	}{
		{
			name: "list multiple rest apis",
			mocks: func(client *awstest.MockFakeApiGateway, store *cache.MockCache) {
				client.On("GetRestApisPages",
					&apigateway.GetRestApisInput{},
					mock.MatchedBy(func(callback func(res *apigateway.GetRestApisOutput, lastPage bool) bool) bool {
						callback(&apigateway.GetRestApisOutput{
							Items: apis[:3],
						}, false)
						callback(&apigateway.GetRestApisOutput{
							Items: apis[3:],
						}, true)
						return true
					})).Return(nil).Once()

				store.On("Get", "apigatewayListAllRestApis").Return(nil).Times(1)
				store.On("Put", "apigatewayListAllRestApis", apis).Return(false).Times(1)
			},
			want: apis,
		},
		{
			name: "should hit cache",
			mocks: func(client *awstest.MockFakeApiGateway, store *cache.MockCache) {
				store.On("Get", "apigatewayListAllRestApis").Return(apis).Times(1)
			},
			want: apis,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &cache.MockCache{}
			client := &awstest.MockFakeApiGateway{}
			tt.mocks(client, store)
			r := &apigatewayRepository{
				client: client,
				cache:  store,
			}
			got, err := r.ListAllRestApis()
			assert.Equal(t, tt.wantErr, err)

			changelog, err := diff.Diff(got, tt.want)
			assert.Nil(t, err)
			if len(changelog) > 0 {
				for _, change := range changelog {
					t.Errorf("%s: %s -> %s", strings.Join(change.Path, "."), change.From, change.To)
				}
				t.Fail()
			}
			store.AssertExpectations(t)
			client.AssertExpectations(t)
		})
	}
}

func Test_apigatewayRepository_ListAllRestApiAuthorizers(t *testing.T) {
	apis := []*apigateway.RestApi{
		{Id: aws.String("restapi1")},
		{Id: aws.String("restapi2")},
	}

	apiAuthorizers := []*apigateway.Authorizer{
		{Id: aws.String("resource1")},
		{Id: aws.String("resource2")},
		{Id: aws.String("resource3")},
		{Id: aws.String("resource4")},
	}

	tests := []struct {
		name    string
		mocks   func(client *awstest.MockFakeApiGateway, store *cache.MockCache)
		want    []*apigateway.Authorizer
		wantErr error
	}{
		{
			name: "list multiple rest api authorizers",
			mocks: func(client *awstest.MockFakeApiGateway, store *cache.MockCache) {
				client.On("GetAuthorizers",
					&apigateway.GetAuthorizersInput{
						RestApiId: aws.String("restapi1"),
					}).Return(&apigateway.GetAuthorizersOutput{Items: apiAuthorizers[:2]}, nil).Once()

				client.On("GetAuthorizers",
					&apigateway.GetAuthorizersInput{
						RestApiId: aws.String("restapi2"),
					}).Return(&apigateway.GetAuthorizersOutput{Items: apiAuthorizers[2:]}, nil).Once()

				store.On("Get", "apigatewayListAllRestApiAuthorizers_api_restapi1").Return(nil).Times(1)
				store.On("Put", "apigatewayListAllRestApiAuthorizers_api_restapi1", apiAuthorizers[:2]).Return(false).Times(1)
				store.On("Get", "apigatewayListAllRestApiAuthorizers_api_restapi2").Return(nil).Times(1)
				store.On("Put", "apigatewayListAllRestApiAuthorizers_api_restapi2", apiAuthorizers[2:]).Return(false).Times(1)
			},
			want: apiAuthorizers,
		},
		{
			name: "should hit cache",
			mocks: func(client *awstest.MockFakeApiGateway, store *cache.MockCache) {
				store.On("Get", "apigatewayListAllRestApiAuthorizers_api_restapi1").Return(apiAuthorizers[:2]).Times(1)
				store.On("Get", "apigatewayListAllRestApiAuthorizers_api_restapi2").Return(apiAuthorizers[2:]).Times(1)
			},
			want: apiAuthorizers,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &cache.MockCache{}
			client := &awstest.MockFakeApiGateway{}
			tt.mocks(client, store)
			r := &apigatewayRepository{
				client: client,
				cache:  store,
			}
			got, err := r.ListAllRestApiAuthorizers(apis)
			assert.Equal(t, tt.wantErr, err)

			changelog, err := diff.Diff(got, tt.want)
			assert.Nil(t, err)
			if len(changelog) > 0 {
				for _, change := range changelog {
					t.Errorf("%s: %s -> %s", strings.Join(change.Path, "."), change.From, change.To)
				}
				t.Fail()
			}
			store.AssertExpectations(t)
			client.AssertExpectations(t)
		})
	}
}
