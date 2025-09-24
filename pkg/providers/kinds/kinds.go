package kinds

type ProviderKind string

const (
	ProviderKindMock       ProviderKind = "mock"
	ProviderKindGitlab     ProviderKind = "gitlab"
	ProviderKindGithub     ProviderKind = "github"
	ProviderKindGoogle     ProviderKind = "google"
	ProviderKindTeleport   ProviderKind = "teleport"
	ProviderKindAWS        ProviderKind = "aws"
	ProviderKindCloudflare ProviderKind = "cloudflare"
)

var AllProviderKinds = []ProviderKind{
	ProviderKindMock,
	ProviderKindGitlab,
	ProviderKindGithub,
	ProviderKindGoogle,
	ProviderKindTeleport,
	ProviderKindAWS,
	ProviderKindCloudflare,
}
