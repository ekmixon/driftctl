package remote

import (
	"testing"

	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/cloudskiff/driftctl/mocks"
	"github.com/cloudskiff/driftctl/pkg/filter"
	"github.com/cloudskiff/driftctl/pkg/remote/alerts"
	"github.com/cloudskiff/driftctl/pkg/remote/aws"
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	"github.com/cloudskiff/driftctl/pkg/remote/common"
	remoteerr "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	testresource "github.com/cloudskiff/driftctl/test/resource"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestApiGatewayRestApi(t *testing.T) {
	dummyError := errors.New("this is an error")

	tests := []struct {
		test           string
		mocks          func(*repository.MockApiGatewayRepository, *mocks.AlerterInterface)
		assertExpected func(t *testing.T, got []*resource.Resource)
		wantErr        error
	}{
		{
			test: "no api gateway rest apis",
			mocks: func(repository *repository.MockApiGatewayRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllRestApis").Return([]*apigateway.RestApi{}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
		},
		{
			test: "multiple api gateway rest apis",
			mocks: func(repository *repository.MockApiGatewayRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllRestApis").Return([]*apigateway.RestApi{
					{Id: awssdk.String("3of73v5ob4")},
					{Id: awssdk.String("1jitcobwol")},
				}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 2)

				assert.Equal(t, got[0].ResourceId(), "3of73v5ob4")
				assert.Equal(t, got[0].ResourceType(), resourceaws.AwsApiGatewayRestApiResourceType)

				assert.Equal(t, got[1].ResourceId(), "1jitcobwol")
				assert.Equal(t, got[1].ResourceType(), resourceaws.AwsApiGatewayRestApiResourceType)
			},
		},
		{
			test: "cannot list api gateway rest apis",
			mocks: func(repository *repository.MockApiGatewayRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllRestApis").Return(nil, dummyError)
				alerter.On("SendAlert", resourceaws.AwsApiGatewayRestApiResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayRestApiResourceType, resourceaws.AwsApiGatewayRestApiResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: remoteerr.NewResourceListingError(dummyError, resourceaws.AwsApiGatewayRestApiResourceType),
		},
	}

	providerVersion := "3.19.0"
	schemaRepository := testresource.InitFakeSchemaRepository("aws", providerVersion)
	resourceaws.InitResourcesMetadata(schemaRepository)
	factory := terraform.NewTerraformResourceFactory(schemaRepository)

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			scanOptions := ScannerOptions{}
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository.MockApiGatewayRepository{}
			c.mocks(fakeRepo, alerter)

			var repo repository.ApiGatewayRepository = fakeRepo

			remoteLibrary.AddEnumerator(aws.NewApiGatewayRestApiEnumerator(repo, factory))

			testFilter := &filter.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, scanOptions, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}

			c.assertExpected(tt, got)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
			testFilter.AssertExpectations(tt)
		})
	}
}

func TestApiGatewayResource(t *testing.T) {
	dummyError := errors.New("this is an error")
	apis := []*apigateway.RestApi{
		{Id: awssdk.String("3of73v5ob4")},
		{Id: awssdk.String("1jitcobwol")},
	}

	tests := []struct {
		test           string
		mocks          func(*repository.MockApiGatewayRepository, *mocks.AlerterInterface)
		assertExpected func(t *testing.T, got []*resource.Resource)
		wantErr        error
	}{
		{
			test: "no api gateway resources",
			mocks: func(repo *repository.MockApiGatewayRepository, alerter *mocks.AlerterInterface) {
				repo.On("ListAllRestApis").Return(apis, nil)
				repo.On("ListAllRestApiResources", apis).Return([]*repository.RestApiResource{}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
		},
		{
			test: "multiple api gateway resources",
			mocks: func(repo *repository.MockApiGatewayRepository, alerter *mocks.AlerterInterface) {
				repo.On("ListAllRestApis").Return(apis, nil)
				repo.On("ListAllRestApiResources", apis).Return([]*repository.RestApiResource{
					{
						Resource:  apigateway.Resource{Id: awssdk.String("21zk4y"), Path: awssdk.String("/")},
						RestApiId: awssdk.String("3of73v5ob4"),
					},
					{
						Resource:  apigateway.Resource{Id: awssdk.String("2ltv32p058"), Path: awssdk.String("/")},
						RestApiId: awssdk.String("1jitcobwol"),
					},
				}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 2)

				assert.Equal(t, got[0].ResourceId(), "21zk4y")
				assert.Equal(t, got[0].ResourceType(), resourceaws.AwsApiGatewayResourceResourceType)

				assert.Equal(t, got[1].ResourceId(), "2ltv32p058")
				assert.Equal(t, got[1].ResourceType(), resourceaws.AwsApiGatewayResourceResourceType)
			},
		},
		{
			test: "cannot list rest apis",
			mocks: func(repo *repository.MockApiGatewayRepository, alerter *mocks.AlerterInterface) {
				repo.On("ListAllRestApis").Return(nil, dummyError)
				alerter.On("SendAlert", resourceaws.AwsApiGatewayResourceResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayResourceResourceType, resourceaws.AwsApiGatewayRestApiResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayResourceResourceType, resourceaws.AwsApiGatewayRestApiResourceType),
		},
		{
			test: "cannot list api gateway resources",
			mocks: func(repo *repository.MockApiGatewayRepository, alerter *mocks.AlerterInterface) {
				repo.On("ListAllRestApis").Return(apis, nil)
				repo.On("ListAllRestApiResources", apis).Return(nil, dummyError)
				alerter.On("SendAlert", resourceaws.AwsApiGatewayResourceResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayResourceResourceType, resourceaws.AwsApiGatewayResourceResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: remoteerr.NewResourceListingError(dummyError, resourceaws.AwsApiGatewayResourceResourceType),
		},
	}

	providerVersion := "3.19.0"
	schemaRepository := testresource.InitFakeSchemaRepository("aws", providerVersion)
	resourceaws.InitResourcesMetadata(schemaRepository)
	factory := terraform.NewTerraformResourceFactory(schemaRepository)

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			scanOptions := ScannerOptions{}
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository.MockApiGatewayRepository{}
			c.mocks(fakeRepo, alerter)

			var repo repository.ApiGatewayRepository = fakeRepo

			remoteLibrary.AddEnumerator(aws.NewApiGatewayResourceEnumerator(repo, factory))

			testFilter := &filter.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, scanOptions, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}

			c.assertExpected(tt, got)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
			testFilter.AssertExpectations(tt)
		})
	}
}
