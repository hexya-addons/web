// Copyright 2016 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package controllers

import (
	"net/http"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/hexya-addons/web/webtypes"
	"github.com/hexya-erp/hexya/src/models"
	"github.com/hexya-erp/hexya/src/models/security"
	"github.com/hexya-erp/hexya/src/models/types"
	"github.com/hexya-erp/hexya/src/server"
	"github.com/hexya-erp/pool/h"
	"github.com/hexya-erp/pool/q"
	"github.com/spf13/viper"
)

// UserCompanies for the se
type UserCompanies struct {
	CurrentCompany   webtypes.RecordIDWithName   `json:"current_company"`
	AllowedCompanies []webtypes.RecordIDWithName `json:"allowed_companies"`
}

// SessionInfo gathers all information about the current session
type SessionInfo struct {
	SessionID                int64                  `json:"session_id"`
	UID                      int64                  `json:"uid"`
	UserContext              map[string]interface{} `json:"user_context"`
	DB                       string                 `json:"db"`
	UserName                 string                 `json:"username"`
	CompanyID                int64                  `json:"company_id"`
	Name                     string                 `json:"name"`
	IsSystem                 bool                   `json:"is_system"`
	IsAdmin                  bool                   `json:"is_admin"`
	PartnerDisplayName       string                 `json:"partner_display_name"`
	PartnerID                int64                  `json:"partner_id"`
	UserCompanies            UserCompanies          `json:"user_companies"`
	WebBaseURL               string                 `json:"web.base.url"`
	ShowEffect               bool                   `json:"show_effect"`
	DisplaySwitchCompanyMenu bool                   `json:"display_switch_company_menu"`
	CacheHashes              map[string]string      `json:"cache_hashes"`
}

// GetSessionInfoStruct returns a struct with information about the given session
func GetSessionInfoStruct(sess sessions.Session) *SessionInfo {
	var (
		userContext   *types.Context
		companyID     int64
		userCompanies UserCompanies
		userName      string
	)
	if sess.Get("uid") != nil {
		models.ExecuteInNewEnvironment(security.SuperUserID, func(env models.Environment) {
			user := h.User().Search(env, q.User().ID().Equals(sess.Get("uid").(int64)))
			userContext = user.ContextGet()
			companyID = user.Company().ID()
			userName = user.Name()
			userCompanies = UserCompanies{
				CurrentCompany: webtypes.RecordIDWithName{
					ID:   companyID,
					Name: user.Company().Name(),
				},
			}
			for _, company := range user.Companies().Records() {
				userCompanies.AllowedCompanies = append(userCompanies.AllowedCompanies, webtypes.RecordIDWithName{
					ID:   company.ID(),
					Name: company.Name(),
				})
			}
		})
		return &SessionInfo{
			SessionID:     sess.Get("ID").(int64),
			UID:           sess.Get("uid").(int64),
			UserContext:   userContext.ToMap(),
			DB:            viper.GetString("DB.Name"),
			UserName:      sess.Get("login").(string),
			CompanyID:     companyID,
			Name:          userName,
			UserCompanies: userCompanies,
			CacheHashes: map[string]string{
				"load_menus":   "1",
				"qweb":         "2",
				"translations": "3",
			},
		}
	}
	return nil
}

// GetSessionInfo returns the session information to the client
func GetSessionInfo(c *server.Context) {
	c.RPC(http.StatusOK, GetSessionInfoStruct(c.Session()))
}

// Modules returns the list of installed modules to the client
func Modules(c *server.Context) {
	mods := make([]string, len(server.Modules))
	for i, m := range server.Modules {
		mods[i] = m.Name
	}
	c.RPC(http.StatusOK, mods)
}

// Logout the current user and redirect to login page
func Logout(c *server.Context) {
	sess := c.Session()
	sess.Delete("uid")
	sess.Delete("ID")
	sess.Delete("login")
	sess.Save()
	redirect := c.DefaultQuery("redirect", "/web/login")
	c.Redirect(http.StatusSeeOther, redirect)
}

// ChangePasswordData is the params format passed to ChangePassword controller
type ChangePasswordData struct {
	Fields []struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	} `json:"fields"`
}

// ChangePassword is called by the client to change the current user password
func ChangePassword(c *server.Context) {
	uid := c.Session().Get("uid").(int64)
	var params ChangePasswordData
	c.BindRPCParams(&params)
	var oldPassword, newPassword, confirmPassword string
	for _, d := range params.Fields {
		switch d.Name {
		case "old_pwd":
			oldPassword = d.Value
		case "new_password":
			newPassword = d.Value
		case "confirm_pwd":
			confirmPassword = d.Value
		}
	}
	res := make(gin.H)
	err := models.ExecuteInNewEnvironment(uid, func(env models.Environment) {
		rs := h.User().NewSet(env)
		if strings.TrimSpace(oldPassword) == "" ||
			strings.TrimSpace(newPassword) == "" ||
			strings.TrimSpace(confirmPassword) == "" {
			log.Panic(rs.T("You cannot leave any password empty."))
		}
		if newPassword != confirmPassword {
			log.Panic(rs.T("The new password and its confirmation must be identical."))
		}
		if rs.ChangePassword(oldPassword, newPassword) {
			res["new_password"] = newPassword
			return
		}
		log.Panic(rs.T("Error, password not changed !"))
	})
	c.RPC(http.StatusOK, res, err)
}
