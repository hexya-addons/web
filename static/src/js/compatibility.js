// ------------------------------------------------------------------------------
// Compatibility with Hexya v8.
// 
// With the new module system, no global variable can (and should) be accessed
// in hexyaerp.  This file exports everything, to mimic the previous global
// namespace structure.  This is only supposed to be used by 3rd parties to 
// facilitate migration.  Hexya addons should not use the 'hexyaerp' variable at
// all.
// ------------------------------------------------------------------------------
hexya.define('web.compatibility', function (require) {
"use strict";

var ActionManager = require('web.ActionManager');
var core = require('web.core');
var data = require('web.data');
var Dialog = require('web.Dialog');
var FavoriteMenu = require('web.FavoriteMenu');
var form_common = require('web.form_common');
var formats = require('web.formats');
var FormView = require('web.FormView');
var form_relational = require('web.form_relational'); // necessary
var form_widgets = require('web.form_widgets'); // necessary
var framework = require('web.framework');
var ListView = require('web.ListView');
var Model = require('web.DataModel');
var pyeval = require('web.pyeval');
var Registry = require('web.Registry');
var SearchView = require('web.SearchView');
var session = require('web.session');
var Sidebar = require('web.Sidebar');
var SystrayMenu = require('web.SystrayMenu');
var time = require('web.time');
var UserMenu = require('web.UserMenu');
var utils = require('web.utils');
var View = require('web.View');
var ViewManager = require('web.ViewManager');
var WebClient = require('web.WebClient');
var Widget = require('web.Widget');

var client_started = $.Deferred();

var OldRegistry = Registry.extend({
    add: function (key, path) {
    },
    get_object: function (key) {
        return get_object(this.map[key]);
    },
});

window.hexyaerp = window.hexyaerp || {};

$.Mutex = utils.Mutex;
hexyaerp._session_id = "instance0";
hexyaerp._t = core._t;
hexyaerp.get_cookie = utils.get_cookie;

hexyaerp.qweb = core.qweb;
hexyaerp.session = session;

hexyaerp.web = hexyaerp.web || {};
hexyaerp.web._t = core._t;
hexyaerp.web._lt = core._lt;

hexyaerp.web.ActionManager = ActionManager;
hexyaerp.web.auto_str_to_date = time.auto_str_to_date;
hexyaerp.web.blockUI = framework.blockUI;
hexyaerp.web.BufferedDataSet = data.BufferedDataSet;
hexyaerp.web.bus = core.bus;
hexyaerp.web.Class = core.Class;
hexyaerp.web.client_actions = make_old_registry(core.action_registry);
hexyaerp.web.CompoundContext = data.CompoundContext;
hexyaerp.web.CompoundDomain = data.CompoundDomain;
hexyaerp.web.DataSetSearch = data.DataSetSearch;
hexyaerp.web.DataSet = data.DataSet;
hexyaerp.web.date_to_str = time.date_to_str;
hexyaerp.web.Dialog = Dialog;
hexyaerp.web.DropMisordered = utils.DropMisordered;

hexyaerp.web.form = hexyaerp.web.form || {};
hexyaerp.web.form.AbstractField = form_common.AbstractField;
hexyaerp.web.form.compute_domain = data.compute_domain;
hexyaerp.web.form.DefaultFieldManager = form_common.DefaultFieldManager;
hexyaerp.web.form.FieldChar = core.form_widget_registry.get('char');
hexyaerp.web.form.FieldFloat = core.form_widget_registry.get('float');
hexyaerp.web.form.FieldStatus = core.form_widget_registry.get('statusbar');
hexyaerp.web.form.FieldMany2ManyTags = core.form_widget_registry.get('many2many_tags');
hexyaerp.web.form.FieldMany2One = core.form_widget_registry.get('many2one');
hexyaerp.web.form.FormWidget = form_common.FormWidget;
hexyaerp.web.form.tags = make_old_registry(core.form_tag_registry);
hexyaerp.web.form.widgets = make_old_registry(core.form_widget_registry);
hexyaerp.web.form.custom_widgets = make_old_registry(core.form_custom_registry);

hexyaerp.web.format_value = formats.format_value;
hexyaerp.web.FormView = FormView;

hexyaerp.web.json_node_to_xml = utils.json_node_to_xml;

hexyaerp.web.ListView = ListView;
hexyaerp.web.Model = Model;
hexyaerp.web.normalize_format = time.strftime_to_moment_format;
hexyaerp.web.py_eval = pyeval.py_eval;
hexyaerp.web.pyeval = pyeval;
hexyaerp.web.qweb = core.qweb;

hexyaerp.web.Registry = OldRegistry;

hexyaerp.web.search = {};
hexyaerp.web.search.FavoriteMenu = FavoriteMenu;
hexyaerp.web.SearchView = SearchView;
hexyaerp.web.Sidebar = Sidebar;
hexyaerp.web.str_to_date = time.str_to_date;
hexyaerp.web.str_to_datetime = time.str_to_datetime;
hexyaerp.web.SystrayItems = SystrayMenu.Items;
hexyaerp.web.unblockUI = framework.unblockUI;
hexyaerp.web.UserMenu = UserMenu;
hexyaerp.web.View = View;
hexyaerp.web.ViewManager = ViewManager;
hexyaerp.web.views = make_old_registry(core.view_registry);
hexyaerp.web.WebClient = WebClient;
hexyaerp.web.Widget = Widget;

hexyaerp.Widget = hexyaerp.web.Widget;
hexyaerp.Widget.prototype.session = session;


WebClient.include({
    init: function () {
        hexyaerp.client = this;
        hexyaerp.webclient = this;
        start_modules();
        client_started.resolve();
        this._super.apply(this, arguments);
    },
});


function make_old_registry(registry) {
    return {
        add: function (key, path) {
            client_started.done(function () {
                registry.add(key, get_object(path));
            });
        },
    };
}
function get_object(path) {
    var object_match = hexyaerp;
    path = path.split('.');
    // ignore first section
    for(var i=1; i<path.length; ++i) {
        object_match = object_match[path[i]];
    }
    return object_match;
}

/**
 * Hexya instance constructor
 *
 * @param {Array|String} modules list of modules to initialize
 */
var inited = false;
function start_modules (modules) {
    if (modules === undefined) {
        modules = hexya._modules;
    }
    modules = _.without(modules, "web");
    if (inited) {
        throw new Error("Hexya was already inited");
    }
    inited = true;
    for(var i=0; i < modules.length; i++) {
        var fct = hexyaerp[modules[i]];
        if (typeof(fct) === "function") {
            hexyaerp[modules[i]] = {};
            for (var k in fct) {
                hexyaerp[modules[i]][k] = fct[k];
            }
            fct(hexyaerp, hexyaerp[modules[i]]);
        }
    }
    hexyaerp._modules = ['web'].concat(modules);
    return hexyaerp;
};

// Monkey-patching of the ListView for backward compatibiliy of the colors and
// fonts row's attributes, as they are deprecated in 9.0.
ListView.include({
    willStart: function() {
        if (this.fields_view.arch.attrs.colors) {
            this.colors = _(this.fields_view.arch.attrs.colors.split(';')).chain()
                .compact()
                .map(function(color_pair) {
                    var pair = color_pair.split(':'),
                        color = pair[0],
                        expr = pair[1];
                    return [color, py.parse(py.tokenize(expr)), expr];
                }).value();
        }

        if (this.fields_view.arch.attrs.fonts) {
            this.fonts = _(this.fields_view.arch.attrs.fonts.split(';')).chain().compact()
                .map(function(font_pair) {
                    var pair = font_pair.split(':'),
                        font = pair[0],
                        expr = pair[1];
                    return [font, py.parse(py.tokenize(expr)), expr];
                }).value();
        }

        return this._super();
    },
    /**
     * Returns the style for the provided record in the current view (from the
     * ``@colors`` and ``@fonts`` attributes)
     *
     * @param {Record} record record for the current row
     * @returns {String} CSS style declaration
     */
    style_for: function (record) {
        var len, style= '';

        var context = _.extend({}, record.attributes, {
            uid: session.uid,
            current_date: moment().format('YYYY-MM-DD')
            // TODO: time, datetime, relativedelta
        });
        var i;
        var pair;
        var expression;
        if (this.fonts) {
            for(i=0, len=this.fonts.length; i<len; ++i) {
                pair = this.fonts[i];
                var font = pair[0];
                expression = pair[1];
                if (py.PY_isTrue(py.evaluate(expression, context))) {
                    switch(font) {
                    case 'bold':
                        style += 'font-weight: bold;';
                        break;
                    case 'italic':
                        style += 'font-style: italic;';
                        break;
                    case 'underline':
                        style += 'text-decoration: underline;';
                        break;
                    }
                }
            }
        }
 
        if (!this.colors) { return style; }
        for(i=0, len=this.colors.length; i<len; ++i) {
            pair = this.colors[i];
            var color = pair[0];
            expression = pair[1];
            if (py.PY_isTrue(py.evaluate(expression, context))) {
                return style += 'color: ' + color + ';';
            }
        }
        return style;
     },
});

// IE patch
//-------------------------------------------------------------------------
if (typeof(console) === "undefined") {
    // Even IE9 only exposes console object if debug window opened
    window.console = {};
    ('log error debug info warn assert clear dir dirxml trace group'
        + ' groupCollapsed groupEnd time timeEnd profile profileEnd count'
        + ' exception').split(/\s+/).forEach(function(property) {
            console[property] = _.identity;
    });
}

/**
    Some hack to make placeholders work in ie9.
*/
if (!('placeholder' in document.createElement('input'))) {
    document.addEventListener("DOMNodeInserted",function(event){
        var nodename =  event.target.nodeName.toLowerCase();
        if ( nodename === "input" || nodename === "textarea" ) {
            $(event.target).placeholder();
        }
    });
}

});
