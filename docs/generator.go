// Copyright 2017 Monax Industries Limited
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"text/template"

	commands "github.com/eris-ltd/eris-db/cmd"
	docs "github.com/eris-ltd/eris-db/docs/generator"

	clientCommands "github.com/eris-ltd/eris-db/client/cmd"
	"github.com/eris-ltd/eris-db/definitions"
	"github.com/eris-ltd/eris-db/version"
	"github.com/spf13/cobra"
)

// Repository maintainers should customize the next two lines.
var Description = "Blockchain Client"                                       // should match the docs site name
var RenderDir = fmt.Sprintf("./docs/documentation/db/%s/", version.VERSION) // should be the "shortversion..."

// The below variables should be updated only if necessary.
var Specs = []*docs.Entry{}
var Examples = []*docs.Entry{}
var SpecsDir = "./docs/specs"
var ExamplesDir = "./docs/examples"

type Cmd struct {
	Command     *cobra.Command
	Entry       *docs.Entry
	Description string
}

func RenderFiles(cmdRaw *cobra.Command, tmpl *template.Template) error {
	this_entry := &docs.Entry{
		Title:          cmdRaw.CommandPath(),
		Specifications: Specs,
		Examples:       Examples,
		BaseURL:        strings.Replace(RenderDir, ".", "", 1),
		Template:       tmpl,
		FileName:       docs.GenerateFileName(RenderDir, cmdRaw.CommandPath()),
	}

	cmd := &Cmd{
		Command:     cmdRaw,
		Entry:       this_entry,
		Description: Description,
	}

	for _, command := range cmd.Command.Commands() {
		RenderFiles(command, tmpl)
	}

	if !cmd.Command.HasParent() {
		entries := append(cmd.Entry.Specifications, cmd.Entry.Examples...)
		for _, entry := range entries {
			entry.Specifications = cmd.Entry.Specifications
			entry.Examples = cmd.Entry.Examples
			entry.CmdEntryPoint = cmd.Entry.Title
			entry.BaseURL = cmd.Entry.BaseURL
			if err := docs.RenderEntry(entry); err != nil {
				return err
			}
		}
	}

	outFile, err := os.Create(cmd.Entry.FileName)
	if err != nil {
		return err
	}
	defer outFile.Close()

	err = cmd.Entry.Template.Execute(outFile, cmd)
	if err != nil {
		return err
	}

	return nil
}

func AddClientToDB(dbCmd, clientCmd *cobra.Command) error {
	// formulate the filenames properly
	dbFile := docs.GenerateFileName(RenderDir, dbCmd.CommandPath())
	clFile := docs.GenerateFileName(RenderDir, clientCmd.CommandPath())

	// get the manual additions sorted
	dbAdditions := []byte(fmt.Sprintf("\n# Related Commands\n\n* [%s](%s)", "Eris Client", docs.GenerateURLFromFileName(clFile)))
	clAdditions := []byte(fmt.Sprintf("\n# Related Commands\n\n* [%s](%s)", "Eris DB", docs.GenerateURLFromFileName(dbFile)))

	// read and write the db file
	dbTxt, err := ioutil.ReadFile(dbFile)
	if err != nil {
		return err
	}
	dbTxt = append(dbTxt, dbAdditions...)
	err = ioutil.WriteFile(dbFile, dbTxt, 0644)
	if err != nil {
		return err
	}

	// read and write the client file
	clTxt, err := ioutil.ReadFile(clFile)
	if err != nil {
		return err
	}
	clTxt = append(clTxt, clAdditions...)
	err = ioutil.WriteFile(clFile, clTxt, 0644)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	// Repository maintainers should populate the top level command object.
	erisDbCommand := commands.ErisDbCmd
	do := definitions.NewDo()
	commands.AddGlobalFlags(do)
	commands.AddCommands(do)

	erisClientCommand := clientCommands.ErisClientCmd
	clientCommands.InitErisClientInit()
	clientCommands.AddClientCommands()

	// Make the proper directory.
	var err error
	if _, err = os.Stat(RenderDir); os.IsNotExist(err) {
		err = os.MkdirAll(RenderDir, 0755)
		if err != nil {
			panic(err)
		}
	}

	// Generate specs and examples files.
	Specs, err = docs.GenerateEntries(SpecsDir, (RenderDir + "specifications/"), Description)
	if err != nil {
		panic(err)
	}
	Examples, err = docs.GenerateEntries(ExamplesDir, (RenderDir + "examples/"), Description)
	if err != nil {
		panic(err)
	}

	// Get template from docs generator.
	tmpl, err := docs.GenerateCommandsTemplate()
	if err != nil {
		panic(err)
	}

	// Render the templates.
	if err = RenderFiles(erisDbCommand, tmpl); err != nil {
		panic(err)
	}

	if err = RenderFiles(erisClientCommand, tmpl); err != nil {
		panic(err)
	}

	if err = AddClientToDB(erisDbCommand, erisClientCommand); err != nil {
		panic(err)
	}
}
