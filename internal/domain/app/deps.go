package app

import (
	"log/slog"

	"github.com/raozhaizhu/go-estate/internal/dao/cache"
	db "github.com/raozhaizhu/go-estate/internal/dao/sqlc"
	"github.com/raozhaizhu/go-estate/internal/util"
	"github.com/raozhaizhu/go-estate/internal/worker"
	"github.com/raozhaizhu/go-estate/pkg/async"
	objectStore "github.com/raozhaizhu/go-estate/pkg/object_store"
	"github.com/raozhaizhu/go-estate/pkg/token"
)

type Deps struct {
	Config        util.Config
	Store         db.Store
	Cache         cache.Cache
	ObjectStore   objectStore.StorageService
	TokenMaker    token.Maker
	Distributor   worker.TaskDistributor
	TaskProcessor worker.TaskProcessor
	Logger        *slog.Logger
	AsyncRunner   async.AsyncRunner
}
