// Code generated by go-swagger; DO NOT EDIT.

package ui

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime/middleware"

	strfmt "github.com/go-openapi/strfmt"
)

// NewGetUIFileParams creates a new GetUIFileParams object
// no default values defined in spec.
func NewGetUIFileParams() GetUIFileParams {

	return GetUIFileParams{}
}

// GetUIFileParams contains all the bound params for the get UI file operation
// typically these are obtained from a http.Request
//
// swagger:parameters getUIFile
type GetUIFileParams struct {

	// HTTP Request Object
	HTTPRequest *http.Request `json:"-"`

	/*UI file name
	  Required: true
	  In: path
	*/
	Filename string
}

// BindRequest both binds and validates a request, it assumes that complex things implement a Validatable(strfmt.Registry) error interface
// for simple values it will use straight method calls.
//
// To ensure default values, the struct must have been initialized with NewGetUIFileParams() beforehand.
func (o *GetUIFileParams) BindRequest(r *http.Request, route *middleware.MatchedRoute) error {
	var res []error

	o.HTTPRequest = r

	rFilename, rhkFilename, _ := route.Params.GetOK("filename")
	if err := o.bindFilename(rFilename, rhkFilename, route.Formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

// bindFilename binds and validates parameter Filename from path.
func (o *GetUIFileParams) bindFilename(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: true
	// Parameter is provided by construction from the route

	o.Filename = raw

	return nil
}