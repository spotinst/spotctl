package json

import (
	"encoding/json"
	"io"

	"github.com/spotinst/spotctl/internal/writer"
)

// WriterFormat is the format of this writer.
const WriterFormat writer.Format = "json"

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
	out, err := json.MarshalIndent(obj, "", "    ")
	if err != nil {
		return err
	}

	_, err = x.w.Write(out)
	return err
}
