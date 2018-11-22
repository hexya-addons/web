hexya.define('web.session', function (require) {
"use strict";

var Session = require('web.Session');
var modules = hexya._modules;

var session = new Session(undefined, undefined, {modules: modules, use_cors: false});
session.is_bound = session.session_bind();

return session;

});
