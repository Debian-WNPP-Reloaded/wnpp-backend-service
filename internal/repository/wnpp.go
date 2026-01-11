package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type WNPPItem struct {
	BugID        int       `json:"bug_id"`
	Type         string    `json:"type"`
	Source       string    `json:"source"`

	WNPPPackage  string    `json:"wnpp_package"`
	Arrival      time.Time `json:"arrival"`
	Submitter    string    `json:"submitter"`
	Owner        string    `json:"owner"`
	OwnerName    string    `json:"owner_name"`
	OwnerEmail   string    `json:"owner_email"`
	LastModified time.Time `json:"last_modified"`
	Title        string    `json:"title"`

	Installs     *int      `json:"installs"` // nullable
	Users        *int      `json:"users"`    // nullable
}

type WNPPRepository struct {
	db *pgxpool.Pool
}

func NewWNPPRepository(db *pgxpool.Pool) *WNPPRepository {
	return &WNPPRepository{db: db}
}

func (r *WNPPRepository) List(
	ctx context.Context,
	limit int,
	offset int,
	orderBy string,
	typ string,
) ([]WNPPItem, error) {

	// whitelist ordering (VERY IMPORTANT)
	orderClause := "b.arrival DESC"
	switch orderBy {
	case "arrival":
		orderClause = "b.arrival DESC"
	case "installs":
		orderClause = "p.insts DESC NULLS LAST"
	case "users":
		orderClause = "p.vote DESC NULLS LAST"
	}

	whereClause := `
WHERE
    b.package = 'wnpp'
    AND b.status <> 'done'
`
	args := []any{limit, offset}
	//argPos := 3

	if typ != "" {
		whereClause += " AND b.title ~ $3"
		args = append(args, "^"+typ+": ")
	}

	query := `
SELECT
    b.id                                         AS bug_id,
    substring(b.title, '^([A-Z]{1,3}): .*')      AS type,
    substring(b.title, '^[A-Z]{1,3}: ([^ ]+)')   AS source,
    b.package                                    AS wnpp_package,
    b.arrival,
    b.submitter,
    b.owner,
    b.owner_name,
    b.owner_email,
    b.last_modified,
    b.title,
    p.insts                                      AS installs,
    p.vote                                       AS users
FROM public.bugs b
LEFT JOIN public.popcon p
       ON p.package = substring(
            b.title,
            '^[A-Z]{1,3}: ([^ ]+)'
          )
` + whereClause + `
ORDER BY ` + orderClause + `
LIMIT $1 OFFSET $2
`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]WNPPItem, 0, limit)

	for rows.Next() {
		var item WNPPItem

		err := rows.Scan(
			&item.BugID,
			&item.Type,
			&item.Source,
			&item.WNPPPackage,
			&item.Arrival,
			&item.Submitter,
			&item.Owner,
			&item.OwnerName,
			&item.OwnerEmail,
			&item.LastModified,
			&item.Title,
			&item.Installs,
			&item.Users,
		)
		if err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *WNPPRepository) Count(ctx context.Context, typ string) (int, error) {
	whereClause := `
WHERE
    package = 'wnpp'
    AND status <> 'done'
`
	args := []any{}

	if typ != "" {
		whereClause += " AND title ~ $1"
		args = append(args, "^"+typ+": ")
	}

	query := `
SELECT COUNT(*)
FROM public.bugs
` + whereClause

	var total int
	err := r.db.QueryRow(ctx, query, args...).Scan(&total)
	if err != nil {
		return 0, err
	}

	return total, nil
}
