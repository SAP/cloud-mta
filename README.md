[![CircleCI](https://circleci.com/gh/SAP/cloud-mta.svg?style=svg)](https://circleci.com/gh/SAP/cloud-mta)
[![Go Report Card](https://goreportcard.com/badge/github.com/SAP/cloud-mta)](https://goreportcard.com/report/github.com/SAP/cloud-mta)
[![Coverage Status](https://coveralls.io/repos/github/SAP/cloud-mta/badge.svg?branch=CD)](https://coveralls.io/github/SAP/cloud-mta?branch=CD)
[![contributions welcome](https://img.shields.io/badge/contributions-welcome-brightgreen.svg?style=flat)](https://github.com/SAP/cloud-mta/blob/master/.github/CONTRIBUTING.md)
[![GoDoc](https://godoc.org/github.com/SAP/cloud-mta?status.svg)](https://godoc.org/github.com/SAP/cloud-mta/mta)
![pre-alpha](https://img.shields.io/badge/Release-pre--alpha-orange.svg)
[![dependentbot](https://api.dependabot.com/badges/status?host=github&repo=SAP/cloud-mta)](https://dependabot.com/)
[![REUSE status](https://api.reuse.software/badge/github.com/SAP/cloud-mta)](https://api.reuse.software/info/github.com/SAP/cloud-mta)

                   
# Description

MTA tool for exploring and validating the multitarget application descriptor (`mta.yaml`).

The tool can be used as a Go library or as a command-line tool, also available as an npm package.

## Multitarget Applications

A multitarget application is a package comprised of multiple application and resource modules that have been created using different technologies and deployed to different run-times; however, they have a common life cycle. A user can bundle the modules together using the `mta.yaml` file, describe them along with their inter-dependencies to other modules, services, and interfaces, and package them in an MTA project.

## Go Library

The tool commands (APIs) are used to do the following:

   - Explore the structure of the `mta.yaml` file objects, such as retrieving a list of resources required by a specific module.
   - Validate an `mta.yaml` file against a specified schema version.
   - Ensure semantic correctness of an `mta.yaml` file, such as the uniqueness of module/resources names, the resolution of requires/provides pairs, and so on.
   - Validate the descriptor against the project folder structure, such as the `path` attribute reference in an existing project folder.
   - Get data for constructing a deployment MTA descriptor, such as deployment module types.
   
### Requirements

* [Go (version > 1.13.x)](https://golang.org/dl/) 

### Download and Installation

1.  Set your [workspace](https://golang.org/doc/code.html#Workspaces).

2.  Download and install it:

    ```sh
    $ go get github.com/SAP/cloud-mta/mta
    ```

### Usage

 - Import it into your source code:

    ```go
    import "github.com/SAP/cloud-mta/mta"
    ```

 -  Quick start example:

    ```go
    
    // sets the path to the MTA project.
    mf, _ := ioutil.ReadFile("/path/mta.yaml")
    // Returns an MTA object.
    if err != nil {
    	return err
    }
    // unmarshal MTA content.
    m := Unmarshal(mf)
    if err != nil {
    	return err
    }
    // Returns the module properties.
    module, err := m.GetModuleByName(moduleName)
    if err != nil {
    	return err
    }
    ```

## Command-Line Tool

Some of the tool's features are available as an command-line tool, which can be downloaded from the GitHub releases page or installed as an npm package.

The commands of the CLI tool are used as APIs by other programs, such as the `mta-lib` npm package which exposes Javascript APIs for reading and manipulating the `mta.yaml` file.  

### npm package

#### `mta`
The `mta` npm package installs the executable and allows you to run it from a shell or command line.
You can install it globally via the command:
```shell script
npm install -g mta
```

#### `mta-local`
The `mta-local` npm package exposes the same CLI tool without installing it globally. It is packaged by other libraries, and it provides a way to lazily download the executable according to the current operating system and run it.
You can use it in the following way:
```javascript
// You can use "cross-spawn" library instead of "process" for compatibility to Windows systems
const { spawn } = require("process");
const mtaPath = require("mta-local").paths["mta"];
const childProcess = spawn(mtaPath, args);
// Handle the process events ...
```

#### Packaging with webpack
To use these npm libraries from an application packaged with webpack, you have to copy the `bin/mta` file to the webpack output directory (keeping the same file structure), make it executable and enable `__dirname` to be used.

**Note:** while the packaged `bin/mta` file is already executable, the `CopyWebpackPlugin` loses the executable bits during the copy. See [this issue](https://github.com/webpack-contrib/copy-webpack-plugin/issues/35).

For example, if the results are in the `dist` folder, add this configuration inside your webpack configuration file:
```javascript
const path = require("path");
const fs = require("fs");
const CopyWebpackPlugin = require("copy-webpack-plugin");
const config = {
    // ...
    node: {
        __dirname: false,
    },
    plugins: [
        new CopyWebpackPlugin({
            patterns: [
                {
                  from: path.join(require.resolve("mta-local"), "..", "bin"),
                  to: path.resolve(__dirname, "dist", "bin"),
                }
            ]
        }),
        function (compiler) {
          compiler.hooks.done.tap("ExecuteChmodOnBinMta", () => {
            fs.chmodSync(path.resolve(__dirname, "dist", "bin", "mta"), "755");
          });
        }
    ]
};
```

**Note:** if you did not previously use `copy-webpack-plugin` you will need to add it to the `devDependencies` in your `package.json` file.

## Contributions

Contributions are greatly appreciated.
See [CONTRIBUTING.md](https://github.com/SAP/cloud-mta/blob/master/.github/CONTRIBUTING.md) for details.

## Known Issues

No known major issues.  To report a new issue, please use our GitHub bug tracking system.

## Support

Please follow our [issue template](https://github.com/SAP/cloud-mta/blob/master/.github/ISSUE_TEMPLATE/bug_report.md) on how to report an issue.

