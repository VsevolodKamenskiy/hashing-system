package storage

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	Pool *pgxpool.Pool
}

func New(ctx context.Context, dsn string) (*Store, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return &Store{Pool: pool}, nil
}

type HashRow struct {
	ID   int64
	Hash string
}

func (s *Store) Close() {
	if s.Pool != nil {
		s.Pool.Close()
	}
}

func (s *Store) InsertHashes(ctx context.Context, hashes []string) ([]HashRow, error) {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	rows := make([]HashRow, 0, len(hashes))
	for _, h := range hashes {
		var id int64
		if err := tx.QueryRow(ctx, `INSERT INTO hashes (hash) VALUES ($1) RETURNING id`, h).Scan(&id); err != nil {
			return nil, err
		}
		rows = append(rows, HashRow{ID: id, Hash: h})
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return rows, nil
}

func (s *Store) GetByIDs(ctx context.Context, ids []int64) ([]HashRow, error) {
	if len(ids) == 0 {
		return nil, errors.New("empty ids")
	}
	// ANY($1) работает и с массивом в pgx
	rows, err := s.Pool.Query(ctx, `SELECT id, hash FROM hashes WHERE id = ANY($1) ORDER BY id`, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []HashRow
	for rows.Next() {
		var r HashRow
		if err := rows.Scan(&r.ID, &r.Hash); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
