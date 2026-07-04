package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/EdgarOrtegaRamirez/mdtable/internal/table"
	"github.com/spf13/cobra"
)

var version = "dev"

func main() {
	root := &cobra.Command{
		Use:   "mdtable",
		Short: "Markdown table toolkit — parse, sort, filter, merge, convert",
		Long: `mdtable is a command-line toolkit for working with markdown tables.
It can parse, sort, filter, merge, diff, and convert tables between
markdown, CSV, JSON, HTML, and TSV formats.`,
		Version: version,
	}

	// Read input from file or stdin
	root.PersistentFlags().StringP("file", "f", "", "Input file (default: stdin)")

	root.AddCommand(
		newSortCmd(),
		newFilterCmd(),
		newMergeCmd(),
		newDiffCmd(),
		newConvertCmd(),
		newStatsCmd(),
		newHeadCmd(),
		newTailCmd(),
		newUniqueCmd(),
		newColumnsCmd(),
	)

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func readInput(cmd *cobra.Command) (*table.Table, error) {
	file, _ := cmd.Flags().GetString("file")

	var r io.Reader
	if file != "" {
		f, err := os.Open(file)
		if err != nil {
			return nil, fmt.Errorf("opening file: %w", err)
		}
		defer f.Close()
		r = f
	} else {
		// Check if stdin has data
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) != 0 {
			return nil, fmt.Errorf("no input provided. Use -f <file> or pipe data via stdin")
		}
		r = os.Stdin
	}

	tbl, err := table.ParseReader(r)
	if err != nil {
		return nil, fmt.Errorf("parsing table: %w", err)
	}
	return tbl, nil
}

func newSortCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sort [column]",
		Short: "Sort table by column",
		Long:  "Sort table rows by a column. Numeric columns are sorted numerically.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			tbl, err := readInput(cmd)
			if err != nil {
				return err
			}

			desc, _ := cmd.Flags().GetBool("desc")
			col := args[0]

			idx := tbl.ColumnIndex(col)
			if idx < 0 {
				return fmt.Errorf("column %q not found", col)
			}

			tbl.SortByColumn(idx, desc)
			fmt.Print(tbl.Render())
			return nil
		},
	}

	cmd.Flags().BoolP("desc", "d", false, "Sort in descending order")
	return cmd
}

func newFilterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "filter [column] [pattern]",
		Short: "Filter table rows by regex pattern",
		Long:  "Keep only rows where the specified column matches the regex pattern.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			tbl, err := readInput(cmd)
			if err != nil {
				return err
			}

			col := args[0]
			pattern := args[1]

			idx := tbl.ColumnIndex(col)
			if idx < 0 {
				return fmt.Errorf("column %q not found", col)
			}

			not, _ := cmd.Flags().GetBool("not")
			var result *table.Table
			if not {
				result, err = tbl.FilterNot(idx, pattern)
			} else {
				result, err = tbl.Filter(idx, pattern)
			}
			if err != nil {
				return fmt.Errorf("filter: %w", err)
			}

			fmt.Print(result.Render())
			return nil
		},
	}

	cmd.Flags().Bool("not", false, "Invert match (exclude matching rows)")
	return cmd
}

func newMergeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "merge [other-file]",
		Short: "Merge two tables by key column",
		Long:  "Join two tables by a key column using inner, left, right, or full join.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			tbl, err := readInput(cmd)
			if err != nil {
				return err
			}

			otherFile := args[0]
			f, err := os.Open(otherFile)
			if err != nil {
				return fmt.Errorf("opening other file: %w", err)
			}
			defer f.Close()

			other, err := table.ParseReader(f)
			if err != nil {
				return fmt.Errorf("parsing other table: %w", err)
			}

			key, _ := cmd.Flags().GetString("key")
			otherKey, _ := cmd.Flags().GetString("other-key")
			if otherKey == "" {
				otherKey = key
			}

			modeStr, _ := cmd.Flags().GetString("mode")
			var mode table.MergeMode
			switch strings.ToLower(modeStr) {
			case "left":
				mode = table.MergeLeft
			case "right":
				mode = table.MergeRight
			case "full":
				mode = table.MergeFull
			default:
				mode = table.MergeInner
			}

			merged, err := table.Merge(tbl, other, key, otherKey, mode)
			if err != nil {
				return fmt.Errorf("merge: %w", err)
			}

			fmt.Print(merged.Render())
			return nil
		},
	}

	cmd.Flags().StringP("key", "k", "", "Column name to join on (required)")
	cmd.Flags().String("other-key", "", "Column name in other table (default: same as --key)")
	cmd.Flags().StringP("mode", "m", "inner", "Join mode: inner, left, right, full")
	cmd.MarkFlagRequired("key")

	return cmd
}

func newDiffCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "diff [other-file]",
		Short: "Compare two tables and show differences",
		Long:  "Compare two tables row by row and show additions, removals, and modifications.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			old, err := readInput(cmd)
			if err != nil {
				return err
			}

			otherFile := args[0]
			f, err := os.Open(otherFile)
			if err != nil {
				return fmt.Errorf("opening other file: %w", err)
			}
			defer f.Close()

			new, err := table.ParseReader(f)
			if err != nil {
				return fmt.Errorf("parsing other table: %w", err)
			}

			diffs, err := table.Diff(old, new)
			if err != nil {
				return fmt.Errorf("diff: %w", err)
			}

			fmt.Print(table.RenderDiff(diffs, old.Headers))

			stats := table.ComputeDiffStats(diffs)
			fmt.Fprintf(os.Stderr, "\n--- Summary ---\n")
			fmt.Fprintf(os.Stderr, "Added:    %d\n", stats.Added)
			fmt.Fprintf(os.Stderr, "Removed:  %d\n", stats.Removed)
			fmt.Fprintf(os.Stderr, "Modified: %d\n", stats.Modified)
			fmt.Fprintf(os.Stderr, "Equal:    %d\n", stats.Equal)

			return nil
		},
	}

	return cmd
}

func newConvertCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "convert [format]",
		Short: "Convert table to another format",
		Long:  "Convert a markdown table to CSV, JSON, HTML, or TSV format.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			tbl, err := readInput(cmd)
			if err != nil {
				return err
			}

			format := strings.ToLower(args[0])
			switch format {
			case "csv":
				fmt.Print(tbl.RenderCSV())
			case "json":
				fmt.Print(tbl.RenderJSON())
			case "html":
				fmt.Print(tbl.RenderHTML())
			case "tsv":
				fmt.Print(tbl.RenderTSV())
			case "markdown", "md":
				fmt.Print(tbl.Render())
			default:
				return fmt.Errorf("unsupported format %q (use csv, json, html, tsv, or markdown)", format)
			}
			return nil
		},
	}

	return cmd
}

func newStatsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Show column statistics",
		Long:  "Compute and display statistics for each column (count, min, max, mean, median, etc.).",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			tbl, err := readInput(cmd)
			if err != nil {
				return err
			}

			col, _ := cmd.Flags().GetString("column")
			if col != "" {
				idx := tbl.ColumnIndex(col)
				if idx < 0 {
					return fmt.Errorf("column %q not found", col)
				}
				stats := tbl.ColumnStats(idx)
				printStats(stats)
			} else {
				for _, s := range tbl.StatsSummary() {
					printStats(s)
					fmt.Println()
				}
			}
			return nil
		},
	}

	cmd.Flags().StringP("column", "c", "", "Show stats for a specific column")
	return cmd
}

func printStats(s table.Stats) {
	fmt.Printf("Column: %s\n", s.Column)
	if s.IsNum {
		fmt.Printf("  Type:   numeric\n")
	} else {
		fmt.Printf("  Type:   text\n")
	}
	fmt.Printf("  Count:  %d\n", s.Count)
	fmt.Printf("  Unique: %d\n", s.Unique)
	fmt.Printf("  Empty:  %d\n", s.Empty)
	if s.IsNum {
		fmt.Printf("  Min:    %s\n", s.Min)
		fmt.Printf("  Max:    %s\n", s.Max)
		fmt.Printf("  Sum:    %.4f\n", s.Sum)
		fmt.Printf("  Mean:   %.4f\n", s.Mean)
		fmt.Printf("  Median: %.4f\n", s.Median)
		fmt.Printf("  StdDev: %.4f\n", s.StdDev)
	} else {
		fmt.Printf("  Min:    %s\n", s.Min)
		fmt.Printf("  Max:    %s\n", s.Max)
	}
}

func newHeadCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "head [n]",
		Short: "Show first N rows",
		Long:  "Display the first N rows of the table.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			tbl, err := readInput(cmd)
			if err != nil {
				return err
			}

			n, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid number: %w", err)
			}

			fmt.Print(tbl.Head(n).Render())
			return nil
		},
	}
}

func newTailCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "tail [n]",
		Short: "Show last N rows",
		Long:  "Display the last N rows of the table.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			tbl, err := readInput(cmd)
			if err != nil {
				return err
			}

			n, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid number: %w", err)
			}

			fmt.Print(tbl.Tail(n).Render())
			return nil
		},
	}
}

func newUniqueCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unique [column]",
		Short: "Remove duplicate rows",
		Long:  "Remove duplicate rows from the table. Optionally deduplicate by a specific column.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			tbl, err := readInput(cmd)
			if err != nil {
				return err
			}

			if len(args) > 0 {
				idx := tbl.ColumnIndex(args[0])
				if idx < 0 {
					return fmt.Errorf("column %q not found", args[0])
				}
				fmt.Print(tbl.UniqueByColumn(idx).Render())
			} else {
				fmt.Print(tbl.Unique().Render())
			}
			return nil
		},
	}

	return cmd
}

func newColumnsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "columns",
		Short: "List table columns",
		Long:  "Display column names, types, and alignment.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			tbl, err := readInput(cmd)
			if err != nil {
				return err
			}

			for i, h := range tbl.Headers {
				colType := "text"
				if tbl.IsNumericColumn(i) {
					colType = "numeric"
				}
				align := "left"
				switch tbl.Alignment[i] {
				case table.AlignCenter:
					align = "center"
				case table.AlignRight:
					align = "right"
				}
				fmt.Printf("%-20s  %-10s  %s\n", h, colType, align)
			}
			return nil
		},
	}

	return cmd
}
