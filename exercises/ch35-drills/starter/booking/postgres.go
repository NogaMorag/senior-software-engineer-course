package booking

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Schema creates the bookings table. The tests apply it to a fresh schema.
// TODO (drill 5): add the constraint that makes overlapping bookings of one resource
// impossible, whatever the application does. Check what the GiST index needs for the
// equality on resource.
const Schema = `
CREATE TABLE IF NOT EXISTS bookings (
    id       bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    resource text NOT NULL,
    who      text NOT NULL,
    during   tstzrange NOT NULL
);
`

// Postgres stores bookings in PostgreSQL.
type Postgres struct {
	DB *pgxpool.Pool
}

// Book reserves resource for [start, end) or returns ErrConflict or ErrInvalid.
func (p *Postgres) Book(ctx context.Context, resource, who string, start, end time.Time) (Booking, error) {
	panic("TODO: drill 5, see chapter 35, section 35.6")
}

// Cancel removes a booking.
func (p *Postgres) Cancel(ctx context.Context, id int64) error {
	panic("TODO: drill 5, see chapter 35, section 35.6")
}

// List returns the bookings of resource ordered by start.
func (p *Postgres) List(ctx context.Context, resource string) ([]Booking, error) {
	panic("TODO: drill 5, see chapter 35, section 35.6")
}
