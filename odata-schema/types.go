package odataschema

import "encoding/xml"

type PropertyRef struct {
	XMLName xml.Name `xml:"PropertyRef"`
	Name    string   `xml:"Name,attr"`
}

type TypeRef struct {
	XMLName xml.Name `xml:"TypeRef"`
	Type    string   `xml:"Type,attr"`
}

type Property struct {
	XMLName  xml.Name `xml:"Property"`
	Name     string   `xml:"Name,attr"`
	Type     string   `xml:"Type,attr"`
	TypeRef  *TypeRef `xml:"TypeRef"`
	Nullable bool     `xml:"Nullable,attr"`
	// DefaultValue string   `xml:"DefaultValue,attr"`
	// MaxLength    string   `xml:"MaxLength,attr"`
	// FixedLength  string   `xml:"FixedLength,attr"`
	// Precision    string   `xml:"Precision,attr"`
	// Scale        string   `xml:"Scale,attr"`
	// Unicode      string   `xml:"Unicode,attr"`
	// Collation    string   `xml:"Collation,attr"`
	// SRID         string   `xml:"SRID,attr"`
}

type NavigationProperty struct {
	XMLName  xml.Name `xml:"NavigationProperty"`
	Name     string   `xml:"Name,attr"`
	Type     string   `xml:"Type,attr"`
	Nullable *bool    `xml:"Nullable,attr"`
	// ToRole   string   `xml:"ToRole,attr"`
}

type NavigationPropertyBinding struct {
	XMLName xml.Name `xml:"NavigationPropertyBinding"`
	Path    string   `xml:"Path,attr"`
	Target  string   `xml:"Target,attr"`
}

type EnumTypeMember struct {
	XMLName xml.Name `xml:"Member"`
	Name    string   `xml:"Name,attr"`
	Value   string   `xml:"Value,attr"`
}

type EnumType struct {
	XMLName        xml.Name         `xml:"EnumType"`
	Name           string           `xml:"Name,attr"`
	UnderlyingType string           `xml:"UnderlyingType,attr"`
	IsFlags        bool             `xml:"IsFlags,attr"`
	Members        []EnumTypeMember `xml:"Member"`
}

type ComplexType struct {
	XMLName              xml.Name             `xml:"ComplexType"`
	Name                 string               `xml:"Name,attr"`
	BaseType             *string              `xml:"BaseType"`
	Abstract             bool                 `xml:"Abstract,attr"`
	Properties           []Property           `xml:"Property"`
	NavigationProperties []NavigationProperty `xml:"NavigationProperty"`
}

type EntityType struct {
	XMLName              xml.Name             `xml:"EntityType"`
	Name                 string               `xml:"Name,attr"`
	HasStream            bool                 `xml:"HasStream,attr"`
	BaseType             *string              `xml:"BaseType,attr"`
	Abstract             bool                 `xml:"Abstract,attr"`
	OpenType             bool                 `xml:"OpenType"`
	Key                  *[]PropertyRef       `xml:">PropertyRef"`
	Properties           []Property           `xml:"Property"`
	NavigationProperties []NavigationProperty `xml:"NavigationProperty"`
}

type EntitySet struct {
	XMLName                    xml.Name                    `xml:"EntitySet"`
	Name                       string                      `xml:"Name,attr"`
	EntityType                 string                      `xml:"EntityType,attr"`
	NavigationPropertyBindings []NavigationPropertyBinding `xml:"NavigationPropertyBinding"`
}

type EntityContainer struct {
	Name            string           `xml:"Name,attr"`
	EntitySets      []EntitySet      `xml:"EntitySet"`
	FunctionImports []FunctionImport `xml:"FunctionImport"`
	ActionImports   []ActionImport   `xml:"ActionImport"`
}

type Schema struct {
	XMLName         xml.Name         `xml:"Schema"`
	Namespace       string           `xml:"Namespace,attr"`
	Alias           *string          `xml:"Alias,attr"`
	EntityContainer *EntityContainer `xml:"EntityContainer"`
	EntityTypes     []EntityType     `xml:"EntityType"`
	ComplexTypes    []ComplexType    `xml:"ComplexType"`
	EnumTypes       []EnumType       `xml:"EnumType"`
	Functions       []Function       `xml:"Function"`
	Actions         []Action         `xml:"Action"`
}

type DataServices struct {
	XMLName xml.Name `xml:"DataServices"`
	Schemas []Schema `xml:"Schema"`
}

type EdmxDocument struct {
	XMLName      xml.Name     `xml:"Edmx"`
	DataServices DataServices `xml:"DataServices"`
}

type ReturnType struct {
	XMLName  xml.Name `xml:"ReturnType"`
	Type     string   `xml:",attr"`
	Nullable bool     `xml:",attr"`
}

type Parameter struct {
	XMLName xml.Name `xml:"Parameter"`
	Name    string   `xml:",attr"`
	Type    string   `xml:",attr"`
}

type Function struct {
	XMLName       xml.Name    `xml:"Function"`
	Name          string      `xml:",attr"`
	IsBound       bool        `xml:",attr"`
	EntitySetPath *string     `xml:",attr"`
	IsComposable  bool        `xml:",attr"`
	Parameters    []Parameter `xml:"Parameter"`
	ReturnType    ReturnType  `xml:"ReturnType"`
}

type Action struct {
	XMLName       xml.Name    `xml:"Action"`
	Name          string      `xml:",attr"`
	EntitySetPath *string     `xml:",attr"`
	IsBound       bool        `xml:",attr"`
	Parameters    []Parameter `xml:"Parameter"`
	ReturnType    *ReturnType `xml:"ReturnType"`
}

type FunctionImport struct {
	XMLName   xml.Name `xml:"FunctionImport"`
	Name      string   `xml:",attr"`
	EntitySet string   `xml:",attr"`
	Function  string   `xml:",attr"`
}

type ActionImport struct {
	XMLName   xml.Name `xml:"ActionImport"`
	Name      string   `xml:",attr"`
	EntitySet string   `xml:",attr"`
	Action    string   `xml:",attr"`
}
