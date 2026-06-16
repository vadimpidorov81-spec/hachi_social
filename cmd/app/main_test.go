package main

import (
	"io/fs"
	"testing"

	"github.com/hachisocial/hachisocial/web"
)

func TestEmbeddedFrontendAssets(t *testing.T) {
	t.Parallel()

	for _, name := range []string{"index.html", "styles.css", "app.js"} {
		info, err := fs.Stat(web.Assets, name)
		if err != nil {
			t.Fatalf("embedded asset %q is unavailable: %v", name, err)
		}
		if info.Size() == 0 {
			t.Fatalf("embedded asset %q is empty", name)
		}
	}
}
