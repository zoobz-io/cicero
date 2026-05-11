// Package main is the entry point for the public API binary.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"github.com/zoobz-io/aegis"
	sessionpb "github.com/zoobz-io/aegis/proto/session/v1"
	"github.com/zoobz-io/aperture"
	"github.com/zoobz-io/astql/postgres"
	"github.com/zoobz-io/capitan"
	"github.com/zoobz-io/grub"
	grubredis "github.com/zoobz-io/grub/redis"
	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/cicero/api/contracts"
	"github.com/zoobz-io/cicero/api/handlers"
	"github.com/zoobz-io/cicero/api/wire"
	"github.com/zoobz-io/cicero/config"
	"github.com/zoobz-io/cicero/events"
	"github.com/zoobz-io/cicero/internal/auth"
	intotel "github.com/zoobz-io/cicero/internal/otel"
	"github.com/zoobz-io/cicero/models"
	"github.com/zoobz-io/cicero/services"
	"github.com/zoobz-io/cicero/stores"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	log.Println("starting...")
	ctx := context.Background()

	// Initialize sum service and registry.
	svc := sum.New()
	k := sum.Start()

	// =========================================================================
	// 1. Load Configuration
	// =========================================================================

	if err := sum.Config[config.App](ctx, k, nil); err != nil {
		return fmt.Errorf("failed to load app config: %w", err)
	}
	if err := sum.Config[config.Database](ctx, k, nil); err != nil {
		return fmt.Errorf("failed to load database config: %w", err)
	}
	if err := sum.Config[config.Translator](ctx, k, nil); err != nil {
		return fmt.Errorf("failed to load translator config: %w", err)
	}
	if err := sum.Config[config.Redis](ctx, k, nil); err != nil {
		return fmt.Errorf("failed to load redis config: %w", err)
	}
	if err := sum.Config[config.Mesh](ctx, k, nil); err != nil {
		return fmt.Errorf("failed to load mesh config: %w", err)
	}

	// =========================================================================
	// 2. Connect to Infrastructure
	// =========================================================================

	dbCfg := sum.MustUse[config.Database](ctx)
	db, err := sqlx.Connect("postgres", dbCfg.DSN())
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer func() { _ = db.Close() }()
	log.Println("database connected")
	capitan.Emit(ctx, events.StartupDatabaseConnected)

	redisCfg := sum.MustUse[config.Redis](ctx)
	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisCfg.Addr,
		Password: redisCfg.Password,
		DB:       redisCfg.DB,
	})
	defer func() { _ = redisClient.Close() }()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to redis: %w", err)
	}
	log.Println("redis connected")

	// =========================================================================
	// 3. Create Stores
	// =========================================================================

	renderer := postgres.New()
	redisProvider := grubredis.New(redisClient)
	translationCache := grub.NewStore[models.Translation](redisProvider)
	allStores := stores.New(db, renderer, translationCache)

	// =========================================================================
	// 4. Create Clients and Services
	// =========================================================================

	translatorCfg := sum.MustUse[config.Translator](ctx)
	translateSvc := services.NewTranslateService(translatorCfg.Addr)
	defer func() { _ = translateSvc.Close() }()

	// =========================================================================
	// 5. Register Contracts
	// =========================================================================

	sum.Register[contracts.Sources](k, allStores.Sources)
	sum.Register[contracts.Translations](k, allStores.Translations)
	sum.Register[contracts.Translator](k, translateSvc)

	// =========================================================================
	// 6. Register Boundaries
	// =========================================================================

	if boundaryErr := models.RegisterBoundaries(k); boundaryErr != nil {
		return fmt.Errorf("failed to register model boundaries: %w", boundaryErr)
	}
	if boundaryErr := wire.RegisterBoundaries(k); boundaryErr != nil {
		return fmt.Errorf("failed to register wire boundaries: %w", boundaryErr)
	}

	// =========================================================================
	// 7. Freeze Registry
	// =========================================================================

	sum.Freeze(k)
	capitan.Emit(ctx, events.StartupServicesReady)

	// =========================================================================
	// 8. Initialize Observability (OTEL + Aperture)
	// =========================================================================

	otelEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if otelEndpoint == "" {
		otelEndpoint = "localhost:4318"
	}
	serviceName := os.Getenv("OTEL_SERVICE_NAME")
	if serviceName == "" {
		serviceName = "cicero"
	}

	otelProviders, err := intotel.New(ctx, intotel.Config{
		Endpoint:    otelEndpoint,
		ServiceName: serviceName,
	})
	if err != nil {
		return fmt.Errorf("failed to create otel providers: %w", err)
	}
	defer func() { _ = otelProviders.Shutdown(ctx) }()
	log.Println("observability initialized")
	capitan.Emit(ctx, events.StartupOTELReady)

	ap, err := aperture.New(
		capitan.Default(),
		otelProviders.Log,
		otelProviders.Metric,
		otelProviders.Trace,
	)
	if err != nil {
		return fmt.Errorf("failed to create aperture: %w", err)
	}
	defer ap.Close()
	capitan.Emit(ctx, events.StartupApertureReady)

	// =========================================================================
	// 9. Aegis Mesh Node
	// =========================================================================

	meshCfg := sum.MustUse[config.Mesh](ctx)

	keychain := aegis.NewFileKeychain(meshCfg.CertDir)
	admin, err := aegis.NewAdminFromKeychain(ctx, keychain, meshCfg.ID)
	if err != nil {
		return fmt.Errorf("failed to create mesh admin: %w", err)
	}

	node, err := aegis.NewNodeBuilder().
		WithID(meshCfg.ID).
		WithName(meshCfg.Name).
		WithAddress(meshCfg.Addr()).
		WithCertDir(meshCfg.CertDir).
		WithAdmin(admin).
		Build()
	if err != nil {
		return fmt.Errorf("failed to build mesh node: %w", err)
	}

	if err := node.StartServer(); err != nil {
		return fmt.Errorf("failed to start mesh server: %w", err)
	}
	defer func() { _ = node.Shutdown() }()
	log.Println("mesh node started")

	// =========================================================================
	// 10. Authentication via Janus
	// =========================================================================

	pool := aegis.NewServiceClientPool(node)
	sessionClient := aegis.NewServiceClient(pool, "session", "v1", sessionpb.NewSessionServiceClient)
	svc.Engine().WithAuthenticator(auth.NewAuthenticator(sessionClient))
	log.Println("authenticator configured")

	// =========================================================================
	// 11. Register Handlers and Run
	// =========================================================================

	svc.Handle(handlers.All()...)

	appCfg := sum.MustUse[config.App](ctx)
	capitan.Emit(ctx, events.StartupServerListening, events.StartupPortKey.Field(appCfg.Port))
	log.Printf("starting server on port %d...", appCfg.Port)
	return svc.Run("", appCfg.Port)
}
