"use strict";

const webpack = require("webpack");
const path = require("path");
const fs = require("fs");
const distPath = path.resolve(__dirname, "dist");

/**@type {import('webpack').Configuration}*/
const config = {
  target: "node", // vscode extensions run in a Node.js-context https://webpack.js.org/configuration/node/

  entry: {
    "index.js": "./index.js",
    "bin/mta": "./bin/mta", // mta binary runner
  },
  output: {
    // the bundle is stored in the 'dist' folder (check package.json), https://webpack.js.org/configuration/output/
    path: distPath,
    filename: "[name]",
    libraryTarget: "commonjs2",
  },
  module: {
    rules: [
      {
        // Shebangs are not supported by webpack
        test: /bin[/\\]mta$/g,
        loader: "shebang-loader",
      },
      // Fix path to unpacked binary used in bin/mta (the path is wrong because the end result is placed
      // inside dist/bin and binwrap expects it to be in bin - otherwise the pre-installed mta executable will not be used)
      {
        test: /bin[/\\]mta$/g,
        loader: "string-replace-loader",
        options: {
          search: 'path.join(__dirname, "..", "unpacked_bin")',
          replace: 'path.join(__dirname, "..", "..", "unpacked_bin")',
          strict: true,
        },
      },
      // Fix expressions in require used in bin/mta (file generated by binwrap) and binwrap
      {
        test: /bin[/\\]mta$/g,
        loader: "string-replace-loader",
        options: {
          search: 'require(path.join(__dirname, "..", "package.json"))',
          replace: 'require("../package.json")',
          strict: true,
        },
      },
      {
        test: /bin[/\\]mta$/g,
        loader: "string-replace-loader",
        options: {
          search: 'require(path.join(__dirname, "..", packageInfo.main))',
          replace: 'require("../index.js")',
          strict: true,
        },
      },
      {
        test: /binwrap[/\\]install\.js$/g,
        loader: "string-replace-loader",
        options: {
          search: /require\(path\.join\(__dirname, "(.+)"\)\)/g,
          replace: 'require("./$1")',
          strict: true,
        },
      },
      {
        test: /binwrap[/\\]index\.js$/g,
        loader: "string-replace-loader",
        options: {
          search: /require\(path\.join\(__dirname, "(.+)"\)\)/g,
          replace: 'require("./$1")',
          strict: true,
        },
      },
    ],
  },
  node: {
    // Use the actual __dirname so we can resolve the path to bin/mta from index.js
    // and __filename so bin/mta will write the correct log message when downloading mta executable
    __dirname: false,
    __filename: false,
  },
  plugins: [
    // Add the shebang on the bin/mta file since it's removed by the shebang-loader
    new webpack.BannerPlugin({
      banner: "#!/usr/bin/env node",
      raw: true,
      entryOnly: true,
      test: "bin/mta",
    }),
    function (compiler) {
      compiler.hooks.done.tap("ExecuteChmodOnBinMta", () => {
        // bin/mta should be executable
        fs.chmodSync(path.resolve(distPath, "bin", "mta"), "755");
      });
    }
  ],
};
module.exports = config;
