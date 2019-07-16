package table

type (
	// Table is a tabular representation of a set of resources.
	Table struct {
		// Columns represents a collection of column names (i.e. header line).
		Columns []Column `json:"columns"`

		// Rows represents a collection of resources.
		Rows []Row `json:"rows"`
	}

	// Column represents an individual column in a table.
	Column struct {
		// Name is a human readable name for the column.
		Name string `json:"name"`

		// Priority is an integer defining the relative importance of this column
		// compared to others. Lower numbers are considered higher priority.
		Priority int32 `json:"priority"`
	}

	// Row represents an individual row in a table.
	Row struct {
		// Cells will be as wide as the column definitions array and may contain
		// strings, numbers, booleans, simple maps, lists, or null.
		Cells []interface{} `json:"cells"`

		// for internal use only (sorting, etc.)
		cells map[int32]interface{}
	}
)

// columnSorter implements the sort.Interface interface.
type columnSorter struct {
	columns []Column
}

func (s *columnSorter) Len() int           { return len(s.columns) }
func (s *columnSorter) Swap(i, j int)      { s.columns[i], s.columns[j] = s.columns[j], s.columns[i] }
func (s *columnSorter) Less(i, j int) bool { return s.columns[i].Priority < s.columns[j].Priority }
