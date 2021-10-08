package gqlschema

import (
	"fmt"
	"strings"

	mschema "github.com/kinvey/odata-schema/mediation-schema"
)

const indentationSize = 4

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

// A field with no type?
type Directive struct {
	Name   string
	Fields []Field
}

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
	Type   string
	Fields *[]Field
	Element
}

type usedTypeDesc struct {
	QualifiedName string
	MedSchemaType mschema.Type
	GqlField      Definition
}

func newBackendDirective(product string, collection string, method string, endpoint string) Directive {
	fields := []Field{
		{
			Type:    product,
			Element: Element{Name: "product"},
		},
		{
			Type:    collection,
			Element: Element{Name: "collection"},
		},
	}

	if method != "" {
		fields = append(fields, Field{Type: method, Element: Element{Name: "method"}})
	}

	if endpoint != "" {
		fields = append(fields, Field{Type: endpoint, Element: Element{Name: "endpoint"}})
	}

	return Directive{
		Name:   "backend",
		Fields: fields,
	}
}

func stringifyFields(fields []Field, joiner string, indentLevels int) string {
	sb := &strings.Builder{}
	count := len(fields)

	for i, field := range fields {
		fmt.Fprintf(sb, "%s%s", strings.Repeat(" ", indentLevels*indentationSize), field.String(false, true))
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
			sb.WriteString(dir.String(false))
		}
		sb.WriteString(" ")
	}

	sb.WriteString("{\n")

	if def.Fields != nil {
		sb.WriteString(stringifyFields(*def.Fields, "\n", 1))
		sb.WriteString("\n")
	}

	sb.WriteString("}")

	return sb.String()
}

// TODO: quote values :(
func (directive *Directive) String(quoteValues bool) string {
	sb := &strings.Builder{}

	fmt.Fprintf(sb, "@%s(", directive.Name)

	for i, field := range directive.Fields {
		fmt.Fprintf(sb, "%s", field.String(quoteValues, false))
		if i < len(directive.Fields)-1 {
			sb.WriteString(", ")
		}
	}

	sb.WriteString(")")

	return sb.String()
}

// TODO: Refactor printWhenNoValue :(
func (field *Field) String(quoteValue bool, printWhenNoValue bool) string {
	requiredFlag := ""
	// TODO: This generates a wrong schema
	if field.Required {
		requiredFlag = "!"
	}

	fieldType := field.Type
	if fieldType != "" {
		if quoteValue || strings.Contains(fieldType, "-") || strings.Contains(fieldType, " ") {
			fieldType = fmt.Sprintf(`"%s"`, fieldType)
		}
		fieldType = fmt.Sprintf(": %s", fieldType)
	} else if !printWhenNoValue {
		return ""
	}

	argumentsStr := ""

	if field.Arguments != nil {
		argumentsStr = fmt.Sprintf("(%s)", stringifyFields(*field.Arguments, ", ", 0))
	}

	fieldStr := fmt.Sprintf("%s%s%s%s", field.Name, argumentsStr, fieldType, requiredFlag)
	elStr := field.Element.String()
	result := strings.Replace(elStr, field.Name, fieldStr, 1)

	return result
}

func (element *Element) String() string {
	sb := &strings.Builder{}

	sb.WriteString(element.Name)

	if element.Directives != nil {
		for _, dir := range *element.Directives {
			fmt.Fprintf(sb, " %s", dir.String(true))
		}
	}

	return sb.String()
}

func (dec *DirectiveDeclaration) String() string {
	sb := &strings.Builder{}

	applications := strings.Join(dec.Applications, " | ")
	fmt.Fprintf(sb, "directive %s on %s", dec.Directive.String(false), applications)

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
