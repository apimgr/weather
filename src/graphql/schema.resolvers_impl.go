package graphql

// This file contains helper functions for resolver implementations
// The actual resolver implementations are in schema.resolvers.go

import (
	"context"
	"fmt"
	"time"
)

// Helper function to get user ID from context with error handling
func getUserIDFromContextWithError(ctx context.Context) (int, error) {
	userID, ok := ctx.Value("user_id").(int)
	if !ok {
		return 0, fmt.Errorf("unauthorized: user not found in context")
	}
	return userID, nil
}

// Helper function to check if user is admin
func isAdminFromContext(ctx context.Context) bool {
	role, ok := ctx.Value("user_role").(string)
	return ok && role == "admin"
}

// Helper function to get IP from context
func getIPFromContext(ctx context.Context) string {
	ip, ok := ctx.Value("client_ip").(string)
	if !ok {
		return "unknown"
	}
	return ip
}

// Helper function to create a string pointer
func stringPtr(s string) *string {
	return &s
}

// Helper function to create an int pointer
func intPtr(i int) *int {
	return &i
}

// Helper function to create a float64 pointer
func float64Ptr(f float64) *float64 {
	return &f
}

// Helper function to create a time pointer
func timePtr(t time.Time) *time.Time {
	return &t
}
