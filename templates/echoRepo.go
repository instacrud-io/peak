package templates

// EchoUsecaseTmpl : is tmpl to create app
var EchoRepoTmpl = `// Code generated by Mestri; DO NOT EDIT.
// This file was generated by Mestri robots at
// {{ .Timestamp }}
package {{ .Entity }}
import (
	"context"
	"database/sql"
	"fmt"

	"github.com/sirupsen/logrus"
    "github.com/heraju/mestri/app/helpers"
    
    _ "github.com/lib/pq"
)

type {{.ModelName}}PgRepository struct {
	Conn *sql.DB
}

// Reader : Reader is the model to read from DB
type Reader struct {
	{{ range $key, $value := .Attributes }}
   		{{ $value.ModelName  }} {{ $value.DataType }} "json:'{{$key}},{{ $value.KeyType }}'"
	{{ end }}
}


// NewPgRepository will create an object that represent the article.Repository interface
func NewPgRepository(Conn *sql.DB) Repository {
	return &{{.ModelName}}PgRepository{Conn}
}

func (m *{{.ModelName}}PgRepository) fetch(ctx context.Context, query string, args ...interface{}) (result []Entity, err error) {
	rows, err := m.Conn.QueryContext(ctx, query, args...)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}

	defer func() {
		errRow := rows.Close()
		if errRow != nil {
			logrus.Error(errRow)
		}
	}()

	result = make([]Entity, 0)
	for rows.Next() {
		t := Reader{}
		err = rows.Scan(
			{{ range $key, $value := .Attributes }}
				&t.{{ $value.ModelName  }},
			{{ end }}
			)
		en := Entity{
			{{ range $key, $value := .Attributes }}
				{{ $value.ModelName  }}: t.{{ $value.ModelName  }}.{{ $value.ReaderType  }},
			{{ end }}
		}
		if err != nil {
			logrus.Error(err)
			return nil, err
		}
        
		result = append(result, en)
	}

	return result, nil
}

func (m *{{.ModelName}}PgRepository) Index(ctx context.Context, cursor string, num int64) (res []Entity, nextCursor string, err error) {
	query := "SELECT {{.Fields}} FROM {{ .Entity }} "

	//decodedCursor, err := helpers.DecodeCursor(cursor)
	//if err != nil && cursor != "" {
	//	return nil, "", helpers.ErrBadParamInput
	//}
	
	res, err = m.fetch(ctx, query)
	if err != nil {
		return nil, "", err
	}

	if len(res) == int(num) {
		nextCursor = "next"//helpers.EncodeCursor(res[len(res)-1].Id)
	}

	return
}

func (m *{{.ModelName}}PgRepository) Get(ctx context.Context, id {{.IdType}}) (res Entity, err error) {
	query := "SELECT {{.Fields}} FROM {{ .Entity }} WHERE ID = $1"

	list, err := m.fetch(ctx, query, id)
	if err != nil {
		return Entity{}, err
	}

	if len(list) > 0 {
		res = list[0]
	} else {
		return res, helpers.ErrNotFound
	}

	return
}


func (m *{{.ModelName}}PgRepository) Create(ctx context.Context, en *Entity) (err error) {
	query := "INSERT INTO {{ .Entity }}({{.InsFdName}}) VALUES ({{.InsFields}});"
	stmt, err := m.Conn.PrepareContext(ctx, query)
	if err != nil {
		return
	}

	_, err = stmt.ExecContext(ctx,
	{{ range $key, $value := .InsAttributes }}
		en.{{ $value.ModelName  }},
	{{ end }})

	if err != nil {
		return
	}
	
	return
}

func (m *{{.ModelName}}PgRepository) Update(ctx context.Context, en *Entity, id {{.IdType}}) (err error) {
	query := "UPDATE {{ .Entity }} set {{ .CrFields }} WHERE {{.IdName}} = ${{.UpAttrLen}}"

	stmt, err := m.Conn.PrepareContext(ctx, query)
	if err != nil {
		return
	}

	res, err := stmt.ExecContext(ctx, 
	{{ range $key, $value := .UpdateAttributes }}
		en.{{ $value.ModelName  }},
	{{ end }}
		en.{{.IdModelName}},
	)
	
	if err != nil {
		return
	}
	affect, err := res.RowsAffected()
	if err != nil {
		return
	}
	if affect != 1 {
		err = fmt.Errorf("Weird  Behavior. Total Affected: %d", affect)
		return
	}

	return
}

func (m *{{.ModelName}}PgRepository) Delete(ctx context.Context, id {{.IdType}}) (err error) {
	query := "DELETE FROM {{ .Entity }} WHERE id = ?"

	stmt, err := m.Conn.PrepareContext(ctx, query)
	if err != nil {
		return
	}

	res, err := stmt.ExecContext(ctx, id)
	if err != nil {
		return
	}

	rowsAfected, err := res.RowsAffected()
	if err != nil {
		return
	}

	if rowsAfected != 1 {
		err = fmt.Errorf("Weird  Behavior. Total Affected: %d", rowsAfected)
		return
	}

	return
}
`
