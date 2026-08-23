package main

import (
	"context"
	"errors"
	"github.com/VanceMichael/go-base-railbond-g06/internal/auth"
	"github.com/VanceMichael/go-base-railbond-g06/internal/consignment"
	"github.com/VanceMichael/go-base-railbond-g06/internal/httpapi"
	"github.com/VanceMichael/go-base-railbond-g06/internal/storage"
	"github.com/VanceMichael/go-base-railbond-g06/internal/tenant"
	"github.com/VanceMichael/go-base-railbond-g06/internal/train"
	"github.com/VanceMichael/go-base-railbond-g06/internal/worker"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	path := os.Getenv("DATABASE_URL")
	if path == "" {
		path = "railbond.db"
	}
	st, err := storage.Open(path)
	if err != nil {
		logger.Error("open", "error", err)
		os.Exit(1)
	}
	defer st.Close()
	migration := filepath.Join("migrations", "001_initial.sql")
	if err := st.Migrate(context.Background(), migration); err != nil {
		logger.Error("migration", "error", err)
		os.Exit(1)
	}
	as := auth.Service{Store: st}
	ts := tenant.Service{Store: st}
	tr := train.Service{Store: st}
	cs := consignment.Service{Store: st}
	api := &httpapi.Server{Store: st, Auth: as, Tenant: ts, Train: tr, Consignment: cs}
	api.Ready.Store(true)
	runtime := &worker.Runtime{Store: st}
	runtime.Start(context.Background())
	defer runtime.Stop(context.Background())
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	srv := &http.Server{Addr: ":" + port, Handler: httpapi.Recover(api.Routes()), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 20 * time.Second, WriteTimeout: 20 * time.Second, IdleTimeout: 60 * time.Second}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	go func() {
		logger.Info("server listening", "addr", srv.Addr)
		if e := srv.ListenAndServe(); e != nil && !errors.Is(e, http.ErrServerClosed) {
			logger.Error("serve", "error", e)
			os.Exit(1)
		}
	}()
	<-ctx.Done()
	shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdown); err != nil {
		logger.Error("shutdown", "error", err)
	}
}
