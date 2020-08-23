package resolver

import (
	"encoding/json"
	"strings"

	"github.com/SAP/cloud-mta/mta"
)

type vcapServices map[string][]VcapService

//VcapService - vcap service struct
type VcapService struct {
	Name         string   `json:"name"`
	InstanceName string   `json:"instance_name"`
	Label        string   `json:"label"`
	Tags         []string `json:"tags"`
	Plan         string   `json:"plan"`
}

const tagResourceNamePrefix = "mta-resource-name:"

// ResolveContext holds context info during resolving of properties
type ResolveContext struct {
	global    map[string]string
	modules   map[string]map[string]string
	resources map[string]map[string]string
}

func (m *MTAResolver) addServiceNames(module *mta.Module) {
	vcapServices := m.getvcapServicesFromEnv()
	if vcapServices == nil {
		return
	}

	for _, resource := range m.Resources {
		resCtx := m.context.resources[resource.Name]

		//try to find the service-name in VCAP_SERVICES
		serviceName := findServiceInvcapServices(vcapServices, resource)
		if len(serviceName) > 0 {
			resCtx["service-name"] = serviceName
		}
	}
}

func (m *MTAResolver) getvcapServicesFromEnv() *vcapServices {
	vcap := m.context.global["VCAP_SERVICES"]
	if len(vcap) > 0 {
		var vcapSrv vcapServices
		err := json.Unmarshal([]byte(vcap), &vcapSrv)
		if err == nil {
			return &vcapSrv

		}
	}
	return nil
}

func findServiceInvcapServices(vcapServices *vcapServices, resource *mta.Resource) string {
	//look for the resource name as a tag in the service instances:
	for _, vcapServiceArray := range *vcapServices {
		for _, vcapService := range vcapServiceArray {
			for _, tag := range vcapService.Tags {
				pos := strings.Index(tag, tagResourceNamePrefix)
				if pos == 0 && tag[len(tagResourceNamePrefix):] == resource.Name {
					return vcapService.Name
				}
			}
		}
	}
	return ""
}
