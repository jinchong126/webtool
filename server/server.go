package server

import (
	"context"
)

func ServerStart(ctx context.Context, s *Server, codes *ResponseCodes) error {
	if s.Type == "echo" {
		return echoStart(ctx, s)
	} else {
		return wafStart(ctx, s, codes)
	}
}
