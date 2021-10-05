// Copyright 2021 NetApp, Inc. All Rights Reserved.

package values

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"gopkg.in/yaml.v3"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"
)

func Merge(values ...string) (string, error) {
	list := make([]map[string]interface{}, 0, len(values))
	for _, value := range values {
		m := make(map[string]interface{})
		r := bytes.NewReader([]byte(value))
		d := yamlutil.NewYAMLOrJSONDecoder(r, r.Len())
		if err := d.Decode(&m); err != nil {
			if err == io.EOF {
				continue
			}
			return "", fmt.Errorf("failed to unmarshal merged values: %w", err)
		}
		list = append(list, m)
	}
	b, err := yaml.Marshal(mergeMaps(list...))
	if err != nil {
		return "", fmt.Errorf("failed to marshal merged values: %w", err)
	}
	return string(b), nil
}

func mergeMaps(maps ...map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{})
	for _, m := range maps {
		for k, v := range m {
			if v, ok := v.(map[string]interface{}); ok {
				if bv, ok := out[k]; ok {
					if bv, ok := bv.(map[string]interface{}); ok {
						out[k] = mergeMaps(bv, v)
						continue
					}
				}
			}
			out[k] = v
		}
	}
	return out
}

func decode(values string, dest interface{}) error {
	if len(values) == 0 {
		return nil
	}
	r := bytes.NewReader([]byte(values))
	d := yamlutil.NewYAMLOrJSONDecoder(r, r.Len())
	return d.Decode(dest)
}

func complete(values interface{}) bool {
	type validator interface {
		Valid() bool
	}
	v, ok := values.(validator)
	if ok {
		return v.Valid()
	}
	return false
}

func buildMerge(ctx context.Context, values string, builder Builder) (string, error) {
	v, err := builder.Build(ctx)
	if err != nil {
		return "", fmt.Errorf("unable to build values: %w", err)
	}
	if len(v) > 0 {
		values, err = Merge(v, values)
		if err != nil {
			return "", fmt.Errorf("unable to merge values: %w", err)
		}
	}
	return values, nil
}

func build(ctx context.Context, values string, builder Builder, dest interface{}) (string, error) {
	err := decode(values, dest)
	if err != nil {
		return "", err
	}
	if !complete(dest) {
		values, err = buildMerge(ctx, values, builder)
		if err != nil {
			return "", err
		}
	}
	return values, nil
}
