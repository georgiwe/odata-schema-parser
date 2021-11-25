package mediationschema

import (
	"errors"
	"strings"

	ods "github.com/kinvey/odata-schema/odata-schema"
	"github.com/kinvey/odata-schema/utils"
)

var duplicateSitefinityEnums = []string{
	"Telerik.Sitefinity.Forms.Model.ConditionOperator",
	"Telerik.Sitefinity.Forms.Model.FormRuleAction",
	"Telerik.Sitefinity.Web.Api.Strategies.Pages.PageType",
	"Telerik.Sitefinity.Pages.Model.PageTemplateFramework",
}

type edmObjects struct {
	entityTypes     map[string]*ods.EntityType
	complexTypes    map[string]*ods.ComplexType
	enumTypes       map[string]*ods.EnumType
	functions       map[string]*ods.Function
	functionImports map[string]*ods.FunctionImport
	actions         map[string]*ods.Action
	actionImports   map[string]*ods.ActionImport
	entityContainer *ods.EntityContainer
}

func addToEntityTypes(objects edmObjects, schema *ods.Schema, entityType ods.EntityType) error {
	namespacedName, aliasedName := formQualifiedName(schema, entityType.Name)
	if _, ok := objects.entityTypes[namespacedName]; ok {
		return ErrDuplicateDefinition.WithMessagef("duplicate entity type definition for entity type '%s'", namespacedName)
	}

	objects.entityTypes[namespacedName] = &entityType
	if aliasedName != "" {
		if _, ok := objects.entityTypes[aliasedName]; ok {
			return ErrDuplicateDefinition.WithMessagef("duplicate entity type definition for entity type alias '%s'", aliasedName)
		}

		objects.entityTypes[aliasedName] = &entityType
	}

	return nil
}

func addToComplexTypes(objects edmObjects, schema *ods.Schema, complexType ods.ComplexType) error {
	namespacedName, aliasedName := formQualifiedName(schema, complexType.Name)
	if _, ok := objects.complexTypes[namespacedName]; ok {
		return ErrDuplicateDefinition.WithMessagef("duplicate complex type definition for complex type '%s'", namespacedName)
	}

	objects.complexTypes[namespacedName] = &complexType
	if aliasedName != "" {
		if _, ok := objects.complexTypes[aliasedName]; ok {
			return ErrDuplicateDefinition.WithMessagef("duplicate complex type definition for complex type alias '%s'", aliasedName)
		}

		objects.complexTypes[aliasedName] = &complexType
	}

	return nil
}

func addToEnumTypes(objects edmObjects, schema *ods.Schema, enumtype ods.EnumType) error {
	namespacedName, aliasedName := formQualifiedName(schema, enumtype.Name)
	if _, ok := objects.enumTypes[namespacedName]; ok {
		return ErrDuplicateDefinition.WithMessagef("duplicate enum type definition for enum type '%s'", namespacedName)
	}

	objects.enumTypes[namespacedName] = &enumtype
	if aliasedName != "" {
		if _, ok := objects.enumTypes[aliasedName]; ok {
			return ErrDuplicateDefinition.WithMessagef("duplicate enum type definition for enum type alias '%s'", aliasedName)
		}

		objects.enumTypes[aliasedName] = &enumtype
	}

	return nil
}

func addToFunctions(objects edmObjects, schema *ods.Schema, function ods.Function) error {
	namespacedName, aliasedName := formQualifiedName(schema, function.Name)
	if _, ok := objects.functions[namespacedName]; !ok {
		objects.functions[namespacedName] = &function
	}

	if aliasedName != "" {
		if _, ok := objects.functions[aliasedName]; !ok {
			objects.functions[aliasedName] = &function
		}
	}

	return nil
}

func addToActions(objects edmObjects, schema *ods.Schema, action ods.Action) error {
	namespacedName, aliasedName := formQualifiedName(schema, action.Name)
	if _, ok := objects.actions[namespacedName]; !ok {
		objects.actions[namespacedName] = &action
	}

	if aliasedName != "" {
		if _, ok := objects.actions[aliasedName]; !ok {
			objects.actions[aliasedName] = &action
		}
	}

	return nil
}

func extractObjects(backendName string, edm *ods.EdmxDocument) (*edmObjects, error) {
	objects := edmObjects{
		entityTypes:     make(map[string]*ods.EntityType),
		complexTypes:    make(map[string]*ods.ComplexType),
		enumTypes:       make(map[string]*ods.EnumType),
		functions:       make(map[string]*ods.Function),
		actions:         make(map[string]*ods.Action),
		functionImports: make(map[string]*ods.FunctionImport),
		actionImports:   make(map[string]*ods.ActionImport),
		entityContainer: nil,
	}

	for _, schema := range edm.DataServices.Schemas {
		if schema.EntityContainer != nil {
			objects.entityContainer = schema.EntityContainer
		}
		for _, entityType := range schema.EntityTypes {
			if err := addToEntityTypes(objects, &schema, entityType); err != nil {
				return nil, err
			}
		}
		for _, complexType := range schema.ComplexTypes {
			if err := addToComplexTypes(objects, &schema, complexType); err != nil {
				return nil, err
			}
		}
		for _, enumType := range schema.EnumTypes {
			if err := addToEnumTypes(objects, &schema, enumType); err != nil {
				// TODO: Log that this happened
				if namespacedName, _ := formQualifiedName(&schema, enumType.Name); !(backendName == "sitefinity" && errors.Is(err, ErrDuplicateDefinition) && utils.SliceContainsString(duplicateSitefinityEnums, namespacedName)) {
					return nil, err
				}
			}
		}
		for _, function := range schema.Functions {
			if err := addToFunctions(objects, &schema, function); err != nil {
				// TODO: handle function overloads
				if !strings.HasPrefix(err.Error(), "duplicate function definition") {
					return nil, err
				}
			}
		}
		for _, action := range schema.Actions {
			if err := addToActions(objects, &schema, action); err != nil {
				// TODO: handle action overloads
				if !strings.HasPrefix(err.Error(), "duplicate action definition") {
					return nil, err
				}
			}
		}
	}

	for _, functionImport := range objects.entityContainer.FunctionImports {
		objects.functionImports[functionImport.Function] = &functionImport
	}

	for _, actionImport := range objects.entityContainer.ActionImports {
		objects.actionImports[actionImport.Action] = &actionImport
	}

	return &objects, nil
}
