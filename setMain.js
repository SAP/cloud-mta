const { join } = require("path");
const { readFileSync, writeFileSync } = require("fs");

// Set the "main" of the package,
// If a parameter is passed, we set it to the parameter value.
// Otherwise we set it to the default value (index.js).

const pkgJsonPath = join(__dirname, "package.json");
// Read the original representation of the pkg.json
const pkgJsonOrgStr = readFileSync(pkgJsonPath, "utf8");
const pkgJson = JSON.parse(pkgJsonOrgStr);

pkgJson.main = process.argv[2] || "index.js";

writeFileSync(pkgJsonPath, JSON.stringify(pkgJson, undefined, 2));