package azurerm

import (
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/cloudskiff/driftctl/pkg/alerter"
	"github.com/cloudskiff/driftctl/pkg/output"
	"github.com/cloudskiff/driftctl/pkg/remote/azurerm/repository"
	"github.com/cloudskiff/driftctl/pkg/remote/cache"
	"github.com/cloudskiff/driftctl/pkg/remote/common"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/azurerm"
	"github.com/cloudskiff/driftctl/pkg/terraform"
)

func Init(
	version string,
	alerter *alerter.Alerter,
	providerLibrary *terraform.ProviderLibrary,
	remoteLibrary *common.RemoteLibrary,
	progress output.Progress,
	resourceSchemaRepository *resource.SchemaRepository,
	factory resource.ResourceFactory,
	configDir string) error {

	if version == "" {
		version = "2.71.0"
	}

	provider, err := NewAzureTerraformProvider(version, progress, configDir)
	if err != nil {
		return err
	}
	err = provider.Init()
	if err != nil {
		return err
	}

	providerConfig := provider.GetConfig()
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return err
	}
	con := arm.NewDefaultConnection(cred, nil)

	c := cache.New(100)

	storageAccountRepo := repository.NewStorageRepository(con, providerConfig, c)
	armResourcesRepo := repository.NewArmResourcesRepository(con, providerConfig, c)

	providerLibrary.AddProvider(terraform.AZURE, provider)

	remoteLibrary.AddEnumerator(NewAzurermStorageAccountEnumerator(storageAccountRepo, factory))
	remoteLibrary.AddEnumerator(NewAzurermStorageContainerEnumerator(storageAccountRepo, factory))
	remoteLibrary.AddEnumerator(NewAzurermResourceGroupEnumerator(armResourcesRepo, factory))

	err = resourceSchemaRepository.Init(terraform.AZURE, version, provider.Schema())
	if err != nil {
		return err
	}
	azurerm.InitResourcesMetadata(resourceSchemaRepository)

	return nil
}
