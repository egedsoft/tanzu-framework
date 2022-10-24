// Code generated by go-swagger; DO NOT EDIT.

package azure

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	middleware "github.com/go-openapi/runtime/middleware"
)

// GetAzureVnetsHandlerFunc turns a function with the right signature into a get azure vnets handler
type GetAzureVnetsHandlerFunc func(GetAzureVnetsParams) middleware.Responder

// Handle executing the request and returning a response
func (fn GetAzureVnetsHandlerFunc) Handle(params GetAzureVnetsParams) middleware.Responder {
	return fn(params)
}

// GetAzureVnetsHandler interface for that can handle valid get azure vnets params
type GetAzureVnetsHandler interface {
	Handle(GetAzureVnetsParams) middleware.Responder
}

// NewGetAzureVnets creates a new http.Handler for the get azure vnets operation
func NewGetAzureVnets(ctx *middleware.Context, handler GetAzureVnetsHandler) *GetAzureVnets {
	return &GetAzureVnets{Context: ctx, Handler: handler}
}

/*
GetAzureVnets swagger:route GET /api/providers/azure/resourcegroups/{resourceGroupName}/vnets azure getAzureVnets

Retrieve list of Azure virtual networks in a resource group
*/
type GetAzureVnets struct {
	Context *middleware.Context
	Handler GetAzureVnetsHandler
}

func (o *GetAzureVnets) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		r = rCtx
	}
	var Params = NewGetAzureVnetsParams()

	if err := o.Context.BindValidRequest(r, route, &Params); err != nil { // bind params
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}

	res := o.Handler.Handle(Params) // actually handle the request

	o.Context.Respond(rw, r, route.Produces, route, res)

}