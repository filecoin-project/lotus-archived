package leveldbbs

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/filecoin-project/lotus/lib/blockstore"
)

func TestLevelDBBlockstore(t *testing.T) {
	ts := &blockstore.TestSuite{
		NewBlockstore:  newBlockstore(),
		OpenBlockstore: openBlockstore(),
	}
	ts.RunTests(t, "single")
}

func newBlockstore() func(tb testing.TB) (bs blockstore.Blockstore, path string) {
	return func(tb testing.TB) (bs blockstore.Blockstore, path string) {
		tb.Helper()

		path, err := ioutil.TempDir("", "")
		if err != nil {
			tb.Fatal(err)
		}

		db, err := Open(path, DefaultOptions())
		if err != nil {
			tb.Fatal(err)
		}

		tb.Cleanup(func() {
			_ = os.RemoveAll(path)
		})

		return db, path
	}
}

func openBlockstore() func(tb testing.TB, path string) (bs blockstore.Blockstore, err error) {
	return func(tb testing.TB, path string) (bs blockstore.Blockstore, err error) {
		tb.Helper()
		return Open(path, DefaultOptions())
	}
}
