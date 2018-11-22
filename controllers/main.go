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
	// CommonLess is the list of Less assets to import by the web client
	// that are common to the frontend and the backend. All less assets are
	// cat'ed together in the given order before being compiled.
	CommonLess = []string{
		"/static/web/src/less/fonts.less",
		"/static/web/src/less/navbar.less",
		"/static/web/src/less/mimetypes.less",
		"/static/web/src/less/animation.less",
		"/static/web/src/less/backend_variables.less",
	}
	// CommonCSS is the list of CSS files to include without compilation both
	// for the frontend and the backend.
	CommonCSS = []string{
		"/static/web/lib/jquery.ui/jquery-ui.css",
		"/static/web/lib/fontawesome/css/font-awesome.css",
		"/static/web/lib/bootstrap-datetimepicker/css/bootstrap-datetimepicker.css",
		"/static/web/lib/select2/select2.css",
		"/static/web/lib/select2-bootstrap-css/select2-bootstrap.css",
	}
	// CommonJS is the list of JavaScript assets to import by the web client
	// that are common to the frontend and the backend
	CommonJS = []string{
		"/static/web/lib/es5-shim/es5-shim.min.js",
		"/static/web/lib/underscore/underscore.js",
		"/static/web/lib/underscore.string/lib/underscore.string.js",
		"/static/web/lib/moment/moment.js",
		"/static/web/lib/jquery/jquery.js",
		"/static/web/lib/jquery.ui/jquery-ui.js",
		"/static/web/lib/jquery/jquery.browser.js",
		"/static/web/lib/jquery.blockUI/jquery.blockUI.js",
		"/static/web/lib/jquery.hotkeys/jquery.hotkeys.js",
		"/static/web/lib/jquery.placeholder/jquery.placeholder.js",
		"/static/web/lib/jquery.form/jquery.form.js",
		"/static/web/lib/jquery.ba-bbq/jquery.ba-bbq.js",
		"/static/web/lib/jquery.mjs.nestedSortable/jquery.mjs.nestedSortable.js",
		"/static/web/lib/bootstrap/js/affix.js",
		"/static/web/lib/bootstrap/js/alert.js",
		"/static/web/lib/bootstrap/js/button.js",
		"/static/web/lib/bootstrap/js/carousel.js",
		"/static/web/lib/bootstrap/js/collapse.js",
		"/static/web/lib/bootstrap/js/dropdown.js",
		"/static/web/lib/bootstrap/js/modal.js",
		"/static/web/lib/bootstrap/js/tooltip.js",
		"/static/web/lib/bootstrap/js/popover.js",
		"/static/web/lib/bootstrap/js/scrollspy.js",
		"/static/web/lib/bootstrap/js/tab.js",
		"/static/web/lib/bootstrap/js/transition.js",
		"/static/web/lib/qweb/qweb2.js",
		"/static/web/src/js/boot.js",
		"/static/web/src/js/config.js",
		"/static/web/src/js/framework/class.js",
		"/static/web/src/js/framework/translation.js",
		"/static/web/src/js/framework/ajax.js",
		"/static/web/src/js/framework/time.js",
		"/static/web/src/js/framework/mixins.js",
		"/static/web/src/js/framework/widget.js",
		"/static/web/src/js/framework/registry.js",
		"/static/web/src/js/framework/session.js",
		"/static/web/src/js/framework/model.js",
		"/static/web/src/js/framework/dom_utils.js",
		"/static/web/src/js/framework/utils.js",
		"/static/web/src/js/framework/qweb.js",
		"/static/web/src/js/framework/bus.js",
		"/static/web/src/js/services/core.js",
		"/static/web/src/js/framework/dialog.js",
		"/static/web/src/js/framework/local_storage.js",
		"/static/web/lib/bootstrap-datetimepicker/src/js/bootstrap-datetimepicker.js",
		"/static/web/lib/select2/select2.js",
	}
	// BackendLess is the list of Less assets to import by the web client
	// that are specific to the backend. All less assets are
	// cat'ed together in the given order before being compiled.
	BackendLess = []string{
		"/static/web/src/less/import_bootstrap.less",
		"/static/web/src/less/bootstrap_overridden.less",
		"/static/web/src/less/webclient_extra.less",
		"/static/web/src/less/webclient_layout.less",
		"/static/web/src/less/webclient.less",
		"/static/web/src/less/datepicker.less",
		"/static/web/src/less/progress_bar.less",
		"/static/web/src/less/dropdown.less",
		"/static/web/src/less/tooltip.less",
		"/static/web/src/less/debug_manager.less",
		"/static/web/src/less/control_panel.less",
		"/static/web/src/less/control_panel_layout.less",
		"/static/web/src/less/views.less",
		"/static/web/src/less/pivot_view.less",
		"/static/web/src/less/graph_view.less",
		"/static/web/src/less/tree_view.less",
		"/static/web/src/less/form_view_layout.less",
		"/static/web/src/less/form_view.less",
		"/static/web/src/less/list_view.less",
		"/static/web/src/less/search_view.less",
		"/static/web/src/less/modal.less",
		"/static/web/src/less/data_export.less",
		"/static/web/src/less/switch_company_menu.less",
		"/static/web/src/less/dropdown_extra.less",
		"/static/web/src/less/views_extra.less",
		"/static/web/src/less/form_view_extra.less",
		"/static/web/src/less/form_view_layout_extra.less",
		"/static/web/src/less/search_view_extra.less",
		"/static/web/src/less/main.less",
		"/static/web/src/less/responsive_navbar.less",
		"/static/web/src/less/app_drawer.less",
		"/static/web/src/less/responsive_form_view.less",
		"/static/web/src/less/responsive_variables.less",
		"/static/web/src/less/drawer.less",
		"/static/web/src/less/backend_variables.less",
		"/static/web/src/less/bootswatch.less",
		"/static/web/src/less/style.less",
		"/static/web/src/less/sidebar.less",
	}
	// BackendCSS is the list of CSS files to include without compilation for
	// the backend.
	BackendCSS = []string{
		"/static/web/lib/nvd3/nv.d3.css",
		"/static/web/lib/jquery.drawer/css/drawer.3.2.2.css",
	}
	// BackendJS is the list of JavaScript assets to import by the web client
	// that are specific to the backend.
	BackendJS = []string{
		"/static/web/lib/jquery.scrollTo/jquery.scrollTo.js",
		"/static/web/lib/nvd3/d3.v3.js",
		"/static/web/lib/nvd3/nv.d3.js",
		"/static/web/lib/backbone/backbone.js",
		"/static/web/lib/fuzzy-master/fuzzy.js",
		"/static/web/lib/py.js/lib/py.js",
		"/static/web/lib/jquery.ba-bbq/jquery.ba-bbq.js",
		"/static/web/src/js/framework/data_model.js",
		"/static/web/src/js/framework/formats.js",
		"/static/web/src/js/framework/view.js",
		"/static/web/src/js/framework/pyeval.js",
		"/static/web/src/js/action_manager.js",
		"/static/web/src/js/control_panel.js",
		"/static/web/src/js/view_manager.js",
		"/static/web/src/js/abstract_web_client.js",
		"/static/web/src/js/web_client.js",
		"/static/web/src/js/framework/data.js",
		"/static/web/src/js/compatibility.js",
		"/static/web/src/js/framework/misc.js",
		"/static/web/src/js/framework/crash_manager.js",
		"/static/web/src/js/framework/data_manager.js",
		"/static/web/src/js/services/crash_manager.js",
		"/static/web/src/js/services/data_manager.js",
		"/static/web/src/js/services/session.js",
		"/static/web/src/js/widgets/auto_complete.js",
		"/static/web/src/js/widgets/change_password.js",
		"/static/web/src/js/widgets/debug_manager.js",
		"/static/web/src/js/widgets/data_export.js",
		"/static/web/src/js/widgets/date_picker.js",
		"/static/web/src/js/widgets/loading.js",
		"/static/web/src/js/widgets/notification.js",
		"/static/web/src/js/widgets/sidebar.js",
		"/static/web/src/js/widgets/priority.js",
		"/static/web/src/js/widgets/progress_bar.js",
		"/static/web/src/js/widgets/pager.js",
		"/static/web/src/js/widgets/systray_menu.js",
		"/static/web/src/js/widgets/switch_company_menu.js",
		"/static/web/src/js/widgets/user_menu.js",
		"/static/web/src/js/menu.js",
		"/static/web/src/js/views/list_common.js",
		"/static/web/src/js/views/list_view.js",
		"/static/web/src/js/views/form_view.js",
		"/static/web/src/js/views/form_common.js",
		"/static/web/src/js/views/form_widgets.js",
		"/static/web/src/js/views/form_upgrade_widgets.js",
		"/static/web/src/js/views/form_relational_widgets.js",
		"/static/web/src/js/views/list_view_editable.js",
		"/static/web/src/js/views/pivot_view.js",
		"/static/web/src/js/views/graph_view.js",
		"/static/web/src/js/views/graph_widget.js",
		"/static/web/src/js/views/search_view.js",
		"/static/web/src/js/views/search_filters.js",
		"/static/web/src/js/views/search_inputs.js",
		"/static/web/src/js/views/search_menus.js",
		"/static/web/src/js/views/tree_view.js",
		"/static/web/src/js/apps.js",
		"/static/web/lib/bililite-range/bililiteRange.2.6.js",
		"/static/web/lib/jquery.sendkeys/jquery.sendkeys.4.js",
		"/static/web/lib/iscroll/iscroll-probe.5.2.0.js",
		"/static/web/lib/jquery.drawer/js/drawer.3.2.2.js",
		"/static/web/src/js/web_responsive.js",
		"/static/web/src/js/sidebar.js",
		"/static/web/src/js/sidebar-toggle.js",
	}
	// FrontendLess is the list of Less assets to import by the web client
	// that are specific to the frontend. All less assets are
	// cat'ed together in the given order before being compiled.
	FrontendLess = []string{
		"/static/web/src/less/import_bootstrap.less",
	}
	// FrontendCSS is the list of CSS files to include without compilation for
	// the frontend.
	FrontendCSS []string
	// FrontendJS is the list of JavaScript assets to import by the web client
	// that are specific to the frontend.
	FrontendJS = []string{
		"/static/web/src/js/services/session.js",
	}
	// LessHelpers are less files that must be imported for compiling any assets
	LessHelpers = []string{
		"/static/web/lib/bootstrap/less/variables.less",
		"/static/web/lib/bootstrap/less/mixins/vendor-prefixes.less",
		"/static/web/lib/bootstrap/less/mixins/buttons.less",
		"/static/web/src/less/variables.less",
		"/static/web/src/less/utils.less",
	}
	// FrontendContext is the base context to update when rendering
	// a frontend HTML template.
	FrontendContext = hweb.Context{
		"commonCSS":           CommonCSS,
		"commonCompiledCSS":   commonCSSRoute,
		"commonJS":            CommonJS,
		"frontendCSS":         FrontendCSS,
		"frontendCompiledCSS": frontendCSSRoute,
		"frontendJS":          FrontendJS,
	}
)

func init() {
	log = logging.GetLogger("web/controllers")
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
			webClient.AddController(http.MethodGet, "/qweb", QWeb)
			webClient.AddController(http.MethodGet, "/locale", LoadLocale)
			webClient.AddController(http.MethodGet, "/locale/:lang", LoadLocale)
			webClient.AddController(http.MethodPost, "/translations", BootstrapTranslations)
			webClient.AddController(http.MethodPost, "/bootstrap_translations", BootstrapTranslations)
			webClient.AddController(http.MethodPost, "/csslist", CSSList)
			webClient.AddController(http.MethodPost, "/jslist", JSList)
			webClient.AddController(http.MethodPost, "/version_info", VersionInfo)
		}
		dataset := web.AddGroup("/dataset")
		{
			dataset.AddController(http.MethodPost, "/call_kw/*path", CallKW)
			dataset.AddController(http.MethodPost, "/search_read", SearchRead)
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
