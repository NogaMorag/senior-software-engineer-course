package grpcapi_test

import (
	"context"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"

	catalogv1 "example.com/ch09-catalog-api/gen/catalog/v1"
	"example.com/ch09-catalog-api/internal/catalog"
	"example.com/ch09-catalog-api/internal/grpcapi"
)

// dial starts a real gRPC server on an in-memory listener and returns a client connected to it.
// bufconn gives the full stack (HTTP/2 framing, protobuf encoding, status codes) without a port.
func dial(t *testing.T, store catalog.Store) catalogv1.CatalogServiceClient {
	t.Helper()
	lis := bufconn.Listen(1 << 20)
	srv := grpc.NewServer()
	catalogv1.RegisterCatalogServiceServer(srv, &grpcapi.Server{Store: store})
	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(srv.Stop)

	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return lis.DialContext(ctx) }),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	return catalogv1.NewCatalogServiceClient(conn)
}

func TestPing(t *testing.T) {
	c := dial(t, catalog.NewMemory(nil))
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	resp, err := c.Ping(ctx, &catalogv1.PingRequest{})
	if err != nil || resp.GetMessage() != "pong" {
		t.Fatalf("Ping = %v, %v; want pong", resp, err)
	}
}

func TestGetItem(t *testing.T) {
	store := catalog.NewMemory(nil)
	it, err := store.Create(context.Background(), "MUG-001", "Blue mug", 4500)
	if err != nil {
		t.Fatal(err)
	}
	c := dial(t, store)

	tests := []struct {
		name     string
		id       string
		wantCode codes.Code
	}{
		{"found", it.ID, codes.OK},
		{"not found", "itm_99999999", codes.NotFound},
		{"malformed id", "42", codes.InvalidArgument},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			resp, err := c.GetItem(ctx, &catalogv1.GetItemRequest{Id: tt.id})
			if got := status.Code(err); got != tt.wantCode {
				t.Fatalf("code = %s, want %s (err %v)", got, tt.wantCode, err)
			}
			if tt.wantCode != codes.OK {
				return
			}
			got := resp.GetItem()
			if got.GetId() != it.ID || got.GetSku() != "MUG-001" || got.GetPriceCents() != 4500 ||
				got.GetCreatedAtUnixMs() != it.CreatedAt.UnixMilli() {
				t.Fatalf("item = %v, want %+v", got, it)
			}
		})
	}
}

func TestDeadlineIsEnforced(t *testing.T) {
	c := dial(t, catalog.NewMemory(nil))
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()
	if _, err := c.Ping(ctx, &catalogv1.PingRequest{}); status.Code(err) != codes.DeadlineExceeded {
		t.Fatalf("expired deadline: code %s, want DeadlineExceeded", status.Code(err))
	}
}
