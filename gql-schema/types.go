package gqlschema

import (
	"fmt"
	"strings"
)

type Schema struct {
	Query    Definition
	Mutation Definition
	Types    []Definition
}

type Directive struct {
	Name   string
	Fields []Field
}

type Element struct {
	Name       string
	Directives *[]Directive
}

type Field struct {
	Type     string
	Required bool
	Element
}

// TODO: name this
type Function struct {
	Arguments  []Field
	ReturnType string
	Element
}

type Definition struct {
	Type      string
	Fields    *[]Field
	Functions *[]Function
	Elements  *[]Element
	Element
}

// TODO: Indentation
func (def *Definition) String() string {
	sb := &strings.Builder{}

	fmt.Fprintf(sb, "%s %s ", def.Type, def.Name)

	if def.Directives != nil {
		for _, dir := range *def.Directives {
			sb.WriteString(dir.String())
		}
		sb.WriteString(" ")
	}

	sb.WriteString("{\n")

	if def.Fields != nil {
		for _, field := range *def.Fields {
			fmt.Fprintf(sb, "  %s\n", field.String())
		}
	}

	if def.Elements != nil {
		for _, element := range *def.Elements {
			fmt.Fprintf(sb, "  %s\n", element.String())
		}
	}

	sb.WriteString("}")

	return sb.String()
}

func (directive *Directive) String() string {
	sb := &strings.Builder{}

	fmt.Fprintf(sb, "@%s(", directive.Name)

	for i, field := range directive.Fields {
		fmt.Fprintf(sb, "%s", field.String())
		if i < len(directive.Fields)-1 {
			sb.WriteString(", ")
		}
	}

	sb.WriteString(")")

	return sb.String()
}

func (field *Field) String() string {
	sb := &strings.Builder{}
	requiredFlag := ""
	if field.Required {
		requiredFlag = "!"
	}

	// fmt.Fprintf(sb, "%s: %s%s", field.Name, field.Type, requiredFlag)
	fieldStr := fmt.Sprintf("%s: %s%s", field.Name, field.Type, requiredFlag)
	elStr := field.Element.String()
	result := strings.Replace(elStr, field.Name, fieldStr, 1)
	sb.WriteString(result)

	// if field.Directives != nil {
	// 	for _, dir := range *field.Directives {
	// 		fmt.Fprintf(sb, " %s", dir.String())
	// 	}
	// }

	return sb.String()
}

func (element *Element) String() string {
	sb := &strings.Builder{}

	sb.WriteString(element.Name)

	if element.Directives != nil {
		for _, dir := range *element.Directives {
			fmt.Fprintf(sb, " %s", dir.String())
		}
	}

	return sb.String()
}

func (schema *Schema) String() string {
	sb := &strings.Builder{}

	for _, def := range schema.Types {
		fmt.Fprintf(sb, "%s\n\n", def.String())
	}

	return sb.String()
}
