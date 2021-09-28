package main

import (
	"encoding/json"
	"os"

	mediationschema "github.com/kinvey/odata-schema/mediation-schema"
	odataschema "github.com/kinvey/odata-schema/odata-schema"
)

func main() {
	edm, _ := odataschema.Parse("./schemas/sitefinity.xml")
	odataService, _ := mediationschema.Parse(edm)
	bytes, _ := json.MarshalIndent(odataService, "", "  ")
	// fmt.Println(string(bytes))
	os.WriteFile("./out.json", bytes, 0644)
}
