// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package controllers

import (
	"encoding/base64"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/hexya-erp/hexya/src/models"
	"github.com/hexya-erp/hexya/src/models/security"
	"github.com/hexya-erp/hexya/src/server"
	"github.com/hexya-erp/hexya/src/tools/assets"
	"github.com/hexya-erp/pool/h"
	"github.com/hexya-erp/pool/q"
)

func getAssetTempFile(asset string) string {
	return filepath.Join(os.TempDir(), strings.Replace(asset, "/", "_", -1))
}

func createCSSAssets(files []string, asset string, includePaths ...string) {
	var readers []io.Reader
	for _, fileName := range files {
		filePath := filepath.Join(server.ResourceDir, fileName)
		f, err := os.Open(filePath)
		if err != nil {
			log.Panic("Error while reading less file", "filename", fileName, "error", err)
		}
		readers = append(readers, f, strings.NewReader("\n"))
	}
	tmpFile, err := os.Create(getAssetTempFile(asset))
	defer tmpFile.Close()
	if err != nil {
		log.Panic("Error while opening asset file", "error", err)
	}
	err = assets.CompileLess(io.MultiReader(readers...), tmpFile, includePaths...)
	if err != nil {
		tmpFile.Close()
		os.Remove(getAssetTempFile(asset))
		log.Panic("Error while generating asset file", "error", err)
	}
}

// AssetsCommonCSS returns the compiled CSS for the common assets
func AssetsCommonCSS(c *server.Context) {
	fName := getAssetTempFile(commonCSSRoute)
	if _, err := os.Stat(fName); err != nil {
		createCSSAssets(append(LessHelpers, CommonLess...), commonCSSRoute)
	}
	c.File(fName)
}

// AssetsBackendCSS returns the compiled CSS for the backend assets
func AssetsBackendCSS(c *server.Context) {
	fName := getAssetTempFile(backendCSSRoute)
	if _, err := os.Stat(fName); err != nil {
		bootstrapDir := filepath.Join(server.ResourceDir, "static", "web", "lib", "bootstrap", "less")
		createCSSAssets(append(LessHelpers, BackendLess...), backendCSSRoute, bootstrapDir)
	}
	c.File(fName)
}

// AssetsFrontendCSS returns the compiled CSS for the frontend assets
func AssetsFrontendCSS(c *server.Context) {
	fName := getAssetTempFile(frontendCSSRoute)
	if _, err := os.Stat(fName); err != nil {
		bootstrapDir := filepath.Join(server.ResourceDir, "static", "web", "lib", "bootstrap", "less")
		createCSSAssets(append(LessHelpers, FrontendLess...), frontendCSSRoute, bootstrapDir)
	}
	c.File(fName)
}

// Dashboard returns the dashboard image of the company or the default one
func Dashboard(c *server.Context) {
	checkUser(c.Session().Get("uid").(int64))
	var image []byte
	models.ExecuteInNewEnvironment(security.SuperUserID, func(env models.Environment) {
		user := h.User().Search(env, q.User().ID().Equals(c.Session().Get("uid").(int64)))
		if user.Company().DashboardBackground() == "" {
			return
		}
		image, _ = base64.StdEncoding.DecodeString(user.Company().DashboardBackground())
	})
	if len(image) == 0 {
		c.Redirect(http.StatusFound, "/static/web/src/img/material-background.jpg")
		return
	}
	c.Data(http.StatusOK, "image", image)
}
