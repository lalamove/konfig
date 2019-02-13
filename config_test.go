package konfig

import (
	"errors"
	"log"
	"os"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/lalamove/nui/nlogger"
	"github.com/stretchr/testify/require"
)

type DummyLoader struct {
	DataToLoad    [][2]string
	maxRetry      int
	retryDelay    time.Duration
	stopOnFailure bool
	err           bool
}

func (d *DummyLoader) Load(s Values) error {
	if d.err {
		return errors.New("")
	}
	for _, dl := range d.DataToLoad {
		log.Print("setting data", dl[0], dl[1])
		s.Set(dl[0], dl[1])
	}
	log.Print("running loader")
	return nil
}

func (d *DummyLoader) Name() string {
	return "dummy"
}

func (d *DummyLoader) MaxRetry() int {
	return d.maxRetry
}

func (d *DummyLoader) StopOnFailure() bool {
	return d.stopOnFailure
}

func (d *DummyLoader) RetryDelay() time.Duration {
	return d.retryDelay
}

func TestSingleton(t *testing.T) {
	t.Run(
		"Instance",
		func(t *testing.T) {
			Init(DefaultConfig())
			var v = c
			var i = Instance()
			require.Equal(t, v, i)
		},
	)

	t.Run(
		"RegisterCloser",
		func(t *testing.T) {
			Init(DefaultConfig())
			var v = instance()
			RegisterCloser(nil)
			require.Equal(t, Closers{nil}, v.Closers)
		},
	)
}

func TestConfigWatcherLoader(t *testing.T) {
	var testCases = []struct {
		name       string
		setUp      func(ctrl *gomock.Controller)
		asserts    func(t *testing.T)
		strictKeys []string
		loadErr    bool
		watchErr   bool
	}{
		{
			name: "OneLoaderStrictKeysNoWatcher",
			setUp: func(ctrl *gomock.Controller) {
				RegisterLoader(
					&DummyLoader{
						[][2]string{
							[2]string{
								"foo", "bar",
							},
						},
						1,
						3 * time.Second,
						false,
						false,
					},
				)
			},
			asserts: func(t *testing.T) {
				require.Equal(t, "bar", MustString("foo"))
			},
			strictKeys: []string{
				"foo",
			},
		},
		{
			name:    "OneLoaderStrictKeysErrNoWatcher",
			loadErr: true,
			setUp: func(ctrl *gomock.Controller) {
				RegisterLoader(
					&DummyLoader{
						[][2]string{
							[2]string{
								"foo", "bar",
							},
						},
						1,
						3 * time.Second,
						false,
						false,
					},
				)
			},
			asserts: func(t *testing.T) {
				require.Equal(t, "bar", MustString("foo"))
			},
			strictKeys: []string{
				"bar",
			},
		},
		{
			name: "OneLoaderNoWatcherOneError",
			setUp: func(ctrl *gomock.Controller) {

				// set our expectations
				var l = NewMockLoader(ctrl)

				l.EXPECT().MaxRetry().MinTimes(1).Return(2)
				l.EXPECT().RetryDelay().MinTimes(1).Return(1 * time.Millisecond)

				gomock.InOrder(
					l.EXPECT().Load(Values{}).Return(errors.New("")),
					l.EXPECT().Load(Values{}).Return(nil),
				)

				RegisterLoader(
					l,
				)
			},
			asserts: func(t *testing.T) {},
		},
		{
			name: "OneLoaderNoWatcherOneError",
			setUp: func(ctrl *gomock.Controller) {

				// set our expectations
				var l = NewMockLoader(ctrl)

				l.EXPECT().MaxRetry().MinTimes(1).Return(2)
				l.EXPECT().RetryDelay().MinTimes(1).Return(1 * time.Millisecond)

				gomock.InOrder(
					l.EXPECT().Load(Values{}).Return(errors.New("")),
					l.EXPECT().Load(Values{}).Return(nil),
				)

				RegisterLoader(
					l,
				)
			},
			asserts: func(t *testing.T) {},
		},

		{
			name: "OneLoaderNoWatcherErrorMaxRetry",
			setUp: func(ctrl *gomock.Controller) {

				// set our expectations
				var l = NewMockLoader(ctrl)

				l.EXPECT().MaxRetry().MinTimes(1).Return(2)
				l.EXPECT().RetryDelay().MinTimes(1).Return(1 * time.Millisecond)

				gomock.InOrder(
					l.EXPECT().Load(Values{}).Return(errors.New("")),
					l.EXPECT().Load(Values{}).Return(errors.New("")),
					l.EXPECT().Load(Values{}).Return(errors.New("")),
					l.EXPECT().StopOnFailure().Return(false),
				)

				RegisterLoader(
					l,
				)
			},
			asserts: func(t *testing.T) {},
			loadErr: true,
		},
		{
			name: "OneWatcherLoaderError",
			setUp: func(ctrl *gomock.Controller) {
				// set our expectations
				var wl = NewMockWatcher(ctrl)
				var l = NewMockLoader(ctrl)
				var c = make(chan struct{}, 1)
				var d = make(chan struct{})

				l.EXPECT().MaxRetry().MinTimes(1).Return(2)
				l.EXPECT().RetryDelay().MinTimes(1).Return(1 * time.Millisecond)

				gomock.InOrder(
					l.EXPECT().Load(Values{}).Return(nil),
					l.EXPECT().Load(Values{}).Return(errors.New("")),
					l.EXPECT().Load(Values{}).Return(errors.New("")),
					l.EXPECT().Load(Values{}).Return(errors.New("")),
					l.EXPECT().StopOnFailure().Return(true),
				)
				wl.EXPECT().Start().Times(1).Return(nil)
				wl.EXPECT().Done().MinTimes(1).Return(d)
				wl.EXPECT().Watch().Return(c)
				wl.EXPECT().Close().Return(nil)
				// register the loader
				RegisterLoaderWatcher(
					NewLoaderWatcher(
						l,
						wl,
					),
				)

				// write to the watch chan
				c <- struct{}{}
			},
			asserts: func(t *testing.T) {
				// we don't assert anything as we set expectations on the mock
				// we make it wait long enough
				time.Sleep(100 * time.Millisecond)
			},
		},
		{
			name: "MultiWatcherLoadersError",
			setUp: func(ctrl *gomock.Controller) {
				// set our expectations
				var wl1 = NewMockWatcher(ctrl)
				var l = NewMockLoader(ctrl)
				var c = make(chan struct{}, 1)
				var d = make(chan struct{})

				var wl2 = NewMockWatcher(ctrl)
				var l2 = NewMockLoader(ctrl)
				var c2 = make(chan struct{}, 1)
				var d2 = make(chan struct{})

				l.EXPECT().MaxRetry().MinTimes(1).Return(2)
				l.EXPECT().RetryDelay().MinTimes(1).Return(1 * time.Millisecond)

				l2.EXPECT().Load(Values{}).MinTimes(1).Return(nil)

				gomock.InOrder(
					l.EXPECT().Load(Values{}).Return(nil),
					l.EXPECT().Load(Values{}).Return(errors.New("")),
					l.EXPECT().Load(Values{}).Return(errors.New("")),
					l.EXPECT().Load(Values{}).Return(errors.New("")),
					l.EXPECT().StopOnFailure().Return(true),
				)
				wl2.EXPECT().Start().Times(1).Return(nil)
				wl2.EXPECT().Done().MinTimes(1).Return(d2)
				wl2.EXPECT().Watch().MinTimes(1).Return(c2)
				wl2.EXPECT().Close().Times(1).Return(nil)

				wl1.EXPECT().Start().Times(1).Return(nil)
				wl1.EXPECT().Done().MinTimes(1).Return(d)
				wl1.EXPECT().Watch().Return(c)
				wl1.EXPECT().Close().Return(nil)

				// register the loader
				RegisterLoaderWatcher(
					NewLoaderWatcher(
						l,
						wl1,
					),
					func(Store) error {
						return nil
					},
				)
				RegisterLoaderWatcher(
					NewLoaderWatcher(
						l2,
						wl2,
					),
					func(Store) error {
						return nil
					},
				)

				// close the watch chan so that it always goes through
				close(c)
				close(c2)
			},
			asserts: func(t *testing.T) {
				log.Print("sleeping")
				time.Sleep(200 * time.Millisecond)
			},
		},
		{
			name: "MultiWatcherLoadersLoaderHooksError",
			setUp: func(ctrl *gomock.Controller) {
				// set our expectations
				var wl1 = NewMockWatcher(ctrl)
				var l = NewMockLoader(ctrl)
				var c = make(chan struct{}, 1)
				var d = make(chan struct{})

				var wl2 = NewMockWatcher(ctrl)
				var l2 = NewMockLoader(ctrl)
				var c2 = make(chan struct{}, 1)
				var d2 = make(chan struct{})

				l2.EXPECT().Load(Values{}).MinTimes(1).Return(nil)

				gomock.InOrder(
					l.EXPECT().Load(Values{}).Return(nil),
					l.EXPECT().Load(Values{}).Return(nil),
					l.EXPECT().StopOnFailure().Return(false),
				)
				wl2.EXPECT().Start().Times(1).Return(nil)
				wl2.EXPECT().Done().MinTimes(1).Return(d2)
				wl2.EXPECT().Watch().MinTimes(1).Return(c2)

				wl1.EXPECT().Start().Times(1).Return(nil)
				wl1.EXPECT().Done().MinTimes(1).Return(d)
				wl1.EXPECT().Watch().MinTimes(1).Return(c)

				var i int
				// register the loader
				RegisterLoaderWatcher(
					NewLoaderWatcher(
						l,
						wl1,
					),
					func(Store) error {
						if i == 0 {
							i++
							return nil
						}
						return errors.New("err")
					},
				)
				RegisterLoaderWatcher(
					NewLoaderWatcher(
						l2,
						wl2,
					),
					func(Store) error {
						return nil
					},
				)

				// close the watch chan so that it always goes through
				c <- struct{}{}
			},
			asserts: func(t *testing.T) {
				log.Print("sleeping")
				time.Sleep(200 * time.Millisecond)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// init the test
			reset()
			Init(DefaultConfig())
			var c = instance()
			c.cfg.NoExitOnError = true

			// setup the test
			var ctrl = gomock.NewController(t)
			defer ctrl.Finish()
			testCase.setUp(ctrl)

			if testCase.strictKeys != nil {
				c.Strict(testCase.strictKeys...)
			}

			// run the test
			var err = Load()
			if testCase.loadErr {
				require.NotNil(t, err, "there should be an error")
				return
			}

			require.Nil(t, err, "there should be no error")
			err = Watch()
			if testCase.watchErr {
				require.Nil(t, err, "there should be no error")
			}
			testCase.asserts(t)
			log.Print("test done")
		})
	}

}

func TestStoreMisc(t *testing.T) {
	t.Run(
		"Strict",
		func(t *testing.T) {
			reset()

			Init(DefaultConfig())
			Strict("test", "test1")

			var c = instance()

			require.Equal(t, []string{"test", "test1"}, c.strictKeys)
		},
	)

	t.Run(
		"Name",
		func(t *testing.T) {
			var s = New(&Config{
				Name: "test",
			})
			require.Equal(t, "test", s.Name())
		},
	)

	t.Run(
		"SetLogger",
		func(t *testing.T) {
			reset()
			Init(DefaultConfig())

			var l = nlogger.New(os.Stdout, "")

			SetLogger(l)

			var c = instance()

			require.True(t, c.cfg.Logger == l)
		},
	)
}

func TestRegisterLoaderHooks(t *testing.T) {
	var ctrl = gomock.NewController(t)
	defer ctrl.Finish()
	reset()
	var c = instance()

	var l = RegisterLoader(
		NewMockLoader(ctrl),
	)

	require.Equal(t, 0, len(c.WatcherLoaders[0].loaderHooks))

	l.AddHooks(
		func(Store) error { return nil },
	)

	require.Equal(t, 1, len(c.WatcherLoaders[0].loaderHooks))
}

type TestCloser struct {
	err    error
	closed bool
}

func (t *TestCloser) Close() error {
	t.closed = true
	return t.err
}

func TestStop(t *testing.T) {
	t.Run(
		"no error, 2 closers",
		func(t *testing.T) {
			var c = New(DefaultConfig()).(*store)
			var testCloser = &TestCloser{}
			c.RegisterCloser(testCloser)
			c.cfg.NoExitOnError = true
			c.stop()
			require.Equal(t, true, testCloser.closed)
		},
	)

	t.Run(
		"no error, 2 closers",
		func(t *testing.T) {
			var c = New(DefaultConfig())
			var testCloser = &TestCloser{
				err: errors.New("foo"),
			}
			c.RegisterCloser(testCloser)
			c.(*store).cfg.NoExitOnError = true
			c.(*store).stop()
			require.Equal(t, true, testCloser.closed)
		},
	)
	reset()
}

func TestGet(t *testing.T) {
	reset()
	Init(DefaultConfig())
	Set("FOO", "BAR")
	require.Equal(t, "BAR", Get("FOO"))
	require.Equal(t, nil, Get("IDONOTEXIST"))
	reset()
}
