package app

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/belito3/go-web-api/app/api"
	"github.com/belito3/go-web-api/app/config"
	"github.com/belito3/go-web-api/app/repository/impl"
	"github.com/belito3/go-web-api/app/util"
	"github.com/belito3/go-web-api/pkg/logger"
	"go.uber.org/dig"
)

type options struct {
	AppConf config.AppConfiguration
	Version string
}

// Option
type Option func(*options)

// SetConfig
func SetAppConfig(s config.AppConfiguration) Option {
	return func(o *options) {
		o.AppConf = s
	}
}

// SetVersion
func SetVersion(s string) Option {
	return func(o *options) {
		o.Version = s
	}
}

// Run server
func Run(ctx context.Context, opts ...Option) error {
	var state int32 = 1
	sc := make(chan os.Signal, 1)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be catch, so don't need add it
	// SIGHUP: (signal hang up) sent to a process when its controlling terminal is closed, such as daemons
	// SIGINT: Ctrl-C sends an INT signal ("interrupt")
	// SIGTERM: signal is sent to a proc ess to request its  termination, allows process releasing releasing resources and saving state
	// SIGKILL: sent to a process to cause it to terminate immediately (kill), can't perform any clean-up upon receiving this signal
	// SIGQUIT: when user requests that the process quit and perform a core dump
	signal.Notify(sc, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	cleanFunc, err := Init(ctx, opts...)
	if err != nil {
		return err
	}

EXIT:
	for {
		sig := <-sc
		logger.Printf(ctx, "Received a signal[%s]", sig.String())
		switch sig {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			atomic.CompareAndSwapInt32(&state, 1, 0)
			break EXIT
		case syscall.SIGHUP:
		default:
			break EXIT
		}
	}

	cleanFunc()
	logger.Printf(ctx, "Service exit")
	time.Sleep(time.Second)
	os.Exit(int(atomic.LoadInt32(&state)))
	return nil
}


// Init
func Init(ctx context.Context, opts ...Option) (func(), error) {
	var o options
	for _, opt := range opts {
		opt(&o)
	}

	config.PrintWithJSON(o.AppConf)
	logger.Printf(ctx, "Service started, running mode：%s，version number：%s，process number：%d", o.AppConf.RunMode, o.Version, os.Getpid())

	// Initialize trace_id for node that app is running
	// TODO: uuid, object, snowflake
	util.InitID(o.AppConf)

	// Init logger
	setupLogger(o.AppConf)

	container, containerCall := BuildContainer(o.AppConf)

	httpServerCleanFunc := InitHTTPServer(ctx, container, o.AppConf)

	return func() {
		httpServerCleanFunc()
		containerCall()
	}, nil
}

func BuildContainer(conf config.AppConfiguration) (*dig.Container, func()) {
	// Inject store and api with container
	container := dig.New()

	// store DB
	storeCall, err := InitStore(container, conf)
	handleError(err)

	return container, func() {
		if storeCall != nil {
			storeCall()
		}
	}
}

func InitHTTPServer(ctx context.Context, container *dig.Container, conf config.AppConfiguration) func() {
	cfg := conf.HTTP
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	server := api.NewServer(conf, container)
	srv := &http.Server{
		Addr:    addr,
		Handler: server.InitGinEngine(),
		//ReadTimeout: 5 * time.Second,
		//WriteTimeout: 10 * time.Second,
		//IdleTimeout: 15 * time.Second,
	}

	go func() {
		logger.Printf(ctx, "HTTP server is running at %s.", addr)
		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	return func() {
		//TODO: Wait for interrupt signal to gracefully shutdown the app with
		// a timeout
		ctx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(cfg.ShutdownTimeout))
		defer cancel()

		srv.SetKeepAlivesEnabled(false)
		if err := srv.Shutdown(ctx); err != nil {
			logger.Errorf(ctx, err.Error())
		}
	}
}

func InitStore(container *dig.Container, conf config.AppConfiguration) (func(), error) {
	// Init dbsql db
	cfg2 := conf.DBSQL
	sqlDB, sqlDBCall, err := impl.NewDB(&impl.Config{
		DriverName:   cfg2.DriverName,
		DSN:          cfg2.DSN(),
		MaxLifetime:  cfg2.MaxLifeTime,
		MaxIdleConns: cfg2.MaxIdleConns,
		MaxOpenConns: cfg2.MaxOpenConns})
	if err != nil {
		return nil, err
	}

	_ = container.Provide(func() *sql.DB {
		return sqlDB
	})

	// TODO: gen unique client id
	ctx := context.Background()
	clientId := util.NewID()

	logger.Infof(ctx, "client id: %v", clientId)
	_ = impl.Inject(container)

	return func() {
		sqlDBCall()
	}, err
}


func handleError(err error) {
	if err != nil {
		panic(err)
	}
}
