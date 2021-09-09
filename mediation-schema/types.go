package mediationschema

type Service struct {
	Name        string
	Collections []Collection
	Endpoints   []Endpoint
	Structures  []Structure
	Enums       []Enum
	Invocations []Invocation
}

type Invocation struct {
	Arguments  map[string]Property
	ReturnType string
}

type Collection struct {
	Key []string
	Structure
	HasStream bool `json:",omitempty"`
}

type Structure struct {
	Name       string
	Properties map[string]EntityProperty
}

type Endpoint struct {
	Name       string
	Properties []EntityProperty
}

type Enum struct {
	Name       string
	ValuesType string
	IsFlags    bool
	Members    map[string]string // TODO: in MC-CSDL enum members are order-comparable, here they are not - do we need them to be
}

type Property struct {
	Type         string
	PropertyType string
	IsCollection bool `json:",omitempty"`
}

type EnumProperty struct {
	EntityProperty
	Value string
}

type EntityProperty struct {
	Property
	RelatedTo *string `json:",omitempty"`
}
