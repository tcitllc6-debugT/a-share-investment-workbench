package update

import "context"

type Result struct {
	Available bool
	Version   string
	URL       string
	Source    string
	Notes     string
	SHA256    string
	Size      int64
	AssetName string
}

type Provider interface {
	Check(ctx context.Context, currentVersion string) (Result, error)
}

type DisabledProvider struct{}

func (DisabledProvider) Check(ctx context.Context, currentVersion string) (Result, error) {
	return Result{Available: false, Source: "disabled"}, nil
}
