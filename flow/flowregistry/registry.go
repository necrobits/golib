package flowregistry

import (
	"encoding/json"
	"fmt"

	"github.com/necrobits/x/flow"
)

var (
	// The global registry
	globalRegistry = NewDataRegistry()
)

// The data registry contains the actual FlowData struct for each FlowType
type DataRegistry struct {
	registryData map[flow.FlowType]flow.FlowData
}

func NewDataRegistry() *DataRegistry {
	return &DataRegistry{
		registryData: make(map[flow.FlowType]flow.FlowData),
	}
}

func Global() *DataRegistry {
	return globalRegistry
}

func (r *DataRegistry) Register(flowType flow.FlowType, data flow.FlowData) {
	r.registryData[flowType] = data
}

func (r *DataRegistry) Get(flowType flow.FlowType) flow.FlowData {
	if data, ok := r.registryData[flowType]; ok {
		return data
	}
	return nil
}

func (r *DataRegistry) Unmarshal(flowType flow.FlowType, flowDataJson json.RawMessage) (flow.FlowData, error) {
	if data, ok := r.registryData[flowType]; ok {
		err := json.Unmarshal(flowDataJson, &data)
		return data, err
	}
	return nil, fmt.Errorf("flow type %s not registered", flowType)
}

func (r *DataRegistry) DecodeSnapshot(snapshot *flow.Snapshot) (*flow.Snapshot, error) {
	flowType := flow.FlowType(snapshot.Type)
	flowData, err := r.Unmarshal(flowType, snapshot.EncodedData)
	if err != nil {
		return snapshot, err
	}
	snapshot.Data = flowData
	return snapshot, nil
}
