package resolver

import (
	"encoding/json"
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/types"

	"github.com/SAP/cloud-mta/internal/fs"
	"github.com/SAP/cloud-mta/mta"
)

func callResolveAndGetOutput(wd, moduleName, yamlPath string, extensions []string, envFileName string) (ResolveResult, []string) {
	result, messages, err := Resolve(wd, moduleName, yamlPath, extensions, envFileName)
	Ω(err).Should(Succeed())
	return result, messages
}

func callResolveAndValidateOutput(wd, moduleName, yamlPath string, extensions []string, envFile string, expected ResolveResult, expectedMessagesMatcher GomegaMatcher) {
	actualResult, actualMessages := callResolveAndGetOutput(wd, moduleName, yamlPath, extensions, envFile)
	Ω(actualResult).Should(Equal(expected))
	Ω(actualMessages).Should(expectedMessagesMatcher)
}

var _ = Describe("Resolve", func() {
	expected := ResolveResult{
		Properties: map[string]string{
			`prop1`:                             `no_placeholders`,
			`prop2`:                             `1000m`,
			`prop3`:                             `["1000m","1m"]`,
			`prop4`:                             `{"p1":"1000m","p2":"1m"}`,
			`prop5`:                             `1`,
			`prop6`:                             `{"1":"1000m"}`,
			`prop7`:                             `~{eb-msahaa/heap`,
			`prop8`:                             `[[{"a":["a1",{"a2-key":"a2-value"}]}]]`,
			`prop9`:                             `vvv`,
			`prop10`:                            `${env_var0}`,
			`prop11`:                            `2G`,
			`JBP_CONFIG_companyJVM`:             `[ memory_calculator: { memory_sizes: { heap: 1000m, stack: 1m, metaspace: 150m } } ]`,
			`JBP_CONFIG_companyJVM1`:            `[ memory_calculator: { memory_sizes: { heap: 1000m, stack: 1m, metaspace: 150m } } ]`,
			`JBP_CONFIG_RESOURCE_CONFIGURATION`: `[tomcat/webapps/ROOT/META-INF/context.xml: {"service_name_for_DefaultDB" : "ed-aaa-service"}]`,
			`bbb_service`:                       `ed-bbb-service`,
		},
		Messages: []string{
			`Missing env_var0`,
		}}

	It("resolves from environment variables and default env file when env file is not sent", func() {
		wd := getTestPath("test-project")
		yamlPath := getTestPath("test-project", "mta.yaml")
		envGetter = mockEnvGetterWithVcapServices
		callResolveAndValidateOutput(wd, "eb-java", yamlPath, nil, "", expected, BeEmpty())
	})
	It("resolves vcap_services from env file", func() {
		wd := getTestPath("test-project")
		yamlPath := getTestPath("test-project", "mta.yaml")
		envGetter = func() []string {
			return []string{"health-check-type=http"}
		}
		callResolveAndValidateOutput(wd, "eb-java", yamlPath, nil, ".env_with_vcap", expected, BeEmpty())
	})

	// If this fail you may have forgot to apply a patch
	// see the `patches` directory in the root of this repo
	It("resolves vcap_services from large env file ", func() {
		wd := getTestPath("test-project")
		yamlPath := getTestPath("test-project", "mtaLargeEnv.yaml")
		expected := ResolveResult{
			Properties: map[string]string{
				`SERVICE_REPLACEMENTS`: `[{"key":"ServiceName_1","service":"BAMBA"}]`,
				`TARGET_CONTAINER`:     `BAMBA`,
			},
			Messages: []string{}}
		callResolveAndValidateOutput(wd, "db", yamlPath, nil, ".envLarge", expected, BeEmpty())
	})

	It("uses mta.yaml folder as the working dir when working dir is not sent", func() {
		yamlPath := getTestPath("test-project", "mta.yaml")
		envGetter = mockEnvGetterExtWithVcapServices
		callResolveAndValidateOutput("", "eb-java", yamlPath, nil, "", expected, BeEmpty())
	})
	It("resolves from env file when it's different from the default name (.env)", func() {
		wd := getTestPath("test-project")
		yamlPath := getTestPath("test-project", "mta.yaml")
		envGetter = mockEnvGetterExtWithVcapServices
		expectedResolve := ResolveResult{
			Properties: map[string]string{
				`prop1`:                             `no_placeholders`,
				`prop2`:                             `1000m`,
				`prop3`:                             `["1000m","1m"]`,
				`prop4`:                             `{"p1":"1000m","p2":"1m"}`,
				`prop5`:                             `1`,
				`prop6`:                             `{"1":"1000m"}`,
				`prop7`:                             `~{eb-msahaa/heap`,
				`prop8`:                             `[[{"a":["a1",{"a2-key":"a2-value"}]}]]`,
				`prop9`:                             `newValue`,
				`prop10`:                            `${env_var0}`,
				`prop11`:                            `2G`,
				`JBP_CONFIG_companyJVM`:             `[ memory_calculator: { memory_sizes: { heap: 1000m, stack: 1m, metaspace: 150m } } ]`,
				`JBP_CONFIG_companyJVM1`:            `[ memory_calculator: { memory_sizes: { heap: 1000m, stack: 1m, metaspace: 150m } } ]`,
				`JBP_CONFIG_RESOURCE_CONFIGURATION`: `[tomcat/webapps/ROOT/META-INF/context.xml: {"service_name_for_DefaultDB" : "ed-aaa-service"}]`,
				`bbb_service`:                       `ed-bbb-service`,
			},
			Messages: []string{`Missing env_var0`},
		}
		callResolveAndValidateOutput(wd, "eb-java", yamlPath, nil, ".env2", expectedResolve, BeEmpty())
	})
	It("resolves from .env file with absolute path", func() {
		wd := getTestPath("test-project")
		yamlPath := getTestPath("test-project", "mta.yaml")
		envPath := getTestPath("test-project", "srv", ".env")
		envGetter = mockEnvGetterExtWithVcapServices
		callResolveAndValidateOutput(wd, "eb-java", yamlPath, nil, envPath, expected, BeEmpty())
	})
	It("resolves service name from mta.yaml when vcap_services variable is not defined", func() {
		yamlPath := getTestPath("test-project", "mta.yaml")
		envGetter = mockEnvGetterExt
		// No default service name - returned value is the same as defined in the property (variable reference)
		expected.Properties["JBP_CONFIG_RESOURCE_CONFIGURATION"] = strings.Replace(expected.Properties["JBP_CONFIG_RESOURCE_CONFIGURATION"], "ed-aaa-service", "${service-name}", -1)
		// Default service name defined in the mta.yaml
		expected.Properties["bbb_service"] = strings.Replace(expected.Properties["bbb_service"], "ed-bbb-service", "ed-bbb-param", -1)
		// Message about missing service name variable
		expected.Messages = append(expected.Messages, `Missing ed-aaa/service-name`)
		callResolveAndValidateOutput("", "eb-java", yamlPath, nil, "", expected, BeEmpty())
	})
	It("returns error when module name is empty", func() {
		_, _, err := Resolve("", "", getTestPath("test-project", "mta.yaml"), nil, "")
		Ω(err).Should(HaveOccurred())
		Ω(err.Error()).Should(Equal(emptyModuleNameMsg))
	})
	It("returns error when module does not exist", func() {
		_, _, err := Resolve("", "aaa", getTestPath("test-project", "mta.yaml"), nil, "")

		Ω(err).Should(HaveOccurred())
		Ω(err.Error()).Should(Equal(fmt.Sprintf(moduleNotFoundMsg, "aaa")))
	})
	It("returns error when mta yaml path is not found", func() {
		path := getTestPath("test-project", "mtaNotExist.yaml")
		_, _, err := Resolve("", "eb-java", path, nil, "")
		Ω(err).Should(HaveOccurred())
		Ω(err.Error()).Should(ContainSubstring(fmt.Sprintf(fs.PathNotFoundMsg, path)))
	})
	It("returns error when mta.yaml is invalid", func() {
		path := getTestPath("test-project", "mtaBad.yaml")
		_, _, err := Resolve("", "eb-java", path, nil, "")
		Ω(err).Should(HaveOccurred())
		Ω(err.Error()).Should(ContainSubstring(fmt.Sprintf(mta.UnmarshalFailsMsg, path)))
	})
	Describe("resolve with extensions", func() {
		var getExpectedResolve = func() ResolveResult {
			// This is the result expected to be returned for mtaExtTest.yaml with .envExtTest with env from mockEnvGetterExt and without extensions
			return ResolveResult{
				Properties: map[string]string{
					`requiredStringProp`:       `value in string: testResource1Service`,
					`requiredEnvProp`:          `param1_value`,
					`requiredPropValue`:        `resource2-defaultServiceName`,
					`hardCodedProp`:            `no_placeholders`,
					`stringProp`:               `value1`,
					`arrayProp`:                `["value1","value2"]`,
					`mapProp`:                  `{"field1":"value1","field2":"value2"}`,
					`hardCodedNumericProp`:     `1`,
					`mapPropWithNumericKey`:    `{"1":"value1"}`,
					"badReferenceProp":         "~{testResource3/prop1",
					`hardCodedNestedArrayProp`: `[[{"a":["a1",{"a2-key":"a2-value"}]}]]`,
					`extEnvProp`:               `vvv`,
					`unknownProp`:              `${env_var0}`,
					`paramProp`:                `/health`,
					`nestedProp`:               `[ memory_calculator: { memory_sizes: { heap: value1, stack: 1m, metaspace: 150m } } ]`,
				},
				Messages: []string{`Missing env_var0`},
			}
		}
		It("resolves variables defined in mta.yaml when the sent extensions don't exist", func() {
			wd := getTestPath("test-project")
			mtaPath := getTestPath("test-project", "mtaExtTest.yaml")
			mtaExtPath := getTestPath("test-project", "nonExisting.mtaext")
			envGetter = mockEnvGetterExt
			expectedResolve := getExpectedResolve()
			callResolveAndValidateOutput(wd, "testModule", mtaPath, []string{mtaExtPath}, ".envExtTest", expectedResolve, ContainElement(ContainSubstring(fs.PathNotFoundMsg, mtaExtPath)))
		})
		It("resolves variables defined in mta.yaml when the sent extensions are invalid", func() {
			wd := getTestPath("test-project")
			mtaPath := getTestPath("test-project", "mtaExtTest.yaml")
			mtaExtPath := getTestPath("test-project", "invalid.mtaext")
			envGetter = mockEnvGetterExt
			expectedResolve := getExpectedResolve()
			callResolveAndValidateOutput(wd, "testModule", mtaPath, []string{mtaExtPath}, ".envExtTest", expectedResolve, ContainElement(ContainSubstring("testModule_doesNotExist")))
		})
		It("resolves variables defined in the merged mta when extensions are sent", func() {
			wd := getTestPath("test-project")
			mtaPath := getTestPath("test-project", "mtaExtTest.yaml")
			mtaExtPath := getTestPath("test-project", "valid1.mtaext")
			envGetter = mockEnvGetterExt
			expectedResolve := getExpectedResolve()
			expectedResolve.Properties[`hardCodedProp`] = `no_placeholders_fromExt1`
			expectedResolve.Properties[`stringProp`] = `value2`
			expectedResolve.Properties[`mapProp`] = `{"field1":"hardCodedValue_fromExt1","field2":"value2"}`
			expectedResolve.Properties[`requiredPropValue`] = `resource2-defaultServiceName_fromExt1`
			callResolveAndValidateOutput(wd, "testModule", mtaPath, []string{mtaExtPath}, ".envExtTest", expectedResolve, BeEmpty())
		})
		It("resolves variables defined in the merged mta when extensions are partially invalid", func() {
			wd := getTestPath("test-project")
			mtaPath := getTestPath("test-project", "mtaExtTest.yaml")
			mtaExtPath2 := getTestPath("test-project", "valid2.mtaext")
			mtaExtPath3 := getTestPath("test-project", "invalid3.mtaext")
			envGetter = mockEnvGetterExt
			expectedResolve := getExpectedResolve()
			expectedResolve.Properties[`hardCodedProp`] = `no_placeholders_fromExt3`
			// The service-name in testResource2 is not overwritten because of the invalid module name
			callResolveAndValidateOutput(wd, "testModule", mtaPath, []string{mtaExtPath2, mtaExtPath3}, ".envExtTest", expectedResolve, ContainElement(ContainSubstring("testModule_doesNotExist")))
		})
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
			context: &ResolveContext{
				modules:   map[string]map[string]string{},
				resources: map[string]map[string]string{},
			},
		}
		source := mtaSource{
			Parameters: map[string]interface{}{
				"param1": "value1",
			},
		}
		res := resolver.getParameter(nil, &source, nil, "param1")
		Ω(res).Should(Equal("value1"))
	})
	It("parameter found in source parameters but isn't a string", func() {
		resolver := MTAResolver{
			context: &ResolveContext{
				modules:   map[string]map[string]string{},
				resources: map[string]map[string]string{},
			},
		}
		source := mtaSource{
			Parameters: map[string]interface{}{
				"param1": struct {
					a string
				}{a: "a"},
			},
		}
		res := resolver.getParameter(nil, &source, nil, "param1")
		Ω(res).Should(Equal("${param1}"))
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
			context: &ResolveContext{},
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
		resolver := MTAResolver{}
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
			messages: []string{},
		}
		source := mtaSource{
			Name: "module1",
		}

		res := resolver.getParameter(nil, &source, nil, "param2")
		Ω(res).Should(Equal("${param2}"))
		Ω(resolver.messages).Should(Equal([]string{"Missing module1/param2"}))
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
		resolver := MTAResolver{messages: []string{}}
		res := resolver.getVariableValue(nil, nil, "var_without_prefix")
		Ω(res).Should(Equal("~{var_without_prefix}"))
		Ω(resolver.messages).Should(Equal([]string{fmt.Sprintf(missingPrefixMsg, "var_without_prefix")}))
	})
	It("missing configuration", func() {
		resolver := MTAResolver{messages: []string{}}
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
		Ω(resolver.messages).Should(Equal([]string{"Missing configuration id/var"}))
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
