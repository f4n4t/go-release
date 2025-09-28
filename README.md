# go-release

Go library for parsing and checking a scene or p2p release.

# Usage

```go
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/f4n4t/go-release"
)

func main() {
	releaseService := release.NewServiceBuilder().WithSkipMediaInfo(true).Build()
	releaseInfo, err := releaseService.Parse("./Example.Release-Group")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	out, _ := json.MarshalIndent(releaseInfo, "", "\t")

	fmt.Println(string(out))

	if releaseInfo.SfvCount > 0 {
		if err := releaseService.CheckSFV(releaseInfo, true); err != nil {
			fmt.Println("sfv check failed:", err)
			os.Exit(1)
		}
	}

	if releaseInfo.HasExtensions("zip") {
		if err := releaseService.CheckZip(releaseInfo, true); err != nil {
			fmt.Println("zip check failed:", err)
			os.Exit(1)
		}
	}
}
```
