package server

import (
	"context"
	"fmt"
	"time"
	"webtool/pkg/logger"

	"github.com/valyala/fasthttp"
	"golang.org/x/sync/errgroup"
)

func echoStart(ctx context.Context, s *Server) error {
	srv := &fasthttp.Server{
		Name: fmt.Sprintf("%s-%d", "echo", time.Now().Unix()),
		//Logger:  logger.StdLoggerWithFrameCount(3),
		Handler: echoHandle,
	}

	wg, ctx := errgroup.WithContext(ctx)
	wg.Go(func() error {
		logger.Debug().
			Str("type", s.Type).
			Str("addr", s.Addr).
			Msg("strart")
		if err := srv.ListenAndServe(s.Addr); err != nil {
			logger.Error().Msg(err.Error())
			return err
		}
		return nil
	})

	wg.Go(func() error {
		<-ctx.Done()
		if err := srv.Shutdown(); err != nil {
			fmt.Printf("Server:%s shutdown failed: %v\n", s.Addr, err)
			return fmt.Errorf("server:%s shutdown failed: %v", s.Addr, err)
		}
		fmt.Println("Server Stopped")
		return nil
	})
	return wg.Wait()
}

func echoHandle(ctx *fasthttp.RequestCtx) {
	fmt.Fprintf(ctx, "echo!\n\n")

	fmt.Fprintf(ctx, "Request method is %q\n", ctx.Method())
	fmt.Fprintf(ctx, "RequestURI is %q\n", ctx.RequestURI())
	fmt.Fprintf(ctx, "Requested path is %q\n", ctx.Path())
	fmt.Fprintf(ctx, "Host is %q\n", ctx.Host())
	fmt.Fprintf(ctx, "Query string is %q\n", ctx.QueryArgs())
	fmt.Fprintf(ctx, "User-Agent is %q\n", ctx.UserAgent())
	fmt.Fprintf(ctx, "Connection has been established at %s\n", ctx.ConnTime())
	fmt.Fprintf(ctx, "Request has been started at %s\n", ctx.Time())
	fmt.Fprintf(ctx, "Serial request number for the current connection is %d\n", ctx.ConnRequestNum())
	fmt.Fprintf(ctx, "Your ip is %q\n\n", ctx.RemoteIP())

	fmt.Fprintf(ctx, "Raw request is:\n---CUT---\n%s\n---CUT---", &ctx.Request)

	ctx.SetContentType("text/plain; charset=utf8")

	// Set arbitrary headers
	ctx.Response.Header.Set("X-My-Header", "my-header-value")
	// ctx.Request.re

	// Set cookies
	var c fasthttp.Cookie
	c.SetKey("cookie-name")
	c.SetValue("cookie-value")
	ctx.Response.Header.SetCookie(&c)
	ctx.SetStatusCode(200)
}
