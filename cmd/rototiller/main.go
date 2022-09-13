// @title        Rototiller
// @version      1.0
// @description  Geospatial data transformation service.

// @contact.name   logsquaredn
// @contact.url    https://rototiller.logsquaredn.io/
// @contact.email  logsquaredn@gmail.com

// @schemes  http
// @host     localhost:8080

// @securityDefinitions.apikey  ApiKeyAuth
// @in                          header
// @name                        X-API-Key

package main

import (
	"context"
	"os"

	"github.com/gin-gonic/gin"
	command "github.com/logsquaredn/rototiller/pkg/command/d"

	_ "github.com/logsquaredn/rototiller/docs"
	_ "gocloud.dev/blob/fileblob"
	_ "gocloud.dev/blob/s3blob"
)

func main() {
	gin.SetMode(gin.ReleaseMode)

	if err := command.NewRoot().ExecuteContext(context.Background()); err != nil {
		os.Exit(1)
	}

	os.Exit(0)
}
