package main

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
)

// loggingInterceptor is gRPC middleware — runs before and after every RPC call.
// This is the gRPC equivalent of HTTP middleware you wrote in Phase 1.
func loggingInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	start := time.Now()

	// Call the actual handler
	resp, err := handler(ctx, req)

	duration := time.Since(start)

	if err != nil {
		log.Printf("❌ gRPC %s | %v | %s", info.FullMethod, err, duration)
	} else {
		log.Printf("✅ gRPC %s | OK | %s", info.FullMethod, duration)
	}

	return resp, err
}
