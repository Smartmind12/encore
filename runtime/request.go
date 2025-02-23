package encore

import (
	"reflect"
	"time"

	"encore.dev/appruntime/model"
)

var applicationStartTime = time.Now()

// APIDesc describes the API endpoint being called.
type APIDesc struct {
	// RequestType specifies the type of the request payload,
	// or nil if the endpoint has no request payload or is Raw.
	RequestType reflect.Type

	// ResponseType specifies the type of the response payload,
	// or nil if the endpoint has no response payload or is Raw.
	ResponseType reflect.Type

	// Raw specifies whether the endpoint is a Raw endpoint.
	Raw bool
}

// Request provides metadata about how and why the currently running code was started.
//
// The current request can be returned by calling CurrentRequest()
type Request struct {
	Type    RequestType // What caused this request to start
	Started time.Time   // What time the trigger occurred

	// APICall specific parameters.
	// These will be empty for operations with a type not APICall
	API        *APIDesc   // Metadata about the API endpoint being called
	Service    string     // Which service is processing this request
	Endpoint   string     // Which API endpoint is being called
	Path       string     // What was the path made to the API server
	PathParams PathParams // If there are path parameters, what are they?

	// Payload is the decoded request payload or Pub/Sub message payload,
	// or nil if the API endpoint has no request payload or the endpoint is raw.
	Payload any
}

// RequestType describes how the currently running code was triggered
type RequestType string

const (
	None          RequestType = "none"           // There was no external trigger which caused this code to run. Most likely it was triggered by a package level init function.
	APICall       RequestType = "api-call"       // The code was triggered via an API call to a service
	PubSubMessage RequestType = "pubsub-message" // The code was triggered by a PubSub subscriber
)

// PathParams contains the path parameters parsed from the request path.
// The ordering of the parameters in the path will be maintained from the URL.
type PathParams []PathParam

// PathParam represents a parsed path parameter.
type PathParam struct {
	Name  string // the name of the path parameter, without leading ':' or '*'.
	Value string // the parsed path parameter value.
}

// Get returns the value of the path parameter with the given name.
// If no such parameter exists it reports "".
func (p PathParams) Get(name string) string {
	for _, param := range p {
		if param.Name == name {
			return param.Value
		}
	}

	return ""
}

func (mgr *Manager) CurrentRequest() *Request {
	req := mgr.rt.Current().Req
	if req == nil {
		return &Request{
			Type:    None,
			Started: applicationStartTime,
		}
	}

	opType := None
	switch req.Type {
	case model.RPCCall, model.AuthHandler:
		opType = APICall
	case model.PubSubMessage:
		opType = PubSubMessage
	}

	pathParams := make(PathParams, len(req.PathSegments))
	for i, param := range req.PathSegments {
		pathParams[i].Name = param.Name
		pathParams[i].Value = param.Value
	}

	var api *APIDesc
	if d := req.RPCDesc; d != nil {
		api = &APIDesc{
			RequestType:  d.RequestType,
			ResponseType: d.ResponseType,
			Raw:          d.Raw,
		}
	}

	return &Request{
		Type:       opType,
		API:        api,
		Service:    req.Service,
		Endpoint:   req.Endpoint,
		Started:    req.Start,
		Path:       req.Path,
		PathParams: pathParams,
		Payload:    req.Payload,
	}
}
