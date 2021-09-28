package remote

import (
	"context"
	"testing"

	asset "cloud.google.com/go/asset/apiv1"
	"cloud.google.com/go/storage"
	"github.com/cloudskiff/driftctl/mocks"
	"github.com/cloudskiff/driftctl/pkg/filter"
	"github.com/cloudskiff/driftctl/pkg/remote/alerts"
	"github.com/cloudskiff/driftctl/pkg/remote/cache"
	"github.com/cloudskiff/driftctl/pkg/remote/common"
	remoteerr "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/remote/google"
	"github.com/cloudskiff/driftctl/pkg/remote/google/repository"
	"github.com/cloudskiff/driftctl/pkg/resource"
	googleresource "github.com/cloudskiff/driftctl/pkg/resource/google"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/cloudskiff/driftctl/test"
	"github.com/cloudskiff/driftctl/test/goldenfile"
	testgoogle "github.com/cloudskiff/driftctl/test/google"
	testresource "github.com/cloudskiff/driftctl/test/resource"
	terraform2 "github.com/cloudskiff/driftctl/test/terraform"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	assetpb "google.golang.org/genproto/googleapis/cloud/asset/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGoogleStorageBucket(t *testing.T) {

	cases := []struct {
		test             string
		dirName          string
		response         []*assetpb.ResourceSearchResult
		responseErr      error
		setupAlerterMock func(alerter *mocks.AlerterInterface)
		wantErr          error
	}{
		{
			test:     "no storage buckets",
			dirName:  "google_storage_bucket_empty",
			response: []*assetpb.ResourceSearchResult{},
			wantErr:  nil,
		},
		{
			test:    "multiples storage buckets",
			dirName: "google_storage_bucket",
			response: []*assetpb.ResourceSearchResult{
				{
					AssetType:   "storage.googleapis.com/Bucket",
					DisplayName: "driftctl-unittest-1",
				},
				{
					AssetType:   "storage.googleapis.com/Bucket",
					DisplayName: "driftctl-unittest-2",
				},
				{
					AssetType:   "storage.googleapis.com/Bucket",
					DisplayName: "driftctl-unittest-3",
				},
			},
			wantErr: nil,
		},
		{
			test:        "cannot list storage buckets",
			dirName:     "google_storage_bucket_empty",
			responseErr: status.Error(codes.PermissionDenied, "The caller does not have permission"),
			setupAlerterMock: func(alerter *mocks.AlerterInterface) {
				alerter.On(
					"SendAlert",
					"google_storage_bucket",
					alerts.NewRemoteAccessDeniedAlert(
						common.RemoteGoogleTerraform,
						remoteerr.NewResourceListingError(
							status.Error(codes.PermissionDenied, "The caller does not have permission"),
							"google_storage_bucket",
						),
						alerts.EnumerationPhase,
					),
				).Once()
			},
			wantErr: nil,
		},
	}

	providerVersion := "3.78.0"
	resType := resource.ResourceType(googleresource.GoogleStorageBucketResourceType)
	schemaRepository := testresource.InitFakeSchemaRepository("google", providerVersion)
	googleresource.InitResourcesMetadata(schemaRepository)
	factory := terraform.NewTerraformResourceFactory(schemaRepository)
	deserializer := resource.NewDeserializer(factory)

	for _, c := range cases {
		t.Run(c.test, func(tt *testing.T) {
			shouldUpdate := c.dirName == *goldenfile.Update

			scanOptions := ScannerOptions{Deep: true}
			providerLibrary := terraform.NewProviderLibrary()
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			if c.setupAlerterMock != nil {
				c.setupAlerterMock(alerter)
			}

			var assetClient *asset.Client
			if !shouldUpdate {
				var err error
				assetClient, err = testgoogle.NewFakeAssetServer(c.response, c.responseErr)
				if err != nil {
					tt.Fatal(err)
				}
			}

			realProvider, err := terraform2.InitTestGoogleProvider(providerLibrary, providerVersion)
			if err != nil {
				tt.Fatal(err)
			}
			provider := terraform2.NewFakeTerraformProvider(realProvider)
			provider.WithResponse(c.dirName)

			// Replace mock by real resources if we are in update mode
			if shouldUpdate {
				ctx := context.Background()
				assetClient, err = asset.NewClient(ctx)
				if err != nil {
					tt.Fatal(err)
				}
				err = realProvider.Init()
				if err != nil {
					tt.Fatal(err)
				}
				provider.ShouldUpdate()
			}

			repo := repository.NewAssetRepository(assetClient, realProvider.GetConfig(), cache.New(0))

			remoteLibrary.AddEnumerator(google.NewGoogleStorageBucketEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(resType, common.NewGenericDetailsFetcher(resType, provider, deserializer))

			testFilter := &filter.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, scanOptions, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}
			alerter.AssertExpectations(tt)
			testFilter.AssertExpectations(tt)
			test.TestAgainstGoldenFile(got, resType.String(), c.dirName, provider, deserializer, shouldUpdate, tt)
		})
	}
}

func TestGoogleStorageBucketIAMBinding(t *testing.T) {

	cases := []struct {
		test                  string
		dirName               string
		assetRepositoryMock   func(assetRepository *repository.MockAssetRepository)
		storageRepositoryMock func(storageRepository *repository.MockStorageRepository)
		responseErr           error
		setupAlerterMock      func(alerter *mocks.AlerterInterface)
		wantErr               error
	}{
		{
			test:    "no storage buckets",
			dirName: "google_storage_bucket_empty",
			assetRepositoryMock: func(assetRepository *repository.MockAssetRepository) {
				assetRepository.On("SearchAllBuckets").Return([]*assetpb.ResourceSearchResult{}, nil)
			},
			wantErr: nil,
		},
		{
			test:    "multiples storage buckets, no bindings",
			dirName: "google_storage_bucket_binding_empty",
			assetRepositoryMock: func(assetRepository *repository.MockAssetRepository) {
				assetRepository.On("SearchAllBuckets").Return([]*assetpb.ResourceSearchResult{
					{
						AssetType:   "storage.googleapis.com/Bucket",
						DisplayName: "dctlgstoragebucketiambinding-1",
					},
					{
						AssetType:   "storage.googleapis.com/Bucket",
						DisplayName: "dctlgstoragebucketiambinding-2",
					},
				}, nil)
			},
			storageRepositoryMock: func(storageRepository *repository.MockStorageRepository) {
				storageRepository.On("ListAllBindings", "dctlgstoragebucketiambinding-1").Return(map[string][]string{}, nil)
				storageRepository.On("ListAllBindings", "dctlgstoragebucketiambinding-2").Return(map[string][]string{}, nil)
			},
			wantErr: nil,
		},
		{
			test:    "Cannot list bindings",
			dirName: "google_storage_bucket_binding_listing_error",
			assetRepositoryMock: func(assetRepository *repository.MockAssetRepository) {
				assetRepository.On("SearchAllBuckets").Return([]*assetpb.ResourceSearchResult{
					{
						AssetType:   "storage.googleapis.com/Bucket",
						DisplayName: "dctlgstoragebucketiambinding-1",
					},
				}, nil)
			},
			storageRepositoryMock: func(storageRepository *repository.MockStorageRepository) {
				storageRepository.On("ListAllBindings", "dctlgstoragebucketiambinding-1").Return(
					map[string][]string{},
					errors.New("googleapi: Error 403: driftctl-acc-circle@driftctl-qa-1.iam.gserviceaccount.com does not have storage.buckets.getIamPolicy access to the Google Cloud Storage bucket., forbidden"))
			},
			setupAlerterMock: func(alerter *mocks.AlerterInterface) {
				alerter.On(
					"SendAlert",
					"google_storage_bucket_iam_binding",
					alerts.NewRemoteAccessDeniedAlert(
						common.RemoteGoogleTerraform,
						remoteerr.NewResourceListingError(
							errors.New("googleapi: Error 403: driftctl-acc-circle@driftctl-qa-1.iam.gserviceaccount.com does not have storage.buckets.getIamPolicy access to the Google Cloud Storage bucket., forbidden"),
							"google_storage_bucket_iam_binding",
						),
						alerts.EnumerationPhase,
					),
				).Once()
			},
			wantErr: nil,
		},
		{
			test:    "multiples storage buckets, multiple bindings",
			dirName: "google_storage_bucket_binding_multiple",
			assetRepositoryMock: func(assetRepository *repository.MockAssetRepository) {
				assetRepository.On("SearchAllBuckets").Return([]*assetpb.ResourceSearchResult{
					{
						AssetType:   "storage.googleapis.com/Bucket",
						DisplayName: "dctlgstoragebucketiambinding-1",
					},
					{
						AssetType:   "storage.googleapis.com/Bucket",
						DisplayName: "dctlgstoragebucketiambinding-2",
					},
				}, nil)
			},
			storageRepositoryMock: func(storageRepository *repository.MockStorageRepository) {
				storageRepository.On("ListAllBindings", "dctlgstoragebucketiambinding-1").Return(map[string][]string{
					"roles/storage.admin":        {"user:elie.charra@cloudskiff.com"},
					"roles/storage.objectViewer": {"user:william.beuil@cloudskiff.com"},
				}, nil)

				storageRepository.On("ListAllBindings", "dctlgstoragebucketiambinding-2").Return(map[string][]string{
					"roles/storage.admin":        {"user:william.beuil@cloudskiff.com"},
					"roles/storage.objectViewer": {"user:elie.charra@cloudskiff.com"},
				}, nil)
			},
			wantErr: nil,
		},
	}

	providerVersion := "3.78.0"
	resType := resource.ResourceType(googleresource.GoogleStorageBucketIamBindingResourceType)
	schemaRepository := testresource.InitFakeSchemaRepository("google", providerVersion)
	googleresource.InitResourcesMetadata(schemaRepository)
	factory := terraform.NewTerraformResourceFactory(schemaRepository)
	deserializer := resource.NewDeserializer(factory)

	for _, c := range cases {
		t.Run(c.test, func(tt *testing.T) {
			repositoryCache := cache.New(100)

			shouldUpdate := c.dirName == *goldenfile.Update

			scanOptions := ScannerOptions{Deep: true}
			providerLibrary := terraform.NewProviderLibrary()
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			if c.setupAlerterMock != nil {
				c.setupAlerterMock(alerter)
			}

			storageRepo := &repository.MockStorageRepository{}
			if c.storageRepositoryMock != nil {
				c.storageRepositoryMock(storageRepo)
			}
			var storageRepository repository.StorageRepository = storageRepo
			if shouldUpdate {
				storageClient, err := storage.NewClient(context.Background())
				if err != nil {
					panic(err)
				}
				storageRepository = repository.NewStorageRepository(storageClient, repositoryCache)
			}

			assetRepo := &repository.MockAssetRepository{}
			if c.assetRepositoryMock != nil {
				c.assetRepositoryMock(assetRepo)
			}
			var assetRepository repository.AssetRepository = assetRepo

			realProvider, err := terraform2.InitTestGoogleProvider(providerLibrary, providerVersion)
			if err != nil {
				tt.Fatal(err)
			}
			provider := terraform2.NewFakeTerraformProvider(realProvider)
			provider.WithResponse(c.dirName)

			remoteLibrary.AddEnumerator(google.NewGoogleStorageBucketIamBindingEnumerator(assetRepository, storageRepository, factory))

			testFilter := &filter.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, scanOptions, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, c.wantErr, err)
			if err != nil {
				return
			}
			alerter.AssertExpectations(tt)
			testFilter.AssertExpectations(tt)
			test.TestAgainstGoldenFile(got, resType.String(), c.dirName, provider, deserializer, shouldUpdate, tt)
		})
	}
}
