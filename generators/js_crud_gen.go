// The following directive is necessary to make the package coherent:

// +build ignore

// This program generates contributors.go. It can be invoked by running
// go generate
package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/heraju/mestri"

	_ "github.com/lib/pq"
)

type Model struct {
	Package   string
	ModelName string
}

type Attr struct {
	DataType   string
	ModelName  string
	KeyType    string
	ReaderType string
}

var root = "express"
var rootDir = "express/"

func main() {
	db, err := sql.Open("postgres", mestri.PsqlInfo)
	die(err)
	defer db.Close()

	rows, err := db.Query("SELECT table_name FROM information_schema.tables WHERE table_schema = 'public'")

	entities := make([]string, 0)
	models := make([]Model, 0)
	cleanApp()
	for rows.Next() {
		var table_name string
		err := rows.Scan(&table_name)
		die(err)
		buildEntity(table_name, db)
		entities = append(entities, table_name)
		models = append(models, Model{Package: table_name, ModelName: toCamelCase(table_name)})

	}
	buildApp(entities, models)
	fmt.Print(entities)
}

func cleanApp() {
	_, err := os.Stat(root)
	if os.IsNotExist(err) {
		errDir := os.MkdirAll(root, 0755)
		if errDir != nil {
			log.Fatal(err)
		}

	} else {
		os.RemoveAll(rootDir)
		errDir := os.MkdirAll(root, 0755)
		if errDir != nil {
			log.Fatal(err)
		}
	}
}

func buildApp(entities []string, models []Model) bool {
	buidPackageFile()
	buidDbFile()
	fileName := rootDir + "/app.js"
	f, err := os.Create(fileName)
	die(err)
	defer f.Close()
	appTemplate := template.Must(template.ParseFiles("templates/express/app.tmpl"))
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

func buidPackageFile() bool {
	fileName := rootDir + "/package.json"
	f, err := os.Create(fileName)
	die(err)
	defer f.Close()
	appTemplate := template.Must(template.ParseFiles("templates/express/package.tmpl"))
	appTemplate.Execute(f, struct{}{})
	return true
}

func buidDbFile() bool {
	dirPath := rootDir + "/db/"
	err := os.Mkdir(dirPath, 0755)
	die(err)
	fileName := rootDir + "db/index.js"
	f, err := os.Create(fileName)
	die(err)
	defer f.Close()
	appTemplate := template.Must(template.ParseFiles("templates/express/db.tmpl"))
	appTemplate.Execute(f, struct{}{})
	return true
}

func getIdName(table_name string, db *sql.DB) string {
	var key_column string
	sqlStmt := `select kcu.column_name as key_column
	from information_schema.table_constraints tco
	join information_schema.key_column_usage kcu 
	  on kcu.constraint_name = tco.constraint_name
	  and kcu.constraint_schema = tco.constraint_schema
	  and kcu.constraint_name = tco.constraint_name
	where tco.constraint_type = 'PRIMARY KEY' and kcu.table_schema = 'public' and kcu.table_name = $1`
	row := db.QueryRow(sqlStmt, table_name)
	err := row.Scan(&key_column)
	if err != nil {
		return ""
	}
	return key_column
}

func buildEntity(table_name string, db *sql.DB) bool {

	fmt.Println("Building CRUD For --- : ", table_name)

	idName := getIdName(table_name, db)

	attr, err := db.Query("select column_name, data_type, column_default from information_schema.columns where table_name = $1 order by column_name", table_name)
	die(err)
	entity := make(map[string]Attr)
	var id_data_type string
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
		case "text":
			data_type_map = "NullString"
			key_type = "string"
		case "integer":
			data_type_map = "NullInt64"
			key_type = "int64"
		case "timestamp with time zone":
			data_type_map = "NullString"
			key_type = "string"
		case "json":
			data_type_map = "NullString"
			key_type = "string"
		}
		if column_name == idName {
			id_data_type = key_type

			if column_default.String != "" {
				is_id_auto = true
			}
		}
		entity[column_name] = Attr{ReaderType: toCamelCase(key_type), DataType: data_type_map, ModelName: toCamelCase(column_name), KeyType: key_type}
	}

	buildPgRepo(table_name, entity, id_data_type, idName, is_id_auto)
	return true
}

func buildPgRepo(table_name string, attributes map[string]Attr, id_data_type string, idName string, is_id_auto bool) bool {

	dirPath := rootDir + "/controlers/"

	err := os.Mkdir(dirPath, 0755)

	fileName := dirPath + table_name + ".js"
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
	repoTemplate := template.Must(template.ParseFiles("templates/express/entity.tmpl"))
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
		IdModelName:      idName,
	})

	return true
}

// Utility functions
func die(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

var link = regexp.MustCompile("(^[A-Za-z])|_([A-Za-z])")

func toCamelCase(str string) string {
	return link.ReplaceAllStringFunc(str, func(s string) string {
		return strings.ToUpper(strings.Replace(s, "_", "", -1))
	})
}
