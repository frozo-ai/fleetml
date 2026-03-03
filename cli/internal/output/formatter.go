package output

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
)

// PrintTable prints data as a formatted table.
func PrintTable(headers []string, rows [][]string) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	fmt.Fprintln(w, strings.Join(headers, "\t"))
	fmt.Fprintln(w, strings.Repeat("-", len(strings.Join(headers, "  "))))

	for _, row := range rows {
		fmt.Fprintln(w, strings.Join(row, "\t"))
	}

	w.Flush()
}

// PrintJSON prints data as formatted JSON.
func PrintJSON(data interface{}) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(data)
}

// Success prints a success message.
func Success(msg string) {
	fmt.Printf("  %s\n", msg)
}

// Error prints an error message.
func Error(msg string) {
	fmt.Fprintf(os.Stderr, "  %s\n", msg)
}
