package metrics

import (
	"fmt"
	"io"
	"slices"
	"text/tabwriter"

	"github.com/mvelosop/vloop/internal/intervention"
)

// AgreementRow is one row of the agreement table: the counts of the records
// in one kind or phase (or all of them, for the total row).
type AgreementRow struct {
	Repo              string `json:"repo,omitempty"`
	Name              string `json:"name"`
	N                 int    `json:"n"`
	Recommended       int    `json:"recommended"`
	Share             string `json:"share"`
	OtherOption       int    `json:"other_option"`
	Adjusted          int    `json:"adjusted"`
	Different         int    `json:"different"`
	NoOptions         int    `json:"no_options"`
	AutomatableYes    int    `json:"automatable_yes"`
	AutomatablePartly int    `json:"automatable_partly"`
}

func (r *AgreementRow) add(v intervention.Intervention) {
	r.N++
	switch v.Agreement {
	case "recommended":
		r.Recommended++
	case "other-option":
		r.OtherOption++
	case "adjusted":
		r.Adjusted++
	case "different":
		r.Different++
	case "no-options":
		r.NoOptions++
	}
	switch v.Automatable {
	case "yes":
		r.AutomatableYes++
	case "partly":
		r.AutomatablePartly++
	}
}

func (r *AgreementRow) finish() {
	r.Share = "n/a"
	if d := r.N - r.NoOptions; d > 0 {
		r.Share = fmt.Sprintf("%d%%", (r.Recommended*100+d/2)/d)
	}
}

// AgreementTable counts records by kind or phase (by), one row per value in
// its enum order, then a total row.
func AgreementTable(recs []intervention.Intervention, by string) []AgreementRow {
	names := intervention.Kinds
	if by == "phase" {
		names = intervention.Phases
	}
	rows := make([]AgreementRow, len(names)+1)
	for i, n := range names {
		rows[i].Name = n
	}
	rows[len(names)].Name = "total"
	for _, v := range recs {
		key := v.Kind
		if by == "phase" {
			key = v.Phase
		}
		if i := slices.Index(names, key); i >= 0 {
			rows[i].add(v)
		}
		rows[len(names)].add(v)
	}
	for i := range rows {
		rows[i].finish()
	}
	return rows
}

// WorkspaceAgreementTable is the kind table of each repository in turn, then
// one total row over all of them. recs[i] belongs to repos[i].
func WorkspaceAgreementTable(repos []string, recs [][]intervention.Intervention) []AgreementRow {
	var out []AgreementRow
	var all []intervention.Intervention
	for i, name := range repos {
		rows := AgreementTable(recs[i], "kind")
		for _, r := range rows[:len(rows)-1] {
			r.Repo = name
			out = append(out, r)
		}
		all = append(all, recs[i]...)
	}
	rows := AgreementTable(all, "kind")
	return append(out, rows[len(rows)-1])
}

// PrintAgreementTable writes the table; first is the first column's header
// (kind or phase), and withRepo puts a repo column before it.
func PrintAgreementTable(out io.Writer, rows []AgreementRow, first string, withRepo bool) {
	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	head := "n\trecommended\tshare\tother-option\tadjusted\tdifferent\tno-options\tautomatable-yes\tautomatable-partly"
	if withRepo {
		fmt.Fprintf(w, "repo\t%s\t%s\n", first, head)
	} else {
		fmt.Fprintf(w, "%s\t%s\n", first, head)
	}
	for _, r := range rows {
		if withRepo {
			repo := r.Repo
			if r.Name == "total" {
				repo = ""
			}
			fmt.Fprintf(w, "%s\t", repo)
		}
		fmt.Fprintf(w, "%s\t%d\t%d\t%s\t%d\t%d\t%d\t%d\t%d\t%d\n", r.Name, r.N, r.Recommended, r.Share,
			r.OtherOption, r.Adjusted, r.Different, r.NoOptions, r.AutomatableYes, r.AutomatablePartly)
	}
	w.Flush()
}
