package klfile

import (
	"errors"
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/lalamove/konfig"
	"github.com/lalamove/konfig/mocks"
	"github.com/lalamove/konfig/parser/kpjson"
	"github.com/lalamove/nui/nfs"
	"github.com/stretchr/testify/require"
)

func TestFileLoader(t *testing.T) {
	var testCases = []struct {
		name     string
		fileName string
		setUp    func(ctrl *gomock.Controller, fl *Loader)
		err      bool
	}{
		{
			name:     "BasicNoErrorLoadOnce",
			fileName: "./test",
			setUp: func(ctrl *gomock.Controller, fl *Loader) {
				var fs = nfs.NewMockFileSystem(ctrl)
				var r = ioutil.NopCloser(strings.NewReader(
					"FOO=BAR\nBAR=FOO",
				))
				fs.EXPECT().Open("./test").Return(r, nil)
				fl.fs = fs
				fl.cfg.Files[0].Parser.(*mocks.MockParser).EXPECT().Parse(r, konfig.Values{}).Return(nil)
			},
		},
		{
			name:     "ErrorOnFile",
			fileName: "./test",
			setUp: func(ctrl *gomock.Controller, fl *Loader) {
				var fs = nfs.NewMockFileSystem(ctrl)
				fs.EXPECT().Open("./test").Return(nil, errors.New(""))
				fl.fs = fs
			},
			err: true,
		},
		{
			name:     "ErrorInvalidFormat",
			fileName: "./test",
			setUp: func(ctrl *gomock.Controller, fl *Loader) {
				var fs = nfs.NewMockFileSystem(ctrl)
				var r = ioutil.NopCloser(
					strings.NewReader(`{"test":"test"`),
				)
				fs.EXPECT().Open("./test").Return(
					r,
					nil,
				)
				fl.fs = fs
				fl.cfg.Files[0].Parser.(*mocks.MockParser).
					EXPECT().
					Parse(r, konfig.Values{}).
					Return(errors.New(""))
			},
			err: true,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			konfig.Init(&konfig.Config{})
			var ctrl = gomock.NewController(t)
			defer ctrl.Finish()

			var v = konfig.Values{}

			var fl = New(&Config{
				Files: []File{
					{
						Path:   testCase.fileName,
						Parser: mocks.NewMockParser(ctrl),
					},
				},
			})

			testCase.setUp(ctrl, fl)
			var err = fl.Load(v)
			if testCase.err {
				require.NotNil(t, err, "err should not be nil")
				return
			}
			require.Nil(t, err, "err should be nil")
			require.Equal(t, defaultName, fl.Name())
			require.False(t, fl.StopOnFailure())
		})
	}

}

func TestMaxRetryRetryDelay(t *testing.T) {
	var ctrl = gomock.NewController(t)
	defer ctrl.Finish()

	var fl = New(&Config{
		MaxRetry:   10,
		RetryDelay: 1 * time.Second,
		Files: []File{
			{
				Path:   "dummy",
				Parser: mocks.NewMockParser(ctrl),
			},
		},
	})
	require.Equal(t, 10, fl.MaxRetry())
	require.Equal(t, 1*time.Second, fl.RetryDelay())
}

func TestNewLoader(t *testing.T) {
	t.Run(
		"No parser panics",
		func(t *testing.T) {
			require.Panics(t, func() {
				New(&Config{
					Files: []File{
						{
							Path:   "dummy",
							Parser: nil,
						},
					},
				})
			})
		},
	)

	t.Run(
		"No files panics",
		func(t *testing.T) {
			require.Panics(t, func() {
				var ctrl = gomock.NewController(t)
				defer ctrl.Finish()
				New(&Config{
					Files: []File{},
				})
			})
		},
	)

	t.Run(
		"With watcher",
		func(t *testing.T) {
			var ctrl = gomock.NewController(t)
			defer ctrl.Finish()
			var wl = New(&Config{
				Files: []File{
					{
						Path:   "fileloader_test.go",
						Parser: mocks.NewMockParser(ctrl),
					},
				},
				Watch: true,
			})
			require.NotNil(t, wl.FileWatcher)
		},
	)
}

func TestNewFileLoader(t *testing.T) {
	t.Run(
		"new file loader without watcher",
		func(t *testing.T) {
			var fl = NewFileLoader("config-files", kpjson.Parser, "foo.json", "bar.json")

			require.Equal(
				t,
				"foo.json",
				fl.cfg.Files[0].Path,
			)

			require.Equal(
				t,
				"bar.json",
				fl.cfg.Files[1].Path,
			)
		},
	)

	t.Run(
		"new file loader with watcher",
		func(t *testing.T) {
			var fl = NewFileLoader("config-files", kpjson.Parser, "./fileloader.go", "./fileloader_test.go").WithWatcher()

			require.Equal(
				t,
				"./fileloader.go",
				fl.cfg.Files[0].Path,
			)

			require.Equal(
				t,
				"./fileloader_test.go",
				fl.cfg.Files[1].Path,
			)

			require.NotNil(
				t,
				fl.FileWatcher,
			)
		},
	)
}
