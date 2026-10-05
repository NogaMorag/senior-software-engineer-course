package booking

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Schema creates the bookings table. The exclusion constraint is the whole design: the
// database refuses a row whose resource is equal (=) to an existing row's and whose range
// overlaps (&&) it. Equality on a text column inside a GiST index needs the btree_gist
// extension; the range overlap does not. CREATE EXTENSION needs a privileged role, so in
// production it belongs in a migration run by the owner, not in application start-up.
const Schema = `
CREATE EXTENSION IF NOT EXISTS btree_gist;

CREATE TABLE IF NOT EXISTS bookings (
    id       bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    resource text NOT NULL,
    who      text NOT NULL,
    during   tstzrange NOT NULL CHECK (NOT isempty(during)),
    CONSTRAINT bookings_no_overlap EXCLUDE USING gist (resource WITH =, during WITH &&)
);
`

// exclusionViolation is the SQLSTATE PostgreSQL returns when an exclusion constraint
// rejects a row (class 23, integrity constraint violation).
const exclusionViolation = "23P01"

// Postgres stores bookings in PostgreSQL. Concurrent Book calls from any number of
// processes are safe: the constraint, not the application, decides.
type Postgres struct {
	DB *pgxpool.Pool
}

// Book inserts the row and lets the constraint reject an overlap. There is no
// SELECT-then-INSERT, so there is no window between the check and the write.
func (p *Postgres) Book(ctx context.Context, resource, who string, start, end time.Time) (Booking, error) {
	if err := validate(resource, who, start, end); err != nil {
		return Booking{}, err
	}
	b := Booking{Resource: resource, Who: who, Start: start, End: end}
	err := p.DB.QueryRow(ctx,
		`INSERT INTO bookings (resource, who, during) VALUES ($1, $2, tstzrange($3, $4, '[)')) RETURNING id`,
		resource, who, start, end).Scan(&b.ID)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == exclusionViolation {
		return Booking{}, ErrConflict
	}
	if err != nil {
		return Booking{}, err
	}
	return b, nil
}

// Cancel deletes the booking; the freed range can be booked at once.
func (p *Postgres) Cancel(ctx context.Context, id int64) error {
	tag, err := p.DB.Exec(ctx, `DELETE FROM bookings WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// List returns the bookings of resource ordered by start.
func (p *Postgres) List(ctx context.Context, resource string) ([]Booking, error) {
	rows, err := p.DB.Query(ctx,
		`SELECT id, resource, who, lower(during), upper(during) FROM bookings WHERE resource = $1 ORDER BY lower(during)`,
		resource)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (Booking, error) {
		var b Booking
		err := row.Scan(&b.ID, &b.Resource, &b.Who, &b.Start, &b.End)
		return b, err
	})
}
