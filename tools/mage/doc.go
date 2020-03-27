package mage

/**
 * Panther is a scalable, powerful, cloud-native SIEM written in Golang/React.
 * Copyright (C) 2020 Panther Labs Inc
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/alecthomas/jsonschema"

	"github.com/panther-labs/panther/internal/log_analysis/log_processor/parsers"
	"github.com/panther-labs/panther/internal/log_analysis/log_processor/registry"
	"github.com/panther-labs/panther/tools/cfndoc"
	"github.com/panther-labs/panther/tools/cfngen/gluecf"
)

// Generate all documentation
func Doc() {
	opDocs()
	logDocs()
}

const (
	inventoryDocHeader = `
<!-- This document is generated by "mage doc:cfn". DO NOT EDIT! -->

# Panther Application Run Books

Refer to the 
[Cloud Security](https://docs.runpanther.io/policies/scanning#how-it-works)
and
[Log Analysis](https://docs.runpanther.io/log-analysis/log-processing#how-it-works)
architecture diagrams for context.

Resource names below refer to resources in the Cloud Formation templates in Panther.

Each resource describes its function and failure impacts.

`
)

// generate operational documentation from deployment CloudFormation
func opDocs() {
	logger.Infof("doc: generating operational documentation from cloudformation")
	outDir := filepath.Join("docs", "gitbook", "operations")
	if err := os.MkdirAll(outDir, 0755); err != nil {
		logger.Fatalf("failed to create directory %s: %v", outDir, err)
	}
	inventoryFileName := filepath.Join(outDir, "runbooks.md")

	inventoryFile, err := os.Create(inventoryFileName)
	if err != nil {
		logger.Fatalf("failed to create file %s: %v", inventoryFileName, err)
	}
	defer inventoryFile.Close()

	docs, err := cfndoc.ReadCfn(cfnFiles()...)
	if err != nil {
		logger.Fatalf("failed to generate operational documentation: %v", err)
	}

	var docsBuffer bytes.Buffer
	docsBuffer.WriteString(inventoryDocHeader)
	for _, doc := range docs {
		docsBuffer.WriteString(fmt.Sprintf("## %s\n%s\n\n", doc.Resource, doc.Documentation))
	}
	if _, err = inventoryFile.Write(docsBuffer.Bytes()); err != nil {
		logger.Fatalf("failed to write file %s: %v", inventoryFileName, err)
	}
}

const (
	parserReadmeHeader = `
<!-- This document is generated by "mage doc:logs". DO NOT EDIT! -->

# Supported logs

`
)

func logDocs() {
	logger.Infof("doc: generating documentation on supported logs")
	outDir := filepath.Join("docs", "gitbook", "log-analysis", "log-processing", "supported-logs")
	if err := os.MkdirAll(outDir, 0755); err != nil {
		logger.Fatalf("failed to create directory %s: %v", outDir, err)
	}
	readmeFileName := filepath.Join(outDir, "README.md")

	readmeFile, err := os.Create(readmeFileName)
	if err != nil {
		logger.Fatalf("failed to create file %s: %v", readmeFileName, err)
	}
	defer readmeFile.Close()

	// group the data by category
	tables := registry.AvailableTables()
	logCategories := make(map[string][]string) // category -> logTypes
	for _, table := range tables {
		logType := table.LogType()
		categoryType := strings.Split(logType, ".")
		if len(categoryType) != 2 {
			logger.Fatalf("unexpected logType format: %s", logType)
		}
		logCategories[categoryType[0]] = append(logCategories[categoryType[0]], logType)
	}
	var sortedCategories []string
	for category := range logCategories {
		sortedCategories = append(sortedCategories, category)
		sort.Strings(logCategories[category])
	}
	sort.Strings(sortedCategories)

	// write README.md with table of log descriptions
	var docsBuffer bytes.Buffer
	docsBuffer.WriteString(parserReadmeHeader)
	for _, category := range sortedCategories {
		logTypes := logCategories[category]
		docsBuffer.WriteString(fmt.Sprintf("## %s\n", category))

		// use html table to get needed control

		for _, logType := range logTypes {
			table := registry.AvailableParsers().LookupParser(logType).GlueTableMetadata

			description := html.EscapeString(table.Description())

			docsBuffer.WriteString(fmt.Sprintf("###%s\n%s\n", logType, description))

			// add schema as html table since markdown won't let you embed tables
			docsBuffer.WriteString(`<table>` + "\n")
			docsBuffer.WriteString("<tr><th align=center>Column</th><th align=center>Type</th><th align=center>Required</th><th align=center>Description</th></tr>\n") // nolint
			columns := gluecf.InferJSONColumns(table.EventStruct(), gluecf.GlueMappings...)
			for _, column := range columns {
				docsBuffer.WriteString(fmt.Sprintf("<tr><td valign=top>%s</td><td>%s</td><td valign=top>%s</td><td valign=top>%s</td></tr>\n",
					column.Name,
					formatType(column),
					strconv.FormatBool(column.Required),
					html.EscapeString(column.Comment)))
			}

			docsBuffer.WriteString("</table>\n\n")
		}
	}
	if _, err = readmeFile.Write(docsBuffer.Bytes()); err != nil {
		logger.Fatalf("failed to write file %s: %v", readmeFileName, err)
	}
}

func formatType(col gluecf.Column) string {
	complexTypes := []string{"array", "struct", "map"}
	complex := false
	for _, ct := range complexTypes {
		if strings.HasPrefix(col.Type, ct) {
			complex = true
			break
		}
	}

	// if NOT a complex type we just use the Glue type
	if !complex {
		return col.Type
	}

	// complex Glue types are hard to read, so use JSON schema
	colType := col.Field.Type
	switch colType.String() { // handle special Panther types that will not work with JSON schema
	case reflect.TypeOf(&parsers.PantherAnyString{}).String():
		colType = reflect.TypeOf([]string{}) // slice of strings
	}
	// deference pointers
	if colType.Kind() == reflect.Ptr {
		colType = colType.Elem()
	}
	if colType.Kind() != reflect.Struct { // we need to create a struct for the parser to work
		fields := []reflect.StructField{
			{
				Name: col.Field.Name,
				Type: colType,
				Tag:  col.Field.Tag,
			},
		}
		colType = reflect.StructOf(fields)
	}

	// we need to wrap because schema package needs Name() and PkgPath()
	colSchema := jsonschema.ReflectFromType(&docType{Type: colType, name: col.Name})
	const indent = "&nbsp;&nbsp;&nbsp;&nbsp;"
	var jsonBuffer bytes.Buffer
	for name, schemaType := range colSchema.Definitions {
		// NOTE: we cannot use jsoniter package because it does not support prefix
		props, err := json.MarshalIndent(schemaType.Properties, "<br>", indent)
		if err != nil {
			logger.Fatal(err)
		}
		if (string)(props) != "{}" { // skip empty and column name
			if name != "" && name != col.Name {
				jsonBuffer.WriteString(fmt.Sprintf(`"%s":`, name))
			}
			jsonBuffer.Write(props)
			jsonBuffer.WriteString("<br><br>")
		} else if name == "RFC3339" { // special case for our timestamps embedded in structs
			jsonBuffer.WriteString(fmt.Sprintf(`"%s":`, name))
			jsonBuffer.WriteString("{<br>" + indent + `"type": "timestamp"<br>}`)
			jsonBuffer.WriteString("<br><br>")
		}
	}
	return jsonBuffer.String()
}

type docType struct {
	reflect.Type
	name string
}

func (dt *docType) Name() string {
	return dt.name
}

func (dt *docType) PkgPath() string {
	return "nopath" // this can be any non-empty value for the json parser to work
}
