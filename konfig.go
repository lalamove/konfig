// Package konfig provides a composable, observable and performant
// config handling for Go.
// Written for larger distributed systems where you may have plenty of
// configuration sources - it allows you to compose configurations from
// multiple sources with reload hooks making it simple to build apps that live
// in a highly dynamic environment.
//
// Konfig is built around 4 small interfaces: Loader, Watcher, Parser, Closer
//
// Get started:
//  var configFiles = []klfile.File{
//        {
//            Path:   "./config.json",
//            Parser: kpjson.Parser,
//        },
//  }
//
//  func init() {
//  	konfig.Init(konfig.DefaultConfig())
//  }
//
//  func main() {
//  	// load from json file with a file wacher
//  	konfig.RegisterLoaderWatcher(
//  		klfile.New(&klfile.Config{
//  			Files: configFiles,
//  			Watch: true,
//  		}),
//  		// optionally you can pass config hooks to run when a file is changed
//  		func(c konfig.Store) error {
//  			return nil
//  		},
//  	)
//
//      // Load and start watching
//  	if err := konfig.LoadWatch(); err != nil {
//  		log.Fatal(err)
//  	}
//
//  	// retrieve value from config file
//  	konfig.Bool("debug")
//  }
//
//
package konfig
