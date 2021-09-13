package mediationschema

type Service struct {
	Name        string
	Collections []Collection
	Structures  []Structure
	Enums       []Enum
	Invocations []Invocation
}

type Invocation struct {
	Name             string
	BindingType      string
	BoundTo          *string `json:",omitempty"`
	BoundDataPointer *string `json:",omitempty"`
	Arguments        []Property
	Result           *Property
}

type Collection struct {
	Key        []string
	EntityType string
	HasStream  bool `json:",omitempty"`
	Structure
}

type Structure struct {
	Name       string
	Properties map[string]Property
}

type Enum struct {
	Name       string
	ValuesType string
	IsFlags    bool              `json:",omitempty"`
	Members    map[string]string // TODO: in MC-CSDL enum members are order-comparable, here they are not - do we need them to be
}

type Property struct {
	Type               string
	PropertyType       string
	IsCollection       bool    `json:",omitempty"`
	RelationCollection *string `json:",omitempty"`
}

type EnumProperty struct {
	Property
	Value string
}
