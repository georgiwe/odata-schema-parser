package gqlschema

import (
	"fmt"
	"strings"
)

type DirectiveDeclaration struct {
	Applications []string
	Directive
}

type Schema struct {
	Query                 Definition
	Mutation              Definition
	Types                 []Definition
	DirectiveDeclarations []DirectiveDeclaration
}

type Directive struct {
	Name   string
	Fields []Field
}

// TODO: Merge with Field somehow
type Element struct {
	Name       string
	Directives *[]Directive
}

type Field struct {
	Type      string
	Required  bool
	Arguments *[]Field
	Element
}

type Definition struct {
	Type     string
	Fields   *[]Field
	Elements *[]Element
	Element
}

func stringifyFields(fields []Field, joiner string, indentation int) string {
	sb := &strings.Builder{}
	count := len(fields)

	for i, field := range fields {
		fmt.Fprintf(sb, "%s%s", strings.Repeat(" ", indentation), field.String())
		if i < count-1 {
			sb.WriteString(joiner)
		}
	}

	return sb.String()
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
		sb.WriteString(stringifyFields(*def.Fields, "\n", 2))
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

	fieldType := field.Type
	if strings.Contains(fieldType, " ") {
		fieldType = fmt.Sprintf(`"%s"`, fieldType)
	}

	argumentsStr := ""

	if field.Arguments != nil {
		argumentsStr = fmt.Sprintf("(%s)", stringifyFields(*field.Arguments, ", ", 0))
	}

	fieldStr := fmt.Sprintf("%s%s: %s%s", field.Name, argumentsStr, fieldType, requiredFlag)
	elStr := field.Element.String()
	result := strings.Replace(elStr, field.Name, fieldStr, 1)

	sb.WriteString(result)

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

func (dec *DirectiveDeclaration) String() string {
	sb := &strings.Builder{}

	applications := strings.Join(dec.Applications, " | ")
	fmt.Fprintf(sb, "directive %s on %s", dec.Directive.String(), applications)

	return sb.String()
}

func (schema *Schema) String() string {
	sb := &strings.Builder{}

	for _, declalration := range schema.DirectiveDeclarations {
		sb.WriteString(declalration.String())
		sb.WriteString("\n")
	}

	if len(schema.DirectiveDeclarations) > 0 {
		sb.WriteString("\n")
	}

	sb.WriteString(schema.Query.String())
	sb.WriteString("\n\n")

	sb.WriteString(schema.Mutation.String())
	sb.WriteString("\n\n")

	typesCount := len(schema.Types)
	for i, def := range schema.Types {
		sb.WriteString(def.String())
		sb.WriteString("\n")

		if i < typesCount-1 {
			sb.WriteString("\n")
		}
	}

	return sb.String()
}
