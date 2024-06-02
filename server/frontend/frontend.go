package frontend

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/uclipboard/uclipboard/model"
)

//go:embed dist
var frontend embed.FS

var frontendRoot fs.FS
var assets fs.FS
var logger *logrus.Entry

func InitFrontend() {
	logger = model.NewModuleLogger("frontend")
	var err error
	frontendRoot, err = fs.Sub(frontend, "dist")
	if err != nil {
		logger.Fatal("sub dist error")
		return
	}
	assets, err = fs.Sub(frontendRoot, "assets")
	if err != nil {
		logger.Fatal("sub dist error")
		return
	}
	logger.Info("frontend init success")
}

func AssetsFS() http.FileSystem {
	return http.FS(assets)
}

func FrontendRootFS() http.FileSystem {
	return http.FS(frontendRoot)
}
