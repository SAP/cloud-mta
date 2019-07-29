package mta

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"os"
	"path/filepath"
)

var _ = Describe("Extension MTA", func() {
	var _ = Describe("UnmarshalExt", func() {
		It("Sanity", func() {
			wd, err := os.Getwd()
			Ω(err).Should(Succeed())
			content, err := readFile(filepath.Join(wd, "testdata", "mta.yaml"))
			Ω(err).Should(Succeed())
			m, err := UnmarshalExt(content)
			Ω(err).Should(Succeed())
			Ω(len(m.Modules)).Should(Equal(2))
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
					OverWritable: true,
				},
			},
			"metadata is not specified for the keys": {
				"c": {
					OverWritable: false,
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
							OverWritable: false,
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
							OverWritable: false,
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
								OverWritable: false,
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
							OverWritable: false,
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
						Entry("overrides inner values, adds inner values and adds new values", mapWithOneKey, mapWithTwoKeys, mapWithTwoKeys),
						Entry("copies ext map when original map is nil", nil, mapWithTwoKeys, mapWithTwoKeys),
						Entry("doesn't change original when ext is nil", mapWithTwoKeys, nil, mapWithTwoKeys),
						Entry("returns nil when original and ext are nil", nil, nil, nil),
					)
				}

				DescribeTable("overwritable is false", func(mta map[string]interface{}, ext map[string]interface{}, expected map[string]interface{}) {
					var meta = map[string]MetaData{
						"b": {
							OverWritable: false,
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
							OverWritable: false,
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

				for desc, metadata := range overwritableScenarios {
					DescribeTable(desc, func(mta map[string]interface{}, ext map[string]interface{}, err string, args ...interface{}) {
						checkExtendMapFails(mta, ext, metadata, err, args...)
					},
						Entry("fails when overriding scalar with map", mapWithTwoKeys, mapWithOneKey, overwriteScalarWithStructuredErrorMsg, "e"),
						Entry("fails when overriding map with scalar", mapWithOneKey, mapWithTwoKeys, overwriteStructuredWithScalarErrorMsg, "e"),
					)
				}
			})
		})
	})

	var _ = Describe("Merge", func() {
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
									OverWritable: true,
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
								"provep2": map[string]interface{}{
									"provep2t1": "added",
								},
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
									OverWritable: true,
								},
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
							OverWritable: false,
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
							OverWritable: false,
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
									OverWritable: false,
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
							Active: false,
						},
						{
							Name:   "rc",
							Active: true,
						},
					},
				}
				extMta := EXT{
					Resources: []*ResourceExt{
						{
							Name:   "rc",
							Active: false,
						},
						{
							Name:   "rb",
							Active: true,
						},
					},
				}

				err := Merge(&mtaObj, &extMta)

				Ω(err).Should(Succeed())
				Ω(mtaObj).Should(Equal(MTA{
					Resources: []*Resource{
						{
							Name:   "rb",
							Active: true,
						},
						{
							Name:   "rc",
							Active: false,
						},
					},
				}))
				Ω(extMta).Should(Equal(EXT{
					Resources: []*ResourceExt{
						{
							Name:   "rc",
							Active: false,
						},
						{
							Name:   "rb",
							Active: true,
						},
					},
				}))
			})
			It("fails if there is a resource in the extension that doesn't exist in the original MTA", func() {
				extMta := EXT{
					Resources: []*ResourceExt{
						{
							Name:   "ra",
							Active: true,
						},
					},
				}
				mtaObj := MTA{
					Resources: []*Resource{
						{
							Name:   "rb",
							Active: false,
						},
					},
				}

				err := Merge(&mtaObj, &extMta)

				Ω(err).Should(HaveOccurred())
				Ω(err.Error()).Should(ContainSubstring(unknownResourceErrorMsg, "ra"))
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
						Active: false,
					},
					{
						Name:   "rd",
						Active: true,
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
						Active: true,
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
						Active: true,
					},
					{
						Name:   "rd",
						Active: true,
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
						Active: true,
					},
				},
			}))
		})
	})
})

func checkModuleMerge(mtaModule Module, mtaExtModule ModuleExt, expected Module) {
	mtaObj := MTA{
		Modules: []*Module{
			&mtaModule,
		},
	}
	extMta := EXT{
		Modules: []*ModuleExt{
			&mtaExtModule,
		},
	}

	err := Merge(&mtaObj, &extMta)

	Ω(err).Should(Succeed())
	Ω(mtaObj).Should(Equal(MTA{
		Modules: []*Module{
			&expected,
		},
	}))
}

func checkModuleMergeFails(mtaModule Module, mtaExtModule ModuleExt, msg string, args ...interface{}) {
	mtaObj := MTA{
		Modules: []*Module{
			&mtaModule,
		},
	}
	extMta := EXT{
		Modules: []*ModuleExt{
			&mtaExtModule,
		},
	}

	err := Merge(&mtaObj, &extMta)

	Ω(err).Should(HaveOccurred())
	Ω(err.Error()).Should(ContainSubstring(msg, args...))
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
		if mValue, ok := value.(map[string]interface{}); ok {
			result[key] = copyMap(mValue)
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
