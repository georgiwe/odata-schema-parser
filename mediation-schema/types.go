package mediationschema

import (
	"encoding/json"
)

type Service struct {
	Name        string
	Type        string
	Collections map[string]Collection
	Types       map[string]Type
	Invocations map[string]Invocation
}

type Type struct {
	Kind       string
	EntityType *EntityType `json:",omitempty"`
	Structure  *Structure  `json:",omitempty"`
	Enum       *Enum       `json:",omitempty"`
}

type InvocationArgument struct {
	Name string
	Property
}

type Invocation struct {
	Name             string
	BindingType      string
	BoundTo          *string `json:",omitempty"`
	BoundDataPointer *string `json:",omitempty"`
	Arguments        []InvocationArgument
	Result           *Property
}

type EntityType struct {
	Key        []string
	Streamable bool    `json:",omitempty"`
	BaseType   *string `json:",omitempty"`
	Structure
}

type Collection struct {
	Name       string
	EntityType string
	Streamable bool `json:",omitempty"`
}

type Structure struct {
	Name       string
	OpenType   bool `json:",omitempty"`
	Properties map[string]Property
}

type Enum struct {
	Name        string
	ValuesType  string
	Multiselect bool `json:",omitempty"`
	Members     map[string]string
}

type Property struct {
	ValueType          string
	PropertyType       string
	RelationCollection *string `json:",omitempty"`
	IsCollection       bool    `json:",omitempty"`
	Required           bool    `json:",omitempty"`
}

type Backend struct {
	Name     string
	Services []Service
}

type entityTypeSerializer struct {
	Kind string
	EntityType
}

type structureSerializer struct {
	Kind string
	Structure
}

type enumSerializer struct {
	Kind string
	Enum
}

func (td Type) MarshalJSON() ([]byte, error) {
	switch td.Kind {
	case "EntityType":
		ser := entityTypeSerializer{
			Kind:       td.Kind,
			EntityType: *td.EntityType,
		}
		return json.Marshal(ser)
	case "Structure":
		ser := structureSerializer{
			Kind:      td.Kind,
			Structure: *td.Structure,
		}
		return json.Marshal(ser)
	case "Enum":
		ser := enumSerializer{
			Kind: td.Kind,
			Enum: *td.Enum,
		}
		return json.Marshal(ser)
	}
	panic("should not happen")
}

type kindness struct {
	Kind string
}

func (td *Type) UnmarshalJSON(b []byte) error {
	var k kindness

	if err := json.Unmarshal(b, &k); err != nil {
		return err
	}

	td.Kind = k.Kind

	switch td.Kind {
	default:
		panic("unknown kind - cannot deserialize")
	case "EntityType":
		{
			deser := EntityType{}
			if err := json.Unmarshal(b, &deser); err != nil {
				return err
			}
			td.EntityType = &deser
		}
	case "Structure":
		{
			deser := Structure{}
			if err := json.Unmarshal(b, &deser); err != nil {
				return err
			}
			td.Structure = &deser
		}
	case "Enum":
		{
			deser := Enum{}
			if err := json.Unmarshal(b, &deser); err != nil {
				return err
			}
			td.Enum = &deser
		}
	}

	return nil
}
