package main

import (
	"fmt"
	"go/build"
//	"path/filepath"
	log "github.com/sirupsen/logrus"
)
func main() {
	contextLogger := log.WithFields(log.Fields{
		"pkg": "importer",
	})
	ctx := build.Default
	importPaths := map[string]*build.Package{}
	// this is mocking real input of a list of packages for now
	for _, input := range []string{
		"github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1",
	}{
		pkg, err := ctx.Import(input, ".", build.ImportComment)
		if err != nil {
			contextLogger.Error(err)
		}
		importPaths[input] = pkg
	}
}
