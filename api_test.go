package api

import (
	"context"
	"testing"
)

func TestAPIClient(t *testing.T) {

	ctx := context.Background()

	cl, err := NewAPIClient(ctx, "brooklynintegers://")

	if err != nil {
		t.Fatalf("Failed to create new client, %v", err)
	}

	i, err := cl.NextInt(ctx)

	if err != nil {
		t.Fatalf("Failed to get next int (i), %v", err)
	}

	j, err := cl.NextInt(ctx)

	if err != nil {
		t.Fatalf("Failed to get next int (j), %v", err)
	}

	if i > j {
		t.Fatalf("First int (i) is greater than second int (j)")
	}
}
