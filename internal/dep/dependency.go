package dep

import (
	"bytes"
	"fmt"
	"html/template"
	"net/url"
	"runtime"
)

type dependency struct {
	name               string
	upstreamBinaryName string
	version            string
	url                string
}

func (x *dependency) Name() string { return x.name }

func (x *dependency) UpstreamBinaryName() string {
	if x.upstreamBinaryName != "" {
		return x.upstreamBinaryName
	}
	return x.name
}

func (x *dependency) Version() string { return x.version }

func (x *dependency) URL() (*url.URL, error) {
	tmpl, err := template.New(x.name).Parse(x.url)
	if err != nil {
		return nil, err
	}

	variables := map[string]string{
		"version":   x.version,
		"os":        runtime.GOOS,
		"arch":      runtime.GOARCH,
		"extension": x.Extension(),
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, variables); err != nil {
		return nil, err
	}

	return url.Parse(buf.String())
}

func (x *dependency) Extension() (ext string) {
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}
	return
}

func (x *dependency) Executable() string {
	return fmt.Sprintf("%s%s-%s", x.name, x.Extension(), x.Version())
}
