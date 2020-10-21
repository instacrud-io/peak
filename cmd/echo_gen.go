/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/heraju/mestri"
	"github.com/instacrud-io/peak/templates"
	"github.com/spf13/cobra"
)

// EchoCmd represents the init command
var EchoCmd = &cobra.Command{
	Use:   "echo",
	Short: "ECHO :: => Creating your ECHO APP",
	Long:  `ECHO ==> Creating your ECHO APP`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Creating your APP!!")
		echoinit(cmd)
	},
}

func echoinit(cmd *cobra.Command) error {
	folder, _ := cmd.Flags().GetString("app")

	if folder == "" {
		return errors.New("Your need to inform a path to initialize the git repository")
	}

	command := exec.Command("git", "init", folder)
	err := command.Run()
	if err != nil {
		fmt.Printf("Could not create the directory %v", err)
	}
	fmt.Println("Git repository initialized in " + folder)
	createEchoApp(folder)

	command = exec.Command("cd", folder)
	err = command.Run()
	if err != nil {
		fmt.Printf("Could not move to directory %v", err)
	}
	command = exec.Command("go", "mod", "init")
	err = command.Run()
	if err != nil {
		fmt.Printf("GO MOD failed %v", err)
	}
	command = exec.Command("go", "run", "cmd/start.go")
	err = command.Run()
	if err != nil {
		fmt.Printf("GO run start %v", err)
	}
	return nil
}

func createEchoApp(app string) {
	db, err := sql.Open("postgres", mestri.PsqlInfo)
	die(err)
	defer db.Close()
	err = db.Ping()
	die(err)
	rows, err := db.Query("SELECT table_name FROM information_schema.tables WHERE table_schema = 'public'")
	die(err)
	entities := make([]string, 0)
	models := make([]Model, 0)
	echoCleanApp(app)
	for rows.Next() {
		var table_name string
		err := rows.Scan(&table_name)
		die(err)
		pass := echoBuildEntity(table_name, db, app)
		fmt.Print(table_name)
		if pass {
			entities = append(entities, table_name)
			models = append(models, Model{Package: table_name, ModelName: toCamelCase(table_name)})
		}
	}
	echoBuildApp(entities, models, app)
	buildStart(app)
	buildHelpers(app)
	fmt.Print(entities)
}

func echoCleanApp(app string) {
	_, err := os.Stat(app)
	die(err)
	serviceFolder := app + "/services"
	err = os.MkdirAll(serviceFolder, 0755)
	die(err)
	cmdFolder := app + "/cmd"
	err = os.MkdirAll(cmdFolder, 0755)
	die(err)
	helperFolder := app + "/helpers"
	err = os.MkdirAll(helperFolder, 0755)
	die(err)
}

func echoBuildApp(entities []string, models []Model, app string) bool {
	dirPath := app + "/services/"
	fileName := dirPath + "/services.go"
	f, err := os.Create(fileName)
	die(err)
	defer f.Close()
	appTemplate, err := template.New("entity").Parse(templates.EchoServiceTmpl)
	die(err)
	appTemplate.Execute(f, struct {
		Timestamp time.Time
		Entities  []string
		Models    []Model
	}{
		Timestamp: time.Now(),
		Entities:  entities,
		Models:    models,
	})
	return true
}

func buildStart(app string) bool {
	dirPath := app + "/cmd/"
	fileName := dirPath + "/start.go"
	f, err := os.Create(fileName)
	die(err)
	defer f.Close()
	appTemplate, err := template.New("entity").Parse(templates.EchoStartTmpl)
	die(err)
	appTemplate.Execute(f, struct{}{})
	return true
}

func buildHelpers(app string) bool {
	dirPath := app + "/helpers/"
	fileName := dirPath + "/merrors.go"
	f, err := os.Create(fileName)
	die(err)
	defer f.Close()
	appTemplate, err := template.New("entity").Parse(templates.EchoHelperTmpl)
	die(err)
	appTemplate.Execute(f, struct{}{})
	return true
}

func echoBuildEntity(table_name string, db *sql.DB, app string) bool {
	idName := getIdName(table_name, db)
	if idName == "" {
		fmt.Println("No Primary for", table_name)
		return false
	}

	dirPath := app + "/services/" + table_name
	err := os.Mkdir(dirPath, 0755)
	die(err)
	fmt.Println("Building CRUD For --- : ", table_name)
	fileName := dirPath + "/entity.go"
	f, err := os.Create(fileName)
	die(err)
	defer f.Close()

	entityTemplate, err := template.New("entity").Parse(templates.EchoEntityTmpl)
	die(err)
	attr, err := db.Query("select column_name, data_type, column_default from information_schema.columns where table_name = $1 and table_schema = 'public' order by column_name", table_name)

	entity := make(map[string]Attr)
	var id_data_type string
	var is_string_key bool
	var is_id_auto bool

	for attr.Next() {
		var column_name string
		var data_type string
		var data_type_map string
		var key_type string
		var column_default sql.NullString

		attr.Scan(&column_name, &data_type, &column_default)
		switch data_type {
		case "uuid":
			data_type_map = "NullString"
			key_type = "string"
		case "text", "timestamp without time zone", "numeric", "character varying":
			data_type_map = "NullString"
			key_type = "string"
		case "integer", "bigint":
			data_type_map = "NullInt64"
			key_type = "int64"
		case "timestamp with time zone":
			data_type_map = "NullString"
			key_type = "string"
		case "bit":
			data_type_map = "NullBool"
			key_type = "bool"
		case "json":
			data_type_map = "NullString"
			key_type = "string"
		}
		if column_name == idName {
			id_data_type = key_type
			if key_type == "string" {
				is_string_key = true
			}
			if column_default.String != "" {
				is_id_auto = true
			}
		}
		entity[column_name] = Attr{ReaderType: toCamelCase(key_type), DataType: data_type_map, ModelName: toCamelCase(column_name), KeyType: key_type}
	}

	entityTemplate.Execute(f, struct {
		Timestamp time.Time
		Model     string
		Entity    map[string]Attr
		IdType    string
		IdName    string
	}{
		Timestamp: time.Now(),
		Model:     table_name,
		Entity:    entity,
		IdType:    id_data_type,
		IdName:    idName,
	})
	buildUsecase(table_name, id_data_type, app)
	echoBuildPgRepo(table_name, entity, id_data_type, idName, is_id_auto, app)
	buildHandler(table_name, id_data_type, is_string_key, app)
	return true
}

func buildHandler(table_name string, id_data_type string, is_string_key bool, app string) bool {
	dirPath := app + "/services/" + table_name
	fileName := dirPath + "/handler.go"
	f, err := os.Create(fileName)
	die(err)
	defer f.Close()
	handlerTemplate, err := template.New("entity").Parse(templates.EchoHandlerTmpl)
	die(err)
	handlerTemplate.Execute(f, struct {
		Timestamp  time.Time
		Entity     string
		ModelName  string
		IdType     string
		IsStingKey bool
	}{
		Timestamp:  time.Now(),
		Entity:     table_name,
		ModelName:  toCamelCase(table_name),
		IdType:     id_data_type,
		IsStingKey: is_string_key,
	})
	return true
}

func buildUsecase(table_name string, id_data_type string, app string) bool {
	dirPath := app + "/services/" + table_name
	fileName := dirPath + "/usecase.go"
	f, err := os.Create(fileName)
	die(err)
	defer f.Close()
	useCaseTemplate, err := template.New("entity").Parse(templates.EchoUsecaseTmpl)
	die(err)
	useCaseTemplate.Execute(f, struct {
		Timestamp time.Time
		Entity    string
		ModelName string
		IdType    string
	}{
		Timestamp: time.Now(),
		Entity:    table_name,
		ModelName: toCamelCase(table_name),
		IdType:    id_data_type,
	})
	return true
}

func echoBuildPgRepo(table_name string, attributes map[string]Attr, id_data_type string, idName string, is_id_auto bool, app string) bool {
	dirPath := app + "/services/" + table_name
	err := os.Mkdir(dirPath, 0755)

	fileName := dirPath + "/pgRepository.go"
	f, err := os.Create(fileName)
	die(err)
	defer f.Close()

	keys := make([]string, 0, len(attributes))
	i_keys := make([]string, 0, len(attributes))
	u_keys := make([]string, 0, len(attributes))
	cr_keys := make([]string, 0, len(attributes))
	for k := range attributes {
		keys = append(keys, k)

	}
	fmt.Print(table_name, is_id_auto)
	sort.Strings(keys)
	n := 1
	for i, ke := range keys {
		if ke != idName || !is_id_auto {
			ik := fmt.Sprintf("$%d", (i + 1))
			i_keys = append(i_keys, ik)
			cr_keys = append(cr_keys, ke)
		}
		if ke != idName {
			nk := fmt.Sprintf("$%d", n)
			u_keys = append(u_keys, ke+"="+nk)
			n += 1
		}

	}
	updateAttributes := make(map[string]Attr)
	insAttributes := make(map[string]Attr)
	for key, value := range attributes {
		updateAttributes[key] = value
	}
	if idName != "" {
		delete(updateAttributes, idName)
	}
	if is_id_auto {
		insAttributes = updateAttributes
	} else {
		insAttributes = attributes
	}
	repoTemplate, err := template.New("entity").Parse(templates.EchoRepoTmpl)
	die(err)
	repoTemplate.Execute(f, struct {
		Timestamp        time.Time
		Entity           string
		ModelName        string
		Attributes       map[string]Attr
		Fields           string
		IdType           string
		CrFields         string
		InsFields        string
		IdName           string
		UpdateAttributes map[string]Attr
		InsAttributes    map[string]Attr
		UpAttrLen        int
		IdModelName      string
		InsFdName        string
	}{
		Timestamp:        time.Now(),
		Entity:           table_name,
		ModelName:        toCamelCase(table_name),
		Attributes:       attributes,
		InsAttributes:    insAttributes,
		Fields:           strings.Join(keys, ","),
		IdType:           id_data_type,
		CrFields:         strings.Join(u_keys, ","),
		InsFields:        strings.Join(i_keys, ","),
		InsFdName:        strings.Join(cr_keys, ","),
		IdName:           idName,
		UpdateAttributes: updateAttributes,
		UpAttrLen:        (len(updateAttributes) + 1),
		IdModelName:      toCamelCase(idName),
	})

	return true
}
