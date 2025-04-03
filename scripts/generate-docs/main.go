// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package main

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"text/template"

	"github.com/dicedb/dice/config"
	"github.com/dicedb/dice/internal/cmd"
)

const DocsCommandsDirectory = "docs/src/content/docs/commands"
const ServerConfigDirectory = "docs/src/content/docs/server-configs"

type ServerConfigMeta struct {
	Order 	    int
	Name        string
	Description string
	CLICommand  string
	Default     string
	Type        string
	Values	    []string
}

func mapType(goType string) string {
	switch goType {
	case "int", "int8", "int16", "int32", "int64":
		return "Integer"
	case "uint", "uint8", "uint16", "uint32", "uint64":
		return "Unsigned Integer"
	case "float32", "float64":
		return "Float"
	case "bool":
		return "Boolean"
	case "string":
		return "String"
	default:
		return goType
	}
}

func generateDocs(tmpl *template.Template, c *cmd.CommandMeta) {
	docFile, err := os.Create(fmt.Sprintf("%s/%s.md", DocsCommandsDirectory, c.Name))
	if err != nil {
		fmt.Printf("ERR: error creating file: %v\n", err)
	}
	defer docFile.Close()

	err = tmpl.Execute(docFile, c)
	if err != nil {
		fmt.Printf("ERR: error executing template: %v\n", err)
	}
}

func generateServerParamtersDocs(tmpl *template.Template) {
	typeOf := reflect.TypeOf(&config.DiceDBConfig{}).Elem()
	for i := 0; i < typeOf.NumField(); i++ {
		field := typeOf.Field(i)
		description := field.Tag.Get("description")
		mapstructure := field.Tag.Get("mapstructure")
		defaultValue := field.Tag.Get("default")
		values := strings.Split(field.Tag.Get("values"), ", ")
		if len(values) == 1 && values[0] == "" {
			values = []string{}
		}
		
		docFile, err := os.Create(fmt.Sprintf("%s/%s.md", ServerConfigDirectory, field.Name))
		if err != nil {
			fmt.Printf("ERR: error creating file: %v\n", err)
		}
		defer docFile.Close()

		err = tmpl.Execute(docFile, ServerConfigMeta{i, field.Name, description, "--"+mapstructure, defaultValue, mapType(field.Type.String()), values})
		if err != nil {
			fmt.Printf("ERR: error executing template: %v\n", err)
		}
	}
}



func main() {
	tmpl := template.Must(template.ParseFiles("scripts/generate-docs/doc.tmpl"))
	for _, c := range cmd.CommandRegistry.CommandMetas {
		if c.HelpLong == "" {
			continue
		}

		generateDocs(tmpl, c)
	}
	sctmpl := template.Must(template.ParseFiles("scripts/generate-docs/server-config.tmpl"))
	generateServerParamtersDocs(sctmpl)
}
