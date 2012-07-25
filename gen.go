package main

import (
	"flag"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"
)

func init() {
	cmd := &command{
		short: "generated percentiles based on historical data",
		long:  "afsdf",
		usage: "gen [-days=] [-n=] [-p=] [-c=] <duration> [durations ...]",

		needsBackend: true,

		flags: flag.NewFlagSet("gen", flag.ExitOnError),
		run:   gen,
	}

	cmd.flags.StringVar(&genParams.ptiles, "p", "25,50,75,90", "comma separated list of percentiles")
	cmd.flags.StringVar(&genParams.confs, "c", "50,75,90,99", "comma separated list of confidence intervals")
	cmd.flags.IntVar(&genParams.days, "days", 60, "number of days of history to use")
	cmd.flags.IntVar(&genParams.n, "n", 100000, "number of iterations")

	commands["gen"] = cmd
}

type genParamsType struct {
	ptiles string
	confs  string
	days   int
	n      int
}

var genParams genParamsType

func parseFloats(in string) (iles []float64, err error) {
	ptilestrs := strings.Split(in, ",")
	iles = make([]float64, 0, len(ptilestrs))
	var f float64
	for _, s := range ptilestrs {
		f, err = strconv.ParseFloat(strings.TrimSpace(s), 64)
		if err != nil {
			return
		}
		if f < 0 || f > 100 {
			err = fmt.Errorf("%.2f is not in the range [0,100]", f)
			return
		}
		iles = append(iles, f/100.0)
	}
	sort.Float64s(iles)
	return
}

func gen(c *command) {
	args := c.flags.Args()
	if len(args) < 1 {
		c.Usage(1)
	}

	//parse out the set of durations
	durs := make([]time.Duration, 0, len(args))
	for _, arg := range args {
		d, err := time.ParseDuration(arg)
		if err != nil {
			c.Error(err)
		}
		durs = append(durs, d)
	}

	//parse out the set of percentiles and confidence intervals
	ptiles, err := parseFloats(genParams.ptiles)
	if err != nil {
		return
	}
	confs, err := parseFloats(genParams.confs)
	if err != nil {
		return
	}

	//find the history section we need
	high := time.Now()
	var low time.Time
	if genParams.days > 0 {
		low = high.AddDate(0, 0, -1*genParams.days)
	}

	//grab the tasks in the history range
	tasks, err := defaultBackend.Find("", low, high)
	if err != nil {
		c.Error(err)
	}

	//create the set of ratios
	rs := make([]float64, 0, len(tasks))
	for _, t := range tasks {
		if ratio := t.Ratio(); ratio > 0 {
			rs = append(rs, ratio)
		}
	}

	//seed the generator
	rand.Seed(time.Now().UnixNano())

	//create our result array
	results := make([]time.Duration, 0, genParams.n)
	for i := 0; i < genParams.n; i++ {
		results = append(results, generate(rs, durs))
	}
	sort.Sort(sortedDurations(results))

	//print the standard deviation of our ratios
	var sum, sumsq float64
	for _, f := range rs {
		sum, sumsq = sum+f, sumsq+(f*f)
	}
	lf := float64(len(rs))
	sigma := math.Sqrt(lf*sumsq-sum*sum) / lf
	fmt.Println("Sigma: ", sigma)

	//print off the percentiles
	if len(ptiles) > 0 {
		fmt.Println("Percentiles:")
	}
	for _, ptile := range ptiles {
		i := int(ptile*float64(genParams.n)) - 1
		pcent := ptile * 100.0
		bars := strings.Repeat("|", int(pcent/5.0))
		fmt.Printf("%7.2f [% -20s]: %s\n", pcent, bars, results[i])
	}

	if len(confs) > 0 {
		fmt.Println("Confidences:")
	}
	median := genParams.n / 2
	for _, conf := range confs {
		offset := int(conf * float64(genParams.n/2))
		pcent := conf * 100.0
		low, high := results[median-offset], results[median+offset]
		fmt.Printf("%7.2f (% -20s to % -20s) var: %s\n", pcent, low, high, high-low)
	}
}

type sortedDurations []time.Duration

func (s sortedDurations) Len() int           { return len(s) }
func (s sortedDurations) Less(i, j int) bool { return s[i] < s[j] }
func (s sortedDurations) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func generate(rs []float64, durs []time.Duration) (result time.Duration) {
	rlen := len(rs)
	for _, d := range durs {
		r := rs[rand.Intn(rlen)]
		result += time.Duration(r * float64(d))
	}
	return
}
