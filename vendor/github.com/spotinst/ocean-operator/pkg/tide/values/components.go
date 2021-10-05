// Copyright 2021 NetApp, Inc. All Rights Reserved.

package values

import (
	"context"
)

func ForOceanOperator(ctx context.Context, values string, builder Builder) (string, error) {
	return build(ctx, values, builder, new(valuesOceanOperator))
}

func ForOceanController(ctx context.Context, values string, builder Builder) (string, error) {
	return build(ctx, values, builder, new(valuesOceanController))
}
