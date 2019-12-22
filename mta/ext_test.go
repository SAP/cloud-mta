package mta

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

func boolPtr(b bool) *bool {
	return &b
}
func falsePtr() *bool {
	return boolPtr(false)
}
func truePtr() *bool {
	return boolPtr(true)
}

var _ = Describe("Extension MTA", func() {
	var _ = Describe("UnmarshalExt", func() {
		It("Sanity", func() {
			wd, err := os.Getwd()
			Ω(err).Should(Succeed())
			content, err := readFile(filepath.Join(wd, "testdata", "my.mtaext"))
			Ω(err).Should(Succeed())
			m, err := UnmarshalExt(content)
			Ω(err).Should(Succeed())
			Ω(len(m.Modules)).Should(Equal(1))
			Ω(len(m.Resources)).Should(Equal(1))
			Ω(m.Resources[0].Active).ShouldNot(BeNil())
			Ω(*m.Resources[0].Active).Should(BeFalse())
		})
		It("Invalid content", func() {
			_, err := UnmarshalExt([]byte("wrong mtaExt"))
			Ω(err).Should(HaveOccurred())
		})
	})

	var _ = Describe("extendMap", func() {
		overwritableScenarios := map[string]map[string]MetaData{
			"overwritable is true": {
				"b": {
					OverWritable: truePtr(),
				},
			},
			"metadata is not specified for the keys": {
				"c": {
					OverWritable: falsePtr(),
				},
			},
			"metadata does not exists": nil,
		}

		Context("merging map/scalar value (valid cases)", func() {
			When("merging flat maps of strings", func() {
				mapWithOneKey := map[string]interface{}{
					"b": "xx",
				}
				mapWithTwoKeys := map[string]interface{}{
					"a": "aa",
					"b": "bb",
				}

				for desc, metadata := range overwritableScenarios {
					DescribeTable(desc, func(mta map[string]interface{}, ext map[string]interface{}, expected map[string]interface{}) {
						checkExtendMap(mta, ext, metadata, expected)
					},
						Entry("overrides values", mapWithTwoKeys, mapWithOneKey, map[string]interface{}{
							"a": "aa",
							"b": "xx",
						}),
						Entry("overrides and adds values", mapWithOneKey, mapWithTwoKeys, mapWithTwoKeys),
						Entry("copies ext map when original map is nil", nil, mapWithTwoKeys, mapWithTwoKeys),
						Entry("doesn't change original when ext is nil", mapWithTwoKeys, nil, mapWithTwoKeys),
						Entry("returns nil when original and ext are nil", nil, nil, nil),
					)
				}

				DescribeTable("overwritable is false (valid cases)", func(mta map[string]interface{}, ext map[string]interface{}, expected map[string]interface{}) {
					var meta = map[string]MetaData{
						"b": {
							OverWritable: falsePtr(),
						},
					}
					checkExtendMap(mta, ext, meta, expected)
				},
					Entry("adds values", mapWithTwoKeys, map[string]interface{}{
						"c": "cc",
					}, map[string]interface{}{
						"a": "aa",
						"b": "bb",
						"c": "cc",
					}),
					Entry("copies ext map when original map is nil", nil, mapWithTwoKeys, mapWithTwoKeys),
					Entry("doesn't change original when ext is nil", mapWithTwoKeys, nil, mapWithTwoKeys),
					Entry("returns nil when original and ext are nil", nil, nil, nil),
				)

				It("fails when trying to override non-overwritable field", func() {
					var meta = map[string]MetaData{
						"b": {
							OverWritable: falsePtr(),
						},
					}
					checkExtendMapFails(mapWithTwoKeys, mapWithOneKey, meta, overwriteNonOverwritableErrorMsg, "b")
				})
			})

			When("merging maps with one nested level", func() {
				mapWithOneKey := map[string]interface{}{
					"b": map[string]interface{}{
						"e": "xx",
					},
				}
				mapWithTwoKeys := map[string]interface{}{
					"b": map[string]interface{}{
						"e": "ee",
						"c": "cc",
					},
					"a": "aa",
				}

				for desc, metadata := range overwritableScenarios {
					DescribeTable(desc, func(mta map[string]interface{}, ext map[string]interface{}, expected map[string]interface{}) {
						checkExtendMap(mta, ext, metadata, expected)
					},
						Entry("overrides inner values, adds inner values and adds values", mapWithOneKey, mapWithTwoKeys, mapWithTwoKeys),
						Entry("copies extend file when mta file is empty", nil, mapWithTwoKeys, mapWithTwoKeys),
						Entry("doesn't change mta file when extend is empty", mapWithTwoKeys, nil, mapWithTwoKeys),
						Entry("returns nil when original and ext are nil", nil, nil, nil),
					)
				}

				Context("overwritable is false", func() {
					DescribeTable("overwritable is false", func(mta map[string]interface{}, ext map[string]interface{}, expected map[string]interface{}) {
						var meta = map[string]MetaData{
							"b": {
								OverWritable: falsePtr(),
							},
						}
						checkExtendMap(mta, ext, meta, expected)
					},
						Entry("adds values", mapWithOneKey, map[string]interface{}{
							"a": "aa",
						}, map[string]interface{}{
							"b": map[string]interface{}{
								"e": "xx",
							},
							"a": "aa",
						}),
						Entry("copies extend file when mta file is empty", nil, mapWithTwoKeys, mapWithTwoKeys),
						Entry("doesnt change mta file when extend is empty", mapWithTwoKeys, nil, mapWithTwoKeys),
						Entry("returns nil when original and ext are nil", nil, nil, nil),
					)
				})

				It("fails when trying to override non-overwritable field", func() {
					var meta = map[string]MetaData{
						"b": {
							OverWritable: falsePtr(),
						},
					}
					checkExtendMapFails(mapWithTwoKeys, mapWithOneKey, meta, overwriteNonOverwritableErrorMsg, "b")
				})
			})

			When("merging maps with more than one nested level", func() {
				mapWithOneKey := map[string]interface{}{
					"b": map[string]interface{}{
						"d": map[string]interface{}{
							"e": "xx",
						},
					},
				}
				mapWithTwoKeys := map[string]interface{}{
					"b": map[string]interface{}{
						"d": map[string]interface{}{
							"e": "ee",
							"m": "mm",
						},
						"f": map[string]interface{}{
							"l": "ll",
						},
					},
					"a": "aa",
				}
				interfaceMapWithOneKey := map[string]interface{}{
					"b": map[interface{}]interface{}{
						"d": map[interface{}]interface{}{
							"e": "xx",
						},
					},
				}

				for desc, metadata := range overwritableScenarios {
					DescribeTable(desc, func(mta map[string]interface{}, ext map[string]interface{}, expected map[string]interface{}) {
						checkExtendMap(mta, ext, metadata, expected)
					},
						Entry("overrides inner values", mapWithTwoKeys, mapWithOneKey, map[string]interface{}{
							"b": map[string]interface{}{
								"d": map[string]interface{}{
									"e": "xx",
									"m": "mm",
								},
								"f": map[string]interface{}{
									"l": "ll",
								},
							},
							"a": "aa",
						}),
						Entry("overrides inner values - with interface keys", mapWithTwoKeys, interfaceMapWithOneKey, map[string]interface{}{
							"b": map[string]interface{}{
								"d": map[string]interface{}{
									"e": "xx",
									"m": "mm",
								},
								"f": map[string]interface{}{
									"l": "ll",
								},
							},
							"a": "aa",
						}),
						Entry("overrides inner values, adds inner values and adds new values", mapWithOneKey, mapWithTwoKeys, mapWithTwoKeys),
						Entry("overrides inner values, adds inner values and adds new values - with interface keys", interfaceMapWithOneKey, mapWithTwoKeys, mapWithTwoKeys),
						Entry("copies ext map when original map is nil", nil, mapWithTwoKeys, mapWithTwoKeys),
						Entry("doesn't change original when ext is nil", mapWithTwoKeys, nil, mapWithTwoKeys),
						Entry("returns nil when original and ext are nil", nil, nil, nil),
					)
				}

				DescribeTable("overwritable is false", func(mta map[string]interface{}, ext map[string]interface{}, expected map[string]interface{}) {
					var meta = map[string]MetaData{
						"b": {
							OverWritable: falsePtr(),
						},
					}
					checkExtendMap(mta, ext, meta, expected)
				},
					Entry("adds new values", mapWithOneKey, map[string]interface{}{
						"a": "aa",
					}, map[string]interface{}{
						"b": map[string]interface{}{
							"d": map[string]interface{}{
								"e": "xx",
							},
						},
						"a": "aa",
					}),
					Entry("copies ext map when original map is nil", nil, mapWithTwoKeys, mapWithTwoKeys),
					Entry("doesn't change original when ext is nil", mapWithTwoKeys, nil, mapWithTwoKeys),
					Entry("returns nil when original and ext are nil", nil, nil, nil),
				)

				It("fails when trying to override non-overwritable field", func() {
					var meta = map[string]MetaData{
						"b": {
							OverWritable: falsePtr(),
						},
					}
					checkExtendMapFails(mapWithTwoKeys, mapWithOneKey, meta, overwriteNonOverwritableErrorMsg, "b")
				})
			})

			for desc, metadata := range overwritableScenarios {
				It("merges maps of different scalar and sequence types when "+desc, func() {
					checkExtendMap(map[string]interface{}{
						"a":  1,
						"b":  "s",
						"c1": []string{"a", "b"},
						"d": map[string]interface{}{
							"e": 5,
						},
						"g": nil,
						"h": map[string]interface{}{
							"e": 5,
						},
						"i": nil,
					}, map[string]interface{}{
						"a":  "s",
						"b":  []int{1, 2},
						"c1": nil,
						"d": map[string]interface{}{
							"f": "xx",
						},
						"g": "gg",
						"h": nil,
						"i": map[string]interface{}{
							"e": 5,
						},
					}, metadata, map[string]interface{}{
						"a":  "s",
						"b":  []int{1, 2},
						"c1": nil,
						"d": map[string]interface{}{
							"e": 5,
							"f": "xx",
						},
						"g": "gg",
						"h": nil,
						"i": map[string]interface{}{
							"e": 5,
						},
					})
				})
			}
		})

		When("merging mixed scalar and flat values", func() {
			Context("on the first level", func() {
				mapWithOneKey := map[string]interface{}{
					"b": map[string]interface{}{
						"c": "cc",
					},
				}
				mapWithTwoKeys := map[string]interface{}{
					"a": "aa",
					"b": "bb",
				}

				for desc, metadata := range overwritableScenarios {
					DescribeTable(desc, func(mta map[string]interface{}, ext map[string]interface{}, err string, args ...interface{}) {
						checkExtendMapFails(mta, ext, metadata, err, args...)
					},
						Entry("fails when overriding scalar with map", mapWithTwoKeys, mapWithOneKey, overwriteScalarWithStructuredErrorMsg, "b"),
						Entry("fails when overriding map with scalar", mapWithOneKey, mapWithTwoKeys, overwriteStructuredWithScalarErrorMsg, "b"),
					)
				}
			})

			Context("on a nested level", func() {
				mapWithOneKey := map[string]interface{}{
					"b": map[string]interface{}{
						"d": map[string]interface{}{
							"e": map[string]interface{}{
								"r": "rr",
							},
						},
					},
				}
				mapWithTwoKeys := map[string]interface{}{
					"b": map[string]interface{}{
						"d": map[string]interface{}{
							"e": "ee",
							"m": "mm",
						},
						"f": map[string]interface{}{
							"l": "ll",
						},
					},
					"a": "aa",
				}
				interfaceMapWithOneKey := map[string]interface{}{
					"b": map[string]interface{}{
						"d": map[string]interface{}{
							"e": map[interface{}]interface{}{
								"r": "rr",
							},
						},
					},
				}

				for desc, metadata := range overwritableScenarios {
					DescribeTable(desc, func(mta map[string]interface{}, ext map[string]interface{}, err string, args ...interface{}) {
						checkExtendMapFails(mta, ext, metadata, err, args...)
					},
						Entry("fails when overriding scalar with map", mapWithTwoKeys, mapWithOneKey, overwriteScalarWithStructuredErrorMsg, "e"),
						Entry("fails when overriding map with scalar", mapWithOneKey, mapWithTwoKeys, overwriteStructuredWithScalarErrorMsg, "e"),
						Entry("fails when overriding scalar with map", mapWithTwoKeys, interfaceMapWithOneKey, overwriteScalarWithStructuredErrorMsg, "e"),
						Entry("fails when overriding map with scalar", interfaceMapWithOneKey, mapWithTwoKeys, overwriteStructuredWithScalarErrorMsg, "e"),
					)
				}
			})
		})
	})

	var _ = Describe("Merge", func() {
		It("merges the root parameters", func() {
			mtaObj := MTA{
				Parameters: map[string]interface{}{
					"p1": "the p",
				},
			}
			err := Merge(&mtaObj, &EXT{
				Parameters: map[string]interface{}{
					"p1": "changed",
					"p2": "added",
				},
			})

			Ω(err).Should(Succeed())
			Ω(mtaObj).Should(Equal(MTA{
				Parameters: map[string]interface{}{
					"p1": "changed",
					"p2": "added",
				},
			}))
		})
		It("merges the root parameters when original map is nil", func() {
			mtaObj := MTA{}
			err := Merge(&mtaObj, &EXT{
				Parameters: map[string]interface{}{
					"p2": "added",
				},
			})

			Ω(err).Should(Succeed())
			Ω(mtaObj).Should(Equal(MTA{
				Parameters: map[string]interface{}{
					"p2": "added",
				},
			}))
		})
		It("fails if it can't merge root parameters", func() {
			mtaObj := MTA{
				Parameters: map[string]interface{}{
					"p1p1": "value",
				},
			}
			err := Merge(&mtaObj, &EXT{
				Parameters: map[string]interface{}{
					"p1p1": map[string]interface{}{
						"p1p2c1": "added",
					},
					"p1p2": map[string]interface{}{
						"p1p2c1": "changed",
					},
				},
			})

			Ω(err).Should(HaveOccurred())
			Ω(err.Error()).Should(ContainSubstring(overwriteScalarWithStructuredErrorMsg, "p1p1"))
		})
		It("fails if it can't merge root parameters because of metadata", func() {
			mtaObj := MTA{
				Parameters: map[string]interface{}{
					"p1p1": "value",
				},
				ParametersMetaData: map[string]MetaData{
					"p1p1": {
						OverWritable: falsePtr(),
					},
				},
			}
			err := Merge(&mtaObj, &EXT{
				Parameters: map[string]interface{}{
					"p1p1": "changed",
				},
			})

			Ω(err).Should(HaveOccurred())
			Ω(err.Error()).Should(ContainSubstring(overwriteNonOverwritableErrorMsg, "p1p1"))
		})
		Context("modules", func() {
			It("merges the module properties", func() {
				checkModuleMerge(Module{
					Name: "module1",
					Properties: map[string]interface{}{
						"p1": "the p",
					},
				}, ModuleExt{
					Name: "module1",
					Properties: map[string]interface{}{
						"p1": "changed",
						"p2": "added",
					},
				}, Module{
					Name: "module1",
					Properties: map[string]interface{}{
						"p1": "changed",
						"p2": "added",
					},
				})
			})
			It("merges the module properties when original map is nil", func() {
				checkModuleMerge(Module{
					Name: "module1",
				}, ModuleExt{
					Name: "module1",
					Properties: map[string]interface{}{
						"p2": "added",
					},
				}, Module{
					Name: "module1",
					Properties: map[string]interface{}{
						"p2": "added",
					},
				})
			})
			It("merges the module parameters", func() {
				checkModuleMerge(Module{
					Name: "module1",
					Parameters: map[string]interface{}{
						"p1p1": "the paaram",
					},
				}, ModuleExt{
					Name: "module1",
					Parameters: map[string]interface{}{
						"p1p1": "changed",
						"p1p2": map[string]interface{}{
							"p1p2c1": "added",
						},
					},
				}, Module{
					Name: "module1",
					Parameters: map[string]interface{}{
						"p1p1": "changed",
						"p1p2": map[string]interface{}{
							"p1p2c1": "added",
						},
					},
				})
			})
			It("merges the module parameters when original map is nil", func() {
				checkModuleMerge(Module{
					Name: "module1",
				}, ModuleExt{
					Name: "module1",
					Parameters: map[string]interface{}{
						"p1p1": "added",
					},
				}, Module{
					Name: "module1",
					Parameters: map[string]interface{}{
						"p1p1": "added",
					},
				})
			})
			It("merges the module build parameters", func() {
				checkModuleMerge(Module{
					Name: "module1",
					BuildParams: map[string]interface{}{
						"p1b1": "the paaram",
					},
				}, ModuleExt{
					Name: "module1",
					BuildParams: map[string]interface{}{
						"p1b1": "changed",
						"p1b2": map[string]interface{}{
							"p1b2c1": "added",
						},
					},
				}, Module{
					Name: "module1",
					BuildParams: map[string]interface{}{
						"p1b1": "changed",
						"p1b2": map[string]interface{}{
							"p1b2c1": "added",
						},
					},
				})
			})
			It("merges the module build parameters when original map is nil", func() {
				checkModuleMerge(Module{
					Name: "module1",
				}, ModuleExt{
					Name: "module1",
					BuildParams: map[string]interface{}{
						"p1b1": "added",
					},
				}, Module{
					Name: "module1",
					BuildParams: map[string]interface{}{
						"p1b1": "added",
					},
				})
			})
			Context("hooks", func() {
				It("merges the module hooks parameters", func() {
					checkModuleMerge(Module{
						Name: "module1",
						Hooks: []Hook{
							{
								Name: "h1",
								Parameters: map[string]interface{}{
									"provp1": "the prop",
								},
								ParametersMetaData: map[string]MetaData{
									"provp1": {
										OverWritable: truePtr(),
									},
								},
							},
							{
								Name: "h2",
							},
						},
					}, ModuleExt{
						Name: "module1",
						Hooks: []Hook{
							{
								Name: "h1",
								Parameters: map[string]interface{}{
									"provp1": "changed",
									"provep2": map[string]interface{}{
										"provep2t1": "added",
									},
								},
							},
							{
								Name: "h2",
								Parameters: map[string]interface{}{
									"provp1": "added",
								},
							},
						},
					}, Module{
						Name: "module1",
						Hooks: []Hook{
							{
								Name: "h1",
								Parameters: map[string]interface{}{
									"provp1": "changed",
									"provep2": map[string]interface{}{
										"provep2t1": "added",
									},
								},
								ParametersMetaData: map[string]MetaData{
									"provp1": {
										OverWritable: truePtr(),
									},
								},
							},
							{
								Name: "h2",
								Parameters: map[string]interface{}{
									"provp1": "added",
								},
							},
						},
					})
				})
				It("fails if there is a module hook in the extension that doesn't exist in the original MTA", func() {
					checkModuleMergeFails(Module{
						Name: "module1",
						Hooks: []Hook{
							{
								Name: "h1",
								Parameters: map[string]interface{}{
									"provp1": "the prop",
								},
							},
						},
					}, ModuleExt{
						Name: "module1",
						Hooks: []Hook{
							{
								Name: "h2",
								Parameters: map[string]interface{}{
									"provp1": "changed",
								},
							},
						},
					}, unknownModuleHookErrorMsg, "h2", "module1")
				})
				It("fails if it can't merge module hook parameters", func() {
					checkModuleMergeFails(Module{
						Name: "module1",
						Hooks: []Hook{
							{
								Name: "h1",
								Parameters: map[string]interface{}{
									"provp1":  "value",
									"provep2": "value",
								},
							},
						},
					}, ModuleExt{
						Name: "module1",
						Hooks: []Hook{
							{
								Name: "h1",
								Parameters: map[string]interface{}{
									"provp1": "changed",
									"provep2": map[string]interface{}{
										"provep2t1": "added",
									},
								},
							},
						},
					}, overwriteScalarWithStructuredErrorMsg, "provep2")
				})
				It("fails if it can't merge module hook parameters because of metadata", func() {
					checkModuleMergeFails(Module{
						Name: "module1",
						Hooks: []Hook{
							{
								Name: "h1",
								Parameters: map[string]interface{}{
									"provp1": "value",
								},
								ParametersMetaData: map[string]MetaData{
									"provp1": {
										OverWritable: falsePtr(),
									},
								},
							},
						},
					}, ModuleExt{
						Name: "module1",
						Hooks: []Hook{
							{
								Name: "h1",
								Parameters: map[string]interface{}{
									"provp1": "changed",
								},
							},
						},
					}, overwriteNonOverwritableErrorMsg, "provp1")
				})
				It("merges the module hooks requires parameters", func() {
					checkModuleMerge(Module{
						Name: "module1",
						Hooks: []Hook{
							{
								Name: "h1",
								Requires: []Requires{
									{
										Name: "r1",
										Parameters: map[string]interface{}{
											"provp1": "the prop",
										},
										ParametersMetaData: map[string]MetaData{
											"provp1": {
												OverWritable: truePtr(),
											},
										},
									},
									{
										Name: "r2",
									},
								},
							},
						},
					}, ModuleExt{
						Name: "module1",
						Hooks: []Hook{
							{
								Name: "h1",
								Requires: []Requires{
									{
										Name: "r1",
										Parameters: map[string]interface{}{
											"provp1": "changed",
											"provep2": map[string]interface{}{
												"provep2t1": "added",
											},
										},
									},
									{
										Name: "r2",
										Parameters: map[string]interface{}{
											"provp1": "added",
										},
									},
								},
							},
						},
					}, Module{
						Name: "module1",
						Hooks: []Hook{
							{
								Name: "h1",
								Requires: []Requires{
									{
										Name: "r1",
										Parameters: map[string]interface{}{
											"provp1": "changed",
											"provep2": map[string]interface{}{
												"provep2t1": "added",
											},
										},
										ParametersMetaData: map[string]MetaData{
											"provp1": {
												OverWritable: truePtr(),
											},
										},
									},
									{
										Name: "r2",
										Parameters: map[string]interface{}{
											"provp1": "added",
										},
									},
								},
							},
						},
					})
				})
				It("merges the module hooks requires properties", func() {
					checkModuleMerge(Module{
						Name: "module1",
						Hooks: []Hook{
							{
								Name: "h1",
								Requires: []Requires{
									{
										Name: "r1",
										Properties: map[string]interface{}{
											"provp1": "the prop",
										},
										PropertiesMetaData: map[string]MetaData{
											"provp1": {
												OverWritable: truePtr(),
											},
										},
									},
									{
										Name: "r2",
									},
								},
							},
						},
					}, ModuleExt{
						Name: "module1",
						Hooks: []Hook{
							{
								Name: "h1",
								Requires: []Requires{
									{
										Name: "r1",
										Properties: map[string]interface{}{
											"provp1": "changed",
											"provep2": map[string]interface{}{
												"provep2t1": "added",
											},
										},
									},
									{
										Name: "r2",
										Properties: map[string]interface{}{
											"provp1": "added",
										},
									},
								},
							},
						},
					}, Module{
						Name: "module1",
						Hooks: []Hook{
							{
								Name: "h1",
								Requires: []Requires{
									{
										Name: "r1",
										Properties: map[string]interface{}{
											"provp1": "changed",
											"provep2": map[string]interface{}{
												"provep2t1": "added",
											},
										},
										PropertiesMetaData: map[string]MetaData{
											"provp1": {
												OverWritable: truePtr(),
											},
										},
									},
									{
										Name: "r2",
										Properties: map[string]interface{}{
											"provp1": "added",
										},
									},
								},
							},
						},
					})
				})
				It("fails if there is a module hook requires in the extension that doesn't exist in the original MTA", func() {
					checkModuleMergeFails(Module{
						Name: "module1",
						Hooks: []Hook{
							{
								Name: "h1",
								Requires: []Requires{
									{
										Name: "r1",
										Properties: map[string]interface{}{
											"provp1": "the prop",
										},
									},
								},
							},
						},
					}, ModuleExt{
						Name: "module1",
						Hooks: []Hook{
							{
								Name: "h1",
								Requires: []Requires{
									{
										Name: "r2",
										Properties: map[string]interface{}{
											"provp1": "changed",
										},
									},
								},
							},
						},
					}, unknownModuleHookRequiresErrorMsg, "r2", "h1", "module1")
				})
				It("fails if it can't merge module hook requires properties", func() {
					checkModuleMergeFails(Module{
						Name: "module1",
						Hooks: []Hook{
							{
								Name: "h1",
								Requires: []Requires{
									{
										Name: "p1",
										Properties: map[string]interface{}{
											"provp1":  "value",
											"provep2": "value",
										},
									},
								},
							},
						},
					}, ModuleExt{
						Name: "module1",
						Hooks: []Hook{
							{
								Name: "h1",
								Requires: []Requires{
									{
										Name: "p1",
										Properties: map[string]interface{}{
											"provp1": "changed",
											"provep2": map[string]interface{}{
												"provep2t1": "added",
											},
										},
									},
								},
							},
						},
					}, overwriteScalarWithStructuredErrorMsg, "provep2")
				})
				It("fails if it can't merge module hook requires properties because of metadata", func() {
					checkModuleMergeFails(Module{
						Name: "module1",
						Hooks: []Hook{
							{
								Name: "h1",
								Requires: []Requires{
									{
										Name: "p1",
										Properties: map[string]interface{}{
											"provp1": "value",
										},
										PropertiesMetaData: map[string]MetaData{
											"provp1": {
												OverWritable: falsePtr(),
											},
										},
									},
								},
							},
						},
					}, ModuleExt{
						Name: "module1",
						Hooks: []Hook{
							{
								Name: "h1",
								Requires: []Requires{
									{
										Name: "p1",
										Properties: map[string]interface{}{
											"provp1": "changed",
										},
									},
								},
							},
						},
					}, overwriteNonOverwritableErrorMsg, "provp1")
				})
				It("fails if it can't merge module hook requires parameters", func() {
					checkModuleMergeFails(Module{
						Name: "module1",
						Hooks: []Hook{
							{
								Name: "h1",
								Requires: []Requires{
									{
										Name: "p1",
										Parameters: map[string]interface{}{
											"provp1":  "value",
											"provep2": "value",
										},
									},
								},
							},
						},
					}, ModuleExt{
						Name: "module1",
						Hooks: []Hook{
							{
								Name: "h1",
								Requires: []Requires{
									{
										Name: "p1",
										Parameters: map[string]interface{}{
											"provp1": "changed",
											"provep2": map[string]interface{}{
												"provep2t1": "added",
											},
										},
									},
								},
							},
						},
					}, overwriteScalarWithStructuredErrorMsg, "provep2")
				})
				It("fails if it can't merge module hook requires parameters because of metadata", func() {
					checkModuleMergeFails(Module{
						Name: "module1",
						Hooks: []Hook{
							{
								Name: "h1",
								Requires: []Requires{
									{
										Name: "p1",
										Parameters: map[string]interface{}{
											"provp1": "value",
										},
										ParametersMetaData: map[string]MetaData{
											"provp1": {
												OverWritable: falsePtr(),
											},
										},
									},
								},
							},
						},
					}, ModuleExt{
						Name: "module1",
						Hooks: []Hook{
							{
								Name: "h1",
								Requires: []Requires{
									{
										Name: "p1",
										Parameters: map[string]interface{}{
											"provp1": "changed",
										},
									},
								},
							},
						},
					}, overwriteNonOverwritableErrorMsg, "provp1")
				})
			})
			Context("requires", func() {
				It("merges the module requires parameters", func() {
					checkModuleMerge(Module{
						Name: "module1",
						Requires: []Requires{
							{
								Name: "r1",
								Parameters: map[string]interface{}{
									"provp1": "the prop",
								},
								ParametersMetaData: map[string]MetaData{
									"provp1": {
										OverWritable: truePtr(),
									},
								},
							},
							{
								Name: "r2",
							},
						},
					}, ModuleExt{
						Name: "module1",
						Requires: []Requires{
							{
								Name: "r1",
								Parameters: map[string]interface{}{
									"provp1": "changed",
									"provep2": map[string]interface{}{
										"provep2t1": "added",
									},
								},
							},
							{
								Name: "r2",
								Parameters: map[string]interface{}{
									"provp1": "added",
								},
							},
						},
					}, Module{
						Name: "module1",
						Requires: []Requires{
							{
								Name: "r1",
								Parameters: map[string]interface{}{
									"provp1": "changed",
									"provep2": map[string]interface{}{
										"provep2t1": "added",
									},
								},
								ParametersMetaData: map[string]MetaData{
									"provp1": {
										OverWritable: truePtr(),
									},
								},
							},
							{
								Name: "r2",
								Parameters: map[string]interface{}{
									"provp1": "added",
								},
							},
						},
					})
				})
				It("merges the module requires properties", func() {
					checkModuleMerge(Module{
						Name: "module1",
						Requires: []Requires{
							{
								Name: "r1",
								Properties: map[string]interface{}{
									"provp1": "the prop",
								},
								PropertiesMetaData: map[string]MetaData{
									"provp1": {
									},
								},
							},
							{
								Name: "r2",
							},
						},
					}, ModuleExt{
						Name: "module1",
						Requires: []Requires{
							{
								Name: "r1",
								Properties: map[string]interface{}{
									"provp1": "changed",
									"provep2": map[string]interface{}{
										"provep2t1": "added",
									},
								},
							},
							{
								Name: "r2",
								Properties: map[string]interface{}{
									"provp1": "added",
								},
							},
						},
					}, Module{
						Name: "module1",
						Requires: []Requires{
							{
								Name: "r1",
								Properties: map[string]interface{}{
									"provp1": "changed",
									"provep2": map[string]interface{}{
										"provep2t1": "added",
									},
								},
								PropertiesMetaData: map[string]MetaData{
									"provp1": {
									},
								},
							},
							{
								Name: "r2",
								Properties: map[string]interface{}{
									"provp1": "added",
								},
							},
						},
					})
				})
				It("fails if there is a module requires in the extension that doesn't exist in the original MTA", func() {
					checkModuleMergeFails(Module{
						Name: "module1",
						Requires: []Requires{
							{
								Name: "r1",
								Properties: map[string]interface{}{
									"provp1": "the prop",
								},
							},
						},
					}, ModuleExt{
						Name: "module1",
						Requires: []Requires{
							{
								Name: "r2",
								Properties: map[string]interface{}{
									"provp1": "changed",
								},
							},
						},
					}, unknownModuleRequiresErrorMsg, "r2", "module1")
				})
				It("fails if it can't merge module requires properties", func() {
					checkModuleMergeFails(Module{
						Name: "module1",
						Requires: []Requires{
							{
								Name: "p1",
								Properties: map[string]interface{}{
									"provp1":  "value",
									"provep2": "value",
								},
							},
						},
					}, ModuleExt{
						Name: "module1",
						Requires: []Requires{
							{
								Name: "p1",
								Properties: map[string]interface{}{
									"provp1": "changed",
									"provep2": map[interface{}]interface{}{
										"provep2t1": "added",
									},
								},
							},
						},
					}, overwriteScalarWithStructuredErrorMsg, "provep2")
				})
				It("fails if it can't merge module requires properties because of metadata", func() {
					checkModuleMergeFails(Module{
						Name: "module1",
						Requires: []Requires{
							{
								Name: "p1",
								Properties: map[string]interface{}{
									"provp1": "value",
								},
								PropertiesMetaData: map[string]MetaData{
									"provp1": {
										OverWritable: falsePtr(),
									},
								},
							},
						},
					}, ModuleExt{
						Name: "module1",
						Requires: []Requires{
							{
								Name: "p1",
								Properties: map[string]interface{}{
									"provp1": "changed",
								},
							},
						},
					}, overwriteNonOverwritableErrorMsg, "provp1")
				})
				It("fails if it can't merge module requires parameters", func() {
					checkModuleMergeFails(Module{
						Name: "module1",
						Requires: []Requires{
							{
								Name: "p1",
								Parameters: map[string]interface{}{
									"provp1":  "value",
									"provep2": "value",
								},
							},
						},
					}, ModuleExt{
						Name: "module1",
						Requires: []Requires{
							{
								Name: "p1",
								Parameters: map[string]interface{}{
									"provp1": "changed",
									"provep2": map[string]interface{}{
										"provep2t1": "added",
									},
								},
							},
						},
					}, overwriteScalarWithStructuredErrorMsg, "provep2")
				})
				It("fails if it can't merge module requires parameters because of metadata", func() {
					checkModuleMergeFails(Module{
						Name: "module1",
						Requires: []Requires{
							{
								Name: "p1",
								Parameters: map[string]interface{}{
									"provp1": "value",
								},
								ParametersMetaData: map[string]MetaData{
									"provp1": {
										OverWritable: falsePtr(),
									},
								},
							},
						},
					}, ModuleExt{
						Name: "module1",
						Requires: []Requires{
							{
								Name: "p1",
								Parameters: map[string]interface{}{
									"provp1": "changed",
								},
							},
						},
					}, overwriteNonOverwritableErrorMsg, "provp1")
				})
			})
			Context("module includes", func() {
				It("merges the module includes when original and extension both have includes with different names", func() {
					checkModuleMerge(Module{
						Name: "module1",
						Includes: []Includes{
							{
								Name: "ic1",
								Path: "neverland",
							},
						},
					}, ModuleExt{
						Name: "module1",
						Includes: []Includes{
							{
								Name: "added",
								Path: "added",
							},
						},
					}, Module{
						Name: "module1",
						Includes: []Includes{
							{
								Name: "ic1",
								Path: "neverland",
							},
							{
								Name: "added",
								Path: "added",
							},
						},
					})
				})
				It("merges the module includes when original and extension both have includes with the same name names", func() {
					checkModuleMerge(Module{
						Name: "module1",
						Includes: []Includes{
							{
								Name: "ic1",
								Path: "neverland",
							},
						},
					}, ModuleExt{
						Name: "module1",
						Includes: []Includes{
							{
								Name: "ic1",
								Path: "added",
							},
						},
					}, Module{
						Name: "module1",
						Includes: []Includes{
							{
								Name: "ic1",
								Path: "neverland",
							},
							{
								Name: "ic1",
								Path: "added",
							},
						},
					})
				})
				It("merges the module includes when only original has includes", func() {
					checkModuleMerge(Module{
						Name: "module1",
						Includes: []Includes{
							{
								Name: "ic1",
								Path: "neverland",
							},
						},
					}, ModuleExt{
						Name: "module1",
					}, Module{
						Name: "module1",
						Includes: []Includes{
							{
								Name: "ic1",
								Path: "neverland",
							},
						},
					})
				})
				It("merges the module includes when only extension has includes", func() {
					checkModuleMerge(Module{
						Name: "module1",
					}, ModuleExt{
						Name: "module1",
						Includes: []Includes{
							{
								Name: "added",
								Path: "added",
							},
						},
					}, Module{
						Name: "module1",
						Includes: []Includes{
							{
								Name: "added",
								Path: "added",
							},
						},
					})
				})
				It("merges the module includes when both includes arrays are nil", func() {
					checkModuleMerge(Module{
						Name: "module1",
					}, ModuleExt{
						Name: "module1",
					}, Module{
						Name: "module1",
					})
				})
			})
			It("merges the module provides properties", func() {
				checkModuleMerge(Module{
					Name: "module1",
					Provides: []Provides{
						{
							Name:   "p1",
							Public: true,
							Properties: map[string]interface{}{
								"provp1": "the prop",
							},
							PropertiesMetaData: map[string]MetaData{
								"provp1": {
									OverWritable: truePtr(),
								},
							},
						},
						{
							Name: "p2",
						},
					},
				}, ModuleExt{
					Name: "module1",
					Provides: []Provides{
						{
							Name:   "p1",
							Public: true,
							Properties: map[string]interface{}{
								"provp1": "changed",
								"provep2": map[string]interface{}{
									"provep2t1": "added",
								},
							},
						},
						{
							Name: "p2",
							Properties: map[string]interface{}{
								"provp1": "added",
							},
						},
					},
				}, Module{
					Name: "module1",
					Provides: []Provides{
						{
							Name:   "p1",
							Public: true,
							Properties: map[string]interface{}{
								"provp1": "changed",
								"provep2": map[string]interface{}{
									"provep2t1": "added",
								},
							},
							PropertiesMetaData: map[string]MetaData{
								"provp1": {
									OverWritable: truePtr(),
								},
							},
						},
						{
							Name: "p2",
							Properties: map[string]interface{}{
								"provp1": "added",
							},
						},
					},
				})
			})
			It("fails if there is a module in the extension that doesn't exist in the original MTA", func() {
				checkModuleMergeFails(Module{
					Name: "module2",
					Properties: map[string]interface{}{
						"p1": "the p",
					},
					Parameters: map[string]interface{}{
						"p1p1": "the paaram",
					},
					BuildParams: map[string]interface{}{
						"p1b1": "the paaram",
					},
				}, ModuleExt{
					Name: "module1",
					Properties: map[string]interface{}{
						"p1": "the p",
					},
					Parameters: map[string]interface{}{
						"p1p1": "the paaram",
					},
					BuildParams: map[string]interface{}{
						"p1b1": "the paaram",
					},
				}, unknownModuleErrorMsg, "module1")
			})
			It("fails if it can't merge module properties", func() {
				checkModuleMergeFails(Module{
					Name: "module1",
					Properties: map[string]interface{}{
						"p1": "value",
					},
				}, ModuleExt{
					Name: "module1",
					Properties: map[string]interface{}{
						"p1": map[string]interface{}{
							"p1p2": "changed",
						},
						"p2": "added",
					},
				}, overwriteScalarWithStructuredErrorMsg, "p1")
			})
			It("fails if it can't merge module properties because of metadata", func() {
				checkModuleMergeFails(Module{
					Name: "module1",
					Properties: map[string]interface{}{
						"p1": "value",
					},
					PropertiesMetaData: map[string]MetaData{
						"p1": {
							OverWritable: falsePtr(),
						},
					},
				}, ModuleExt{
					Name: "module1",
					Properties: map[string]interface{}{
						"p1": "changed",
					},
				}, overwriteNonOverwritableErrorMsg, "p1")
			})
			It("fails if it can't merge module parameters", func() {
				checkModuleMergeFails(Module{
					Name: "module1",
					Parameters: map[string]interface{}{
						"p1p1": "value",
					},
				}, ModuleExt{
					Name: "module1",
					Parameters: map[string]interface{}{
						"p1p1": map[string]interface{}{
							"p1p2c1": "added",
						},
						"p1p2": map[string]interface{}{
							"p1p2c1": "changed",
						},
					},
				}, overwriteScalarWithStructuredErrorMsg, "p1p1")
			})
			It("fails if it can't merge module parameters because of metadata", func() {
				checkModuleMergeFails(Module{
					Name: "module1",
					Parameters: map[string]interface{}{
						"p1p1": "value",
					},
					ParametersMetaData: map[string]MetaData{
						"p1p1": {
							OverWritable: falsePtr(),
						},
					},
				}, ModuleExt{
					Name: "module1",
					Parameters: map[string]interface{}{
						"p1p1": "changed",
					},
				}, overwriteNonOverwritableErrorMsg, "p1p1")
			})
			It("fails if it can't merge module build parameters", func() {
				checkModuleMergeFails(Module{
					Name: "module1",
					BuildParams: map[string]interface{}{
						"p1b1": "value",
						"p1b2": "value",
					},
				}, ModuleExt{
					Name: "module1",
					BuildParams: map[string]interface{}{
						"p1b1": "changed",
						"p1b2": map[string]interface{}{
							"p1b2c1": "added",
						},
					},
				}, overwriteScalarWithStructuredErrorMsg, "p1b2")
			})
			It("fails if there is a module provides in the extension that doesn't exist in the original MTA", func() {
				checkModuleMergeFails(Module{
					Name: "module1",
					Provides: []Provides{
						{
							Name:   "p1",
							Public: true,
							Properties: map[string]interface{}{
								"provp1": "the prop",
							},
						},
					},
				}, ModuleExt{
					Name: "module1",
					Provides: []Provides{
						{
							Name:   "p2",
							Public: true,
							Properties: map[string]interface{}{
								"provp1": "changed",
							},
						},
					},
				}, unknownModuleProvidesErrorMsg, "p2", "module1")
			})
			It("fails if it can't merge module provides properties", func() {
				checkModuleMergeFails(Module{
					Name: "module1",
					Provides: []Provides{
						{
							Name:   "p1",
							Public: true,
							Properties: map[string]interface{}{
								"provp1":  "value",
								"provep2": "value",
							},
						},
					},
				}, ModuleExt{
					Name: "module1",
					Provides: []Provides{
						{
							Name:   "p1",
							Public: true,
							Properties: map[string]interface{}{
								"provp1": "changed",
								"provep2": map[string]interface{}{
									"provep2t1": "added",
								},
							},
						},
					},
				}, overwriteScalarWithStructuredErrorMsg, "provep2")
			})
			It("fails if it can't merge module provides properties because of metadata", func() {
				checkModuleMergeFails(Module{
					Name: "module1",
					Provides: []Provides{
						{
							Name:   "p1",
							Public: true,
							Properties: map[string]interface{}{
								"provp1": "value",
							},
							PropertiesMetaData: map[string]MetaData{
								"provp1": {
									OverWritable: falsePtr(),
								},
							},
						},
					},
				}, ModuleExt{
					Name: "module1",
					Provides: []Provides{
						{
							Name:   "p1",
							Public: true,
							Properties: map[string]interface{}{
								"provp1": "changed",
							},
						},
					},
				}, overwriteNonOverwritableErrorMsg, "provp1")
			})
		})

		Context("resources", func() {
			It("overrides the active field", func() {
				mtaObj := MTA{
					Resources: []*Resource{
						{
							Name:   "rb",
							Active: falsePtr(),
						},
						{
							Name:   "rc",
							Active: truePtr(),
						},
					},
				}
				extMta := EXT{
					Resources: []*ResourceExt{
						{
							Name:   "rc",
							Active: falsePtr(),
						},
						{
							Name:   "rb",
							Active: truePtr(),
						},
					},
				}

				err := Merge(&mtaObj, &extMta)

				Ω(err).Should(Succeed())
				Ω(mtaObj).Should(Equal(MTA{
					Resources: []*Resource{
						{
							Name:   "rb",
							Active: truePtr(),
						},
						{
							Name:   "rc",
							Active: falsePtr(),
						},
					},
				}))
				Ω(extMta).Should(Equal(EXT{
					Resources: []*ResourceExt{
						{
							Name:   "rc",
							Active: falsePtr(),
						},
						{
							Name:   "rb",
							Active: truePtr(),
						},
					},
				}))
			})
			It("overrides the active field when one of the fields is nil", func() {
				mtaObj := MTA{
					Resources: []*Resource{
						{
							Name:   "rb",
							Active: nil,
						},
						{
							Name:   "rc",
							Active: nil,
						},
						{
							Name:   "rd",
							Active: falsePtr(),
						},
						{
							Name:   "re",
							Active: truePtr(),
						},
						{
							Name:   "rf",
							Active: nil,
						},
					},
				}
				extMta := EXT{
					Resources: []*ResourceExt{
						{
							Name:   "rc",
							Active: falsePtr(),
						},
						{
							Name:   "rb",
							Active: truePtr(),
						},
						{
							Name:   "rd",
							Active: nil,
						},
						{
							Name:   "re",
							Active: nil,
						},
						{
							Name:   "rf",
							Active: nil,
						},
					},
				}

				err := Merge(&mtaObj, &extMta)

				Ω(err).Should(Succeed())
				Ω(mtaObj).Should(Equal(MTA{
					Resources: []*Resource{
						{
							Name:   "rb",
							Active: truePtr(),
						},
						{
							Name:   "rc",
							Active: falsePtr(),
						},
						{
							Name:   "rd",
							Active: falsePtr(),
						},
						{
							Name:   "re",
							Active: truePtr(),
						},
						{
							Name:   "rf",
							Active: nil,
						},
					},
				}))
				Ω(extMta).Should(Equal(EXT{
					Resources: []*ResourceExt{
						{
							Name:   "rc",
							Active: falsePtr(),
						},
						{
							Name:   "rb",
							Active: truePtr(),
						},
						{
							Name:   "rd",
							Active: nil,
						},
						{
							Name:   "re",
							Active: nil,
						},
						{
							Name:   "rf",
							Active: nil,
						},
					},
				}))
			})
			It("merges the resource properties", func() {
				checkResourceMerge(Resource{
					Name: "resource1",
					Properties: map[string]interface{}{
						"p1": "the p",
					},
				}, ResourceExt{
					Name: "resource1",
					Properties: map[string]interface{}{
						"p1": "changed",
						"p2": "added",
					},
				}, Resource{
					Name: "resource1",
					Properties: map[string]interface{}{
						"p1": "changed",
						"p2": "added",
					},
				})
			})
			It("merges the resource properties when original map is nil", func() {
				checkResourceMerge(Resource{
					Name: "resource1",
				}, ResourceExt{
					Name: "resource1",
					Properties: map[string]interface{}{
						"p2": "added",
					},
				}, Resource{
					Name: "resource1",
					Properties: map[string]interface{}{
						"p2": "added",
					},
				})
			})
			It("merges the resource parameters", func() {
				checkResourceMerge(Resource{
					Name: "resource1",
					Parameters: map[string]interface{}{
						"p1": "the p",
					},
				}, ResourceExt{
					Name: "resource1",
					Parameters: map[string]interface{}{
						"p1": "changed",
						"p2": "added",
					},
				}, Resource{
					Name: "resource1",
					Parameters: map[string]interface{}{
						"p1": "changed",
						"p2": "added",
					},
				})
			})
			It("merges the resource parameters when original map is nil", func() {
				checkResourceMerge(Resource{
					Name: "resource1",
				}, ResourceExt{
					Name: "resource1",
					Parameters: map[string]interface{}{
						"p2": "added",
					},
				}, Resource{
					Name: "resource1",
					Parameters: map[string]interface{}{
						"p2": "added",
					},
				})
			})
			Context("requires", func() {
				It("merges the resource requires parameters", func() {
					checkResourceMerge(Resource{
						Name: "resource1",
						Requires: []Requires{
							{
								Name: "r1",
								Parameters: map[string]interface{}{
									"provp1": "the prop",
								},
								ParametersMetaData: map[string]MetaData{
									"provp1": {
										OverWritable: nil,
									},
								},
							},
							{
								Name: "r2",
							},
						},
					}, ResourceExt{
						Name: "resource1",
						Requires: []Requires{
							{
								Name: "r1",
								Parameters: map[string]interface{}{
									"provp1": "changed",
									"provep2": map[string]interface{}{
										"provep2t1": "added",
									},
								},
							},
							{
								Name: "r2",
								Parameters: map[string]interface{}{
									"provp1": "added",
								},
							},
						},
					}, Resource{
						Name: "resource1",
						Requires: []Requires{
							{
								Name: "r1",
								Parameters: map[string]interface{}{
									"provp1": "changed",
									"provep2": map[string]interface{}{
										"provep2t1": "added",
									},
								},
								ParametersMetaData: map[string]MetaData{
									"provp1": {
										OverWritable: nil,
									},
								},
							},
							{
								Name: "r2",
								Parameters: map[string]interface{}{
									"provp1": "added",
								},
							},
						},
					})
				})
				It("merges the resource requires properties", func() {
					checkResourceMerge(Resource{
						Name: "resource1",
						Requires: []Requires{
							{
								Name: "r1",
								Properties: map[string]interface{}{
									"provp1": "the prop",
								},
								PropertiesMetaData: map[string]MetaData{
									"provp1": {
										OverWritable: truePtr(),
									},
								},
							},
							{
								Name: "r2",
							},
						},
					}, ResourceExt{
						Name: "resource1",
						Requires: []Requires{
							{
								Name: "r1",
								Properties: map[string]interface{}{
									"provp1": "changed",
									"provep2": map[string]interface{}{
										"provep2t1": "added",
									},
								},
							},
							{
								Name: "r2",
								Properties: map[string]interface{}{
									"provp1": "added",
								},
							},
						},
					}, Resource{
						Name: "resource1",
						Requires: []Requires{
							{
								Name: "r1",
								Properties: map[string]interface{}{
									"provp1": "changed",
									"provep2": map[string]interface{}{
										"provep2t1": "added",
									},
								},
								PropertiesMetaData: map[string]MetaData{
									"provp1": {
										OverWritable: truePtr(),
									},
								},
							},
							{
								Name: "r2",
								Properties: map[string]interface{}{
									"provp1": "added",
								},
							},
						},
					})
				})
				It("fails if there is a resource requires in the extension that doesn't exist in the original MTA", func() {
					checkResourceMergeFails(Resource{
						Name: "resource1",
						Requires: []Requires{
							{
								Name: "r1",
								Properties: map[string]interface{}{
									"provp1": "the prop",
								},
							},
						},
					}, ResourceExt{
						Name: "resource1",
						Requires: []Requires{
							{
								Name: "r2",
								Properties: map[string]interface{}{
									"provp1": "changed",
								},
							},
						},
					}, unknownResourceRequiresErrorMsg, "r2", "resource1")
				})
				It("fails if it can't merge resource requires properties", func() {
					checkResourceMergeFails(Resource{
						Name: "resource1",
						Requires: []Requires{
							{
								Name: "p1",
								Properties: map[string]interface{}{
									"provp1":  "value",
									"provep2": "value",
								},
							},
						},
					}, ResourceExt{
						Name: "resource1",
						Requires: []Requires{
							{
								Name: "p1",
								Properties: map[string]interface{}{
									"provp1": "changed",
									"provep2": map[string]interface{}{
										"provep2t1": "added",
									},
								},
							},
						},
					}, overwriteScalarWithStructuredErrorMsg, "provep2")
				})
				It("fails if it can't merge resource requires properties because of metadata", func() {
					checkResourceMergeFails(Resource{
						Name: "resource1",
						Requires: []Requires{
							{
								Name: "p1",
								Properties: map[string]interface{}{
									"provp1": "value",
								},
								PropertiesMetaData: map[string]MetaData{
									"provp1": {
										OverWritable: falsePtr(),
									},
								},
							},
						},
					}, ResourceExt{
						Name: "resource1",
						Requires: []Requires{
							{
								Name: "p1",
								Properties: map[string]interface{}{
									"provp1": "changed",
								},
							},
						},
					}, overwriteNonOverwritableErrorMsg, "provp1")
				})
				It("fails if it can't merge resource requires parameters", func() {
					checkResourceMergeFails(Resource{
						Name: "resource1",
						Requires: []Requires{
							{
								Name: "p1",
								Parameters: map[string]interface{}{
									"provp1":  "value",
									"provep2": "value",
								},
							},
						},
					}, ResourceExt{
						Name: "resource1",
						Requires: []Requires{
							{
								Name: "p1",
								Parameters: map[string]interface{}{
									"provp1": "changed",
									"provep2": map[string]interface{}{
										"provep2t1": "added",
									},
								},
							},
						},
					}, overwriteScalarWithStructuredErrorMsg, "provep2")
				})
				It("fails if it can't merge resource requires parameters because of metadata", func() {
					checkResourceMergeFails(Resource{
						Name: "resource1",
						Requires: []Requires{
							{
								Name: "p1",
								Parameters: map[string]interface{}{
									"provp1": "value",
								},
								ParametersMetaData: map[string]MetaData{
									"provp1": {
										OverWritable: falsePtr(),
									},
								},
							},
						},
					}, ResourceExt{
						Name: "resource1",
						Requires: []Requires{
							{
								Name: "p1",
								Parameters: map[string]interface{}{
									"provp1": "changed",
								},
							},
						},
					}, overwriteNonOverwritableErrorMsg, "provp1")
				})
			})
			It("fails if there is a resource in the extension that doesn't exist in the original MTA", func() {
				checkResourceMergeFails(Resource{
					Name:   "rb",
					Active: falsePtr(),
				}, ResourceExt{
					Name:   "ra",
					Active: truePtr(),
				}, unknownResourceErrorMsg, "ra")
			})
			It("fails if it can't merge resource properties", func() {
				checkResourceMergeFails(Resource{
					Name: "resource1",
					Properties: map[string]interface{}{
						"p1": "value",
					},
				}, ResourceExt{
					Name: "resource1",
					Properties: map[string]interface{}{
						"p1": map[string]interface{}{
							"p1p2": "changed",
						},
						"p2": "added",
					},
				}, overwriteScalarWithStructuredErrorMsg, "p1")
			})
			It("fails if it can't merge resource properties because of metadata", func() {
				checkResourceMergeFails(Resource{
					Name: "resource1",
					Properties: map[string]interface{}{
						"p1": "value",
					},
					PropertiesMetaData: map[string]MetaData{
						"p1": {
							OverWritable: falsePtr(),
						},
					},
				}, ResourceExt{
					Name: "resource1",
					Properties: map[string]interface{}{
						"p1": "changed",
					},
				}, overwriteNonOverwritableErrorMsg, "p1")
			})
			It("fails if it can't merge resource parameters", func() {
				checkResourceMergeFails(Resource{
					Name: "resource1",
					Parameters: map[string]interface{}{
						"p1p1": "value",
					},
				}, ResourceExt{
					Name: "resource1",
					Parameters: map[string]interface{}{
						"p1p1": map[string]interface{}{
							"p1p2c1": "added",
						},
						"p1p2": map[string]interface{}{
							"p1p2c1": "changed",
						},
					},
				}, overwriteScalarWithStructuredErrorMsg, "p1p1")
			})
			It("fails if it can't merge resource parameters because of metadata", func() {
				checkResourceMergeFails(Resource{
					Name: "resource1",
					Parameters: map[string]interface{}{
						"p1p1": "value",
					},
					ParametersMetaData: map[string]MetaData{
						"p1p1": {
							OverWritable: falsePtr(),
						},
					},
				}, ResourceExt{
					Name: "resource1",
					Parameters: map[string]interface{}{
						"p1p1": "changed",
					},
				}, overwriteNonOverwritableErrorMsg, "p1p1")
			})
		})

		It("merges both the modules and the resources in the MTA object", func() {
			mta := MTA{
				Modules: []*Module{
					{
						Name: "module1",
						Properties: map[string]interface{}{
							"p1": "the p",
						},
					},
				},
				Resources: []*Resource{
					{
						Name:   "rb",
						Active: falsePtr(),
					},
					{
						Name:   "rd",
						Active: truePtr(),
					},
				},
			}
			extMta := EXT{
				Modules: []*ModuleExt{
					{
						Name: "module1",
						Properties: map[string]interface{}{
							"p1": "changed",
							"p2": "added",
						},
					},
				},
				Resources: []*ResourceExt{
					{
						Name:   "rb",
						Active: truePtr(),
					},
				},
			}

			err := Merge(&mta, &extMta)

			Ω(err).Should(Succeed())
			Ω(mta).Should(Equal(MTA{
				Modules: []*Module{
					{
						Name: "module1",
						Properties: map[string]interface{}{
							"p1": "changed",
							"p2": "added",
						},
					},
				},
				Resources: []*Resource{
					{
						Name:   "rb",
						Active: truePtr(),
					},
					{
						Name:   "rd",
						Active: truePtr(),
					},
				},
			}))
			// Check the extension didn't change
			Ω(extMta).Should(Equal(EXT{
				Modules: []*ModuleExt{
					{
						Name: "module1",
						Properties: map[string]interface{}{
							"p1": "changed",
							"p2": "added",
						},
					},
				},
				Resources: []*ResourceExt{
					{
						Name:   "rb",
						Active: truePtr(),
					},
				},
			}))
		})

		It("returns error when merge fails on overwriting scalar with structured value", func() {
			content, err := readFile(getTestPath("mta.yaml"))
			Ω(err).Should(Succeed())
			mta, err := Unmarshal(content)
			Ω(err).Should(Succeed())

			// The map in lines 12-13 is returned as map[interface{}]interface{} instead of map[string]interface{}
			extContent, err := readFile(getTestPath("overwrite_error.mtaext"))
			Ω(err).Should(Succeed())
			ext, err := UnmarshalExt(extContent)
			Ω(err).Should(Succeed())

			err = Merge(mta, ext)
			Ω(err).Should(HaveOccurred())
			Ω(err.Error()).Should(ContainSubstring(overwriteScalarWithStructuredErrorMsg, "name"))
		})

	})
})

func checkMerge(mtaObj MTA, extMta EXT, expected MTA) {
	err := Merge(&mtaObj, &extMta)
	Ω(err).Should(Succeed())
	Ω(mtaObj).Should(Equal(expected))
}

func checkMergeFails(mtaObj MTA, extMta EXT, msg string, args ...interface{}) {
	err := Merge(&mtaObj, &extMta)

	Ω(err).Should(HaveOccurred())
	Ω(err.Error()).Should(ContainSubstring(msg, args...))
}

func checkModuleMerge(mtaModule Module, mtaExtModule ModuleExt, expected Module) {
	checkMerge(MTA{
		Modules: []*Module{
			&mtaModule,
		},
	}, EXT{
		Modules: []*ModuleExt{
			&mtaExtModule,
		},
	}, MTA{
		Modules: []*Module{
			&expected,
		},
	})
}

func checkModuleMergeFails(mtaModule Module, mtaExtModule ModuleExt, msg string, args ...interface{}) {
	checkMergeFails(MTA{
		Modules: []*Module{
			&mtaModule,
		},
	}, EXT{
		Modules: []*ModuleExt{
			&mtaExtModule,
		},
	}, msg, args...)
}

func checkResourceMerge(mtaResource Resource, mtaExtResource ResourceExt, expected Resource) {
	checkMerge(MTA{
		Resources: []*Resource{
			&mtaResource,
		},
	}, EXT{
		Resources: []*ResourceExt{
			&mtaExtResource,
		},
	}, MTA{
		Resources: []*Resource{
			&expected,
		},
	})
}

func checkResourceMergeFails(mtaResource Resource, mtaExtResource ResourceExt, msg string, args ...interface{}) {
	checkMergeFails(MTA{
		Resources: []*Resource{
			&mtaResource,
		},
	}, EXT{
		Resources: []*ResourceExt{
			&mtaExtResource,
		},
	}, msg, args...)
}

func checkExtendMap(m map[string]interface{}, ext map[string]interface{}, meta map[string]MetaData, expected map[string]interface{}) {
	// We don't want to change the sent map
	mCopy := copyMap(m)
	metaBeforeExtendMap := copyMetaMap(meta)
	extBeforeExtendMap := copyMap(ext)

	err := extendMap(&mCopy, meta, ext)
	Ω(err).Should(Succeed())
	Ω(meta).Should(Equal(metaBeforeExtendMap))
	Ω(ext).Should(Equal(extBeforeExtendMap))
	Ω(mCopy).Should(Equal(expected))
}

func checkExtendMapFails(m map[string]interface{}, ext map[string]interface{}, meta map[string]MetaData, errorMsg string, args ...interface{}) {
	// We don't want to change the sent map
	mCopy := copyMap(m)
	metaBeforeExtendMap := copyMetaMap(meta)
	extBeforeExtendMap := copyMap(ext)

	err := extendMap(&mCopy, meta, ext)
	Ω(err).ShouldNot(Succeed())
	Ω(err.Error()).Should(ContainSubstring(errorMsg, args...))
	Ω(meta).Should(Equal(metaBeforeExtendMap))
	Ω(ext).Should(Equal(extBeforeExtendMap))
	// Note: mCopy might be changed even if extendMap fails, since the map is merged in-place
}

func copyMap(source map[string]interface{}) map[string]interface{} {
	if source == nil {
		return nil
	}
	result := make(map[string]interface{})
	for key, value := range source {
		if mValue, ok, converted := getMapValue(value); ok {
			if converted {
				result[key] = copyInterfaceMap(value.(map[interface{}]interface{}))
			} else {
				result[key] = copyMap(mValue)
			}
		} else {
			result[key] = value
		}
	}
	return result
}

func copyInterfaceMap(source map[interface{}]interface{}) map[interface{}]interface{} {
	if source == nil {
		return nil
	}
	result := make(map[interface{}]interface{})
	for key, value := range source {
		if mValue, ok, converted := getMapValue(value); ok {
			if converted {
				result[key] = copyInterfaceMap(value.(map[interface{}]interface{}))
			} else {
				result[key] = copyMap(mValue)
			}
		} else {
			result[key] = value
		}
	}
	return result
}

func copyMetaMap(source map[string]MetaData) map[string]MetaData {
	if source == nil {
		return nil
	}
	result := make(map[string]MetaData)
	for key, value := range source {
		result[key] = value
	}
	return result
}
