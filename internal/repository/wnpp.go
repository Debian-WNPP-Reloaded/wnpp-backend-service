package repository

import (
	"context"
	"strconv"
	"strings"
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

	Installs *int `json:"installs"`
	Users    *int `json:"users"`
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
	types []string,
	search string,
) ([]WNPPItem, error) {

	// ---- ORDER BY whitelist
	orderClause := "b.arrival DESC"
	switch orderBy {
	case "arrival":
		orderClause = "b.arrival DESC"
	case "installs":
		orderClause = "p.insts DESC NULLS LAST"
	case "users":
		orderClause = "p.vote DESC NULLS LAST"
	}

	baseWhere := `
WHERE
    b.package = 'wnpp'
    AND b.status <> 'done'
`

	argsBase := []any{}
	argPosBase := 1

	// ---- TYPE FILTER (multi)
	if len(types) > 0 {
		baseWhere += " AND b.title ~ $" + strconv.Itoa(argPosBase)
		argsBase = append(argsBase, "^("+strings.Join(types, "|")+"): ")
		argPosBase++
	}

	// ============================================================
	// 1) PACKAGE NAME PREFIX SEARCH
	// ============================================================
	sourceWhere := baseWhere
	args := append([]any{}, argsBase...)
	argPos := argPosBase

	if search != "" {
		sourceWhere += `
 AND substring(b.title, '^[A-Z]{1,3}: ([^ ]+)') ILIKE $` + strconv.Itoa(argPos)
		args = append(args, search+"%")
		argPos++
	}

	sourceQuery := `
SELECT
    b.id                                       AS bug_id,
    substring(b.title, '^([A-Z]{1,3}): .*')    AS type,
    substring(b.title, '^[A-Z]{1,3}: ([^ ]+)') AS source,
    b.package                                  AS wnpp_package,
    b.arrival,
    b.submitter,
    b.owner,
    b.owner_name,
    b.owner_email,
    b.last_modified,
    b.title,
    p.insts                                    AS installs,
    p.vote                                     AS users
FROM public.bugs b
LEFT JOIN public.popcon p
       ON p.package = substring(
            b.title,
            '^[A-Z]{1,3}: ([^ ]+)'
          )
` + sourceWhere + `
ORDER BY ` + orderClause + `
LIMIT $` + strconv.Itoa(argPos) + `
OFFSET $` + strconv.Itoa(argPos+1)

	items, err := r.runQuery(ctx, sourceQuery, append(args, limit, offset), limit)
	if err != nil {
		return nil, err
	}

	// If package-name matches exist OR no search term → return
	if len(items) > 0 || search == "" {
		return items, nil
	}

	// ============================================================
	// 2) DESCRIPTION SEARCH (fallback)
	// ============================================================
	descWhere := baseWhere
	args = append([]any{}, argsBase...)
	argPos = argPosBase

	descWhere += `
 AND b.title ILIKE '% -- %' || $` + strconv.Itoa(argPos) + ` || '%'`
	args = append(args, search)
	argPos++

	descQuery := `
SELECT
    b.id                                       AS bug_id,
    substring(b.title, '^([A-Z]{1,3}): .*')    AS type,
    substring(b.title, '^[A-Z]{1,3}: ([^ ]+)') AS source,
    b.package                                  AS wnpp_package,
    b.arrival,
    b.submitter,
    b.owner,
    b.owner_name,
    b.owner_email,
    b.last_modified,
    b.title,
    p.insts                                    AS installs,
    p.vote                                     AS users
FROM public.bugs b
LEFT JOIN public.popcon p
       ON p.package = substring(
            b.title,
            '^[A-Z]{1,3}: ([^ ]+)'
          )
` + descWhere + `
ORDER BY ` + orderClause + `
LIMIT $` + strconv.Itoa(argPos) + `
OFFSET $` + strconv.Itoa(argPos+1)

	return r.runQuery(ctx, descQuery, append(args, limit, offset), limit)
}

func (r *WNPPRepository) runQuery(
	ctx context.Context,
	query string,
	args []any,
	limit int,
) ([]WNPPItem, error) {

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]WNPPItem, 0, limit)

	for rows.Next() {
		var item WNPPItem
		if err := rows.Scan(
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
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *WNPPRepository) Count(
	ctx context.Context,
	types []string,
	search string,
) (int, error) {

	baseWhere := `
WHERE
    package = 'wnpp'
    AND status <> 'done'
`

	argsBase := []any{}
	argPosBase := 1

	if len(types) > 0 {
		baseWhere += " AND title ~ $" + strconv.Itoa(argPosBase)
		argsBase = append(argsBase, "^("+strings.Join(types, "|")+"): ")
		argPosBase++
	}

	// ---- 1) package-name count
	if search != "" {
		where := baseWhere + `
 AND substring(title, '^[A-Z]{1,3}: ([^ ]+)') ILIKE $` + strconv.Itoa(argPosBase)

		args := append(argsBase, search+"%")

		var count int
		err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM public.bugs `+where, args...).Scan(&count)
		if err != nil {
			return 0, err
		}
		if count > 0 {
			return count, nil
		}
	}

	// ---- 2) description count fallback
	where := baseWhere
	args := append([]any{}, argsBase...)

	if search != "" {
		where += `
 AND title ILIKE '% -- %' || $` + strconv.Itoa(argPosBase) + ` || '%'`
		args = append(args, search)
	}

	var total int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM public.bugs `+where, args...).Scan(&total)
	return total, err
}
