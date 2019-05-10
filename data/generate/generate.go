package main

import (
	"log"

	"github.com/d2r2/go-rsync/data"
	"github.com/shurcooL/vfsgen"
)

// This application is used to generate "assets_vfsdata.go" file for application Release compilation.
// In Release mode all data files found in "assets" folder is encapsulated to "assets_vfsdata.go".

func main() {
	err := vfsgen.Generate(data.Assets, vfsgen.Options{
		PackageName:  "data",
		BuildTags:    "gorsync_rel",
		VariableName: "Assets",
	})
	if err != nil {
		log.Fatalln(err)
	}
}
