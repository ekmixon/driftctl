package aws

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/cloudskiff/driftctl/pkg/output"
	"github.com/cloudskiff/driftctl/pkg/remote/terraform"
	tf "github.com/cloudskiff/driftctl/pkg/terraform"
)

type awsConfig struct {
	AccessKey     string
	SecretKey     string
	CredsFilename string
	Profile       string
	Token         string
	Region        string `cty:"region"`
	MaxRetries    int

	AssumeRoleARN         string
	AssumeRoleExternalID  string
	AssumeRoleSessionName string
	AssumeRolePolicy      string

	AllowedAccountIds   []string
	ForbiddenAccountIds []string

	Endpoints        map[string]string
	IgnoreTagsConfig map[string]string
	Insecure         bool

	SkipCredsValidation     bool
	SkipGetEC2Platforms     bool
	SkipRegionValidation    bool
	SkipRequestingAccountId bool
	SkipMetadataApiCheck    bool
	S3ForcePathStyle        bool
}

type AWSTerraformProvider struct {
	*terraform.TerraformProvider
	session *session.Session
	name    string
	version string
}

func NewAWSTerraformProvider(version string, progress output.Progress, configDir string) (*AWSTerraformProvider, error) {
	if version == "" {
		version = "3.19.0"
	}
	p := &AWSTerraformProvider{
		version: version,
		name:    "aws",
	}
	installer, err := tf.NewProviderInstaller(tf.ProviderConfig{
		Key:       p.name,
		Version:   version,
		ConfigDir: configDir,
	})
	if err != nil {
		return nil, err
	}
	p.session = session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	tfProvider, err := terraform.NewTerraformProvider(installer, terraform.TerraformProviderConfig{
		Name:         p.name,
		DefaultAlias: *p.session.Config.Region,
		GetProviderConfig: func(alias string) interface{} {
			return awsConfig{
				Region:     alias,
				MaxRetries: 10, // TODO make this configurable
			}
		},
	}, progress)
	if err != nil {
		return nil, err
	}
	p.TerraformProvider = tfProvider
	return p, err
}

func (a *AWSTerraformProvider) Name() string {
	return a.name
}

func (p *AWSTerraformProvider) Version() string {
	return p.version
}
