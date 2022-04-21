package todo

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed docs
var assetsFiles embed.FS

func GetFileSystem() http.FileSystem {
	fsys, err := fs.Sub(assetsFiles, "docs")
	if err != nil {
		panic(err)
	}
	return http.FS(fsys)
}
