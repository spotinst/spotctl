package table

import (
	"fmt"
	"io"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/liggitt/tabwriter"
	"github.com/spotinst/spotctl/internal/writer"
)

// WriterFormat is the format of this writer.
const WriterFormat writer.Format = "table"

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
	table, err := convertObjectToTable(obj)
	if err != nil {
		return err
	}

	return x.printTable(table)
}

func (x *Writer) printTable(table *Table) error {
	tw := NewTabWriter(x.w)
	defer tw.Flush()

	// Header.
	{
		sort.Sort(&columnSorter{table.Columns}) // sort by priority
		cols := make([]string, len(table.Columns))

		for i, col := range table.Columns {
			cols[i] = strings.ToUpper(col.Name)
		}

		fmt.Fprintln(tw, strings.Join(cols, "\t"))
	}

	// Rows.
	{
		for _, row := range table.Rows {
			cells := make([]string, len(row.Cells))

			for i, cell := range row.Cells {
				switch v := cell.(type) {
				case time.Time:
					cells[i] = humanize.Time(v)
				default:
					cells[i] = fmt.Sprintf("%v", cell)
				}
			}

			fmt.Fprintln(tw, strings.Join(cells, "\t"))
		}
	}

	return nil
}

func convertObjectToTable(obj interface{}) (*Table, error) {
	var table Table
	var err error

	table.Columns, err = convertObjectToTableColumns(obj)
	if err != nil {
		return nil, err
	}

	table.Rows, err = convertObjectToTableRows(obj)
	if err != nil {
		return nil, err
	}

	return &table, nil
}

func convertObjectToTableColumns(obj interface{}) ([]Column, error) {
	var columns []Column

	// Convert the interface obj to a reflect.Value s.
	s := reflect.ValueOf(obj)

	// If the object is a slice, we need to handle it by extracting its first element.
	if isList(s) {
		if s.Len() == 0 {
			return columns, nil
		} else {
			s = s.Index(0)
		}

		// Check if the object is a pointer and dereference it if needed.
		if s.Kind() == reflect.Ptr {
			if s.IsNil() {
				return nil, fmt.Errorf("object cannot be null")
			}

			// Dereference.
			s = s.Elem()
		}
	}

	// Convert the value s to a reflect.Type st.
	st := s.Type()

	for i := 0; i < s.NumField(); i++ {
		sv := s.Field(i)
		sf := st.Field(i)

		isUnexported := sf.PkgPath != ""
		if sf.Anonymous {
			typ := sf.Type
			if typ.Kind() == reflect.Ptr {
				typ = typ.Elem()
			}
			if isUnexported && typ.Kind() != reflect.Struct {
				// Ignore embedded fields of unexported non-struct types.
				continue
			}

			// Allow access to unexported fields by creating an addressable copy.
			sfe := reflect.New(sf.Type).Elem()
			sfe.Set(sv)

			// Convert the embedded field.
			cols, err := convertObjectToTableColumns(sv.Interface())
			if err != nil {
				return nil, fmt.Errorf("failed to convert anonymous field %q: %v", sf.Name, err)
			}
			for _, col := range cols {
				columns = append(columns, col)
			}

			// Nothing else to do.
			continue
		} else if isUnexported {
			// Ignore unexported non-embedded fields.
			continue
		}

		tableTag := sf.Tag.Get("table")
		if tableTag == "" {
			continue
		}

		col, err := parseTableTag(tableTag)
		if err != nil {
			return nil, err
		}
		if col == nil {
			continue
		}

		columns = append(columns, *col)
	}

	return columns, nil
}

func convertObjectToTableRows(obj interface{}) ([]Row, error) {
	var rows []Row

	// Convert the interface obj to a reflect.Value s.
	s := reflect.ValueOf(obj)

	if !isList(s) {
		return nil, fmt.Errorf("object must be a list")
	}

	for i := 0; i < s.Len(); i++ {
		row, err := convertObjectToTableRow(s.Index(i).Interface())
		if err != nil {
			return nil, err
		}
		if row == nil {
			continue
		}

		// Store the keys in slice in sorted order.
		keys := make([]int, 0, len(row.cells))
		for key := range row.cells {
			keys = append(keys, int(key))
		}
		sort.Ints(keys)

		// Append all cells.
		for _, key := range keys {
			row.Cells = append(row.Cells, row.cells[int32(key)])
		}

		rows = append(rows, *row)
	}

	return rows, nil
}

func convertObjectToTableRow(obj interface{}) (*Row, error) {
	row := Row{cells: make(map[int32]interface{})}

	// Convert the interface obj to a reflect.Value s.
	s := reflect.ValueOf(obj)

	// Check if the object is a pointer and dereference it if needed.
	if s.Kind() == reflect.Ptr {
		if s.IsNil() {
			return nil, fmt.Errorf("object cannot be null")
		}

		// Dereference.
		s = s.Elem()
	}

	// Convert the value elem to a reflect.Type st.
	st := s.Type()

	for i := 0; i < s.NumField(); i++ {
		sv := s.Field(i)
		sf := st.Field(i)

		isUnexported := sf.PkgPath != ""
		if sf.Anonymous {
			typ := sf.Type
			if typ.Kind() == reflect.Ptr {
				typ = typ.Elem()
			}
			if isUnexported && typ.Kind() != reflect.Struct {
				// Ignore embedded fields of unexported non-struct types.
				continue
			}

			// Allow access to unexported fields by creating an addressable copy.
			sfe := reflect.New(sf.Type).Elem()
			sfe.Set(sv)

			// Convert the embedded field.
			r, err := convertObjectToTableRow(sv.Interface())
			if err != nil {
				return nil, fmt.Errorf("failed to convert anonymous field %q: %v", sf.Name, err)
			}
			for priority, value := range r.cells {
				row.cells[priority] = value
			}

			// Nothing else to do.
			continue
		} else if isUnexported {
			// Ignore unexported non-embedded fields.
			continue
		}

		tableTag := sf.Tag.Get("table")
		if tableTag == "" {
			continue
		}

		col, err := parseTableTag(tableTag)
		if err != nil {
			return nil, err
		}
		if col == nil {
			continue
		}

		row.cells[col.Priority] = sv.Interface()
	}

	return &row, nil
}

// isList returns true if the element is something we can Len().
func isList(v reflect.Value) bool {
	switch v.Type().Kind() {
	case reflect.Array, reflect.Slice:
		return true
	default:
		return false
	}
}

func parseTableTag(value string) (*Column, error) {
	if value == "-" {
		return nil, nil
	}

	i := strings.Index(value, ",")
	if i == -1 || value[:i] == "" {
		return nil, fmt.Errorf("malformed table tag: %s", value)
	}

	p, err := strconv.Atoi(value[:i])
	if err != nil {
		return nil, fmt.Errorf("unable to parse priority: %v", err)
	}

	return &Column{
		Name:     value[i+1:],
		Priority: int32(p),
	}, nil
}

const (
	tabwriterMinWidth = 6
	tabwriterWidth    = 4
	tabwriterPadding  = 3
	tabwriterPadChar  = ' '
	tabwriterFlags    = tabwriter.RememberWidths
)

// NewTabWriter returns a tabwriter that translates tabbed columns in input into
// properly aligned text.
func NewTabWriter(output io.Writer) *tabwriter.Writer {
	return tabwriter.NewWriter(
		output,
		tabwriterMinWidth,
		tabwriterWidth,
		tabwriterPadding,
		tabwriterPadChar,
		tabwriterFlags)
}
