package headers

import (
	"github.com/anghami/protoc-gen-connect-openapi/internal/converter/util"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/utils"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// GetServiceHeaders extracts service-level headers from service options
func GetServiceHeaders(service protoreflect.ServiceDescriptor) []*Header {
	opts := service.Options()
	if !proto.HasExtension(opts, E_Service) {
		return nil
	}

	serviceHeaders, ok := proto.GetExtension(opts, E_Service).(*ServiceHeaders)
	if !ok || serviceHeaders == nil {
		return nil
	}

	return serviceHeaders.RequiredHeaders
}

// GetMethodHeaders extracts method-level headers from method options
func GetMethodHeaders(method protoreflect.MethodDescriptor) []*Header {
	opts := method.Options()
	if !proto.HasExtension(opts, E_Method) {
		return nil
	}

	methodHeaders, ok := proto.GetExtension(opts, E_Method).(*MethodHeaders)
	if !ok || methodHeaders == nil {
		return nil
	}

	return methodHeaders.RequiredHeaders
}

// MergeHeaders combines service-level and method-level headers
// Method-level headers take precedence over service-level headers with the same name
func MergeHeaders(serviceHeaders, methodHeaders []*Header) []*Header {
	if len(serviceHeaders) == 0 {
		return methodHeaders
	}
	if len(methodHeaders) == 0 {
		return serviceHeaders
	}

	// Create a map to track method-level header names for deduplication
	methodHeaderNames := make(map[string]struct{})
	for _, header := range methodHeaders {
		methodHeaderNames[header.Name] = struct{}{}
	}

	// Start with method-level headers
	result := make([]*Header, 0, len(serviceHeaders)+len(methodHeaders))
	result = append(result, methodHeaders...)

	// Add service-level headers that don't conflict with method-level headers
	for _, header := range serviceHeaders {
		if _, exists := methodHeaderNames[header.Name]; !exists {
			result = append(result, header)
		}
	}

	return result
}

// HeadersToParameters converts header definitions to OpenAPI parameters
func HeadersToParameters(headers []*Header) []*v3.Parameter {
	if len(headers) == 0 {
		return nil
	}

	params := make([]*v3.Parameter, 0, len(headers))
	for _, header := range headers {
		param := &v3.Parameter{
			Name:        header.Name,
			In:          "header",
			Description: header.Description,
			Required:    util.BoolPtr(header.Required),
			Deprecated:  header.Deprecated,
			Schema:      headerToSchema(header),
		}
		params = append(params, param)
	}

	return params
}

// headerToSchema converts a header definition to an OpenAPI schema
func headerToSchema(header *Header) *base.SchemaProxy {
	schema := &base.Schema{
		Type: []string{header.Type},
	}

	if header.Format != "" {
		schema.Format = header.Format
	}

	if header.Example != "" {
		schema.Example = utils.CreateStringNode(header.Example)
	}

	return base.CreateSchemaProxy(schema)
}
