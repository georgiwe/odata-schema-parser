package mediationschema

type Service struct {
	Name        string
	Type        string
	Collections map[string]Collection
	EntityTypes map[string]EntityType
	Structures  map[string]Structure
	Enums       map[string]Enum
	Invocations map[string]Invocation
}

type Invocation struct {
	Name             string
	BindingType      string
	BoundTo          *string `json:",omitempty"`
	BoundDataPointer *string `json:",omitempty"`
	Arguments        []Property
	Result           *Property
}

type EntityType struct {
	Key       []string
	HasStream bool `json:",omitempty"`
	Structure
}

type Collection struct {
	Name          string
	EntityType    string
	Relationships map[string]string `json:",omitempty"`
}

type Structure struct {
	Name       string
	Properties map[string]Property
}

type Enum struct {
	Name        string
	ValuesType  string
	Multiselect bool              `json:",omitempty"`
	Members     map[string]string // TODO: in MC-CSDL enum members are order-comparable, here they are not - do we need them to be
}

type Property struct {
	ValueType    string
	PropertyType string
	IsCollection bool `json:",omitempty"`
	// RelationCollection *string `json:",omitempty"`
}
