package web

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"log"
	"net/http"
	"reflect"
	"text/template"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/mosajjal/aio-gw/conf"
)

func RenderTemplate(myTemplate string, values map[string]string) (string, error) {
	t, err := template.New("webpage").Parse(myTemplate)
	if err != nil {
		log.Println(err)
		return "", err
	}
	buf := new(bytes.Buffer)
	t.Execute(buf, values)
	return buf.String(), nil
}

func RenderFieldTemplate(myStruct interface{}, fieldName string, v reflect.Type, prefix string) (string, error) {
	fillValues := map[string]string{"name": prefix + v.Name() + "." + fieldName, "label": fieldName}
	switch expression := conf.GetTagValue(myStruct, fieldName, "inputtype"); expression {
	case "text":
		return RenderTemplate(TextInputTempalte, fillValues)

	case "toggle":
		return RenderTemplate(ToggleTemplate, fillValues)

	case "options":
		return RenderTemplate(OptionsTemplate, fillValues)

	case "file":
		return RenderTemplate(FileTemplate, fillValues)
	}
	return "", nil
}

func RenderChildStruct(inputStruct interface{}, structName string, parentStructName string) (string, error) {
	// fields: name, title and content
	t, err := template.New("card").Parse(ChildCardTemplate)
	if err != nil {
		log.Println(err)
		return "", err
	}
	content := ""
	v := reflect.TypeOf(inputStruct)
	for i := 0; i < v.NumField(); i++ {
		fieldName := v.Field(i).Name
		// fmt.Printf("%#v\n\n", v.Field(i)) //todo:remove
		newContent, err := RenderFieldTemplate(inputStruct, fieldName, v, parentStructName+"."+structName)
		if err != nil {
			log.Println(err)
		}
		content += newContent
	}
	buf := new(bytes.Buffer)
	t.Execute(buf, map[string]string{"title": structName, "name": structName, "content": content})
	return buf.String(), nil
}

func RenderParentStruct(inputStruct interface{}) (string, error) {
	t, err := template.New("parentcard").Parse(ParentCardTemplate)
	if err != nil {
		log.Println(err)
		return "", err
	}
	v := reflect.TypeOf(inputStruct)
	content := ""
	for i := 0; i < v.NumField(); i++ {
		// log.Printf("%##v", x)
		fieldName := v.Field(i).Name
		if v.Field(i).Type.Kind() == reflect.Struct {
			childStruct := reflect.Indirect(reflect.ValueOf(inputStruct)).FieldByName(v.Field(i).Name).Interface()
			newContent, err := RenderChildStruct(childStruct, fieldName, v.Name())
			if err != nil {
				log.Println(err)
			}
			content += newContent
		} else {
			newContent, err := RenderFieldTemplate(inputStruct, fieldName, v, "")
			if err != nil {
				log.Println(err)
			}
			content += newContent
		}
	}
	buf := new(bytes.Buffer)
	t.Execute(buf, map[string]string{"title": v.Name(), "name": v.Name(), "content": content})
	return buf.String(), nil
}

func RenderFirstPage(c echo.Context) error {
	t, err := template.New("webpage").Parse(PageTemplate)
	if err != nil {
		log.Println(err)
		return c.String(http.StatusInternalServerError, "Internal Server Error")
	}
	content := ""

	newContent, _ := RenderParentStruct(conf.GlobalProxySettings)
	content += newContent
	newContent, _ = RenderParentStruct(conf.GlobalServiceSettings)
	content += newContent
	newContent, _ = RenderParentStruct(conf.GlobalUpstreamSettings)
	content += newContent
	newContent, _ = RenderParentStruct(conf.GlobalWebserverSettings)
	content += newContent
	// log.Println(content)
	t.Execute(c.Response().Writer, map[string]string{"content": content})
	return nil
}

func PostWebserverSettings(c echo.Context) error {
	var settings conf.WebserverSettings
	if err := json.NewDecoder(c.Request().Body).Decode(&settings); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Bad Request")
	} else {
		conf.GlobalWebserverSettings = settings
		return c.JSON(http.StatusOK, conf.GlobalWebserverSettings)
	}
}

func GetWebserverSettings(c echo.Context) error {
	return c.JSON(http.StatusOK, conf.GlobalWebserverSettings)
}

func PostUpstreamSettings(c echo.Context) error {
	var settings conf.WebserverSettings
	if err := json.NewDecoder(c.Request().Body).Decode(&settings); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Bad Request")
	} else {
		conf.GlobalWebserverSettings = settings
		return c.JSON(http.StatusOK, conf.GlobalWebserverSettings)
	}
}

func GetUpstreamSettings(c echo.Context) error {
	return c.JSON(http.StatusOK, conf.GlobalWebserverSettings)
}

func PostProxySettings(c echo.Context) error {
	var settings conf.ProxySettings
	if err := json.NewDecoder(c.Request().Body).Decode(&settings); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Bad Request")
	} else {
		conf.GlobalProxySettings = settings
		return c.JSON(http.StatusOK, conf.GlobalProxySettings)
	}
}

func getProxySettings(c echo.Context) error {
	return c.JSON(http.StatusOK, conf.GlobalProxySettings)
}

func PostServiceSettings(c echo.Context) error {
	var settings conf.ServiceSettings
	if err := json.NewDecoder(c.Request().Body).Decode(&settings); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Bad Request")
	} else {
		conf.GlobalServiceSettings = settings
		return c.JSON(http.StatusOK, conf.GlobalServiceSettings)
	}
}

func getServiceSettings(c echo.Context) error {
	return c.JSON(http.StatusOK, conf.GlobalServiceSettings)
}

//go:embed admin.js
var adminJS []byte

func ServeAdminJS(c echo.Context) error {
	return c.Blob(http.StatusOK, "application/javascript", adminJS)
}

func ApplyWebServerSettings(webserverSettings conf.WebserverSettings) error {
	if webserverSettings.Enabled {
		e := echo.New()
		// CORS settings
		e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins: []string{"*"},
		}))

		e.Use(middleware.Logger())
		e.Use(middleware.Recover())
		// paths to serve

		e.GET("/", RenderFirstPage)
		e.GET("/admin.js", ServeAdminJS)
		e.GET("/api/WebserverSettings", GetWebserverSettings)
		e.POST("/api/WebserverSettings", PostWebserverSettings)
		e.GET("/api/UpstreamSettings", GetUpstreamSettings)
		e.POST("/api/UpstreamSettings", PostUpstreamSettings)
		e.GET("/api/ProxySettings", getProxySettings)
		e.POST("/api/ProxySettings", PostProxySettings)
		e.GET("/api/ServiceSettings", getServiceSettings)
		e.POST("/api/ServiceSettings", PostServiceSettings)

		// http.authentication # TODO
		if webserverSettings.Tls.Enabled {
			e.Logger.Fatal(e.StartTLS(webserverSettings.Bind, webserverSettings.Tls.Cert, webserverSettings.Tls.Key))
		} else {
			e.Logger.Fatal(e.Start(webserverSettings.Bind))
		}
	} else {
		log.Println("Admin server is disabled")
	}
	return nil
}
