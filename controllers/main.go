// Copyright 2016 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package controllers

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/hexya-erp/hexya/src/controllers"
	"github.com/hexya-erp/hexya/src/server"
	"github.com/hexya-erp/hexya/src/tools/hweb"
	"github.com/hexya-erp/hexya/src/tools/logging"
)

const (
	commonCSSRoute   = "/web/assets/common.css"
	backendCSSRoute  = "/web/assets/backend.css"
	frontendCSSRoute = "/web/assets/frontend.css"
)

var (
	// AssetsUtils contains the list of scss utilities that must be compiled first
	AssetsUtils = []string{
		"/static/web/lib/bootstrap/scss/_functions.scss",
		"/static/web/lib/bootstrap/scss/_mixins.scss",
		"/static/web/src/scss/bs_mixins_overrides.scss",
		"/static/web/src/scss/utils.scss",
	}
	// AssetsPrimaryVariables contains the list of scss files defining primary variables
	AssetsPrimaryVariables = []string{
		"/static/web/src/scss/primary_variables.scss",
	}
	// AssetsSecondaryVariables contains the list of scss files defining secondary variables
	AssetsSecondaryVariables = []string{
		"/static/web/src/scss/secondary_variables.scss",
	}
	// AssetsBootstrapVariables contains the list of scss files with bootstrap variables
	AssetsBootstrapVariables = []string{
		"/static/web/lib/bootstrap/scss/_variables.scss",
	}
	// AssetsBackendHelper contains the list of scss files that need to override Bootstrap variables in backend
	AssetsBackendHelper = []string{
		"/static/web/src/scss/bootstrap_overridden.scss",
	}
	// AssetsFrontendHelper contains the list of scss files that need to override Bootstrap variables in frontend
	AssetsFrontendHelper = []string{
		"/static/web/src/scss/bootstrap_overridden.scss",
	}
	// AssetsBootstrap contains the list of scss files that import boostrap
	AssetsBootstrap = []string{
		"/static/web/src/scss/import_bootstrap.scss",
		"/static/web/src/scss/bootstrap_review.scss",
	}
	// CommonScss is the list of Scss assets to import by the web client
	// that are common to the frontend and the backend. All Scss assets are
	// cat'ed together in the given order before being compiled.
	//
	// Assets utils will be inserted before this list
	CommonScss = []string{
		"/static/web/lib/tempusdominus/tempusdominus.scss",
		"/static/web/src/scss/fonts.scss",
		"/static/web/src/scss/ui.scss",
		"/static/web/src/scss/ui_extra.scss",
		"/static/web/src/scss/navbar.scss",
		"/static/web/src/scss/mimetypes.scss",
		"/static/web/src/scss/modal.scss",
		"/static/web/src/scss/animation.scss",
		"/static/web/src/scss/rainbow.scss",
		"/static/web/src/scss/datepicker.scss",
		"/static/web/src/scss/daterangepicker.scss",
		"/static/web/src/scss/banner.scss",
		"/static/web/src/scss/colorpicker.scss",
		"/static/web/src/scss/translation_dialog.scss",
		"/static/web/src/scss/keyboard.scss",
		"/static/web/src/scss/name_and_signature.scss",
		"/static/web/src/scss/web.zoomodoo.scss",
		"/static/web/src/scss/fontawesome_overridden.scss",
	}
	// BackendScss is the list of Scss assets to import by the web client
	// that are specific to the backend. All scss assets are
	// cat'ed together in the given order before being compiled.
	//
	// Assets utils will be inserted before this list
	BackendScss = []string{
		"/static/web/src/scss/webclient_extra.scss",
		"/static/web/src/scss/webclient_layout.scss",
		"/static/web/src/scss/webclient.scss",
		"/static/web/src/scss/domain_selector.scss",
		"/static/web/src/scss/model_field_selector.scss",
		"/static/web/src/scss/progress_bar.scss",
		"/static/web/src/scss/dropdown.scss",
		"/static/web/src/scss/dropdown_extra.scss",
		"/static/web/src/scss/tooltip.scss",
		"/static/web/src/scss/switch_company_menu.scss",
		"/static/web/src/scss/debug_manager.scss",
		"/static/web/src/scss/control_panel.scss",
		"/static/web/src/scss/fields.scss",
		"/static/web/src/scss/fields_extra.scss",
		"/static/web/src/scss/views.scss",
		"/static/web/src/scss/pivot_view.scss",
		"/static/web/src/scss/graph_view.scss",
		"/static/web/src/scss/form_view.scss",
		"/static/web/src/scss/form_view_extra.scss",
		"/static/web/src/scss/list_view.scss",
		"/static/web/src/scss/list_view_extra.scss",
		"/static/web/src/scss/kanban_dashboard.scss",
		"/static/web/src/scss/kanban_examples_dialog.scss",
		"/static/web/src/scss/kanban_column_progressbar.scss",
		"/static/web/src/scss/kanban_view.scss",
		"/static/web/src/scss/kanban_view_mobile.scss",
		"/static/web/src/scss/web_calendar.scss",
		"/static/web/src/scss/web_calendar_mobile.scss",
		"/static/web/src/scss/search_view.scss",
		"/static/web/src/scss/search_panel.scss",
		"/static/web/src/scss/search_view_mobile.scss",
		"/static/web/src/scss/dropdown_menu.scss",
		"/static/web/src/scss/search_view_extra.scss",
		"/static/web/src/scss/data_export.scss",
		"/static/web/src/scss/attachment_preview.scss",
		"/static/web/src/scss/notification.scss",
		"/static/web/src/scss/base_document_layout.scss",
		"/static/web/src/scss/ribbon.scss",
		"/static/web/src/scss/base_settings.scss",
		"/static/web/src/scss/report_backend.scss",
	}
	// FrontendScss is the list of Less assets to import by the web client
	// that are specific to the frontend. All scss assets are
	// cat'ed together in the given order before being compiled.
	//
	// Assets utils will be inserted before this list
	FrontendScss = []string{
		"/static/web/src/scss/lazyloader.scss",
		"/static/web/src/scss/navbar_mobile.scss",
		"/static/web/src/scss/notification.scss",
	}
	// FrontendContext is the base context to update when rendering
	// a frontend HTML template.
	FrontendContext = hweb.Context{
		"commonCompiledCSS":   commonCSSRoute,
		"frontendCompiledCSS": frontendCSSRoute,
	}
)

func init() {
	log = logging.GetLogger("web/controllers")
}

// RegisterRoutes register all controllers for the web module.
// This function is called from the web PreInit function so that
// server.ResourceDir is set before calling this function, but
// controllers.Bootstrap is not, yet.
func RegisterRoutes() {
	os.Remove(getAssetTempFile(commonCSSRoute))
	os.Remove(getAssetTempFile(backendCSSRoute))
	os.Remove(getAssetTempFile(frontendCSSRoute))

	root := controllers.Registry
	root.AddController(http.MethodGet, "/", func(c *server.Context) {
		c.Redirect(http.StatusSeeOther, "/web")
	})
	root.AddController(http.MethodGet, "/web/login", LoginGet)
	root.AddController(http.MethodPost, "/web/login", LoginPost)
	root.AddController(http.MethodGet, "/web/binary/company_logo", CompanyLogo)
	assets := root.AddGroup("/web/assets")
	{
		assets.AddController(http.MethodGet, "/common.css", AssetsCommonCSS)
		assets.AddController(http.MethodGet, "/backend.css", AssetsBackendCSS)
		assets.AddController(http.MethodGet, "/frontend.css", AssetsFrontendCSS)
	}

	root.AddStatic("/static", filepath.Join(server.ResourceDir, "static"))
	root.AddController(http.MethodGet, "/dashboard", Dashboard)
	web := root.AddGroup("/web")
	{
		web.AddMiddleWare(LoginRequired)
		web.AddController(http.MethodGet, "/", WebClient)
		web.AddController(http.MethodGet, "/image", Image)
		web.AddController(http.MethodGet, "/menu/:menu_id", MenuImage)

		sess := web.AddGroup("/session")
		{
			sess.AddController(http.MethodPost, "/modules", Modules)
			sess.AddController(http.MethodPost, "/get_session_info", GetSessionInfo)
			sess.AddController(http.MethodGet, "/logout", Logout)
			sess.AddController(http.MethodPost, "/change_password", ChangePassword)
		}

		proxy := web.AddGroup("/proxy")
		{
			proxy.AddController(http.MethodPost, "/load", Load)
		}

		webClient := web.AddGroup("/webclient")
		{
			webClient.AddController(http.MethodGet, "/qweb/:unique", QWeb)
			webClient.AddController(http.MethodGet, "/locale", LoadLocale)
			webClient.AddController(http.MethodGet, "/locale/:lang", LoadLocale)
			webClient.AddController(http.MethodGet, "/translations/:unique", Translations)
			webClient.AddController(http.MethodGet, "/load_menus/:unique", LoadMenus)
			webClient.AddController(http.MethodPost, "/csslist", CSSList)
			webClient.AddController(http.MethodPost, "/jslist", JSList)
			webClient.AddController(http.MethodPost, "/version_info", VersionInfo)
		}
		dataset := web.AddGroup("/dataset")
		{
			dataset.AddController(http.MethodPost, "/call_kw/*path", CallKW)
			dataset.AddController(http.MethodPost, "/search_read", SearchReadController)
			dataset.AddController(http.MethodPost, "/call_button", CallButton)
		}
		action := web.AddGroup("/action")
		{
			action.AddController(http.MethodPost, "/load", ActionLoad)
			action.AddController(http.MethodPost, "/run", ActionRun)
		}
		menu := web.AddGroup("/menu")
		{
			menu.AddController(http.MethodPost, "/load_needaction", MenuLoadNeedaction)
		}
	}
}
