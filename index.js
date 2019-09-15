var binwrap = require('binwrap');
var path = require('path');

var packageInfo = require(path.join(__dirname, 'package.json'));
var version = packageInfo.version;
var root = `https://github.com/SAP/cloud-mta/releases/download/v${version}/cloud-mta_${version}_`;

module.exports = binwrap({
    dirname: __dirname,
    binaries: [
        'mta'
    ],
    urls: {
        'darwin-x64': root + 'Darwin_amd64.tar.gz',
        'linux-x64': root + 'Linux_amd64.tar.gz',
        'win32-x64': root + 'Windows_amd64.tar.gz'
    }
});