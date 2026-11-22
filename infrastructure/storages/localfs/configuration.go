package localfs

import "os"

type Configuration struct {
	Root           *os.Root
	ExpectFullPath *bool
}
