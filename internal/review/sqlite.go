package review

import (
	"context"
	"sync"

	"github.com/saphan-ai/proven/internal/querytpl"
	"github.com/saphan-ai/proven/internal/store"
)

// Opening is delayed until Review has checked the database path against its target.
type lazySQLite struct {
	path string
	once sync.Once
	db   *store.SQLStore
	err  error
}

func (l *lazySQLite) open() { l.once.Do(func() { l.db, l.err = store.OpenSQLite(l.path) }) }
func (l *lazySQLite) Register(ctx context.Context, table store.Table) error {
	l.open()
	if l.err != nil {
		return l.err
	}
	return l.db.Register(ctx, table)
}
func (l *lazySQLite) Run(ctx context.Context, req querytpl.Request, rel querytpl.Relations) (store.Evidence, error) {
	l.open()
	if l.err != nil {
		return store.Evidence{}, l.err
	}
	return l.db.Run(ctx, req, rel)
}
