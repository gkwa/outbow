package options

type Options struct {
	Verbose                 bool
	LogFormat               string
	StorageType             string
	SubsetPercentage        int
	NoRunOsascript          bool
	AllowReviewsLoadSeconds int
}
