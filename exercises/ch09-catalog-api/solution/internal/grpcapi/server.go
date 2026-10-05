// Package grpcapi is the internal gRPC face of the catalog, generated from proto/catalog/v1.
package grpcapi

import (
	"context"
	"errors"
	"strings"

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
	id := req.GetId()
	if !strings.HasPrefix(id, "itm_") || len(id) > 64 {
		return nil, status.Error(codes.InvalidArgument, "id must look like itm_00000001")
	}
	it, err := s.Store.Get(ctx, id)
	if errors.Is(err, catalog.ErrNotFound) {
		return nil, status.Errorf(codes.NotFound, "no item with id %s", id)
	}
	if err != nil {
		return nil, status.Error(codes.Internal, "unexpected error")
	}
	return &catalogv1.GetItemResponse{Item: &catalogv1.Item{
		Id: it.ID, Sku: it.SKU, Name: it.Name, PriceCents: it.PriceCents, CreatedAtUnixMs: it.CreatedAt.UnixMilli(),
	}}, nil
}
