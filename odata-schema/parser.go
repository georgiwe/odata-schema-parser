package odataschema

import (
	"encoding/xml"
	"io/ioutil"
	"os"
)

func Parse(filePath string) (*EdmxDocument, error) {
	xmlFile, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	defer xmlFile.Close()

	bytes, err := ioutil.ReadAll(xmlFile)

	if err != nil {
		return nil, err
	}

	var edm EdmxDocument
	xml.Unmarshal(bytes, &edm)

	return &edm, nil
}
