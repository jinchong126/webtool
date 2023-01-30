package app

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"webtool/pkg/logger"
	srv "webtool/server"

	"github.com/99nil/gopkg/server"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

func NewHttpCmd() *cobra.Command {
	httpCmd := &cobra.Command{
		Use:          "http",
		Short:        "根据配置启动web server",
		Long:         `type: [echo]打印请求和响应信息; [waf]处理waf请求响应回应两包`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			// 根据启动参数初始化配置
			cfg := opt.Config()

			// 初始化日志
			logger.Init("")
			logger.Debug().Msg("webtool starting...")

			// 解析配置
			viper, err := server.ParseConfigWithEnvAlone(opt.ConfigPath, cfg, PROGRAM_NAME)
			logger.Infof("Using config file: %s", viper.ConfigFileUsed())
			if err != nil {
				if _, ok := err.(*os.PathError); !ok {
					return err
				}
				logger.Warnf("not find config file, use default.")
			}
			logger.Infof("Using config: %+v", cfg)

			eg, ctx := errgroup.WithContext(cmd.Context())

			eg.Go(func() error {
				ch := make(chan os.Signal, 1)
				signal.Notify(ch, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)
				select {
				case <-ctx.Done():
					return nil
				case sig := <-ch:
					return fmt.Errorf("server shutdown: %v", sig)
				}
			})

			logger.Infof("codes:%v", opt.Codes)

			for _, v := range cfg.Http.Server {
				serverConfig := v
				eg.Go(func() error { return srv.ServerStart(ctx, &serverConfig, &opt.Codes) })
			}
			return eg.Wait()
		},
	}
	return httpCmd
}
