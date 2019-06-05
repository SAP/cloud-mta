package commands

import (
	"encoding/json"

	"strings"

	"github.com/SAP/cloud-mta/mta"
)

type VcapServices map[string][]VcapService

//VcapService - vcap service struct
type VcapService struct {
	Name         string   `json:"name"`
	InstanceName string   `json:"instance_name"`
	Label        string   `json:"label"`
	Tags         []string `json:"tags"`
	Plan         string   `json:"plan"`
}

const TagResourceNamePrefix = "mta-resource-name:"

type ParamSource struct {
	Parameters map[string]interface{}
}

type ResolveContext struct {
	global    map[string]string
	modules   map[string]map[string]string
	resources map[string]map[string]string
}

func (m *MTAResolver) addServiceNames(module *mta.Module) {

	for _, resource := range m.Resources {
		if m.context.resources[resource.Name] == nil {
			m.context.resources[resource.Name] = map[string]string{}
		}
		resCtx := m.context.resources[resource.Name]

		//try to find the service-name in VCAP_SERVICES
		serviceName := m.getServiceInstanceFromEnv(resource)
		if len(serviceName) > 0 {
			resCtx["service-name"] = serviceName
		}
	}
}

func (m *MTAResolver) getServiceInstanceFromEnv(resource *mta.Resource) string {
	vcap := m.context.global["VCAP_SERVICES"]
	if len(vcap) > 0 {
		var vcapServices VcapServices
		err := json.Unmarshal([]byte(vcap), &vcapServices)
		if err == nil {
			//look for the resource name as a tag in the service instances:
			for _, vcapServiceArray := range vcapServices {
				for _, vcapService := range vcapServiceArray {
					for _, tag := range vcapService.Tags {
						pos := strings.Index(tag, TagResourceNamePrefix)
						if pos == 0 && tag[len(TagResourceNamePrefix):] == resource.Name {
							return vcapService.Name
						}
					}
				}
			}
		}
	}
	return ""
}
