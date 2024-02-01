package old_configmanager

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testCase struct {
	name      string
	cfg       Config
	originCfg Config
	data      map[string]interface{}
	expected  Config
}

func updateConfigTestCases() []*testCase {
	type PrimitiveConfig struct {
		StrField  string `cfg:"str"`
		NumField  int    `cfg:"number"`
		BoolField bool   `cfg:"bool"`
	}
	type ObjectConfig struct {
		ObjectField PrimitiveConfig `cfg:"object"`
	}
	type PointerConfig struct {
		PointerField *string `cfg:"pointer"`
	}
	type ArrayConfig struct {
		ArrayField []string `cfg:"array"`
	}
	type MapConfig struct {
		MapField map[string]string `cfg:"map"`
	}
	type ComplexPointerConfig struct {
		PointerStruct *PrimitiveConfig `cfg:"pointer_struct"`
		PointerArray  *[]string        `cfg:"pointer_array"`
		PointerMap    *map[string]int  `cfg:"pointer_map"`
		DoublePointer **string         `cfg:"double_pointer"`
	}
	type ComplexArrayConfig struct {
		ArrayStruct  []PrimitiveConfig `cfg:"array_struct"`
		ArrayArray   [][]string        `cfg:"array_array"`
		ArrayMap     []map[string]int  `cfg:"array_map"`
		ArrayPointer []*string         `cfg:"array_pointer"`
	}
	type ComplexMapConfig struct {
		MapStruct  map[string]PrimitiveConfig `cfg:"map_struct"`
		MapArray   map[string][]string        `cfg:"map_array"`
		MapMap     map[string]map[string]int  `cfg:"map_map"`
		MapPointer map[string]*string         `cfg:"map_pointer"`
	}

	testCases := []*testCase{
		func() *testCase {
			cfg := PrimitiveConfig{
				StrField:  "str",
				NumField:  123,
				BoolField: true,
			}
			return &testCase{
				name:      "UpdateConfig_Primitive",
				cfg:       cfg,
				originCfg: cloneConfig[PrimitiveConfig](cfg),
				data: map[string]interface{}{
					"test.str":    "newstr",
					"test.number": 456,
					"test.bool":   false,
				},
				expected: PrimitiveConfig{
					StrField:  "newstr",
					NumField:  456,
					BoolField: false,
				},
			}
		}(),
		func() *testCase {
			cfg := ObjectConfig{
				ObjectField: PrimitiveConfig{
					StrField:  "str",
					NumField:  123,
					BoolField: true,
				},
			}
			return &testCase{
				name:      "UpdateConfig_Object",
				cfg:       cfg,
				originCfg: cloneConfig[ObjectConfig](cfg),
				data: map[string]interface{}{
					"test.object.str":    "newstr",
					"test.object.number": 456,
					"test.object.bool":   false,
				},
				expected: ObjectConfig{
					ObjectField: PrimitiveConfig{
						StrField:  "newstr",
						NumField:  456,
						BoolField: false,
					},
				},
			}
		}(),
		func() *testCase {
			str := "str"
			cfg := PointerConfig{
				PointerField: &str,
			}
			return &testCase{
				name:      "UpdateConfig_Pointer",
				cfg:       cfg,
				originCfg: cloneConfig[PointerConfig](cfg),
				data: map[string]interface{}{
					"test.pointer": ptr("newStr"),
				},
				expected: PointerConfig{
					PointerField: ptr("newStr"),
				},
			}
		}(),
		func() *testCase {
			cfg := PointerConfig{}
			return &testCase{
				name:      "UpdateConfig_NilPointer",
				cfg:       cfg,
				originCfg: cloneConfig[PointerConfig](cfg),
				data: map[string]interface{}{
					"test.pointer": ptr("str"),
				},
				expected: PointerConfig{
					PointerField: ptr("str"),
				},
			}
		}(),
		func() *testCase {
			cfg := ArrayConfig{
				ArrayField: []string{"str1", "str2"},
			}
			return &testCase{
				name:      "UpdateConfig_Array_ExistedElement",
				cfg:       cfg,
				originCfg: cloneConfig[ArrayConfig](cfg),
				data: map[string]interface{}{
					"test.array.0": "newstr1",
				},
				expected: ArrayConfig{
					ArrayField: []string{"newstr1", "str2"},
				},
			}
		}(),
		func() *testCase {
			cfg := ArrayConfig{
				ArrayField: []string{"str1"},
			}
			return &testCase{
				name:      "UpdateConfig_Array_NewElement",
				cfg:       cfg,
				originCfg: cloneConfig[ArrayConfig](cfg),
				data: map[string]interface{}{
					"test.array.1": "newstr",
				},
				expected: ArrayConfig{
					ArrayField: []string{"str1", "newstr"},
				},
			}
		}(),
		func() *testCase {
			cfg := ArrayConfig{}
			return &testCase{
				name:      "UpdateConfig_NilArray",
				cfg:       cfg,
				originCfg: cloneConfig[ArrayConfig](cfg),
				data: map[string]interface{}{
					"test.array.0": "str1",
					"test.array.1": "str2",
				},
				expected: ArrayConfig{
					ArrayField: []string{"str1", "str2"},
				},
			}
		}(),
		func() *testCase {
			cfg := MapConfig{
				MapField: map[string]string{
					"key1": "str1",
				},
			}
			return &testCase{
				name:      "UpdateConfig_Map_ExistedKey",
				cfg:       cfg,
				originCfg: cloneConfig[MapConfig](cfg),
				data: map[string]interface{}{
					"test.map.key1": "newstr1",
				},
				expected: MapConfig{
					MapField: map[string]string{
						"key1": "newstr1",
					},
				},
			}
		}(),
		func() *testCase {
			cfg := MapConfig{
				MapField: map[string]string{},
			}
			return &testCase{
				name:      "UpdateConfig_Map_NewKey",
				cfg:       cfg,
				originCfg: cloneConfig[MapConfig](cfg),
				data: map[string]interface{}{
					"test.map.key1": "str1",
				},
				expected: MapConfig{
					MapField: map[string]string{
						"key1": "str1",
					},
				},
			}
		}(),
		func() *testCase {
			cfg := MapConfig{}
			return &testCase{
				name:      "UpdateConfig_NilMap",
				cfg:       cfg,
				originCfg: cloneConfig[MapConfig](cfg),
				data: map[string]interface{}{
					"test.map.key1": "str1",
				},
				expected: MapConfig{
					MapField: map[string]string{
						"key1": "str1",
					},
				},
			}
		}(),
		func() *testCase {
			cfg := ComplexPointerConfig{
				PointerStruct: &PrimitiveConfig{
					StrField:  "str",
					NumField:  123,
					BoolField: true,
				},
				PointerArray: &[]string{"str1", "str2"},
				PointerMap: &map[string]int{
					"key1": 123,
				},
				DoublePointer: ptr(ptr("str")),
			}
			return &testCase{
				name:      "UpdateConfig_ComplexPointer",
				cfg:       cfg,
				originCfg: cloneConfig[ComplexPointerConfig](cfg),
				data: map[string]interface{}{
					"test.pointer_struct.str":    "newstr",
					"test.pointer_struct.number": 456,
					"test.pointer_struct.bool":   false,
					"test.pointer_array.0":       "newstr1",
					"test.pointer_map.key1":      456,
					"test.double_pointer":        ptr(ptr("newstr")),
				},
				expected: ComplexPointerConfig{
					PointerStruct: &PrimitiveConfig{
						StrField:  "newstr",
						NumField:  456,
						BoolField: false,
					},
					PointerArray: &[]string{"newstr1", "str2"},
					PointerMap: &map[string]int{
						"key1": 456,
					},
					DoublePointer: ptr(ptr("newstr")),
				},
			}
		}(),
		func() *testCase {
			cfg := ComplexArrayConfig{
				ArrayStruct: []PrimitiveConfig{
					{
						StrField:  "str",
						NumField:  123,
						BoolField: true,
					},
				},
				ArrayArray: [][]string{
					{"str1", "str2"},
				},
				ArrayMap: []map[string]int{
					{
						"key1": 123,
					},
				},
				ArrayPointer: []*string{
					ptr("str"),
				},
			}
			return &testCase{
				name:      "UpdateConfig_ComplexArray",
				cfg:       cfg,
				originCfg: cloneConfig[ComplexArrayConfig](cfg),
				data: map[string]interface{}{
					"test.array_struct.0.str":    "newstr",
					"test.array_struct.0.number": 456,
					"test.array_struct.0.bool":   false,
					"test.array_array.0.0":       "newstr1",
					"test.array_map.0.key1":      456,
					"test.array_pointer.0":       ptr("newstr"),
				},
				expected: ComplexArrayConfig{
					ArrayStruct: []PrimitiveConfig{
						{
							StrField:  "newstr",
							NumField:  456,
							BoolField: false,
						},
					},
					ArrayArray: [][]string{
						{"newstr1", "str2"},
					},
					ArrayMap: []map[string]int{
						{
							"key1": 456,
						},
					},
					ArrayPointer: []*string{
						ptr("newstr"),
					},
				},
			}
		}(),
		func() *testCase {
			cfg := ComplexMapConfig{
				MapStruct: map[string]PrimitiveConfig{
					"key1": {
						StrField:  "str",
						NumField:  123,
						BoolField: true,
					},
				},
				MapArray: map[string][]string{
					"key1": {"str1", "str2"},
				},
				MapMap: map[string]map[string]int{
					"key1": {
						"key1": 123,
					},
				},
				MapPointer: map[string]*string{
					"key1": ptr("str"),
				},
			}
			return &testCase{
				name:      "UpdateConfig_ComplexMap",
				cfg:       cfg,
				originCfg: cloneConfig[ComplexMapConfig](cfg),
				data: map[string]interface{}{
					"test.map_struct.key1.str":    "newstr",
					"test.map_struct.key1.number": 456,
					"test.map_struct.key1.bool":   false,
					"test.map_array.key1.0":       "newstr1",
					"test.map_map.key1.key1":      456,
					"test.map_pointer.key1":       ptr("newstr"),
				},
				expected: ComplexMapConfig{
					MapStruct: map[string]PrimitiveConfig{
						"key1": {
							StrField:  "newstr",
							NumField:  456,
							BoolField: false,
						},
					},
					MapArray: map[string][]string{
						"key1": {"newstr1", "str2"},
					},
					MapMap: map[string]map[string]int{
						"key1": {
							"key1": 456,
						},
					},
					MapPointer: map[string]*string{
						"key1": ptr("newstr"),
					},
				},
			}
		}(),
	}

	return testCases
}

func TestUpdateConfig(t *testing.T) {
	testCases := updateConfigTestCases()
	for _, tc := range testCases {

		t.Run(tc.name, func(t *testing.T) {
			cfg := tc.cfg

			mng := NewManager(&NewManagerOpts{
				Cfg:    cfg,
				CfgKey: "test",
			})
			err := mng.UpdateConfig(&UpdateConfigOpts{
				Data: tc.data,
			})
			assert.NoError(t, err)
			got := mng.Config()

			assert.Equal(t, tc.expected, got)
			assert.Equal(t, tc.originCfg, cfg)
		})
	}
}

func TestUpdateConfigWithNodes(t *testing.T) {
	type NodesConfig struct {
		Name       string                 `cfg:"name"`
		Node       interface{}            `cfg:"node"`
		MapNode    map[string]interface{} `cfg:"map_node"`
		NestedNode interface{}            `cfg:"nested_node"`
	}
	type NodeConfig struct {
		Name string `cfg:"name"`
	}
	type NestedNodeConfig struct {
		Name string      `cfg:"name"`
		Node interface{} `cfg:"node"`
	}

	nodeCfg := NodeConfig{
		Name: "node",
	}
	node1OfMapNodeCfg := NodeConfig{
		Name: "node1_of_map_node",
	}
	node2OfMapNodeCfg := NodeConfig{
		Name: "node2_of_map_node",
	}
	mapNodeCfg := map[string]interface{}{
		"node1": node1OfMapNodeCfg,
		"node2": node2OfMapNodeCfg,
	}
	nodeOfNestedNodeCfg := NodeConfig{
		Name: "node_of_nested_node",
	}
	nestedNodeCfg := NestedNodeConfig{
		Name: "nested_node",
		Node: nodeOfNestedNodeCfg,
	}
	cfg := NodesConfig{
		Name:       "name",
		Node:       nodeCfg,
		MapNode:    mapNodeCfg,
		NestedNode: nestedNodeCfg,
	}

	mng := NewManager(&NewManagerOpts{
		Cfg:    cfg,
		CfgKey: "test",
	})
	nodeMng := NewManager(&NewManagerOpts{
		Cfg:    nodeCfg,
		CfgKey: "test",
	})
	node1OfMapNodeMng := NewManager(&NewManagerOpts{
		Cfg:    node1OfMapNodeCfg,
		CfgKey: "test",
	})
	node2OfMapNodeMng := NewManager(&NewManagerOpts{
		Cfg:    node2OfMapNodeCfg,
		CfgKey: "test",
	})
	nodeOfNestedNodeCfgMng := NewManager(&NewManagerOpts{
		Cfg:    nodeOfNestedNodeCfg,
		CfgKey: "test",
	})
	nestedNodeMng := NewManager(&NewManagerOpts{
		Cfg:    nestedNodeCfg,
		CfgKey: "test",
	})

	nestedNodeMng.RegisterNode("test.node", nodeOfNestedNodeCfgMng)
	mng.RegisterNode("test.node", nodeMng)
	mng.RegisterNode("test.map_node.node1", node1OfMapNodeMng)
	mng.RegisterNode("test.map_node.node2", node2OfMapNodeMng)
	mng.RegisterNode("test.nested_node", nestedNodeMng)

	data := map[string]interface{}{
		"test.name":                "newname",
		"test.node.name":           "newnode",
		"test.map_node.node1.name": "newnode1",
		"test.map_node.node2": NodeConfig{
			Name: "newnode2",
		},
		"test.nested_node.name":      "newnestednode",
		"test.nested_node.node.name": "newnodeofnestednode",
	}

	err := mng.UpdateConfig(&UpdateConfigOpts{
		Data: data,
	})

	// Assert no error
	assert.NoError(t, err)

	// Assert that all changes are applied
	assert.Equal(t, "newname", mng.Config().(NodesConfig).Name)
	assert.Equal(t, "newnode", mng.Config().(NodesConfig).Node.(NodeConfig).Name)
	assert.Equal(t, "newnode1", mng.Config().(NodesConfig).MapNode["node1"].(NodeConfig).Name)
	assert.Equal(t, "newnode2", mng.Config().(NodesConfig).MapNode["node2"].(NodeConfig).Name)
	assert.Equal(t, "newnestednode", mng.Config().(NodesConfig).NestedNode.(NestedNodeConfig).Name)
	assert.Equal(t, "newnodeofnestednode", mng.Config().(NodesConfig).NestedNode.(NestedNodeConfig).Node.(NodeConfig).Name)

	// Assert that all original configs are not changed
	assert.Equal(t, "name", cfg.Name)
	assert.Equal(t, "node", cfg.Node.(NodeConfig).Name)
	assert.Equal(t, "node1_of_map_node", cfg.MapNode["node1"].(NodeConfig).Name)
	assert.Equal(t, "node2_of_map_node", cfg.MapNode["node2"].(NodeConfig).Name)
	assert.Equal(t, "nested_node", cfg.NestedNode.(NestedNodeConfig).Name)
	assert.Equal(t, "node_of_nested_node", cfg.NestedNode.(NestedNodeConfig).Node.(NodeConfig).Name)
}

func cloneConfig[T Config](cfg T) T {
	var cloneCfg T

	cfgBytes, err := json.Marshal(cfg)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(cfgBytes, &cloneCfg)
	if err != nil {
		panic(err)
	}

	return cloneCfg
}

func ptr[T any](t T) *T {
	return &t
}
