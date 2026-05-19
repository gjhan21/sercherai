package router

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"

	"sercherai/backend/internal/growth/handler"
	"sercherai/backend/internal/growth/repo"
	"sercherai/backend/internal/growth/service"
	"sercherai/backend/internal/platform/config"
	"sercherai/backend/internal/platform/storage"
	"sercherai/backend/internal/platform/worker"
)

func Register(r *gin.Engine) {
	cfg := config.Load()

	var growthRepo repo.GrowthRepo
	var redisClient *redis.Client
	db, err := storage.NewMySQL(cfg)
	if err != nil {
		log.Printf("mysql unavailable, fallback to in-memory repo: %v", err)
		growthRepo = repo.NewInMemoryGrowthRepo()
	} else {
		var rErr error
		redisClient, rErr = storage.NewRedis(cfg)
		if rErr != nil {
			log.Printf("redis unavailable, continue without redis cache: %v", rErr)
		}
		growthRepo = repo.NewMySQLGrowthRepo(db, redisClient, cfg)
	}

	growthSvc := service.NewGrowthService(growthRepo)
	userGrowthHandler := handler.NewUserGrowthHandler(growthSvc, cfg)
	adminHandlers := handler.NewAdminHandlers(growthSvc, cfg)
	authHandler := handler.NewAuthHandler(
		cfg.JWTSecret,
		cfg.JWTExpireSeconds,
		cfg.JWTRefreshExpireSeconds,
		cfg.LoginFailThreshold,
		cfg.LoginIPFailThreshold,
		cfg.LoginIPPhoneThreshold,
		cfg.LoginLockSeconds,
		cfg.AllowMockLogin,
		cfg.AppEnv == "dev",
		db,
		redisClient,
	)

	if db != nil {
		worker.StartAll(growthSvc)
	}

	v1 := r.Group("/api/v1")
	{
		registerAuthRoutes(v1, authHandler, &cfg, db)
		registerUserRoutes(v1, userGrowthHandler, &cfg)
		registerAdminRoutes(v1, adminHandlers, &cfg, db)
		registerPublicRoutes(r, v1, userGrowthHandler, adminHandlers)
	}
}
