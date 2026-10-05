// Package grpcapi is the internal gRPC face of the catalog, generated from proto/catalog/v1.
package grpcapi

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	catalogv1 "example.com/ch09-catalog-api/gen/catalog/v1"
	"example.com/ch09-catalog-api/internal/catalog"
)

// Server implements catalogv1.CatalogServiceServer over the same store as the HTTP API.
// Embedding the Unimplemented type keeps it compiling when the proto gains a method.
type Server struct {
	catalogv1.UnimplementedCatalogServiceServer
	Store catalog.Store
}

// Ping answers at once.
func (s *Server) Ping(context.Context, *catalogv1.PingRequest) (*catalogv1.PingResponse, error) {
	return &catalogv1.PingResponse{Message: "pong"}, nil
}

// GetItem maps the store's errors to gRPC status codes: the gRPC counterpart of problem details.
func (s *Server) GetItem(ctx context.Context, req *catalogv1.GetItemRequest) (*catalogv1.GetItemResponse, error) {
	// TODO(day 21): reject an id without the "itm_" prefix or longer than 64 bytes with InvalidArgument,
	// map catalog.ErrNotFound to NotFound and other errors to Internal, and convert the item
	// (CreatedAtUnixMs is CreatedAt.UnixMilli()).
	return nil, status.Error(codes.Unimplemented, "TODO(day 21): GetItem")
}
