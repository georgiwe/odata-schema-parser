package main

import (
	"encoding/json"
	"fmt"
	"os"

	gqlschema "github.com/kinvey/odata-schema/gql-schema"
	mediationschema "github.com/kinvey/odata-schema/mediation-schema"
	odataschema "github.com/kinvey/odata-schema/odata-schema"
)

func createMediationSchema(backendName string) error {
	if edm, err := odataschema.Parse(fmt.Sprintf("./schemas/%s.xml", backendName)); err != nil {
		return err
	} else {
		if odataService, err := mediationschema.Parse(backendName, edm); err != nil {
			return err
		} else {
			bytes, _ := json.MarshalIndent(odataService, "", "  ")
			return os.WriteFile(fmt.Sprintf("./schemas/%s-mediation-schema.json", backendName), bytes, 0644)
		}
	}
}

func generateMediationGqlSchema(backendName string) string {
	bytes, _ := os.ReadFile(fmt.Sprintf("./schemas/%s-mediation-schema.json", backendName))
	service := &mediationschema.Service{}

	json.Unmarshal(bytes, service)

	schema := gqlschema.Generate(service)

	os.WriteFile(fmt.Sprintf("./schemas/%s.gql", backendName), []byte(schema), 0644)

	return schema
}

func main() {
	schemaName := "sitefinity"
	if err := createMediationSchema(schemaName); err != nil {
		fmt.Print(err)
	} else {
		gqlSchema := generateMediationGqlSchema(schemaName)
		fmt.Println(len(gqlSchema))
	}
}
