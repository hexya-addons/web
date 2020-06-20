// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/hexya-erp/hexya/src/models/security"
	"github.com/hexya-erp/hexya/src/models/types"
	"github.com/hexya-erp/hexya/src/server"
	"github.com/hexya-erp/hexya/src/tools/hweb"
)

// LoginGet is called when the client calls the login page
func LoginGet(c *server.Context) {
	redirect := c.DefaultQuery("redirect", "/web")
	if c.Session().Get("uid") != nil {
		c.Redirect(http.StatusSeeOther, redirect)
		return
	}
	data := FrontendContext
	siBytes, err := json.Marshal(GetSessionInfoStruct(c.Session()))
	if err != nil {
		c.Error(err)
		return
	}
	data["session_info"] = string(siBytes)
	c.HTML(http.StatusOK, "web.login", data)
}

// LoginPost is called when the client sends credentials
// from the login page
func LoginPost(c *server.Context) {
	login := c.DefaultPostForm("login", "")
	secret := c.DefaultPostForm("password", "")
	uid, err := security.AuthenticationRegistry.Authenticate(login, secret, new(types.Context))
	if err != nil {
		c.HTML(http.StatusOK, "web.login", FrontendContext.Update(hweb.Context{
			"error": "Wrong login or password",
		}))
		return
	}

	sess := c.Session()
	sess.Set("uid", uid)
	sess.Set("login", login)
	// TODO Manage session_id
	sess.Set("ID", int64(1))
	sess.Save()
	redirect := c.DefaultPostForm("redirect", "/web")
	c.Redirect(http.StatusSeeOther, redirect)
}

// LoginRequired is a middleware that redirects to login page
// non logged in users.
func LoginRequired(c *server.Context) {
	if c.Session().Get("uid") == nil || c.Session().Get("uid") == 0 {
		c.Redirect(http.StatusSeeOther, "/web/login")
		c.Abort()
	}
}
