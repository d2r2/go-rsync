//--------------------------------------------------------------------------------------------------
// This file is a part of Gorsync Backup project (backup RSYNC frontend).
// Copyright (c) 2017-2022 Denis Dyakov <denis.dyakov@gma**.com>
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING
// BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
// NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
// DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
//--------------------------------------------------------------------------------------------------

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
