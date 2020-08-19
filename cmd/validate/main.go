package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/xeipuuv/gojsonschema"
)

func main() {
	sp := schemaPath()
	schema := gojsonschema.NewReferenceLoader("file://" + schemaPath())
	fmt.Printf("📖 loading schema from %s\n", sp)

	cp := corpusRootPath()
	fmt.Printf("🏛  walking the corpus at %s\n", cp)

	err := filepath.Walk(cp, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !strings.HasSuffix(path, ".json") {
			return nil
		}
		relPath := strings.Replace(path, cp+"/", "", 1)
		doc := gojsonschema.NewReferenceLoader("file://" + path)
		result, err := gojsonschema.Validate(schema, doc)
		if err != nil {
			return fmt.Errorf("validating vector %s: %w", relPath, err)
		}
		if !result.Valid() {
			fmt.Printf("❌ %s\n", relPath)
			for _, desc := range result.Errors() {
				fmt.Printf("\t- %s\n", desc)
			}
			return fmt.Errorf("validating vector %s", relPath)
		}
		fmt.Printf("✅ %s\n", relPath)
		return nil
	})
	if err != nil {
		panic(fmt.Errorf("walking corpus: %w", err))
	}
}

func rootPath() string {
	_, filename, _, _ := runtime.Caller(0)
	return path.Dir(path.Dir(filename))
}

func schemaPath() string {
	return path.Join(rootPath(), "../schema.json")
}

func corpusRootPath() string {
	return path.Join(rootPath(), "../corpus")
}
