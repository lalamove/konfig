package klhttp

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/lalamove/konfig"
	"github.com/lalamove/konfig/mocks"
	"github.com/lalamove/konfig/watcher/kwpoll"
	"github.com/stretchr/testify/require"
)

type RequestMatcher struct {
	req *http.Request
	msg string
}

func (r *RequestMatcher) Matches(x interface{}) bool {
	if v, ok := x.(*http.Request); ok {
		if v.Method != r.req.Method || v.URL.String() != r.req.URL.String() {
			r.msg = "method or url are different"
			return false
		}

		if r.req.Body != nil && v.Body == nil {
			r.msg = "body are different"
			return false
		}

		if r.req.Body != nil {
			var b, _ = ioutil.ReadAll(r.req.Body)
			var b2, _ = ioutil.ReadAll(v.Body)

			if string(b) != string(b2) {
				r.msg = "body are different"
				return false
			}
		}

		return true
	}
	return false
}

func (r *RequestMatcher) String() string {
	return r.msg
}

func TestLoad(t *testing.T) {
	var testCases = []struct {
		name  string
		setUp func(ctrl *gomock.Controller) *Loader
		err   bool
	}{
		{
			name: "single source no error get request",
			setUp: func(ctrl *gomock.Controller) *Loader {
				var c = mocks.NewMockClient(ctrl)
				var p1 = mocks.NewMockParser(ctrl)

				var hl = New(&Config{
					Client: c,
					Sources: []Source{
						{
							URL:    "http://source.com",
							Parser: p1,
						},
					},
				})

				var r = ioutil.NopCloser(strings.NewReader(``))
				var req, _ = http.NewRequest("GET", "http://source.com", nil)
				c.EXPECT().Do(&RequestMatcher{
					req: req,
				}).Times(1).Return(
					&http.Response{
						StatusCode: 200,
						Body:       r,
					},
					nil,
				)

				p1.EXPECT().Parse(r, konfig.Values{}).Times(1).Return(nil)

				return hl
			},
			err: false,
		},
		{
			name: "multiple sources no error get request",
			setUp: func(ctrl *gomock.Controller) *Loader {
				var c = mocks.NewMockClient(ctrl)
				var p1 = mocks.NewMockParser(ctrl)
				var p2 = mocks.NewMockParser(ctrl)

				var hl = New(&Config{
					Client: c,
					Sources: []Source{
						{
							URL:    "http://source.com",
							Parser: p1,
						},
						{
							Method: http.MethodPost,
							URL:    "http://source.com",
							Parser: p2,
						},
					},
				})

				var r = ioutil.NopCloser(strings.NewReader(``))
				var req1, _ = http.NewRequest("GET", "http://source.com", nil)
				var req2, _ = http.NewRequest("POST", "http://source.com", nil)

				gomock.InOrder(
					c.EXPECT().Do(&RequestMatcher{
						req: req1,
					}).Times(1).Return(
						&http.Response{
							StatusCode: 200,
							Body:       r,
						},
						nil,
					),
					c.EXPECT().Do(&RequestMatcher{
						req: req2,
					}).Times(1).Return(
						&http.Response{
							StatusCode: 200,
							Body:       r,
						},
						nil,
					),
				)

				p1.EXPECT().Parse(r, konfig.Values{}).Times(1).Return(nil)
				p2.EXPECT().Parse(r, konfig.Values{}).Times(1).Return(nil)

				return hl
			},
			err: false,
		},
		{
			name: "multiple sources watch no error get request",
			setUp: func(ctrl *gomock.Controller) *Loader {
				var c = mocks.NewMockClient(ctrl)
				var p1 = mocks.NewMockParser(ctrl)
				var p2 = mocks.NewMockParser(ctrl)

				var r = ioutil.NopCloser(strings.NewReader(``))
				var req1, _ = http.NewRequest("GET", "http://source.com", nil)
				var req2, _ = http.NewRequest("POST", "http://source.com", nil)

				gomock.InOrder(
					c.EXPECT().Do(&RequestMatcher{
						req: req1,
					}).Times(1).Return(
						&http.Response{
							StatusCode: 200,
							Body:       r,
						},
						nil,
					),
					c.EXPECT().Do(&RequestMatcher{
						req: req2,
					}).Times(1).Return(
						&http.Response{
							StatusCode: 200,
							Body:       r,
						},
						nil,
					),
				)

				p1.EXPECT().Parse(r, konfig.Values{}).Times(1).Return(nil)
				p2.EXPECT().Parse(r, konfig.Values{}).Times(1).Return(nil)

				var hl = New(&Config{
					Client: c,
					Watch:  true,
					Rater:  kwpoll.Time(100 * time.Millisecond),
					Sources: []Source{
						{
							URL:    "http://source.com",
							Parser: p1,
						},
						{
							Method: http.MethodPost,
							URL:    "http://source.com",
							Parser: p2,
						},
					},
				})

				gomock.InOrder(
					c.EXPECT().Do(&RequestMatcher{
						req: req1,
					}).Times(1).Return(
						&http.Response{
							StatusCode: 200,
							Body:       r,
						},
						nil,
					),
					c.EXPECT().Do(&RequestMatcher{
						req: req2,
					}).Times(1).Return(
						&http.Response{
							StatusCode: 200,
							Body:       r,
						},
						nil,
					),
				)

				p1.EXPECT().Parse(r, konfig.Values{}).Times(1).Return(nil)
				p2.EXPECT().Parse(r, konfig.Values{}).Times(1).Return(nil)

				return hl
			},
			err: false,
		},
		{
			name: "multiple sources no error get request",
			setUp: func(ctrl *gomock.Controller) *Loader {
				var c = mocks.NewMockClient(ctrl)
				var p1 = mocks.NewMockParser(ctrl)
				var p2 = mocks.NewMockParser(ctrl)

				var hl = New(&Config{
					Client: c,
					Sources: []Source{
						{
							URL:    "http://source.com",
							Parser: p1,
						},
						{
							Method: http.MethodPost,
							URL:    "http://source.com",
							Parser: p2,
						},
					},
				})

				var r = ioutil.NopCloser(strings.NewReader(``))
				var req1, _ = http.NewRequest("GET", "http://source.com", nil)
				var req2, _ = http.NewRequest("POST", "http://source.com", nil)

				gomock.InOrder(
					c.EXPECT().Do(&RequestMatcher{
						req: req1,
					}).Times(1).Return(
						&http.Response{
							StatusCode: 200,
							Body:       r,
						},
						nil,
					),
					c.EXPECT().Do(&RequestMatcher{
						req: req2,
					}).Times(1).Return(
						nil,
						errors.New(""),
					),
				)

				p1.EXPECT().Parse(r, konfig.Values{}).Times(1).Return(nil)

				return hl
			},
			err: true,
		},
		{
			name: "single source error wrong status code",
			setUp: func(ctrl *gomock.Controller) *Loader {
				var c = mocks.NewMockClient(ctrl)
				var p1 = mocks.NewMockParser(ctrl)

				var hl = New(&Config{
					Client: c,
					Sources: []Source{
						{
							URL:        "http://source.com",
							Parser:     p1,
							StatusCode: 201,
						},
					},
				})

				var r = ioutil.NopCloser(strings.NewReader(``))
				var req, _ = http.NewRequest("GET", "http://source.com", nil)
				c.EXPECT().Do(&RequestMatcher{
					req: req,
				}).Times(1).Return(
					&http.Response{
						StatusCode: 400,
						Body:       r,
					},
					nil,
				)
				return hl
			},
			err: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(
			testCase.name,
			func(t *testing.T) {
				var ctrl = gomock.NewController(t)
				defer ctrl.Finish()

				konfig.Init(konfig.DefaultConfig())
				var hl = testCase.setUp(ctrl)

				var err = hl.Load(konfig.Values{})
				if testCase.err {
					require.NotNil(t, err, "err should not be nil")
					return
				}
				require.Nil(t, err, "err should be nil")
			},
		)
	}
}

func TestNew(t *testing.T) {
	t.Run(
		"default http client",
		func(t *testing.T) {
			var ctrl = gomock.NewController(t)
			defer ctrl.Finish()

			var p = mocks.NewMockParser(ctrl)

			var hl = New(&Config{
				Sources: []Source{
					{
						URL:    "http://url.com",
						Parser: p,
					},
				},
			})

			require.Equal(t, http.DefaultClient, hl.cfg.Client)
		},
	)
	t.Run(
		"panic no sources",
		func(t *testing.T) {
			var ctrl = gomock.NewController(t)
			defer ctrl.Finish()

			require.Panics(t, func() {
				New(&Config{
					Sources: []Source{},
				})
			})
		},
	)
}

func TestLoaderMethods(t *testing.T) {

	var ctrl = gomock.NewController(t)
	defer ctrl.Finish()

	var p = mocks.NewMockParser(ctrl)

	var hl = New(&Config{
		Name:          "httploader",
		MaxRetry:      1,
		RetryDelay:    1 * time.Second,
		StopOnFailure: true,
		Sources: []Source{
			{
				URL:    "http://url.com",
				Parser: p,
			},
		},
	})

	require.True(t, hl.StopOnFailure())
	require.Equal(t, "httploader", hl.Name())
	require.Equal(t, 1*time.Second, hl.RetryDelay())
	require.Equal(t, 1, hl.MaxRetry())
}
