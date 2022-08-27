package initializr

const (
	DefaultVersionKey = "version"
	DefaultVersion    = "dev"
	DefaultServiceKey = "service"
	DefaultService    = "app"
)

var (
	Version    = DefaultVersion
	VersionKey = DefaultVersionKey
	Service    = DefaultService
	ServiceKey = DefaultServiceKey
)

//goland:noinspection GoBoolExpressions
func IsDev() bool { return Version == DefaultVersion }
