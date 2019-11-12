package repository

import (
	"context"
	"database/sql"
	qu "github.com/Masterminds/squirrel"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	migrate "github.com/rubenv/sql-migrate"

	"github.com/nettyrnp/ads-crawler/api/sys/entity"
)

type Config struct {
	Driver string
	DSN    string
}

type Repository interface {
	GetPortals(ctx context.Context) ([]*entity.Portal, error)
	GetPortalsExt(ctx context.Context, opts PortalsQueryOpts) ([]*entity.Portal, int, error)
	AddProvider(ctx context.Context, provider *entity.Provider) (int, error)
	DeleteProvider(ctx context.Context, portalID string) error
	GetProvidersByPortal(ctx context.Context, portalName string) ([]*entity.Provider, error)
}

type RDBMSRepository struct {
	Name string
	db   *sql.DB
	Cfg  Config
}

func (r *RDBMSRepository) GetPortals(ctx context.Context) ([]*entity.Portal, error) {
	var portals []*entity.Portal

	execErr := r.runInTx(func(tx *sql.Tx) error {
		selectPortals := qu.StatementBuilder.PlaceholderFormat(qu.Dollar).
			Select("id", "protocol", "canonical_name", "email", "phone", "cert_info", "created_at").From("portal")
		query, args, err := selectPortals.ToSql()
		if err != nil {
			return err
		}
		rows, err := tx.QueryContext(ctx, query, args...)
		if err != nil {
			return err
		}
		portals0, err := scanPortalRows(rows, 0)
		if err != nil {
			return err
		}
		portals = portals0
		return nil

	}, sql.LevelReadCommitted)

	if execErr != nil {
		return nil, execErr
	}
	return portals, nil
}

func (r *RDBMSRepository) GetPortalsExt(ctx context.Context, opts PortalsQueryOpts) ([]*entity.Portal, int, error) {
	var portals []*entity.Portal
	var total int

	execErr := r.runInTx(func(tx *sql.Tx) error {
		ordering := map[entity.PortalSortField]string{
			entity.SortByDomain:       "canonical_name",
			entity.SortByCreationDate: "created_at",
		}
		orderBy := ordering[opts.SortBy] + " ASC"
		if opts.Desc {
			orderBy = ordering[opts.SortBy] + " DESC"
		}
		selectPortals := qu.StatementBuilder.PlaceholderFormat(qu.Dollar).
			Select("id", "protocol", "canonical_name", "email", "phone", "cert_info", "created_at").From("portal")
		if !opts.From.IsZero() && !opts.To.IsZero() {
			selectPortals = selectPortals.Where(qu.And{qu.GtOrEq{"created_at": opts.From}, qu.LtOrEq{"created_at": opts.To}})
		}
		query, args, err := selectPortals.OrderBy(orderBy).Limit(opts.Limit).Offset(opts.Offset).ToSql()
		if err != nil {
			return err
		}
		rows, err := tx.QueryContext(ctx, query, args...)
		if err != nil {
			return err
		}
		portals0, err := scanPortalRows(rows, opts.Limit)
		if err != nil {
			return err
		}

		var total0 int
		selectTotal := qu.StatementBuilder.PlaceholderFormat(qu.Dollar).
			Select("COUNT(id)").
			From("portal")
		if !opts.From.IsZero() && !opts.To.IsZero() {
			selectTotal = selectTotal.Where(qu.And{qu.GtOrEq{"created_at": opts.From}, qu.LtOrEq{"created_at": opts.To}})
		}
		queryRows, args, err := selectTotal.ToSql()
		if err := tx.QueryRowContext(ctx, queryRows, args...).Scan(&total0); err != nil {
			return err
		}

		total = total0
		portals = portals0
		return nil

	}, sql.LevelReadCommitted)

	if execErr != nil {
		return nil, 0, execErr
	}
	return portals, total, nil
}

func (r *RDBMSRepository) GetProvidersByPortal(ctx context.Context, portalName string) ([]*entity.Provider, error) {
	var providers []*entity.Provider

	execErr := r.runInTx(func(tx *sql.Tx) error {
		var portalID int
		selectPortalID := qu.StatementBuilder.PlaceholderFormat(qu.Dollar).
			Select("id").
			From("portal").
			Where(qu.Eq{"canonical_name": portalName}).
			Limit(1)
		queryRows, args, err := selectPortalID.ToSql()
		if err := tx.QueryRowContext(ctx, queryRows, args...).Scan(&portalID); err != nil {
			return err
		}
		if portalID == 0 {
			return errors.Errorf("portal with canonical_name '%v' not found", portalName)
		}

		psql := qu.StatementBuilder.PlaceholderFormat(qu.Dollar)
		query, args, err := psql.Select("id", "domain_name", "account_id", "account_type", "cert_auth_id", "portal_id", "created_at").
			From("provider").
			Where(qu.Eq{"portal_id": portalID}).
			ToSql()
		rows, err := tx.QueryContext(ctx, query, args...)
		if err != nil {
			return err
		}
		providers0, err := scanProviderRows(rows, 0)
		if err != nil {
			return err
		}

		providers = providers0
		return nil

	}, sql.LevelSerializable)

	if execErr != nil {
		return nil, execErr
	}
	return providers, nil
}

func (r *RDBMSRepository) AddProvider(ctx context.Context, provider *entity.Provider) (int, error) {
	var id int

	execErr := r.runInTx(func(tx *sql.Tx) error {
		// Insert
		psql := qu.StatementBuilder.PlaceholderFormat(qu.Dollar)
		query, args, err := psql.Insert("provider").Columns("domain_name", "account_id", "account_type", "cert_auth_id", "portal_id", "created_at", "updated_at").
			Values(provider.DomainName, provider.AccountID, provider.AccountType, provider.CertAuthID, provider.PortalID, provider.CreatedAt, provider.CreatedAt).
			ToSql()
		if err != nil {
			return err
		}
		if _, err = tx.ExecContext(ctx, query, args...); err != nil {
			return err
		}

		// Get last insertedID
		var id0 int
		selectMax := qu.StatementBuilder.PlaceholderFormat(qu.Dollar).
			Select("MAX(id)").
			From("provider")
		queryRows, args, err := selectMax.ToSql()
		if err := tx.QueryRowContext(ctx, queryRows, args...).Scan(&id0); err != nil {
			return err
		}

		id = id0
		return nil

	}, sql.LevelSerializable)

	if execErr != nil {
		return 0, execErr
	}
	return id, nil
}

func (r *RDBMSRepository) DeleteProvider(ctx context.Context, portalName string) error {
	return r.runInTx(func(tx *sql.Tx) error {
		var portalID int
		selectPortalID := qu.StatementBuilder.PlaceholderFormat(qu.Dollar).
			Select("id").
			From("portal").
			Where(qu.Eq{"canonical_name": portalName}).
			Limit(1)
		queryRows, args, err := selectPortalID.ToSql()
		if err := tx.QueryRowContext(ctx, queryRows, args...).Scan(&portalID); err != nil {
			return err
		}
		if portalID == 0 {
			return errors.Errorf("portal with canonical_name '%v' not found", portalName)
		}

		deleteSql := qu.StatementBuilder.PlaceholderFormat(qu.Dollar)
		deleteQuery, args, err := deleteSql.Delete("provider").
			Where(qu.Eq{"portal_id": portalID}).
			ToSql()
		if err != nil {
			return err
		}
		if _, err = tx.ExecContext(ctx, deleteQuery, args...); err != nil {
			return err
		}

		return nil

	}, sql.LevelSerializable)
}

func scanPortalRows(rows *sql.Rows, limit uint64) ([]*entity.Portal, error) {
	portals := make([]*entity.Portal, 0, limit)
	defer rows.Close()
	for rows.Next() {
		e := &entity.Portal{}
		if err := rows.Scan(&e.ID, &e.Protocol, &e.CanonicalName, &e.Email, &e.Phone, &e.CertInfo, &e.CreatedAt); err != nil {
			return nil, err
		}
		portals = append(portals, e)
	}
	return portals, nil
}

func scanProviderRows(rows *sql.Rows, limit uint64) ([]*entity.Provider, error) {
	providers := make([]*entity.Provider, 0, limit)
	defer rows.Close()
	for rows.Next() {
		e := &entity.Provider{}
		if err := rows.Scan(&e.ID, &e.DomainName, &e.AccountID, &e.AccountType, &e.CertAuthID, &e.PortalID, &e.CreatedAt); err != nil {
			return nil, err
		}
		providers = append(providers, e)
	}
	return providers, nil
}

func (r *RDBMSRepository) Init() error {
	var err error
	r.db, err = connect(r.Cfg)
	if err != nil {
		return err
	}
	return nil
}

type dbExecutor func(tx *sql.Tx) error

func (r *RDBMSRepository) runInTx(executor dbExecutor, isoLevel sql.IsolationLevel) error {
	tx, err := r.db.BeginTx(context.Background(), &sql.TxOptions{Isolation: isoLevel})
	if err != nil {
		return err
	}

	if err := executor(tx); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return errors.Wrap(err, rollbackErr.Error())
		}
		return err
	}

	return tx.Commit()
}

func (r *RDBMSRepository) MigrateUp() error {
	_, err := migrate.Exec(r.db, r.Cfg.Driver, migrations, migrate.Up)
	return err
}

func (r *RDBMSRepository) MigrateDown() error {
	_, err := migrate.Exec(r.db, r.Cfg.Driver, migrations, migrate.Down)
	return err
}

func connect(cfg Config) (*sql.DB, error) {
	db, openErr := sql.Open(cfg.Driver, cfg.DSN)
	if openErr != nil {
		return nil, openErr
	}

	if pingErr := db.Ping(); pingErr != nil {
		return nil, pingErr
	}

	return db, nil
}
