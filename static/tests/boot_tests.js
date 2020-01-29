(function () {
"use strict";

hexya.__DEBUG__.didLogInfo.then(function () {
    var modulesInfo = hexya.__DEBUG__.jsModules;

    QUnit.module('Hexya JS Modules');

    QUnit.test('all modules are properly loaded', function (assert) {
        assert.expect(2);

        assert.deepEqual(modulesInfo.missing, [],
            "no js module should be missing");
        assert.deepEqual(modulesInfo.failed, [],
            "no js module should have failed");
    });
});
})();
