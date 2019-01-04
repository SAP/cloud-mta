[![CircleCI](https://circleci.com/gh/SAP/cloud-mta.svg?style=svg)](https://circleci.com/gh/SAP/cloud-mta)
[![Go Report Card](https://goreportcard.com/badge/github.com/SAP/cloud-mta)](https://goreportcard.com/report/github.com/SAP/cloud-mta)
![GitHub license](https://img.shields.io/badge/license-Apache_2.0-blue.svg)
![pre-alpha](https://img.shields.io/badge/Release-pre--alpha-orange.svg)


> <b>Disclaimer</b>: The MTA package is under heavy development and is currently in a `pre-alpha` stage.  Some functionality is still missing and the APIs are subject to change; use at your own risk.
                   
# MTA

MTA tool for exploring and validating the multi-target application descriptor (`mta.yaml`).

The tool commands (APIs) are used to do the following:

   - Explore the structure of the `mta.yaml` file objects, such as retrieving a list of resources required by a specific module.
   - Validate an `mta.yaml` file against a specified schema version.
   - Ensure semantic correctness of an `mta.yaml` file, such as the uniqueness of module/resources names, the resolution of requires/provides pairs, and so on.
   - Validate the descriptor against the project folder structure, such as the `path` attribute reference in an existing project folder.
   - Get data for constructing a deployment MTA descriptor, such as deployment module types.
   

### Multi-Target Applications

A multi-target application is a package comprised of multiple application and resource modules that have been created using different technologies and deployed to different run-times; however, they have a common life cycle. A user can bundle the modules together using the `mta.yaml` file, describe them along with their inter-dependencies to other modules, services, and interfaces, and package them in an MTA project.
 
## Roadmap 

### Milestone 1  (Q1 - 2019)
 
 - [x] Supports the MTA parser 
 - [x] Supports development descriptor schema validations (2.1) 
 - [ ] Supports semantic validations (MTA->project)
 - [x] Supports `path` validation
 - [ ] Add MTA API's usage documentation  
 
### Milestone 2 (Q1 - 2019)
 
- [ ] Supports semantic validations (MTA)
- [ ] Supports uniqueness of module and resource names
- [ ] Supports multiple schema 
- [ ] Supports advanced `mta.yaml` file (3.1, 3.2) schemas

 
### Milestone 3 (Q2 - 2019)

- [ ] Supports updating scenarios, such as add module/resource, add module property, add dependency, and so on
- [ ] Supports placeholder resolution

## Requirements

* [Go (version > 1.11.x)](https://golang.org/dl/) 

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
    // returns the module properties.
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

## License
 
Copyright (c) 2019 SAP SE or an SAP affiliate company. All rights reserved.

This file is licensed under the Apache 2.0 License [except as noted otherwise in the LICENSE file](/LICENSE).

