package mediationschema

type Property struct {
	Type         string
	RelatedTo    *string `json:",omitempty"`
	StructureRef *string `json:",omitempty"`
}

type Structure struct {
	Name       string
	Properties map[string]Property
}

type Collection struct {
	Structure
	Key []string
}

type Endpoint struct {
	Name       string
	Properties []Property
}

type Service struct {
	Name        string
	Collections []Collection
	Endpoints   []Endpoint
	Structures  []Structure
}
