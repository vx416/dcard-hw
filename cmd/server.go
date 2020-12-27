package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/vx416/dcard-work/internal/delivery/restful"

	"github.com/spf13/cobra"
	"github.com/vx416/dcard-work/internal/app"
	"github.com/vx416/dcard-work/internal/ring"
	"github.com/vx416/dcard-work/pkg/config"
	"github.com/vx416/dcard-work/pkg/logging"
	"go.uber.org/fx"
)

var Server = &cobra.Command{
	Run: runServer,
	Use: "server",
}

func runServer(cmd *cobra.Command, args []string) {
	log := logging.Get()
	if r := recover(); r != nil {
		var msg string
		for i := 2; ; i++ {
			_, file, line, ok := runtime.Caller(i)
			if !ok {
				break
			}
			msg = msg + fmt.Sprintf("%s:%d\n", file, line)
		}
		log.Fatalf("%s\n↧↧↧↧↧↧ PANIC ↧↧↧↧↧↧\n%s↥↥↥↥↥↥ PANIC ↥↥↥↥↥↥", r, msg)
	}

	cfg, err := config.Init()
	if err != nil {
		log.Fatalf("server: initialize config failed, err:%+v", err)
		os.Exit(1)
	}

	app := fx.New(
		cfg.ProvideInfra(),
		fx.Provide(
			ring.DefaultConfig,
			ring.New,
			func(r *ring.HashRing) app.Ringer { return r },
			app.New,
			restful.New,
		),
		fx.Invoke(restful.V1Routes),
	)

	ctx := context.Background()
	err = app.Start(ctx)
	if err != nil {
		log.Fatalf("server: start app failed, err:%+v", err)
		os.Exit(1)
	}

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigterm:
		log.Debug("server: shutdown process start")
	}

	stopCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err := app.Stop(stopCtx); err != nil {
		log.Errorf("server: shutdown process failed, err:%+v", err)
	}
}
