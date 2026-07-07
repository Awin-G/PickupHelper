package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"pickup-helper/internal/config"
	"pickup-helper/internal/handler"
	"pickup-helper/internal/repository"
	"pickup-helper/internal/router"
	"pickup-helper/internal/service"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

// Container holds all long-lived application dependencies wired at startup.
type Container struct {
	Cfg *config.Config
	DB  *sqlx.DB
	RDB *redis.Client

	// Repositories (stateless, safe for concurrent use).
	userRepo   repository.UserRepo
	adminRepo  repository.AdminRepo
	runnerRepo repository.RunnerAppRepo
	smsCache   repository.SMSCodeCache
	parcelRepo repository.ParcelRepo
	shelfRepo  repository.ShelfRepo
	pickupRepo repository.PickupLogRepo
	proxyRepo  repository.ProxyOrderRepo
	notifyRepo repository.NotifyRepo

	// Services.
	authSvc    *service.AuthService
	userSvc    *service.UserService
	parcelSvc  *service.ParcelService
	pickupSvc  *service.PickupService
	proxySvc   *service.ProxyService
	shelfSvc   *service.ShelfService
	notifySvc  *service.NotifyService
	statsSvc   *service.StatsService

	// Handlers.
	healthH  *handler.HealthHandler
	authH    *handler.AuthHandler
	userH    *handler.UserHandler
	parcelH  *handler.ParcelHandler
	pickupH  *handler.PickupHandler
	proxyH   *handler.ProxyHandler
	shelfH   *handler.ShelfHandler
	notifyH  *handler.NotifyHandler
	statsH   *handler.StatsHandler
}

// BuildContainer manually wires dependencies (no DI framework).
// Callers must defer container.Close() to release resources.
func BuildContainer(cfg *config.Config) (*Container, error) {
	db, err := repository.NewMySQL(cfg)
	if err != nil {
		return nil, fmt.Errorf("mysql: %w", err)
	}
	rdb, err := repository.NewRedis(cfg)
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("redis: %w", err)
	}

	// Repositories.
	userRepo := repository.NewUserRepo()
	adminRepo := repository.NewAdminRepo()
	runnerRepo := repository.NewRunnerAppRepo()
	smsCache := repository.NewSMSCodeCache(rdb)
	parcelRepo := repository.NewParcelRepo()
	shelfRepo := repository.NewShelfRepo()
	pickupRepo := repository.NewPickupLogRepo()
	proxyRepo := repository.NewProxyOrderRepo()
	notifyRepo := repository.NewNotifyRepo()

	// Services.
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "dev"
	}
	sms := service.NewSMSProvider(env, slog.Default())
	authSvc := service.NewAuthService(userRepo, adminRepo, smsCache, sms, cfg, db)
	userSvc := service.NewUserService(userRepo, runnerRepo, db)
	parcelSvc := service.NewParcelService(parcelRepo, shelfRepo, userRepo, db)
	pickupSvc := service.NewPickupService(parcelRepo, pickupRepo, shelfRepo, userRepo, db)
	proxySvc := service.NewProxyService(proxyRepo, parcelRepo, userRepo, db)
	shelfSvc := service.NewShelfService(shelfRepo, db)
	notifySvc := service.NewNotifyService(notifyRepo, db)
	statsSvc := service.NewStatsService(db)

	// Handlers.
	healthH := handler.NewHealthHandler(db, rdb)
	authH := handler.NewAuthHandler(authSvc)
	userH := handler.NewUserHandler(userSvc)
	parcelH := handler.NewParcelHandler(parcelSvc)
	pickupH := handler.NewPickupHandler(pickupSvc)
	proxyH := handler.NewProxyHandler(proxySvc)
	shelfH := handler.NewShelfHandler(shelfSvc)
	notifyH := handler.NewNotifyHandler(notifySvc)
	statsH := handler.NewStatsHandler(statsSvc)

	return &Container{
		Cfg:        cfg,
		DB:         db,
		RDB:        rdb,
		userRepo:   userRepo,
		adminRepo:  adminRepo,
		runnerRepo: runnerRepo,
		smsCache:   smsCache,
		parcelRepo: parcelRepo,
		shelfRepo:  shelfRepo,
		pickupRepo: pickupRepo,
		proxyRepo:   proxyRepo,
		notifyRepo:  notifyRepo,
		authSvc:     authSvc,
		userSvc:     userSvc,
		parcelSvc:   parcelSvc,
		pickupSvc:   pickupSvc,
		proxySvc:    proxySvc,
		shelfSvc:    shelfSvc,
		notifySvc:   notifySvc,
		healthH:     healthH,
		authH:       authH,
		userH:       userH,
		parcelH:     parcelH,
		pickupH:     pickupH,
		proxyH:      proxyH,
		shelfH:      shelfH,
		notifyH:     notifyH,
		statsH:      statsH,
	}, nil
}

// Handlers returns the router.Handlers bundle for router.Register.
func (c *Container) Handlers() *router.Handlers {
	return &router.Handlers{
		Health: c.healthH,
		Auth:   c.authH,
		User:   c.userH,
		Parcel: c.parcelH,
		Pickup: c.pickupH,
		Proxy:  c.proxyH,
		Shelf:  c.shelfH,
		Notify: c.notifyH,
		Stats:  c.statsH,
	}
}

// Close releases all resources held by the container.
func (c *Container) Close() error {
	var firstErr error
	if c.DB != nil {
		if err := c.DB.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	if c.RDB != nil {
		if err := c.RDB.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	if firstErr != nil {
		return fmt.Errorf("container close: %w", firstErr)
	}
	return nil
}

// PingAll checks DB and Redis reachability. Used by health readiness checks.
func (c *Container) PingAll(ctx context.Context) error {
	if err := repository.PingMySQL(ctx, c.DB); err != nil {
		return fmt.Errorf("mysql ping: %w", err)
	}
	if err := repository.PingRedis(ctx, c.RDB); err != nil {
		return fmt.Errorf("redis ping: %w", err)
	}
	return nil
}
