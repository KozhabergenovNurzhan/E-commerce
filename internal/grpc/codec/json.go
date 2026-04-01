// Package codec provides a JSON codec for gRPC, replacing the default protobuf
// codec. This allows using plain Go structs as request/response types without
// requiring protoc code generation.
package codec

import (
	"encoding/json"

	"google.golang.org/grpc/encoding"
)

// Name is the content-subtype for the JSON codec. Registered under "proto" so
// that standard gRPC clients don't need explicit content-subtype negotiation.
const Name = "proto"

func init() {
	encoding.RegisterCodec(JSONCodec{})
}

// JSONCodec implements encoding.Codec using standard library JSON.
type JSONCodec struct{}

func (JSONCodec) Name() string { return Name }

func (JSONCodec) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (JSONCodec) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
