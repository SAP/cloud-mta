[![CircleCI](https://circleci.com/gh/SAP/cloud-mta.svg?style=svg)](https://circleci.com/gh/SAP/cloud-mta)
[![Go Report Card](https://goreportcard.com/badge/github.com/SAP/cloud-mta)](https://goreportcard.com/report/github.com/SAP/cloud-mta)
[![Coverage Status](https://coveralls.io/repos/github/SAP/cloud-mta/badge.svg?branch=CD)](https://coveralls.io/github/SAP/cloud-mta?branch=CD)
[![contributions welcome](https://img.shields.io/badge/contributions-welcome-brightgreen.svg?style=flat)](https://github.com/SAP/cloud-mta/blob/master/.github/CONTRIBUTING.md)
[![GoDoc](https://godoc.org/github.com/SAP/cloud-mta?status.svg)](https://godoc.org/github.com/SAP/cloud-mta/mta)
![pre-alpha](https://img.shields.io/badge/Release-pre--alpha-orange.svg)
[![REUSE status](https://api.reuse.software/badge/github.com/SAP/cloud-mta)](https://api.reuse.software/info/github.com/SAP/cloud-mta)

                   
## Description

MTA tool for exploring and validating the multitarget application descriptor (`mta.yaml`).

The tool commands (APIs) are used to do the following:

   - Explore the structure of the `mta.yaml` file objects, such as retrieving a list of resources required by a specific module.
   - Validate an `mta.yaml` file against a specified schema version.
   - Ensure semantic correctness of an `mta.yaml` file, such as the uniqueness of module/resources names, the resolution of requires/provides pairs, and so on.
   - Validate the descriptor against the project folder structure, such as the `path` attribute reference in an existing project folder.
   - Get data for constructing a deployment MTA descriptor, such as deployment module types.
   

 ### Multitarget Applications

A multitarget application is a package comprised of multiple application and resource modules that have been created using different technologies and deployed to different run-times; however, they have a common life cycle. A user can bundle the modules together using the `mta.yaml` file, describe them along with their inter-dependencies to other modules, services, and interfaces, and package them in an MTA project.
 

## Requirements

* [Go (version > 1.13.x)](https://golang.org/dl/) 

## Download and Installation

1.  Set your [workspace](https://golang.org/doc/code.html#Workspaces).

2.  Download and install it:

    ```sh
    $ go get github.com/SAP/cloud-mta/mta
    ```

## Usage

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

## Contributions

Contributions are greatly appreciated.
See [CONTRIBUTING.md](https://github.com/SAP/cloud-mta/blob/master/.github/CONTRIBUTING.md) for details.

## Known Issues

No known major issues.  To report a new issue, please use our GitHub bug tracking system.

## Support

Please follow our [issue template](https://github.com/SAP/cloud-mta/blob/master/.github/ISSUE_TEMPLATE/bug_report.md) on how to report an issue.

