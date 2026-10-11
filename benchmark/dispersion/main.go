// Command dispersion writes reports/DISPERSION.md: run-to-run dispersion (each
// run's value and the min–max across runs) for the 7C main comparison, the 7C
// stress test and the v2 stress-storm session. Paper support only (freeze rule):
// it reads committed raw data and never modifies a results directory.
//
// The definitions are those of benchmark/summarize, re-implemented here so that
// summarize stays unchanged:
//   - median latency: nearest-rank p50 over successful requests (CSV);
//   - goodput: successful requests / (last end − first start over ALL requests);
//   - wasted on failed %: signer audit `allow` entries for requests the
//     coordinator failed / all `allow` entries (stressOne in stress.go).
//
// Every value is checked against the committed summary.md before anything is
// written: per-run median and goodput against "All configurations, every run",
// and the median of runs of goodput/peak and wasted % against the stress, N76
// and N77 tables. Any mismatch aborts.
//
// Usage (from benchmark/): go run ./dispersion [-out ../reports/DISPERSION.md]
package main

import (
	"bufio"
	"compress/gzip"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const (
	mainDir   = "results/20260927T131349Z-73be90f-7C"
	stressDir = "results/20260927T191742Z-bc0ff05-7C-stress"
	stormDir  = "results/20261003T174726Z-a000c9b-v2-stress-storm"
)

var mainSystems = []string{"B0", "B1", "T-5region-optimistic", "T-5region-strict", "T-sameregion-optimistic", "T-sameregion-strict"}
var mainConc = []int{1, 10, 20, 50}

type runStat struct {
	n, errs         int
	median, goodput float64
}

func f1(x float64) string {
	if math.IsNaN(x) {
		return "–"
	}
	return fmt.Sprintf("%.1f", x)
}

func f2(x float64) string {
	if math.IsNaN(x) {
		return "–"
	}
	return fmt.Sprintf("%.2f", x)
}

func pct(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return math.NaN()
	}
	k := int(math.Ceil(p/100*float64(len(sorted)))) - 1
	if k < 0 {
		k = 0
	}
	return sorted[k]
}

func median(xs []float64) float64 {
	var v []float64
	for _, x := range xs {
		if !math.IsNaN(x) {
			v = append(v, x)
		}
	}
	if len(v) == 0 {
		return math.NaN()
	}
	sort.Float64s(v)
	if len(v)%2 == 1 {
		return v[len(v)/2]
	}
	return (v[len(v)/2-1] + v[len(v)/2]) / 2
}

func minMax(xs []float64) (float64, float64) {
	lo, hi := math.Inf(1), math.Inf(-1)
	for _, x := range xs {
		if math.IsNaN(x) {
			continue
		}
		lo, hi = math.Min(lo, x), math.Max(hi, x)
	}
	if math.IsInf(lo, 1) {
		return math.NaN(), math.NaN()
	}
	return lo, hi
}

func die(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "dispersion: "+format+"\n", a...)
	os.Exit(1)
}

// runDirs returns the run<k> directories of a results directory, sorted.
func runDirs(dir string) []string {
	cand, _ := filepath.Glob(filepath.Join(dir, "run[0-9]*"))
	var runs []string
	for _, c := range cand {
		if fi, err := os.Stat(c); err == nil && fi.IsDir() {
			runs = append(runs, c)
		}
	}
	sort.Strings(runs)
	if len(runs) == 0 {
		die("no run directories under %s", dir)
	}
	return runs
}

// loadCSV computes per-label statistics of one client CSV (as summarize's load).
func loadCSV(path string) map[string]*runStat {
	f, err := os.Open(path)
	if err != nil {
		die("%v", err)
	}
	defer f.Close()
	r := csv.NewReader(f)
	hdr, err := r.Read()
	if err != nil || strings.Join(hdr, ",") != "label,seq,start_unix_ns,latency_ns,ok,error" {
		die("%s: unexpected header %v", path, hdr)
	}
	out := map[string]*runStat{}
	lat := map[string][]float64{}
	afirst, alast := map[string]int64{}, map[string]int64{}
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			die("%s: %v", path, err)
		}
		l := rec[0]
		st := out[l]
		if st == nil {
			st = &runStat{}
			out[l] = st
		}
		st.n++
		start, _ := strconv.ParseInt(rec[2], 10, 64)
		ln, _ := strconv.ParseInt(rec[3], 10, 64)
		if afirst[l] == 0 || start < afirst[l] {
			afirst[l] = start
		}
		if e := start + ln; e > alast[l] {
			alast[l] = e
		}
		if rec[4] != "1" {
			st.errs++
			continue
		}
		lat[l] = append(lat[l], float64(ln)/1e6)
	}
	for l, st := range out {
		xs := lat[l]
		sort.Float64s(xs)
		st.median = pct(xs, 50)
		if w := float64(alast[l]-afirst[l]) / 1e9; w > 0 {
			st.goodput = float64(len(xs)) / w
		}
	}
	return out
}

// loadDir returns label -> run name -> statistics for every CSV of every run.
func loadDir(dir string) (map[string]map[string]*runStat, []string) {
	runs := runDirs(dir)
	res := map[string]map[string]*runStat{}
	var names []string
	for _, rd := range runs {
		run := filepath.Base(rd)
		names = append(names, run)
		csvs, _ := filepath.Glob(filepath.Join(rd, "*.csv"))
		sort.Strings(csvs)
		for _, p := range csvs {
			for l, st := range loadCSV(p) {
				if res[l] == nil {
					res[l] = map[string]*runStat{}
				}
				res[l][run] = st
			}
		}
	}
	return res, names
}

func readLines(path string) []string {
	var r io.Reader
	f, err := os.Open(path)
	if err != nil {
		g, gerr := os.Open(path + ".gz")
		if gerr != nil {
			return nil
		}
		defer g.Close()
		zr, zerr := gzip.NewReader(g)
		if zerr != nil {
			return nil
		}
		defer zr.Close()
		r = zr
	} else {
		defer f.Close()
		r = f
	}
	var out []string
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 1<<20), 1<<24)
	for sc.Scan() {
		if t := strings.TrimSpace(sc.Text()); t != "" {
			out = append(out, t)
		}
	}
	return out
}

// wastedPct is summarize's stressOne reduced to the wasted-on-failed fraction:
// computed shares (signer audit `allow`) for requests the coordinator failed,
// as % of all computed shares. ok is false without audit logs or requests.
func wastedPct(base string) (float64, bool) {
	failed := map[string]bool{}
	requests := 0
	for _, l := range readLines(base + ".coord.jsonl") {
		var m map[string]any
		if json.Unmarshal([]byte(l), &m) != nil {
			continue
		}
		id, _ := m["request_id"].(string)
		switch m["msg"] {
		case "signed":
			requests++
		case "sign failed":
			requests++
			failed[id] = true
		}
	}
	computed, wasted := 0, 0
	found := false
	for i := 1; i <= 9; i++ {
		lines := readLines(fmt.Sprintf("%s.audit-signer%d.jsonl", base, i))
		if lines == nil {
			continue
		}
		found = true
		for _, l := range lines {
			var e struct {
				Request  string `json:"request_id"`
				Decision string `json:"decision"`
			}
			if json.Unmarshal([]byte(l), &e) != nil {
				continue
			}
			if e.Decision == "allow" {
				computed++
				if failed[e.Request] {
					wasted++
				}
			}
		}
	}
	if !found || requests == 0 {
		return math.NaN(), false
	}
	if computed == 0 {
		return math.NaN(), true
	}
	return 100 * float64(wasted) / float64(computed), true
}

func exists(p string) bool { _, err := os.Stat(p); return err == nil }

func isStorm(runs []string, l string) bool {
	for _, rd := range runs {
		if exists(filepath.Join(rd, l+".pods.jsonl")) || exists(filepath.Join(rd, l+".pods.jsonl.gz")) {
			return true
		}
	}
	return false
}

// summaryRows returns the cells of every markdown table row of summary.md.
func summaryRows(dir string) [][]string {
	b, err := os.ReadFile(filepath.Join(dir, "summary.md"))
	if err != nil {
		die("%v", err)
	}
	var rows [][]string
	for _, ln := range strings.Split(string(b), "\n") {
		if !strings.HasPrefix(ln, "| ") {
			continue
		}
		cs := strings.Split(strings.Trim(ln, "|"), "|")
		for i := range cs {
			cs[i] = strings.TrimSpace(cs[i])
		}
		rows = append(rows, cs)
	}
	return rows
}

var runRe = regexp.MustCompile(`^run\d+$`)

// checkEveryRun compares per-run median and goodput with summary.md's
// "All configurations, every run" table (columns: label, run, N, errors,
// error %, median, p95, p99, mean, stddev, min, max, goodput, ...).
func checkEveryRun(dir string, data map[string]map[string]*runStat) int {
	n := 0
	seen := map[string]bool{}
	for _, cs := range summaryRows(dir) {
		if len(cs) < 13 || !runRe.MatchString(cs[1]) {
			continue
		}
		st := data[cs[0]][cs[1]]
		if st == nil {
			die("%s: summary row %s %s has no CSV data", dir, cs[0], cs[1])
		}
		if got := f1(st.median); got != cs[5] {
			die("%s %s %s: median %s, summary.md has %s", dir, cs[0], cs[1], got, cs[5])
		}
		if got := f1(st.goodput); got != cs[12] {
			die("%s %s %s: goodput %s, summary.md has %s", dir, cs[0], cs[1], got, cs[12])
		}
		seen[cs[0]+"/"+cs[1]] = true
		n++
	}
	for l, byRun := range data {
		for r := range byRun {
			if !seen[l+"/"+r] {
				die("%s: %s %s missing from summary.md every-run table", dir, l, r)
			}
		}
	}
	return n
}

var singleRe = regexp.MustCompile(`^(.*)-c(\d+)$`)

// stressRow is one configuration of a stress test, one value per run.
type stressRow struct {
	variant  string
	conc     int
	gp       []float64 // goodput per run
	ratio    []float64 // goodput / that run's own peak (same variant, all c)
	wasted   []float64 // wasted on failed % per run
	sumRatio float64   // summary definition: median-of-runs goodput / max over c of it
}

// stressRows builds the stress table of a session: configurations of system
// sys with signer audit logs, excluding storm runs. Variant = the part after
// "@" ("" for the 7C stress test).
func stressRows(dir, sys string, data map[string]map[string]*runStat, runNames []string) []*stressRow {
	runs := runDirs(dir)
	var rows []*stressRow
	for l := range data {
		m := singleRe.FindStringSubmatch(l)
		if m == nil || isStorm(runs, l) {
			continue
		}
		base, v := m[1], ""
		if i := strings.LastIndex(base, "@"); i >= 0 {
			base, v = base[:i], base[i+1:]
		}
		if base != sys {
			continue
		}
		c, _ := strconv.Atoi(m[2])
		r := &stressRow{variant: v, conc: c}
		ok := false
		for i, rd := range runs {
			w, has := wastedPct(filepath.Join(rd, l))
			ok = ok || has
			r.wasted = append(r.wasted, w)
			gp := math.NaN()
			if st := data[l][runNames[i]]; st != nil {
				gp = st.goodput
			}
			r.gp = append(r.gp, gp)
		}
		if ok {
			rows = append(rows, r)
		}
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].variant != rows[j].variant {
			return variantOrder(rows[i].variant) < variantOrder(rows[j].variant)
		}
		return rows[i].conc < rows[j].conc
	})
	byVar := map[string][]*stressRow{}
	for _, r := range rows {
		byVar[r.variant] = append(byVar[r.variant], r)
	}
	for _, rs := range byVar {
		peakMed := 0.0
		peakRun := make([]float64, len(runNames))
		for _, r := range rs {
			peakMed = math.Max(peakMed, median(r.gp))
			for i, g := range r.gp {
				if !math.IsNaN(g) {
					peakRun[i] = math.Max(peakRun[i], g)
				}
			}
		}
		for _, r := range rs {
			r.sumRatio = median(r.gp) / peakMed
			for i, g := range r.gp {
				r.ratio = append(r.ratio, g/peakRun[i])
			}
		}
	}
	return rows
}

func variantOrder(v string) string {
	switch v {
	case "", "n48":
		return "0"
	case "v2":
		return "1"
	case "v2nb":
		return "2"
	}
	return "9" + v
}

// checkStress compares the median of runs of goodput/peak and wasted % with the
// summary's stress (7C) or N76 (v2 session) table. Columns: [variant,] c, runs,
// offered, goodput, goodput/peak, errors, shares/request, wasted %, ...
func checkStress(dir string, rows []*stressRow, hasVariant bool) int {
	want := map[string][]string{}
	for _, cs := range summaryRows(dir) {
		off := 0
		if hasVariant {
			off = 1
		}
		if len(cs) < 9+off || cs[len(cs)-1] == "" {
			continue
		}
		if _, err := strconv.Atoi(cs[off]); err != nil {
			continue
		}
		v := ""
		if hasVariant {
			v = cs[0]
		}
		// the stress / N76 tables have 12 / 13 columns with "/"-joined shed cells
		if !strings.Contains(strings.Join(cs, "|"), " / ") || len(cs) < 12 {
			continue
		}
		want[v+"/"+cs[off]] = []string{cs[off+4], cs[off+7]}
	}
	n := 0
	for _, r := range rows {
		w := want[r.variant+"/"+strconv.Itoa(r.conc)]
		if w == nil {
			die("%s: no summary stress row for variant %q c=%d", dir, r.variant, r.conc)
		}
		if got := f2(r.sumRatio); got != w[0] {
			die("%s %q c=%d: goodput/peak %s, summary.md has %s", dir, r.variant, r.conc, got, w[0])
		}
		if got := f1(median(r.wasted)); got != w[1] {
			die("%s %q c=%d: wasted %s, summary.md has %s", dir, r.variant, r.conc, got, w[1])
		}
		n++
	}
	if n != len(want) {
		die("%s: %d stress rows computed, summary.md has %d", dir, n, len(want))
	}
	return n
}

type stormRow struct {
	variant string
	pods    int
	wasted  []float64
}

// stormRows: wasted on failed % per run for every storm configuration of sys,
// checked against the N77 table's last column.
func stormRows(dir, sys string, data map[string]map[string]*runStat) ([]*stormRow, int) {
	runs := runDirs(dir)
	var rows []*stormRow
	for l := range data {
		m := singleRe.FindStringSubmatch(l)
		if m == nil || !isStorm(runs, l) || !strings.HasPrefix(m[1], sys+"@") {
			continue
		}
		c, _ := strconv.Atoi(m[2])
		r := &stormRow{variant: strings.TrimPrefix(m[1], sys+"@"), pods: c}
		for _, rd := range runs {
			w, _ := wastedPct(filepath.Join(rd, l))
			r.wasted = append(r.wasted, w)
		}
		rows = append(rows, r)
	}
	sort.Slice(rows, func(i, j int) bool { return variantOrder(rows[i].variant) < variantOrder(rows[j].variant) })
	want := map[string]string{}
	for _, cs := range summaryRows(dir) {
		// N77 table: variant, pods, runs, issued %, amplification a / p / g, ...
		if _, err := strconv.Atoi(cs[1]); err == nil && len(cs) == 11 && strings.Count(cs[4], " / ") == 2 {
			want[cs[0]+"/"+cs[1]] = cs[10]
		}
	}
	n := 0
	for _, r := range rows {
		k := r.variant + "/" + strconv.Itoa(r.pods)
		w, ok := want[k]
		if !ok {
			die("%s: no N77 row for %s", dir, k)
		}
		if got := f1(median(r.wasted)); got != w {
			die("%s storm %s: wasted %s, summary.md has %s", dir, k, got, w)
		}
		n++
	}
	if n != len(want) {
		die("%s: %d storm rows computed, summary.md has %d", dir, n, len(want))
	}
	return rows, n
}

func join(xs []float64, f func(float64) string) string {
	var s []string
	for _, x := range xs {
		s = append(s, f(x))
	}
	return strings.Join(s, " / ")
}

func rangeCell(xs []float64, f func(float64) string) string {
	lo, hi := minMax(xs)
	return f(lo) + " – " + f(hi)
}

// spreadPct = (max − min) / median of runs, in %.
func spreadPct(xs []float64) string {
	lo, hi := minMax(xs)
	m := median(xs)
	if math.IsNaN(lo) || m == 0 {
		return "–"
	}
	return f1(100 * (hi - lo) / m)
}

func main() {
	out := flag.String("out", "../reports/DISPERSION.md", "output markdown file")
	flag.Parse()

	mainData, mainRuns := loadDir(mainDir)
	stressData, stressRuns := loadDir(stressDir)
	stormData, stormRuns := loadDir(stormDir)
	nMain := checkEveryRun(mainDir, mainData)
	nStressRun := checkEveryRun(stressDir, stressData)
	nStormRun := checkEveryRun(stormDir, stormData)

	sRows := stressRows(stressDir, "T-sameregion-optimistic", stressData, stressRuns)
	nS := checkStress(stressDir, sRows, false)
	vRows := stressRows(stormDir, "T-sameregion-optimistic", stormData, stormRuns)
	nV := checkStress(stormDir, vRows, true)
	stRows, nSt := stormRows(stormDir, "T-sameregion-optimistic", stormData)

	runHdr := func(runs []string) string { return strings.Join(runs, " / ") }

	var b strings.Builder
	b.WriteString("# Run-to-run dispersion (7C main comparison, 7C stress test, v2 stress-storm)\n\n")
	b.WriteString("Generated by `benchmark/dispersion` (`cd benchmark && go run ./dispersion`) from the committed raw data only; nothing is typed by hand. Paper support: no system code was changed (freeze rule), and no results directory or `summary.md` is modified.\n\n")
	b.WriteString("Definitions are those of `benchmark/summarize`: **median** = nearest-rank p50 latency (ms) over the successful requests of one run; **goodput** = successful tokens / (last end − first start over all requests of the run), tokens/s; **wasted on failed %** = computed shares (signer audit `allow` entries) for requests the coordinator failed, as % of all computed shares in that run. **min – max** is across runs; **spread %** = (max − min) / median of runs. The **median of runs** columns are the values `summary.md` reports.\n\n")
	fmt.Fprintf(&b, "**Self-check before writing:** every per-run median and goodput below was recomputed from the CSVs and matched the `All configurations, every run` table of the session's `summary.md` (%d rows in 7C main, %d in 7C stress, %d in v2 stress-storm); the median of runs of goodput/peak and wasted %% matched the 7C stress table (%d rows), the v2 session's N76 table (%d rows) and its N77 storm table (%d rows). Any mismatch aborts the script.\n\n", nMain, nStressRun, nStormRun, nS, nV, nSt)

	// --- 7C main comparison
	fmt.Fprintf(&b, "## 1. 7C main comparison (`benchmark/%s`)\n\n", mainDir)
	nLo, nHi := math.MaxInt, 0
	for _, byRun := range mainData {
		for _, st := range byRun {
			nLo, nHi = min(nLo, st.n), max(nHi, st.n)
		}
	}
	nCell := strconv.Itoa(nLo)
	if nLo != nHi {
		nCell = fmt.Sprintf("%d–%d", nLo, nHi)
	}
	fmt.Fprintf(&b, "N = %s requests per run and configuration in the CSVs; runs %s.\n\n", nCell, runHdr(mainRuns))
	b.WriteString("### Median latency (ms)\n\n")
	fmt.Fprintf(&b, "| System | c | per run (%s) | min – max | median of runs | spread %% |\n|---|---:|---|---|---:|---:|\n", runHdr(mainRuns))
	perRun := func(l string, f func(*runStat) float64) []float64 {
		var xs []float64
		for _, r := range mainRuns {
			if st := mainData[l][r]; st != nil {
				xs = append(xs, f(st))
			} else {
				xs = append(xs, math.NaN())
			}
		}
		return xs
	}
	for _, s := range mainSystems {
		for _, c := range mainConc {
			l := fmt.Sprintf("%s-c%d", s, c)
			if mainData[l] == nil {
				die("7C main: no data for %s", l)
			}
			xs := perRun(l, func(st *runStat) float64 { return st.median })
			fmt.Fprintf(&b, "| %s | %d | %s | %s | %s | %s |\n", s, c, join(xs, f1), rangeCell(xs, f1), f1(median(xs)), spreadPct(xs))
		}
	}
	b.WriteString("\n### Goodput (tokens/s)\n\n")
	fmt.Fprintf(&b, "| System | c | per run (%s) | min – max | median of runs | spread %% |\n|---|---:|---|---|---:|---:|\n", runHdr(mainRuns))
	for _, s := range mainSystems {
		for _, c := range mainConc {
			l := fmt.Sprintf("%s-c%d", s, c)
			xs := perRun(l, func(st *runStat) float64 { return st.goodput })
			fmt.Fprintf(&b, "| %s | %d | %s | %s | %s | %s |\n", s, c, join(xs, f1), rangeCell(xs, f1), f1(median(xs)), spreadPct(xs))
		}
	}

	stressTable := func(rows []*stressRow, runs []string, withVariant bool) {
		cols, align := "| c | goodput per run (/s) | goodput / own-run peak, per run | min – max | goodput / peak (summary.md) | wasted on failed %, per run | min – max | median of runs |", "|---:|---|---|---|---:|---|---|---:|"
		if withVariant {
			cols, align = "| variant "+cols, "|---"+align
		}
		b.WriteString(cols + "\n" + align + "\n")
		for _, r := range rows {
			v := ""
			if withVariant {
				v = "| " + r.variant + " "
			}
			fmt.Fprintf(&b, "%s| %d | %s | %s | %s | %s | %s | %s | %s |\n", v, r.conc, join(r.gp, f1), join(r.ratio, f2), rangeCell(r.ratio, f2),
				f2(r.sumRatio), join(r.wasted, f1), rangeCell(r.wasted, f1), f1(median(r.wasted)))
		}
	}
	peakNote := "**goodput / own-run peak** divides each run's goodput by the highest goodput of that same run (same variant, over all c); **goodput / peak (summary.md)** is the summary's definition: median-of-runs goodput / the highest median-of-runs goodput over c. The two differ when the peak falls at a different c in different runs. Per-run order: %s.\n\n"

	fmt.Fprintf(&b, "\n## 2. 7C stress test (`benchmark/%s`)\n\n", stressDir)
	b.WriteString("T-sameregion-optimistic, CPU-capped signers, SIGNER_MAX_CONCURRENT=1 (a stress test, not a realistic workload). ")
	fmt.Fprintf(&b, peakNote, runHdr(stressRuns))
	stressTable(sRows, stressRuns, false)

	fmt.Fprintf(&b, "\n## 3. v2 stress-storm session (`benchmark/%s`)\n\n", stormDir)
	b.WriteString("### Stress (N76 table: variants n48, v2, v2nb)\n\n")
	fmt.Fprintf(&b, peakNote, runHdr(stormRuns))
	stressTable(vRows, stormRuns, true)
	b.WriteString("\n### Storm (N77 table): wasted on failed %\n\n")
	b.WriteString("A storm run has no goodput/peak (pods retry until issued); only wasted % is reported.\n\n")
	fmt.Fprintf(&b, "| variant | pods | wasted on failed %%, per run (%s) | min – max | median of runs |\n|---|---:|---|---|---:|\n", runHdr(stormRuns))
	for _, r := range stRows {
		fmt.Fprintf(&b, "| %s | %d | %s | %s | %s |\n", r.variant, r.pods, join(r.wasted, f1), rangeCell(r.wasted, f1), f1(median(r.wasted)))
	}

	b.WriteString("\n## Sources\n\n")
	fmt.Fprintf(&b, "- `benchmark/%s/run*/*.csv` (client), checked against `benchmark/%s/summary.md`\n", mainDir, mainDir)
	for _, d := range []string{stressDir, stormDir} {
		fmt.Fprintf(&b, "- `benchmark/%s/run*/*.csv` (client), `*.coord.jsonl.gz` (coordinator) and `*.audit-signer<i>.jsonl.gz` (signer audit), checked against `benchmark/%s/summary.md`\n", d, d)
	}
	if err := os.WriteFile(*out, []byte(b.String()), 0o644); err != nil {
		die("%v", err)
	}
	fmt.Println("wrote", *out)
}
