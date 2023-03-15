package azure_np

import "github.com/spotinst/spotinst-sdk-go/spotinst/util/jsonutil"

// NodePoolProperties region
type NodePoolProperties struct {
	MaxPodsPerNode     *int    `json:"maxPodsPerNode,omitempty"`
	EnableNodePublicIP *bool   `json:"enableNodePublicIP,omitempty"`
	OsDiskSizeGB       *int    `json:"osDiskSizeGB,omitempty"`
	OsDiskType         *string `json:"osDiskType,omitempty"`
	OsType             *string `json:"osType,omitempty"`

	forceSendFields []string
	nullFields      []string
}

func (o NodePoolProperties) MarshalJSON() ([]byte, error) {
	type noMethod NodePoolProperties
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *NodePoolProperties) SetMaxPodsPerNode(v *int) *NodePoolProperties {
	if o.MaxPodsPerNode = v; o.MaxPodsPerNode == nil {
		o.nullFields = append(o.nullFields, "MaxPodsPerNode")
	}
	return o
}

func (o *NodePoolProperties) SetEnableNodePublicIP(v *bool) *NodePoolProperties {
	if o.EnableNodePublicIP = v; o.EnableNodePublicIP == nil {
		o.nullFields = append(o.nullFields, "EnableNodePublicIP")
	}
	return o
}

func (o *NodePoolProperties) SetOsDiskSizeGB(v *int) *NodePoolProperties {
	if o.OsDiskSizeGB = v; o.OsDiskSizeGB == nil {
		o.nullFields = append(o.nullFields, "OsDiskSizeGB")
	}
	return o
}

func (o *NodePoolProperties) SetOsDiskType(v *string) *NodePoolProperties {
	if o.OsDiskType = v; o.OsDiskType == nil {
		o.nullFields = append(o.nullFields, "OsDiskType")
	}
	return o
}

func (o *NodePoolProperties) SetOsType(v *string) *NodePoolProperties {
	if o.OsType = v; o.OsType == nil {
		o.nullFields = append(o.nullFields, "OsType")
	}
	return o
}

// endregion

// NodeCountLimits region
type NodeCountLimits struct {
	MinCount *int `json:"minCount,omitempty"`
	MaxCount *int `json:"maxCount,omitempty"`

	forceSendFields []string
	nullFields      []string
}

func (o NodeCountLimits) MarshalJSON() ([]byte, error) {
	type noMethod NodeCountLimits
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *NodeCountLimits) SetMinCount(v *int) *NodeCountLimits {
	if o.MinCount = v; o.MinCount == nil {
		o.nullFields = append(o.nullFields, "MinCount")
	}
	return o
}

func (o *NodeCountLimits) SetMaxCount(v *int) *NodeCountLimits {
	if o.MaxCount = v; o.MaxCount == nil {
		o.nullFields = append(o.nullFields, "MaxCount")
	}
	return o
}

// endregion

// Strategy region
type Strategy struct {
	SpotPercentage *int  `json:"spotPercentage,omitempty"`
	FallbackToOD   *bool `json:"fallbackToOd,omitempty"`

	forceSendFields []string
	nullFields      []string
}

func (o Strategy) MarshalJSON() ([]byte, error) {
	type noMethod Strategy
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *Strategy) SetSpotPercentage(v *int) *Strategy {
	if o.SpotPercentage = v; o.SpotPercentage == nil {
		o.nullFields = append(o.nullFields, "SpotPercentage")
	}
	return o
}

func (o *Strategy) SetFallbackToOD(v *bool) *Strategy {
	if o.FallbackToOD = v; o.FallbackToOD == nil {
		o.nullFields = append(o.nullFields, "FallbackToOD")
	}
	return o
}

// endregion

// region Taint

type Taint struct {
	Key    *string `json:"key,omitempty"`
	Value  *string `json:"value,omitempty"`
	Effect *string `json:"effect,omitempty"`

	forceSendFields []string
	nullFields      []string
}

func (o Taint) MarshalJSON() ([]byte, error) {
	type noMethod Taint
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *Taint) SetKey(v *string) *Taint {
	if o.Key = v; o.Key == nil {
		o.nullFields = append(o.nullFields, "Key")
	}
	return o
}

func (o *Taint) SetValue(v *string) *Taint {
	if o.Value = v; o.Value == nil {
		o.nullFields = append(o.nullFields, "Value")
	}
	return o
}

func (o *Taint) SetEffect(v *string) *Taint {
	if o.Effect = v; o.Effect == nil {
		o.nullFields = append(o.nullFields, "Effect")
	}
	return o
}

// endregion
