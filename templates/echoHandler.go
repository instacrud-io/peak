package templates

// EchoHandlerTmpl : is tmpl to create app
var EchoHandlerTmpl = `// Code generated by Mestri; DO NOT EDIT.
// This file was generated by Mestri robots at
// {{ .Timestamp }}
package {{ .Entity }}
import (
	"fmt"
    "strconv"
	"net/http"
	"github.com/labstack/echo/v4"
    "github.com/heraju/mestri/app/helpers"
	"github.com/sirupsen/logrus"
	validator "gopkg.in/go-playground/validator.v9"
)

// ResponseError represent the reseponse error struct
type ResponseError struct {
	Message string "json:'message'"
}

// ArticleHandler  represent the httphandler for article
type {{.ModelName}}Handler struct {
	UseCase Usecase
}

// {{.ModelName}}Handler will initialize the articles/ resources endpoint
func New{{.ModelName}}Handler(e *echo.Echo, us Usecase) {
	handler := &{{.ModelName}}Handler{UseCase: us}

    e.GET("/{{.Entity}}", handler.Index)
	e.GET("/{{.Entity}}/:id", handler.Get)
    e.POST("/{{.Entity}}", handler.Create)
	e.DELETE("/{{.Entity}}/:id", handler.Delete)
	e.PUT("/{{.Entity}}/:id", handler.Update)
}

// Index will fetch the {{.Entity}} based on given params
func (a *{{.ModelName}}Handler) Index(c echo.Context) error {
    fmt.Println("INDEX {{.ModelName}} CALLED ")
	numS := c.QueryParam("num")
	num, _ := strconv.Atoi(numS)
	cursor := c.QueryParam("cursor")
	ctx := c.Request().Context()

	listAr, nextCursor, err := a.UseCase.Index(ctx, cursor, int64(num))
	if err != nil {
		return c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
	}

	c.Response().Header().Set("X-Cursor", nextCursor)
	return c.JSON(http.StatusOK, listAr)
}

// Get will get {{.Entity}} by given id
func (a *{{.ModelName}}Handler) Get(c echo.Context) error {
    
	{{if .IsStingKey}} 
		idP := c.Param("id")
	{{else}} 
		idP, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return c.JSON(http.StatusNotFound, helpers.ErrNotFound.Error())
		}
	{{end}}
	
	
	id := {{.IdType}}(idP)
	ctx := c.Request().Context()
    fmt.Println("GET {{.ModelName}} CALLED ")
    fmt.Println(id)
	art, err := a.UseCase.Get(ctx, id)
	if err != nil {
		return c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
	}
	
	return c.JSON(http.StatusOK, art)
}


func getStatusCode(err error) int {
	if err == nil {
		return http.StatusOK
	}

	logrus.Error(err)
	switch err {
	case helpers.ErrInternalServerError:
		return http.StatusInternalServerError
	case helpers.ErrNotFound:
		return http.StatusNotFound
	case helpers.ErrConflict:
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}


func isRequestValid(m *Entity) (bool, error) {
	validate := validator.New()
	err := validate.Struct(m)
	if err != nil {
		return false, err
	}
	return true, nil
}

// Create will store the entity by given request body
func (a *{{.ModelName}}Handler) Create(c echo.Context) (err error) {
	var en Entity
	err = c.Bind(&en)
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, err.Error())
	}

	var ok bool
	if ok, err = isRequestValid(&en); !ok {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	ctx := c.Request().Context()
	err = a.UseCase.Create(ctx, &en)
	if err != nil {
		return c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
	}

	return c.JSON(http.StatusCreated, en)
}

// Delete will delete entity by given param
func (a *{{.ModelName}}Handler) Delete(c echo.Context) error {
	 
	{{if .IsStingKey}} 
		idP := c.Param("id")
	{{else}} 
		idP, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return c.JSON(http.StatusNotFound, helpers.ErrNotFound.Error())
		}
	{{end}}
	
	id := {{.IdType}}(idP)
	ctx := c.Request().Context()
    fmt.Println("GET {{.ModelName}} CALLED ")
    fmt.Println(id)
	uerr := a.UseCase.Delete(ctx, id)
	if uerr != nil {
		return c.JSON(getStatusCode(uerr), ResponseError{Message: uerr.Error()})
	}
	
	return c.NoContent(http.StatusNoContent)
}

// Update will update entity by given params
func (a *{{.ModelName}}Handler) Update(c echo.Context) error {
	 
	{{if .IsStingKey}} 
		idP := c.Param("id")
	{{else}} 
		idP, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return c.JSON(http.StatusNotFound, helpers.ErrNotFound.Error())
		}
	{{end}}
	
	var en Entity
	berr := c.Bind(&en)
	if berr != nil {
		return c.JSON(http.StatusUnprocessableEntity, berr.Error())
	}

	var ok bool
	if ok, berr = isRequestValid(&en); !ok {
		return c.JSON(http.StatusBadRequest, berr.Error())
	}

	id := {{.IdType}}(idP)
	ctx := c.Request().Context()
    fmt.Println("GET {{.ModelName}} CALLED ")
    fmt.Println(id)
	uerr := a.UseCase.Update(ctx, &en, id)
	if uerr != nil {
		return c.JSON(getStatusCode(uerr), ResponseError{Message: uerr.Error()})
	}
	
	return c.NoContent(http.StatusNoContent)
}
`
