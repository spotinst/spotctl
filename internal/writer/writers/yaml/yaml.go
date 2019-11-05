package yaml

import (
	"io"

	"github.com/ghodss/yaml"
	"github.com/spotinst/spotctl/internal/writer"
)

// WriterFormat is the format of this writer.
const WriterFormat writer.Format = "yaml"

func init() {
	writer.Register(WriterFormat, factory)
}

func factory(w io.Writer) (writer.Writer, error) {
	return &Writer{w}, nil
}

type Writer struct {
	w io.Writer
}

func (x *Writer) Write(obj interface{}) error {
	out, err := yaml.Marshal(obj)
	if err != nil {
		return err
	}

	_, err = x.w.Write(out)
	return err
}
