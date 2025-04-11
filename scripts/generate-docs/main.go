// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package main

import (
	"fmt"
	"os"
	"reflect"
	"text/template"

	"github.com/dicedb/dice/config"
	"github.com/dicedb/dice/internal/cmd"
)

const DocsCommandsDirectory = "docs/src/content/docs/commands"
const ServerConfigDirectory = "docs/src/content/docs/server-configs"

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
	serverConifigMeta := make(map[string]string)
	for i := 0; i < typeOf.NumField(); i++ {
		field := typeOf.Field(i)
		serverConifigMeta[field.Name] = string(field.Tag)
	}

	docFile, err := os.Create(fmt.Sprintf("%s/%s.md", ServerConfigDirectory, "Configuration"))
	if err != nil {
		fmt.Printf("ERR: error creating file: %v\n", err)
	}
	defer docFile.Close()

	err = tmpl.Execute(docFile, serverConifigMeta)
	if err != nil {
		fmt.Printf("ERR: error executing template: %v\n", err)
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
