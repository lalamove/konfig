package klreader

import (
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/lalamove/konfig"
	"github.com/lalamove/konfig/mocks"
	"github.com/stretchr/testify/require"
)

func TestLoader(t *testing.T) {
	var ctrl = gomock.NewController(t)
	defer ctrl.Finish()

	var mockParser = mocks.NewMockParser(ctrl)

	var r = strings.NewReader(`foo=bar`)

	var loader = New(&Config{
		Reader:     r,
		Parser:     mockParser,
		RetryDelay: 5 * time.Minute,
		MaxRetry:   5,
	})

	mockParser.EXPECT().Parse(r, konfig.Values{}).Return(nil)

	loader.Load(konfig.Values{})

	require.Equal(t, defaultName, loader.Name())
	require.Equal(t, 5*time.Minute, loader.RetryDelay())
	require.Equal(t, 5, loader.MaxRetry())
	require.False(t, loader.StopOnFailure())
}
