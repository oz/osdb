![travis](https://api.travis-ci.org/oz/osdb.png?branch=master)

This is a Go client library for [OpenSubtitles](http://opensubtitles.org/).

Install with `go get github.com/oz/osdb`

Example usage:

```go
package main

import (
	"github.com/oz/osdb"
	"fmt"
)

func main() {
	hash, err := osdb.Hash("somefile.avi")
	if err != nil {
		panic(err)
	}
	fmt.Println("hash: %x\n", hash)
}
```

# License

BSD, see the LICENSE file.
