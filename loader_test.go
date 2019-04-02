package konfig

import (
	"errors"
	"testing"
	time "time"

	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestLoaderHooksRun(t *testing.T) {
	t.Run(
		"run all hooks no error",
		func(t *testing.T) {
			var i int
			var loaderHooks = LoaderHooks{
				func(Store) error {
					i = i + 1
					return nil
				},
				func(Store) error {
					i = i + 2
					return nil
				},
				func(Store) error {
					i = i + 3
					return nil
				},
			}
			var err = loaderHooks.Run(Instance())
			require.Nil(t, err, "err should be nil")
			require.Equal(t, 6, i, "all hooks should have run")
		},
	)

	t.Run(
		"run one hook and error",
		func(t *testing.T) {
			var i int
			var loaderHooks = LoaderHooks{
				func(Store) error {
					i = i + 1
					return errors.New("err")
				},
				func(Store) error {
					i = i + 2
					return nil
				},
				func(Store) error {
					i = i + 3
					return nil
				},
			}
			var err = loaderHooks.Run(Instance())
			require.NotNil(t, err, "err should not be nil")
			require.Equal(t, 1, i, "one hook should have run")
		},
	)
}

func TestLoaderLoadRetry(t *testing.T) {
	var testCases = []struct {
		name  string
		err   bool
		build func(ctrl *gomock.Controller) *loaderWatcher
	}{
		{
			name: "success, no loader hooks, no retry",
			build: func(ctrl *gomock.Controller) *loaderWatcher {
				var mockW = NewMockWatcher(ctrl)
				var mockL = NewMockLoader(ctrl)
				mockL.EXPECT().Load(Values{}).Return(nil)

				var wl = &loaderWatcher{
					Watcher:     mockW,
					Loader:      mockL,
					loaderHooks: nil,
				}
				return wl
			},
		},
		{
			name: "success, no loader hooks, 1 retry",
			build: func(ctrl *gomock.Controller) *loaderWatcher {
				var mockW = NewMockWatcher(ctrl)
				var mockL = NewMockLoader(ctrl)
				gomock.InOrder(
					mockL.EXPECT().Load(Values{}).Return(errors.New("")),
					mockL.EXPECT().Name().Return("l"),
					mockL.EXPECT().MaxRetry().Return(1),
					mockL.EXPECT().RetryDelay().Return(1*time.Millisecond),
					mockL.EXPECT().Load(Values{}).Return(nil),
				)
				var wl = &loaderWatcher{
					Watcher:     mockW,
					Loader:      mockL,
					loaderHooks: nil,
				}
				return wl
			},
		},
		{
			name: "error, no loader hooks, 1 retry",
			err:  true,
			build: func(ctrl *gomock.Controller) *loaderWatcher {
				var mockW = NewMockWatcher(ctrl)
				var mockL = NewMockLoader(ctrl)
				gomock.InOrder(
					mockL.EXPECT().Load(Values{}).Return(errors.New("")),
					mockL.EXPECT().Name().Return("l"),
					mockL.EXPECT().MaxRetry().Return(1),
					mockL.EXPECT().RetryDelay().Return(1*time.Millisecond),
					mockL.EXPECT().Load(Values{}).Return(errors.New("")),
					mockL.EXPECT().Name().Return("l"),
					mockL.EXPECT().MaxRetry().Return(1),
				)
				var wl = &loaderWatcher{
					Watcher:     mockW,
					Loader:      mockL,
					loaderHooks: nil,
				}
				return wl
			},
		},
		{
			name: "success, 2 loader hooks, 1 retry",
			build: func(ctrl *gomock.Controller) *loaderWatcher {
				var mockW = NewMockWatcher(ctrl)
				var mockL = NewMockLoader(ctrl)
				gomock.InOrder(
					mockL.EXPECT().Load(Values{}).Return(errors.New("")),
					mockL.EXPECT().Name().Return("l"),
					mockL.EXPECT().MaxRetry().Return(1),
					mockL.EXPECT().RetryDelay().Return(1*time.Millisecond),
					mockL.EXPECT().Load(Values{}).Return(nil),
				)
				var wl = &loaderWatcher{
					Watcher: mockW,
					Loader:  mockL,
					loaderHooks: LoaderHooks{
						func(Store) error {
							return nil
						},
						func(Store) error {
							return nil
						},
					},
				}
				return wl
			},
		},
		{
			name: "error, 2 loader hooks, 1 retry",
			err:  true,
			build: func(ctrl *gomock.Controller) *loaderWatcher {
				var mockW = NewMockWatcher(ctrl)
				var mockL = NewMockLoader(ctrl)
				gomock.InOrder(
					mockL.EXPECT().Load(Values{}).Return(errors.New("")),
					mockL.EXPECT().Name().Return("l"),
					mockL.EXPECT().MaxRetry().Return(1),
					mockL.EXPECT().RetryDelay().Return(1*time.Millisecond),
					mockL.EXPECT().Load(Values{}).Return(nil),
				)
				var wl = &loaderWatcher{
					Watcher: mockW,
					Loader:  mockL,
					loaderHooks: LoaderHooks{
						func(Store) error {
							return nil
						},
						func(Store) error {
							return errors.New("")
						},
					},
				}
				return wl
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(
			testCase.name,
			func(t *testing.T) {
				var ctrl = gomock.NewController(t)
				defer ctrl.Finish()

				reset()
				var c = instance()
				c.cfg.NoExitOnError = true
				var err = c.loaderLoadRetry(testCase.build(ctrl), 0)
				if testCase.err {
					require.NotNil(t, err, "err should not be nil")
					return
				}
				require.Nil(t, err, "err should be nil")
			},
		)
	}
}

func TestLoaderLoadRetryStrictKeys(t *testing.T) {
	var ctrl = gomock.NewController(t)
	defer ctrl.Finish()

	var c = instance()
	c.loaded = true
	c.Strict("test")
	c.cfg.NoExitOnError = true

	var mockW = NewMockWatcher(ctrl)
	var mockL = NewMockLoader(ctrl)

	gomock.InOrder(
		mockL.EXPECT().Load(Values{}).Do(func(v Values) {
			v["test"] = "test"
		}).Return(errors.New("")),
		mockL.EXPECT().Name().Return("l"),
		mockL.EXPECT().MaxRetry().Return(1),
		mockL.EXPECT().RetryDelay().Return(1*time.Millisecond),
		mockL.EXPECT().Load(Values{}).Do(func(v Values) {
			v["test"] = "test"
		}).Return(nil),

		mockL.EXPECT().Load(Values{}).Do(func(v Values) {
			v["test"] = "test"
		}).Return(errors.New("")),
		mockL.EXPECT().Name().Return("l"),
		mockL.EXPECT().MaxRetry().Return(1),
		mockL.EXPECT().RetryDelay().Return(1*time.Millisecond),
		mockL.EXPECT().Load(Values{}).Do(func(v Values) {
			v["test2"] = "test"
		}).Return(nil),
	)

	var wl = &loaderWatcher{
		Watcher: mockW,
		Loader:  mockL,
		loaderHooks: LoaderHooks{
			func(Store) error {
				return nil
			},
			func(Store) error {
				return nil
			},
		},
	}

	var err = c.loaderLoadRetry(wl, 0)
	require.Nil(t, err, "err should be nil")

	err = c.loaderLoadRetry(wl, 0)
	require.NotNil(t, err, "err should not be nil")
}

func TestLoaderLoadRetryKeyHooks(t *testing.T) {
	var ctrl = gomock.NewController(t)
	defer ctrl.Finish()

	reset()
	var c = New(DefaultConfig())
	c.loaded = true
	c.cfg.NoExitOnError = true

	var ranTest int
	c.RegisterKeyHook(
		"test",
		func(c Store) error {
			ranTest++
			return nil
		},
	)
	c.RegisterKeyHook(
		"test",
		func(c Store) error {
			ranTest++
			return nil
		},
	)

	var ranFoo int
	c.RegisterKeyHook(
		"foo",
		func(c Store) error {
			ranFoo++
			return nil
		},
	)

	var ranErr int
	c.RegisterKeyHook(
		"err.",
		func(c Store) error {
			ranErr++
			return errors.New("")
		},
	)

	var mockW = NewMockWatcher(ctrl)
	var mockL = NewMockLoader(ctrl)

	gomock.InOrder(
		mockL.EXPECT().Load(Values{}).Do(func(v Values) {
			v["test"] = "test"
			v["foo"] = "bar"
		}).Return(nil),
		mockL.EXPECT().Load(Values{}).Do(func(v Values) {
			v["test"] = "test"
			v["foo"] = "bar"
		}).Return(nil),
		mockL.EXPECT().Load(Values{}).Do(func(v Values) {
			v["test"] = "test"
			v["test.foo"] = "foo"
			v["foo"] = "bar"
		}).Return(nil),
		mockL.EXPECT().Load(Values{}).Do(func(v Values) {
			v["test"] = "test"
			v["test.foo"] = "foo"
			v["foo.test"] = "barr"
			v["foo"] = "barr"
		}).Return(nil),
		mockL.EXPECT().Load(Values{}).Do(func(v Values) {
			v["test"] = "test"
			v["test.foo"] = "foo"
		}).Return(nil),
		mockL.EXPECT().Load(Values{}).Do(func(v Values) {
			v["test"] = "test"
			v["test.foo"] = "foo"
			v["err.test"] = ""
		}).Return(nil),
	)

	var wl = &loaderWatcher{
		Watcher: mockW,
		Loader:  mockL,
		loaderHooks: LoaderHooks{
			func(Store) error {
				return nil
			},
			func(Store) error {
				return nil
			},
		},
	}

	var err = c.loaderLoadRetry(wl, 0)
	require.Nil(t, err, "err should be nil")

	err = c.loaderLoadRetry(wl, 0)
	require.Nil(t, err, "err should not be nil")

	err = c.loaderLoadRetry(wl, 0)
	require.Nil(t, err, "err should not be nil")

	err = c.loaderLoadRetry(wl, 0)
	require.Nil(t, err, "err should not be nil")

	err = c.loaderLoadRetry(wl, 0)
	require.Nil(t, err, "err should not be nil")

	err = c.loaderLoadRetry(wl, 0)
	require.NotNil(t, err, "err should not be nil")

	require.Equal(t, 4, ranTest, "ranTest should be 2")
	require.Equal(t, 3, ranFoo, "ranFoo should be 3")
	require.Equal(t, 1, ranErr, "ranErr should be 1")
}

func TestLoaderLoadWatch(t *testing.T) {
	var testCases = []struct {
		name      string
		err       bool
		maxPanics int
		build     func(ctrl *gomock.Controller) *loaderWatcher
	}{
		{
			name: "success, no errors",
			build: func(ctrl *gomock.Controller) *loaderWatcher {
				var mockW = NewMockWatcher(ctrl)
				var mockL = NewMockLoader(ctrl)

				mockL.EXPECT().Name().MinTimes(1).Return("test")
				mockL.EXPECT().Load(Values{}).Return(nil)
				mockW.EXPECT().Start().Return(nil)
				mockW.EXPECT().Watch().Return(nil)
				mockW.EXPECT().Done().Return(nil)

				var wl = &loaderWatcher{
					Watcher:     mockW,
					Loader:      mockL,
					loaderHooks: nil,
				}
				return wl
			},
		},
		{
			name: "success, errors load",
			err:  true,
			build: func(ctrl *gomock.Controller) *loaderWatcher {
				var mockW = NewMockWatcher(ctrl)
				var mockL = NewMockLoader(ctrl)

				mockL.EXPECT().Name().MinTimes(1).Return("test")
				mockL.EXPECT().Load(Values{}).Times(4).Return(errors.New("some err"))
				mockL.EXPECT().MaxRetry().Times(4).Return(3)
				mockL.EXPECT().RetryDelay().Times(3).Return(50 * time.Millisecond)
				mockL.EXPECT().StopOnFailure().Return(true)
				mockW.EXPECT().Close().MinTimes(1).Return(nil)

				var wl = &loaderWatcher{
					Watcher:     mockW,
					Loader:      mockL,
					loaderHooks: nil,
				}
				return wl
			},
		},
		{
			name: "panic load after watch, maxPanics 0",
			build: func(ctrl *gomock.Controller) *loaderWatcher {
				var mockW = NewMockWatcher(ctrl)
				var mockL = NewMockLoader(ctrl)

				var c = make(chan struct{}, 1)
				var d = make(chan struct{})

				mockL.EXPECT().Name().MinTimes(1).Return("test")
				mockL.EXPECT().Load(Values{}).Return(nil)
				mockW.EXPECT().Start().Return(nil)
				mockW.EXPECT().Watch().MinTimes(1).Return(c)
				mockW.EXPECT().Done().MinTimes(1).Return(d)

				mockL.EXPECT().Load(Values{}).Do(func(Values) error {
					panic(errors.New("some err"))
				}).Return(nil)
				mockL.EXPECT().StopOnFailure().Return(false)
				mockW.EXPECT().Close().Return(nil)

				var wl = &loaderWatcher{
					Watcher:     mockW,
					Loader:      mockL,
					loaderHooks: nil,
				}
				// trigger a watch
				c <- struct{}{}
				return wl
			},
		},
		{

			name:      "panic load after watch, maxPanics 2",
			maxPanics: 2,
			build: func(ctrl *gomock.Controller) *loaderWatcher {
				var mockW = NewMockWatcher(ctrl)
				var mockL = NewMockLoader(ctrl)

				var c = make(chan struct{}, 3)
				var d = make(chan struct{})

				mockL.EXPECT().Name().MinTimes(1).Return("test")
				mockL.EXPECT().Load(Values{}).Return(nil)
				mockW.EXPECT().Start().Return(nil)
				mockW.EXPECT().Watch().MinTimes(1).Return(c)
				mockW.EXPECT().Done().MinTimes(1).Return(d)

				gomock.InOrder(
					mockL.EXPECT().Load(Values{}).Do(func(Values) error {
						panic(errors.New("some err"))
					}).Return(nil),
					mockL.EXPECT().StopOnFailure().Return(false),
					mockL.EXPECT().Load(Values{}).Do(func(Values) error {
						panic(errors.New("some err"))
					}).Return(nil),
					mockL.EXPECT().StopOnFailure().Return(false),
					mockW.EXPECT().Close().Return(nil),
				)

				var wl = &loaderWatcher{
					Watcher:     mockW,
					Loader:      mockL,
					loaderHooks: nil,
				}
				// trigger a watch
				c <- struct{}{}
				c <- struct{}{}
				c <- struct{}{}
				return wl
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(
			testCase.name,
			func(t *testing.T) {
				var ctrl = gomock.NewController(t)
				defer ctrl.Finish()

				var c = New(&Config{
					Metrics:          true,
					MaxWatcherPanics: testCase.maxPanics,
				})

				c.RegisterLoaderWatcher(
					testCase.build(ctrl),
				)
				c.cfg.NoExitOnError = true

				var err = c.LoadWatch()

				if testCase.err {
					require.NotNil(t, err, "err should not be nil")
					return
				}

				require.Nil(t, err, "err should be nil")

				time.Sleep(300 * time.Millisecond)
			},
		)
	}
}
