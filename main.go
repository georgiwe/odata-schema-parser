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
	edm, _ := odataschema.Parse(fmt.Sprintf("./schemas/%s.xml", backendName))
	odataService, _ := mediationschema.Parse(edm)
	bytes, _ := json.MarshalIndent(odataService, "", "  ")
	return os.WriteFile(fmt.Sprintf("./schemas/%s-mediation-schema.json", backendName), bytes, 0644)
}

func generateMediationGqlSchema(backendName string) string {
	bytes, _ := os.ReadFile(fmt.Sprintf("./schemas/%s-mediation-schema.json", backendName))
	service := &mediationschema.Service{}

	json.Unmarshal(bytes, service)

	backend := mediationschema.Backend{
		Name: backendName,
		Services: []mediationschema.Service{
			*service,
		},
	}

	schema := gqlschema.Parse(&backend)

	os.WriteFile(fmt.Sprintf("./schemas/%s.gql", backendName), []byte(schema), 0644)

	return schema
}

func main() {
	schemaName := "sitefinity"
	createMediationSchema(schemaName)
	gqlSchema := generateMediationGqlSchema(schemaName)
	fmt.Println(gqlSchema)
}
