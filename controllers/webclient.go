// Copyright 2016 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package controllers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/hexya-addons/web/scripts"
	"github.com/hexya-erp/hexya/src/actions"
	"github.com/hexya-erp/hexya/src/i18n"
	"github.com/hexya-erp/hexya/src/menus"
	"github.com/hexya-erp/hexya/src/models"
	"github.com/hexya-erp/hexya/src/models/security"
	"github.com/hexya-erp/hexya/src/server"
	"github.com/hexya-erp/hexya/src/tools"
	"github.com/hexya-erp/hexya/src/tools/b64image"
	"github.com/hexya-erp/hexya/src/tools/hweb"
	"github.com/hexya-erp/hexya/src/tools/xmlutils"
	"github.com/hexya-erp/pool/h"
	"github.com/hexya-erp/pool/q"
)

// QWeb returns a concatenation of all client qweb templates
func QWeb(c *server.Context) {
	mods := strings.Split(c.Query("mods"), ",")
	fileNames := tools.ListStaticFiles(server.ResourceDir, "src/xml", mods, true)
	res, _, err := xmlutils.ConcatXML(fileNames)
	if err != nil {
		c.Error(fmt.Errorf("error while generating client side QWeb: %s", err.Error()))
	}
	c.Data(http.StatusOK, "application/xml; charset=utf-8", res)
}

// Translations returns data about the current language
func Translations(c *server.Context) {
	lang := c.Query("lang")
	res := gin.H{
		"lang_parameters": i18n.GetLocale(lang),
		"modules":         scripts.ListModuleTranslations(lang),
		"multi_lang":      true,
	}
	c.JSON(http.StatusOK, res)
}

// LoadMenus returns the menus of the application as JSON
func LoadMenus(c *server.Context) {
	lang := GetSessionInfoStruct(c.Session()).UserContext["lang"].(string)
	var allRootMenuIds []int64
	for _, menu := range menus.Registry.All() {
		allRootMenuIds = append(allRootMenuIds, menu.ID)
	}
	rootMenu := gin.H{
		"name":         "root",
		"parent_id":    parentTuple{"-1", ""},
		"children":     getMenuTree(menus.Registry.Menus, lang),
		"all_menu_ids": allRootMenuIds,
	}
	c.JSON(http.StatusOK, rootMenu)
}

// CSSList returns the list of CSS files
func CSSList(c *server.Context) {
	Params := struct {
		Mods string `json:"mods"`
	}{}
	c.BindRPCParams(&Params)
	mods := strings.Split(Params.Mods, ",")
	fileNames := tools.ListStaticFiles(server.ResourceDir, "src/css", mods, false)
	c.RPC(http.StatusOK, fileNames)
}

// JSList returns the list of JS files
func JSList(c *server.Context) {
	Params := struct {
		Mods string `json:"mods"`
	}{}
	c.BindRPCParams(&Params)
	mods := strings.Split(Params.Mods, ",")
	fileNames := tools.ListStaticFiles(server.ResourceDir, "src/js", mods, false)
	c.RPC(http.StatusOK, fileNames)
}

// VersionInfo returns server version information to the client
func VersionInfo(c *server.Context) {
	data := gin.H{
		"server_serie":        "0.9beta",
		"server_version_info": []int8{0, 9, 0, 0, 0},
		"server_version":      "0.9beta",
		"protocol":            1,
	}
	c.RPC(http.StatusOK, data)
}

// LoadLocale returns the locale's JS file
func LoadLocale(c *server.Context) {
	lang := c.Param("lang")
	var outstr string
	langFull := strings.ToLower(strings.Replace(lang, "_", "-", -1))
	jsPath := filepath.Join(server.ResourceDir, "static", "web", "lib", "moment", "locale")
	jsPathFull := filepath.Join(jsPath, fmt.Sprintf("%s.js", langFull))
	content, err := ioutil.ReadFile(jsPathFull)
	if err != nil {
		langShort := strings.Split(lang, "_")[0]
		jsPathShort := filepath.Join(jsPath, fmt.Sprintf("%s.js", langShort))
		content, _ = ioutil.ReadFile(jsPathShort)
	}
	if len(content) > 2 {
		outstr = string(content)
	}
	c.Header("Content-Type", "application/javascript")
	c.String(http.StatusOK, outstr)
}

// A Menu is the representation of a single menu item
type Menu struct {
	ID          int64                `json:"id"`
	XMLID       string               `json:"xmlid"`
	Name        string               `json:"name"`
	Children    []Menu               `json:"children"`
	Action      actions.ActionString `json:"action"`
	Parent      parentTuple          `json:"parent_id"`
	Sequence    uint8                `json:"sequence"`
	WebIconData string               `json:"web_icon_data"`
}

// a parentTuple is an array of two strings that marshals itself as "false" if empty
type parentTuple [2]string

// MarshalJSON method for the parentTuple type. Marshals itself as "false" if empty.
func (pt parentTuple) MarshalJSON() ([]byte, error) {
	if pt == [2]string{} {
		return json.Marshal(false)
	}
	return json.Marshal([2]string(pt))
}

// getMenuTree returns a tree of menus with all their descendants
// from a given slice of menus.Menu objects.
func getMenuTree(menus []*menus.Menu, lang string) []Menu {
	res := make([]Menu, len(menus))
	for i, m := range menus {
		// We deliberately instantiate an empty slice so that it gets JSON marshalled as []
		children := []Menu{}
		if m.HasChildren {
			children = getMenuTree(m.Children.Menus, lang)
		}
		name := m.Name
		if lang != "" {
			name = m.TranslatedName(lang)
		}
		parent := parentTuple{}
		if m.Parent != nil {
			var parentName string
			for cur := m.Parent; cur.Parent != nil; cur = cur.Parent {
				parentName = fmt.Sprintf("%s/%s", cur.Parent.Name, parentName)
			}
			parent = parentTuple{m.ParentID, strings.TrimSuffix(parentName, "/")}
		}
		var aString actions.ActionString
		if m.Action != nil {
			aString = m.Action.ActionString()
		}
		var iconData string
		if m.WebIcon != "" {
			var err error
			iconData, err = b64image.ReadAll(filepath.Join(server.ResourceDir, m.WebIcon))
			if err != nil {
				log.Warn("error while loading menu image", "menu", m.Name, "image", m.WebIcon, "error", err)
			}
		}
		res[i] = Menu{
			ID:          m.ID,
			XMLID:       m.XMLID,
			Parent:      parent,
			Name:        name,
			Action:      aString,
			Children:    children,
			Sequence:    m.Sequence,
			WebIconData: iconData,
		}
	}
	return res
}

// WebClient is the controller for the application main page
func WebClient(c *server.Context) {
	var lang string
	if c.Session().Get("uid") != nil {
		models.ExecuteInNewEnvironment(security.SuperUserID, func(env models.Environment) {
			user := h.User().Search(env, q.User().ID().Equals(c.Session().Get("uid").(int64)))
			lang = user.ContextGet().GetString("lang")
		})
	}

	siBytes, err := json.Marshal(GetSessionInfoStruct(c.Session()))
	if err != nil {
		c.Error(err)
		return
	}
	modBytes, err := json.Marshal(server.Modules.Names())
	if err != nil {
		c.Error(err)
		return
	}

	data := hweb.Context{
		"modules":            string(modBytes),
		"session_info":       string(siBytes),
		"debug":              "",
		"commonCompiledCSS":  commonCSSRoute,
		"backendCompiledCSS": backendCSSRoute,
	}
	templateName := strings.TrimPrefix(path.Join(lang, "web.webclient_bootstrap"), "/")
	c.HTML(http.StatusOK, templateName, data)
}
