package envlocations

type EnvLocationsInterface interface {
	LoadDotEnv() error
	Get(string) string
	GetOrDefault(string, string) string
}