// Code generated by go-swagger; DO NOT EDIT.

package vsphere

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	middleware "github.com/go-openapi/runtime/middleware"
)

// SetVSphereEndpointHandlerFunc turns a function with the right signature into a set v sphere endpoint handler
type SetVSphereEndpointHandlerFunc func(SetVSphereEndpointParams) middleware.Responder

// Handle executing the request and returning a response
func (fn SetVSphereEndpointHandlerFunc) Handle(params SetVSphereEndpointParams) middleware.Responder {
	return fn(params)
}

// SetVSphereEndpointHandler interface for that can handle valid set v sphere endpoint params
type SetVSphereEndpointHandler interface {
	Handle(SetVSphereEndpointParams) middleware.Responder
}

// NewSetVSphereEndpoint creates a new http.Handler for the set v sphere endpoint operation
func NewSetVSphereEndpoint(ctx *middleware.Context, handler SetVSphereEndpointHandler) *SetVSphereEndpoint {
	return &SetVSphereEndpoint{Context: ctx, Handler: handler}
}

/*
SetVSphereEndpoint swagger:route POST /api/providers/vsphere vsphere setVSphereEndpoint

Validate and set vSphere credentials
*/
type SetVSphereEndpoint struct {
	Context *middleware.Context
	Handler SetVSphereEndpointHandler
}

func (o *SetVSphereEndpoint) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		r = rCtx
	}
	var Params = NewSetVSphereEndpointParams()

	if err := o.Context.BindValidRequest(r, route, &Params); err != nil { // bind params
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}

	res := o.Handler.Handle(Params) // actually handle the request

	o.Context.Respond(rw, r, route.Produces, route, res)

}