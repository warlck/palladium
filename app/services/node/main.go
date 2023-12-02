package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ardanlabs/conf/v3"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/warlck/palladium/app/services/node/handlers"
	"github.com/warlck/palladium/foundation/blockchain/database"
	"github.com/warlck/palladium/foundation/blockchain/genesis"
	"github.com/warlck/palladium/foundation/blockchain/state"
	"github.com/warlck/palladium/foundation/blockchain/storage/disk"
	"github.com/warlck/palladium/foundation/logger"
	"go.uber.org/zap"
)

// build is the git version of this program. It is set using build flags in the makefile.
var build = "develop"

func main() {

	// Construct the application logger.
	log, err := logger.New("NODE")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer log.Sync()

	// Perform the startup and shutdown sequence.
	if err := run(log); err != nil {
		log.Errorw("startup", "ERROR", err)
		log.Sync()
		os.Exit(1)
	}
}

func run(log *zap.SugaredLogger) error {

	// =========================================================================
	// Configuration

	// This is all the configuration for the application and the default values.
	// Configuration values will be passed through the application as individual
	// values.
	cfg := struct {
		conf.Version
		Web struct {
			ReadTimeout     time.Duration `conf:"default:5s"`
			WriteTimeout    time.Duration `conf:"default:10s"`
			IdleTimeout     time.Duration `conf:"default:120s"`
			ShutdownTimeout time.Duration `conf:"default:20s"`
			DebugHost       string        `conf:"default:0.0.0.0:7080"`
			PublicHost      string        `conf:"default:0.0.0.0:8080"`
			PrivateHost     string        `conf:"default:0.0.0.0:9080"`
		}
		State struct {
			Beneficiary string `conf:"default:testminer"`
			DBPath      string `conf:"default:zblock/testminer/"`
		}
		NameService struct {
			Folder string `conf:"default:zblock/accounts/"`
		}
	}{
		Version: conf.Version{
			Build: build,
			Desc:  "copyright information here",
		},
	}

	// Parse will set the defaults and then look for any overriding values
	// in environment variables and command line flags.
	const prefix = "NODE"
	help, err := conf.Parse(prefix, &cfg)
	if err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			fmt.Println(help)
			return nil
		}
		return fmt.Errorf("parsing config: %w", err)
	}

	// =========================================================================
	// App Starting

	fmt.Println(`    _____      _ _           _ _                  `)
	fmt.Println(`   |  __ \    | | |         | (_)                 `)
	fmt.Println(`   | |__) |_ _| | | __ _  __| |_ _   _ _ __ ___   `)
	fmt.Println("   |  ___/ _` | | |/ _` |/ _` | | | | | '_ ` _ \\ ")
	fmt.Println(`   | |  | (_| | | | (_| | (_| | | |_| | | | | | | `)
	fmt.Println(`   |_|   \__,_|_|_|\__,_|\__,_|_|\__,_|_| |_| |_| `)
	fmt.Print("\n")

	log.Infow("starting service", "version", build)
	defer log.Infow("shutdown complete")

	// Display the current configuration to the logs.
	out, err := conf.String(&cfg)
	if err != nil {
		return fmt.Errorf("generating config for output: %w", err)
	}
	log.Infow("startup", "config", out)

	// =========================================================================
	// Blockchain Support

	// Need to load the private key file for the configured beneficiary so the
	// account can get credited with fees and tips.
	path := fmt.Sprintf("%s%s.ecdsa", cfg.NameService.Folder, cfg.State.Beneficiary)
	privateKey, err := crypto.LoadECDSA(path)
	if err != nil {
		return fmt.Errorf("unable to load private key for node: %w", err)
	}

	// The blockchain packages accept a function of this signature to allow the
	// application to log.
	ev := func(v string, args ...any) {
		const websocketPrefix = "viewer:"

		s := fmt.Sprintf(v, args...)
		log.Infow(s, "traceid", "00000000-0000-0000-0000-000000000000")

	}

	genesis, err := genesis.Load()
	if err != nil {
		return fmt.Errorf("genesis load: %w", err)
	}

	// Construct the use of disk storage.
	storage, err := disk.New(cfg.State.DBPath)
	if err != nil {
		return err
	}

	// The state value represents the blockchain node and manages the blockchain
	// database and provides an API for application support.
	state, err := state.New(state.Config{
		BeneficiaryID: database.PublicKeyToAccountID(privateKey.PublicKey),
		Host:          cfg.Web.PrivateHost,
		Storage:       storage,
		Genesis:       genesis,
		EvHandler:     ev,
	})
	if err != nil {
		return err
	}

	// =========================================================================
	// Service Start/Stop Support

	// Make a channel to listen for an interrupt or terminate signal from the OS.
	// Use a buffered channel because the signal package requires it.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	// Make a channel to listen for errors coming from the listener. Use a
	// buffered channel so the goroutine can exit if we don't collect this error.
	serverErrors := make(chan error, 1)

	// =========================================================================
	// Start Public Service

	log.Infow("startup", "status", "initializing V1 public API support")

	// Construct the mux for the public API calls.
	publicMux := handlers.PublicMux(handlers.MuxConfig{
		Shutdown: shutdown,
		Log:      log,
	})

	// Construct a server to service the requests against the mux.
	public := http.Server{
		Addr:         cfg.Web.PublicHost,
		Handler:      publicMux,
		ReadTimeout:  cfg.Web.ReadTimeout,
		WriteTimeout: cfg.Web.WriteTimeout,
		IdleTimeout:  cfg.Web.IdleTimeout,
		ErrorLog:     zap.NewStdLog(log.Desugar()),
	}

	// Start the service listening for api requests.
	go func() {
		log.Infow("startup", "status", "public api router started", "host", public.Addr)
		serverErrors <- public.ListenAndServe()
	}()

	// =========================================================================
	// Start Private Service

	log.Infow("startup", "status", "initializing V1 private API support")

	// Construct the mux for the private API calls.
	privateMux := handlers.PrivateMux(handlers.MuxConfig{
		Shutdown: shutdown,
		Log:      log,
	})

	// Construct a server to service the requests against the mux.
	private := http.Server{
		Addr:         cfg.Web.PrivateHost,
		Handler:      privateMux,
		ReadTimeout:  cfg.Web.ReadTimeout,
		WriteTimeout: cfg.Web.WriteTimeout,
		IdleTimeout:  cfg.Web.IdleTimeout,
		ErrorLog:     zap.NewStdLog(log.Desugar()),
	}

	// Start the service listening for api requests.
	go func() {
		log.Infow("startup", "status", "private api router started", "host", private.Addr)
		serverErrors <- private.ListenAndServe()
	}()

	// =========================================================================
	// Shutdown

	// Blocking main and waiting for shutdown.
	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)

	case sig := <-shutdown:
		log.Infow("shutdown", "status", "shutdown started", "signal", sig)
		defer log.Infow("shutdown", "status", "shutdown complete", "signal", sig)

		// Give outstanding requests a deadline for completion.
		ctx, cancelPub := context.WithTimeout(context.Background(), cfg.Web.ShutdownTimeout)
		defer cancelPub()

		// Asking listener to shut down and shed load.
		log.Infow("shutdown", "status", "shutdown private API started")
		if err := private.Shutdown(ctx); err != nil {
			private.Close()
			return fmt.Errorf("could not stop private service gracefully: %w", err)
		}

		// Give outstanding requests a deadline for completion.
		ctx, cancelPri := context.WithTimeout(context.Background(), cfg.Web.ShutdownTimeout)
		defer cancelPri()

		// Asking listener to shut down and shed load.
		log.Infow("shutdown", "status", "shutdown public API started")
		if err := public.Shutdown(ctx); err != nil {
			public.Close()
			return fmt.Errorf("could not stop public service gracefully: %w", err)
		}
	}

	return nil
}
