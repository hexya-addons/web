hexya.define('web.public.Notification', function (require) {
'use strict';

var Notification = require('web.Notification');

Notification.include({
    xmlDependencies: ['/static/web/src/xml/notification.xml'],
});
});
