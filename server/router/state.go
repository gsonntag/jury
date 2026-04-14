package router

import (
	"context"
	"server/database"
	"server/judging"
	"server/logging"
	"server/models"
	"sync"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

// State is the shared application state attached to every request via middleware.
type State struct {
	Db      *mongo.Database
	Clock   *models.SafeClock
	Comps   *judging.Comparisons
	Logger  *logging.Logger
	Limiter *Limiter

	// In-memory options cache — avoids a DB round-trip for every judge action.
	optsMu sync.RWMutex
	opts   *models.Options
}

func NewState(db *mongo.Database, clock *models.SafeClock, comps *judging.Comparisons, logger *logging.Logger, limiter *Limiter, opts *models.Options) *State {
	return &State{
		Db:      db,
		Clock:   clock,
		Comps:   comps,
		Logger:  logger,
		Limiter: limiter,
		opts:    opts,
	}
}

// GetCachedOptions returns the in-memory options without hitting the database.
func (s *State) GetCachedOptions() *models.Options {
	s.optsMu.RLock()
	defer s.optsMu.RUnlock()
	return s.opts
}

// SetCachedOptions replaces the in-memory options cache.
func (s *State) SetCachedOptions(o *models.Options) {
	s.optsMu.Lock()
	defer s.optsMu.Unlock()
	s.opts = o
}

// ReloadOptions re-reads options from the database and refreshes the cache.
// Call this after any operation that modifies the options document.
func (s *State) ReloadOptions(ctx context.Context) error {
	opts, err := database.GetOptions(s.Db, ctx)
	if err != nil {
		return err
	}
	s.SetCachedOptions(opts)
	return nil
}

func GetState(ctx *gin.Context) *State {
	state := ctx.MustGet("state").(*State)
	return state
}
