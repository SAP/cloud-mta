const { join } = require("path");
const { readFileSync, writeFileSync } = require("fs");

// Convert the package.json to the mta-local package.
// This package is only used for downloading and running the mta executable locally and not for installing it.

const pkgJsonPath = join(__dirname, "package.json");
// Read the original representation of the pkg.json
const pkgJsonOrgStr = readFileSync(pkgJsonPath, "utf8");
const pkgJson = JSON.parse(pkgJsonOrgStr);

// This package doesn't download the binary on install, and doesn't add scripts in the ".bin" folder.
pkgJson.name = "mta-local";
delete pkgJson.bin;
delete pkgJson.scripts.install;

writeFileSync(pkgJsonPath, JSON.stringify(pkgJson, undefined, 2));