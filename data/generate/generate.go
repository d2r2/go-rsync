package main

import (
	"log"

	"github.com/d2r2/go-rsync/data"
	"github.com/shurcooL/vfsgen"
)

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
