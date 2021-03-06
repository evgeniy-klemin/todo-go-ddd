// Package ports provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.8.2 DO NOT EDIT.
package ports

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/deepmap/oapi-codegen/pkg/runtime"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v4"
)

// ServerInterface represents all server handlers.
type ServerInterface interface {
	// Get all items
	// (GET /items)
	GetItems(ctx echo.Context, params GetItemsParams) error
	// Create New User
	// (POST /items)
	PostItems(ctx echo.Context) error
	// Get Item Info by Item ID
	// (GET /items/{item_id})
	GetItemsItemId(ctx echo.Context, itemId ItemId) error
	// Update Item
	// (PATCH /items/{item_id})
	PatchItemsItemid(ctx echo.Context, itemId ItemId) error
}

// ServerInterfaceWrapper converts echo contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler ServerInterface
}

// GetItems converts echo context to params.
func (w *ServerInterfaceWrapper) GetItems(ctx echo.Context) error {
	var err error

	// Parameter object where we will unmarshal all parameters from the context
	var params GetItemsParams
	// ------------- Optional query parameter "_per_page" -------------

	err = runtime.BindQueryParameter("form", true, false, "_per_page", ctx.QueryParams(), &params.PerPage)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter _per_page: %s", err))
	}

	// ------------- Optional query parameter "_page" -------------

	err = runtime.BindQueryParameter("form", true, false, "_page", ctx.QueryParams(), &params.Page)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter _page: %s", err))
	}

	// ------------- Optional query parameter "_sort" -------------

	err = runtime.BindQueryParameter("form", true, false, "_sort", ctx.QueryParams(), &params.Sort)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter _sort: %s", err))
	}

	// ------------- Optional query parameter "_fields" -------------

	err = runtime.BindQueryParameter("form", true, false, "_fields", ctx.QueryParams(), &params.Fields)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter _fields: %s", err))
	}

	// ------------- Optional query parameter "done" -------------

	err = runtime.BindQueryParameter("form", true, false, "done", ctx.QueryParams(), &params.Done)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter done: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.GetItems(ctx, params)
	return err
}

// PostItems converts echo context to params.
func (w *ServerInterfaceWrapper) PostItems(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.PostItems(ctx)
	return err
}

// GetItemsItemId converts echo context to params.
func (w *ServerInterfaceWrapper) GetItemsItemId(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "item_id" -------------
	var itemId ItemId

	err = runtime.BindStyledParameterWithLocation("simple", false, "item_id", runtime.ParamLocationPath, ctx.Param("item_id"), &itemId)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter item_id: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.GetItemsItemId(ctx, itemId)
	return err
}

// PatchItemsItemid converts echo context to params.
func (w *ServerInterfaceWrapper) PatchItemsItemid(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "item_id" -------------
	var itemId ItemId

	err = runtime.BindStyledParameterWithLocation("simple", false, "item_id", runtime.ParamLocationPath, ctx.Param("item_id"), &itemId)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter item_id: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.PatchItemsItemid(ctx, itemId)
	return err
}

// This is a simple interface which specifies echo.Route addition functions which
// are present on both echo.Echo and echo.Group, since we want to allow using
// either of them for path registration
type EchoRouter interface {
	CONNECT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	DELETE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	GET(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	HEAD(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	OPTIONS(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PATCH(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	POST(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PUT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	TRACE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
}

// RegisterHandlers adds each server route to the EchoRouter.
func RegisterHandlers(router EchoRouter, si ServerInterface) {
	RegisterHandlersWithBaseURL(router, si, "")
}

// Registers handlers, and prepends BaseURL to the paths, so that the paths
// can be served under a prefix.
func RegisterHandlersWithBaseURL(router EchoRouter, si ServerInterface, baseURL string) {

	wrapper := ServerInterfaceWrapper{
		Handler: si,
	}

	router.GET(baseURL+"/items", wrapper.GetItems)
	router.POST(baseURL+"/items", wrapper.PostItems)
	router.GET(baseURL+"/items/:item_id", wrapper.GetItemsItemId)
	router.PATCH(baseURL+"/items/:item_id", wrapper.PatchItemsItemid)

}

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"H4sIAAAAAAAC/9xYa2/bNhf+KwTf96NsS46zrhoKrNfB6M3oBRvWBQEtHVlsJFIlqSRe4P8+8KKLZTqx",
	"0xYr9sWwpMNz43Oec8gbnPCy4gyYkji+wRURpAQFwjydV2QF+k8KMhG0UpQzHOMFWQFidbkEgQNM9asv",
	"NYg1DjAjJeDYLQywTHIoidWQkbpQOI4CXJJrWtYljqMwDHBJmXsKsFpXejllClYg8GYT4PMKxB4/nvKa",
	"KUQVlBJVIJAz6neo0eJ1ahpue3W3W9rqOU13nZorKNH8WeNHRVTeudGsCrCALzUVkOJYiRr6TmVclETh",
	"GNe1kXTGpRKUrfBGG7fCZoueC/GEpAu9bXYDBa9AKArmKaNQpOafydIeAf1nYCXAJUjpsj70oPWJLz9D",
	"onD3gghB1nevpqrQr54LwQX2aNM5XBCV5Hp9Sdmi53MUDEJIOevbWXJeAGFajU26J7aKS2q368a3t41/",
	"nRv7fORS7eb0vmY7SHyyOs4Grmhrezx5B7LiTHqK5ANPuakRPExcIoAoSM+J8qzKAaVEAVI5UUjlYFSg",
	"KyKRWzbGQQdVLTpStIRdvAbtBm1beMYZoKwgq25Jb+t8lfWR0S81IJoCUzSjIFDGhfFtRS+BGQ+1V3BN",
	"yspkLZpFKZlNyWg2TR6OZg/CaLR8sExGaXJ6ssxmYXSazvph+CuuA1Kn2WQjwoY1XgFbqbzhjVt3fcCj",
	"7ouJQ3Kh+s5Hd1JQHy80HaJlH1LkAVCROOgI4/8CMhzj/026TjFxBDTZgt6gcjpDQ3rQDAZJLahav9eK",
	"LBwV2GIyug0cgAjo8UOuVGX5j7KMGwRzpkhiVkFJaIFjDJcrYHQ9viigpOzXlX49TnjZcfBzK4FeGgkD",
	"0K0sKJ2Fx4s57sLRr3CAL0FIKxONQ72QV8BIRXGMT8bhONQVRlRuopm0+VuBp75+A4VIUbTZ1nVJ9Ld5",
	"ar/O3Yd+Q/7k34xOZNK1uU1wgHAjuO3bey4UWq6Rax57GqrDa9e3uuIYaZmgxf1uC9sx+Q6UoHAJKAGh",
	"CGV32G6/+qwfbfwFLRQIHbFhKr9N98lnUDdwD4ltzmyJglRPeLpu8ArMwIFUVUETs+WTz9KyQ6d8ULkb",
	"W+62nAyopmF4gELno1mxjbj4Bl+SogYDqn4jwNMwejgKfx5NZx+iWRydxtPpn7gh8YwUEiw7H0quvcnH",
	"EGZHh9HmbBPg36nKh9u+5d13srzpb+ZdHNdxmdmKbfy8fYkDnANJ3dT8irIL79BMmdkeVFB2sQ+9f9Vh",
	"eJJY9rAV+ijS76Y/tbX9aBoaKfhFQPFIQpH5UI7/GC0OmN53MLetAcQeLXtm79vVfeCKFCOz1Nd9FClQ",
	"0um9VZ1WOAsj35RAapVzQf+GFBuhk12hF1wsaZqCof9TW0uDMZ4pEIwU6D2ISxDIzqoGNXVZErH2kLgi",
	"K2masXk+s73fE+lTU2+IIAZX7dSy3QH0rNe0gONIpF/zzpCZ/rqaaiZUe1SJ+keKQ2rBjKGeMtDvzTTG",
	"INHTv2g6SDumPV7MkeJugNwKf5feonuFmu7G+q+R23FZvY1gzD65AC2kjyV/uE5oCikqyTXSPqPCzay9",
	"RDXc+6k9Fdrweoc5Pe26pYhKpAdeJNflkhcS63bXCR4Dqq0jrCf811RKylbonZt30ZzZiV1//gFowFXZ",
	"G7hCH6UdWAdEsAncPDi5cTcAm72TYTsKmUNXFyniWe8cppumfir18VQnh9prh/HeWVL/zO3Nw1fPEf/F",
	"Inv78htiaRbOdiXecIVe8JqlX9107B0Ty7geWrsLJ0//Oer00NxO6VKumuuXQSoqdyuwg03CEFxTqRow",
	"enqaVtlisb0Fu19re03EhQU9kai5YWjxaF/owfzY5mbC9nU3/QHVUo867S2K7ma1yYivhR1bWTa3toWh",
	"6J71ZW8Tf7zy0syIXITfu9CMrW9TbQ7vzYXKkNfNTYZebOurFoW7p4gnk4InpMi5VPFJGIamqNz6m37O",
	"defc/BMAAP//NDbyzYcXAAA=",
}

// GetSwagger returns the content of the embedded swagger specification file
// or error if failed to decode
func decodeSpec() ([]byte, error) {
	zipped, err := base64.StdEncoding.DecodeString(strings.Join(swaggerSpec, ""))
	if err != nil {
		return nil, fmt.Errorf("error base64 decoding spec: %s", err)
	}
	zr, err := gzip.NewReader(bytes.NewReader(zipped))
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %s", err)
	}
	var buf bytes.Buffer
	_, err = buf.ReadFrom(zr)
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %s", err)
	}

	return buf.Bytes(), nil
}

var rawSpec = decodeSpecCached()

// a naive cached of a decoded swagger spec
func decodeSpecCached() func() ([]byte, error) {
	data, err := decodeSpec()
	return func() ([]byte, error) {
		return data, err
	}
}

// Constructs a synthetic filesystem for resolving external references when loading openapi specifications.
func PathToRawSpec(pathToFile string) map[string]func() ([]byte, error) {
	var res = make(map[string]func() ([]byte, error))
	if len(pathToFile) > 0 {
		res[pathToFile] = rawSpec
	}

	return res
}

// GetSwagger returns the Swagger specification corresponding to the generated code
// in this file. The external references of Swagger specification are resolved.
// The logic of resolving external references is tightly connected to "import-mapping" feature.
// Externally referenced files must be embedded in the corresponding golang packages.
// Urls can be supported but this task was out of the scope.
func GetSwagger() (swagger *openapi3.T, err error) {
	var resolvePath = PathToRawSpec("")

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.ReadFromURIFunc = func(loader *openapi3.Loader, url *url.URL) ([]byte, error) {
		var pathToFile = url.String()
		pathToFile = path.Clean(pathToFile)
		getSpec, ok := resolvePath[pathToFile]
		if !ok {
			err1 := fmt.Errorf("path not found: %s", pathToFile)
			return nil, err1
		}
		return getSpec()
	}
	var specData []byte
	specData, err = rawSpec()
	if err != nil {
		return
	}
	swagger, err = loader.LoadFromData(specData)
	if err != nil {
		return
	}
	return
}
