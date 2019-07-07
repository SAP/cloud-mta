package resolver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/SAP/cloud-mta/mta"
)

func callResolveAndGetOutput(wd, moduleName, yamlPath string) string {
	reader, writer, err := os.Pipe()
	Ω(err).Should(Succeed())
	stdout := os.Stdout
	stderr := os.Stderr
	defer func() {
		os.Stdout = stdout
		os.Stderr = stderr
		log.SetOutput(os.Stderr)
	}()
	os.Stdout = writer
	os.Stderr = writer
	log.SetOutput(writer)
	out := make(chan string)
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		var buf bytes.Buffer
		wg.Done()
		io.Copy(&buf, reader)
		out <- buf.String()
	}()
	wg.Wait()
	err = Resolve(wd, moduleName, yamlPath)
	Ω(err).Should(Succeed())
	writer.Close()
	return <-out
}

func getExpected(expected []string) string {
	reader, writer, err := os.Pipe()
	Ω(err).Should(Succeed())
	stdout := os.Stdout
	defer func() {
		os.Stdout = stdout
	}()
	os.Stdout = writer
	log.SetOutput(writer)
	out := make(chan string)
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		var buf bytes.Buffer
		wg.Done()
		io.Copy(&buf, reader)
		out <- buf.String()
	}()
	wg.Wait()
	for _, exp := range expected {
		fmt.Println(exp)
	}
	writer.Close()
	return <-out
}

func callResolveAndValidateOutput(wd, moduleName, yamlPath string, expected []string) {
	actualStr := callResolveAndGetOutput(wd, moduleName, yamlPath)
	expectedStr := getExpected(expected)
	Ω(len(actualStr)).Should(Equal(len(expectedStr)))
	for _, exp := range expected {
		Ω(actualStr).Should(ContainSubstring(exp))
	}
}

var _ = Describe("Resolve", func() {
	expected := []string{
		`prop1=no_placeholders`,
		`prop2=1000m`,
		`prop3=["1000m","1m"]`,
		`prop4={"p1":"1000m","p2":"1m"}`,
		`prop5=1`,
		`prop6={"1":"1000m"}`,
		`prop7=~{eb-msahaa/heap`,
		`prop8=[[{"a":["a1",{"a2-key":"a2-value"}]}]]`,
		`prop9=vvv`,
		`prop10=${env_var0}`,
		`prop11=2G`,
		`JBP_CONFIG_companyJVM=[ memory_calculator: { memory_sizes: { heap: 1000m, stack: 1m, metaspace: 150m } } ]`,
		`JBP_CONFIG_companyJVM1=[ memory_calculator: { memory_sizes: { heap: 1000m, stack: 1m, metaspace: 150m } } ]`,
		`JBP_CONFIG_RESOURCE_CONFIGURATION=[tomcat/webapps/ROOT/META-INF/context.xml: {"service_name_for_DefaultDB" : "ed-aaa-service"}]`,
	}

	It("Sanity", func() {
		wd := getTestPath("test-project")
		yamlPath := getTestPath("test-project", "mta.yaml")
		envGetter = mockEnvGetterWithVcapServices
		callResolveAndValidateOutput(wd, "eb-java", yamlPath, expected)
	})
	It("Sanity - working dir not provided", func() {
		yamlPath := getTestPath("test-project", "mta.yaml")
		envGetter = mockEnvGetterExtWithVcapServices
		callResolveAndValidateOutput("", "eb-java", yamlPath, expected)

	})
	It("Sanity - working dir not provided, no VCAP services", func() {
		yamlPath := getTestPath("test-project", "mta.yaml")
		envGetter = mockEnvGetterExt
		expected[len(expected)-1] = strings.Replace(expected[len(expected)-1], "ed-aaa-service", "${service-name}", -1)
		callResolveAndValidateOutput("", "eb-java", yamlPath, expected)
	})
	It("empty module name", func() {
		err := Resolve("", "", getTestPath("test-project", "mta.yaml"))
		Ω(err).Should(HaveOccurred())
		Ω(err.Error()).Should(Equal(emptyModuleNameMsg))
	})
	It("module not exists", func() {
		err := Resolve("", "aaa", getTestPath("test-project", "mta.yaml"))
		Ω(err).Should(HaveOccurred())
		Ω(err.Error()).Should(Equal(fmt.Sprintf(moduleNotFoundMsg, "aaa")))
	})
	It("mta yaml path not found", func() {
		path := getTestPath("test-project", "mtaNotExist.yaml")
		err := Resolve("", "eb-java", path)
		Ω(err).Should(HaveOccurred())
		Ω(err.Error()).Should(ContainSubstring(fmt.Sprintf(pathNotFoundMsg, path)))
	})
	It("failure on unmarshal", func() {
		path := getTestPath("test-project", "mtaBad.yaml")
		err := Resolve("", "eb-java", path)
		Ω(err).Should(HaveOccurred())
		Ω(err.Error()).Should(ContainSubstring(fmt.Sprintf(unmarshalFailsMsg, path)))
	})
})

var _ = Describe("getPropertiesAsEnvVar", func() {
	It("fails on marshalling", func() {
		mod := mta.Module{
			Properties: map[string]interface{}{
				"a": func() {},
			},
		}
		_, err := getPropertiesAsEnvVar(&mod)
		Ω(err).Should(HaveOccurred())
		Ω(err.Error()).Should(Equal(fmt.Sprintf(marshalFailsMag, "a")))
	})
	It("required groups defined", func() {
		mod := mta.Module{
			Requires: []mta.Requires{
				{
					Group: "group1",
					Properties: map[string]interface{}{
						"prop1": "value1",
						"prop2": "value2",
					},
				},
				{
					Group: "group1",
					Properties: map[string]interface{}{
						"prop3": "value3",
					},
				},
				{
					Properties: map[string]interface{}{
						"prop4": "value4",
					},
				},
			},
		}
		props, err := getPropertiesAsEnvVar(&mod)
		Ω(err).Should(Succeed())
		Ω(len(props)).Should(Equal(2))
		Ω(props["group1"]).Should(Equal(`[{"prop1":"value1","prop2":"value2"},{"prop3":"value3"}]`))
		Ω(props["prop4"]).Should(Equal(`value4`))
	})
})

var _ = Describe("convertToString", func() {
	It("fails to marshal function", func() {
		_, success := convertToString(func() {})
		Ω(success).Should(BeFalse())
	})
	It("map conversion", func() {
		str, success := convertToString(map[string]interface{}{
			"field1": "value1",
			"field2": "value2",
		})
		Ω(success).Should(BeTrue())
		Ω(str).Should(Equal(`{"field1":"value1","field2":"value2"}`))
	})
})

var _ = Describe("getParameter", func() {
	It("parameter found in source parameters", func() {
		resolver := MTAResolver{

		}
		source := mtaSource{
			Parameters: map[string]interface{}{
				"param1": "value1",
			},
		}
		res := resolver.getParameter(nil, &source, nil, "param1")
		Ω(res).Should(Equal("value1"))
	})
	It("parameter found in the module (defined by the source) of the context", func() {
		resolver := MTAResolver{
			context: &ResolveContext{
				modules: map[string]map[string]string{
					"module1": {
						"param1": "value1",
					},
				},
			},
		}
		source := mtaSource{
			Name: "module1",
		}
		res := resolver.getParameter(nil, &source, nil, "param1")
		Ω(res).Should(Equal("value1"))
	})
	It("parameter found in the requires", func() {
		resolver := MTAResolver{
			context: &ResolveContext{
			},
		}
		req := mta.Requires{
			Parameters: map[string]interface{}{
				"param1": "value1",
			},
		}

		res := resolver.getParameter(nil, nil, &req, "param1")
		Ω(res).Should(Equal("value1"))
	})
	It("parameter found in the module (defined by the source module) of the context", func() {
		resolver := MTAResolver{
			context: &ResolveContext{
				modules: map[string]map[string]string{
					"module1": {
						"param1": "value1",
					},
				},
			},
		}
		mod := mta.Module{
			Name: "module1",
		}

		res := resolver.getParameter(&mod, nil, nil, "param1")
		Ω(res).Should(Equal("value1"))
	})
	It("parameter found on the MTA root scope", func() {
		resolver := MTAResolver{
		}
		resolver.Parameters = map[string]interface{}{
			"param1": "value1",
		}

		res := resolver.getParameter(nil, nil, nil, "param1")
		Ω(res).Should(Equal("value1"))
	})
	It("parameter is missing", func() {
		resolver := MTAResolver{
			context: &ResolveContext{
				modules: map[string]map[string]string{
					"module1": {
						"param1": "value1",
					},
				},
			},
		}
		source := mtaSource{
			Name: "module1",
		}

		res := resolver.getParameter(nil, &source, nil, "param2")
		Ω(res).Should(Equal("${param2}"))
	})
})

var _ = Describe("findProvider", func() {
	It("provider not found", func() {
		resolver := MTAResolver{}
		Ω(resolver.findProvider("provider")).Should(BeNil())
	})
})

var _ = Describe("resolvePlaceholdersString", func() {
	It("all placeholders resolved", func() {
		resolver := MTAResolver{
			context: &ResolveContext{
				global: map[string]string{
					"p1": "value1",
					"p2": "value2",
				},
			},
		}
		value := resolver.resolvePlaceholdersString(nil, nil, nil, "${p1}/${p2}")
		Ω(value).Should(Equal("value1/value2"))
	})
})

var _ = Describe("resolvePlaceholders", func() {
	It("value is a map[interface{}]interface{} ", func() {
		resolver := MTAResolver{
			context: &ResolveContext{
				global: map[string]string{
					"p1": "value1",
					"p2": "value2",
				},
			},
		}
		value := map[interface{}]interface{}{
			"abc": "${p1}",
		}
		res := resolver.resolvePlaceholders(nil, nil, nil, value)
		Ω(res).Should(BeEquivalentTo(map[string]interface{}{"abc": "value1"}))
	})
})

var _ = Describe("getVariableValue", func() {
	It("missing required prefix", func() {
		resolver := MTAResolver{
		}
		res := resolver.getVariableValue(nil, nil, "var_without_prefix")
		Ω(res).Should(Equal("~{var_without_prefix}"))
	})
	It("missing configuration", func() {
		resolver := MTAResolver{
		}
		resolver.Resources = []*mta.Resource{
			{
				Name:       "provider",
				Type:       "configuration",
				Properties: map[string]interface{}{},
				Parameters: map[string]interface{}{
					"provider-id": "id",
				},
			},
		}
		res := resolver.getVariableValue(nil, nil, "provider/var")
		Ω(res).Should(Equal("~{var}"))
	})
})

var _ = Describe("parseNextVariable", func() {
	It("double end sign", func() {
		start, end, whole := parseNextVariable(0, "a ~{{var1}}", "~")
		Ω(start).Should(Equal(2))
		Ω(end).Should(Equal("{var1}"))
		Ω(whole).Should(BeFalse())
	})
})

func mockEnvGetter() []string {
	return []string{"health-check-type=http"}
}

func mockEnvGetterWithVcapServices() []string {
	vs := vcapServices{"aaa": []VcapService{
		{Name: "ed-aaa-service", InstanceName: "aaa", Label: "aaa", Plan: "aaa", Tags: []string{"mta-resource-name:ed-aaa"}},
		{Name: "ed-bbb-service", InstanceName: "bbb", Label: "bbb", Plan: "bbb", Tags: []string{"mta-resource-name:ed-bbb"}},
	}}
	vsByte, _ := json.Marshal(vs)
	vsStr := string(vsByte)
	return []string{"health-check-type=http", "VCAP_SERVICES=" + vsStr}
}

// when working dir not provided .env file not found and it's variable are missing
// to complete the list of environment variables this function is used
func mockEnvGetterExtWithVcapServices() []string {
	result := mockEnvGetterWithVcapServices()
	result = append(result, "env_var1=vvv")
	return result
}

func mockEnvGetterExt() []string {
	result := mockEnvGetter()
	result = append(result, "env_var1=vvv")
	return result
}
