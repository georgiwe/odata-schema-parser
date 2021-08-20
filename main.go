package main

import (
	"encoding/json"
	"fmt"

	mediationschema "github.com/kinvey/odata-schema/mediation-schema"
	odataschema "github.com/kinvey/odata-schema/odata-schema"
)

func main() {
	edm, _ := odataschema.Parse("./schema.xml")
	schema, _ := mediationschema.Parse(edm)
	bytes, _ := json.MarshalIndent(schema, "", "  ")
	fmt.Println(string(bytes))
	// fmt.Println(schema.Endpoints)
}
