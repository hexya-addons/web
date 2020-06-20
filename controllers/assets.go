// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package controllers

import (
	"encoding/base64"
	"fmt"
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

const bootstrapScssPath = "static/web/lib/bootstrap/scss"

func getAssetTempFile(asset string) string {
	return filepath.Join(os.TempDir(), strings.Replace(asset, "/", "_", -1))
}

// createCompiledCSSFile compiles the SCSS files into fName file
func createCompiledCSSFile(fName string, files []string, includePaths ...string) {
	tmpFile, err := os.Create(fName)
	if err != nil {
		log.Panic("Error while opening asset file", "error", err)
	}
	defer tmpFile.Close()
	var readers []io.Reader
	for _, fileName := range files {
		filePath := filepath.Join(server.ResourceDir, fileName)
		f, err := os.Open(filePath)
		if err != nil {
			log.Panic("Error while reading less file", "filename", fileName, "error", err)
		}
		readers = append(readers, strings.NewReader(fmt.Sprintf("/* ====importing %s==== */\n", fileName)), f, strings.NewReader("\n"))
	}
	err = assets.ScssCompiler{}.Compile(io.MultiReader(readers...), tmpFile, includePaths...)
	if err != nil {
		log.Panic("Error while creating CSS assets file", "error", err)
	}
}

// AssetsCommonCSS returns the compiled CSS for the common assets
// It creates the temp file on the fly if it doesn't exist yet
func AssetsCommonCSS(c *server.Context) {
	fName := getAssetTempFile(commonCSSRoute)
	if _, err := os.Stat(fName); err != nil {
		var commonList []string
		for _, l := range [][]string{AssetsUtils, AssetsPrimaryVariables, AssetsSecondaryVariables, AssetsBootstrapVariables, CommonScss} {
			commonList = append(commonList, l...)
		}
		bootstrapScssDir := filepath.Join(server.ResourceDir, bootstrapScssPath)
		createCompiledCSSFile(fName, commonList, bootstrapScssDir)
	}
	c.File(fName)
}

// AssetsBackendCSS returns the compiled CSS for the backend assets
func AssetsBackendCSS(c *server.Context) {
	fName := getAssetTempFile(backendCSSRoute)
	if _, err := os.Stat(fName); err != nil {
		var backendList []string
		for _, l := range [][]string{AssetsUtils, AssetsPrimaryVariables, AssetsSecondaryVariables, AssetsBackendHelper, AssetsBootstrapVariables, AssetsBootstrap, BackendScss} {
			backendList = append(backendList, l...)
		}
		bootstrapScssDir := filepath.Join(server.ResourceDir, bootstrapScssPath)
		createCompiledCSSFile(fName, backendList, bootstrapScssDir)
	}
	c.File(fName)
}

// AssetsFrontendCSS returns the compiled CSS for the frontend assets
func AssetsFrontendCSS(c *server.Context) {
	fName := getAssetTempFile(frontendCSSRoute)
	if _, err := os.Stat(fName); err != nil {
		var frontendList []string
		for _, l := range [][]string{AssetsUtils, AssetsPrimaryVariables, AssetsSecondaryVariables, AssetsFrontendHelper, AssetsBootstrapVariables, AssetsBootstrap, FrontendScss} {
			frontendList = append(frontendList, l...)
		}
		bootstrapScssDir := filepath.Join(server.ResourceDir, bootstrapScssPath)
		createCompiledCSSFile(fName, frontendList, bootstrapScssDir)
	}
	c.File(fName)
}

// Dashboard returns the dashboard image of the company or the default one
func Dashboard(c *server.Context) {
	CheckUser(c.Session().Get("uid").(int64))
	var image []byte
	models.ExecuteInNewEnvironment(security.SuperUserID, func(env models.Environment) {
		user := h.User().Search(env, q.User().ID().Equals(c.Session().Get("uid").(int64)))
		if user.Company().DashboardBackground() == "" {
			return
		}
		image, _ = base64.StdEncoding.DecodeString(user.Company().DashboardBackground())
	})
	if len(image) == 0 {
		c.Redirect(http.StatusFound, "/static/web/src/img/material-background.png")
		return
	}
	c.Data(http.StatusOK, "image", image)
}
