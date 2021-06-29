[![libstorage](https://circleci.com/gh/libstorage/libstoragemgmt-golang.svg?style=svg)](https://github.com/libstorage/libstoragemgmt-golang)


A golang library for libStorageMgmt client and plugins.

A client example listing systems
```go
package main

import (
	"fmt"

	lsm "github.com/libstorage/libstoragemgmt-golang"
)

func main() {
	// Ignoring errors for succinctness
	c, _ := lsm.Client("sim://", "", 30000)
	systems, _ := c.Systems()
	for _, s := range systems {
		fmt.Printf("ID: %s, Name:%s, Version: %s\n", s.ID, s.Name, s.FwVersion)
	}
}

```
For an example of how to write a libStorageMgmt plugin using this, please see
the example:
[https://github.com/tasleson/simgo](https://github.com/tasleson/simgo)
