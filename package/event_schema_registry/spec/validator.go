package spec

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type validator struct {
}

func NewValidator() *validator {
	return &validator{}
}

func (v *validator) Validate(bytes []byte, domain, eventName string, version string) error {
	elems := []string{
		"package",
		"event_schema_registry",
		"spec",
		"schemas",
		domain,
		strings.ReplaceAll(eventName, ".", "_"),
		fmt.Sprintf("%s.json", version),
	}

	schemaPath := filepath.Join(elems...)
	schemaFile, err := os.Open(schemaPath)
	if err != nil {
		return err
	}

	defer schemaFile.Close()

	// no errors
	return nil
}
