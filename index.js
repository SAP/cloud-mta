var binwrap = require('binwrap');

var packageInfo = require("./package.json");
var version = packageInfo.version;
var root = `https://github.com/SAP/cloud-mta/releases/download/v${version}/cloud-mta_${version}_`;

module.exports = binwrap({
    dirname: __dirname,
    binaries: [
        'mta'
    ],
    urls: {
        'darwin-arm64': root + 'Darwin_arm64.tar.gz',
        'darwin-x64': root + 'Darwin_amd64.tar.gz',
        'linux-x64': root + 'Linux_amd64.tar.gz',
        'win32-x64': root + 'Windows_amd64.tar.gz'
    }
});