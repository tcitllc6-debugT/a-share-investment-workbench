package update

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDisabledProviderReturnsNoUpdate(t *testing.T) {
	p := DisabledProvider{}
	result, err := p.Check(context.Background(), "1.0.0")

	require.NoError(t, err)
	require.False(t, result.Available)
	require.Equal(t, "disabled", result.Source)
}
