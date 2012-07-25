package main

import (
	"flag"
	"fmt"
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
		usage: "gen [-days=] [-n=] [-p=] <duration> [durations ...]",

		needsBackend: true,

		flags: flag.NewFlagSet("gen", flag.ExitOnError),
		run:   gen,
	}

	cmd.flags.StringVar(&genParams.ptiles, "p", "25,50,75,90", "comma separated list of percentiles")
	cmd.flags.IntVar(&genParams.days, "days", 60, "number of days of history to use")
	cmd.flags.IntVar(&genParams.n, "n", 100000, "number of iterations")

	commands["gen"] = cmd
}

type genParamsType struct {
	ptiles string
	days   int
	n      int
}

var genParams genParamsType

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

	//parse out the set of percentiles
	ptilestrs := strings.Split(genParams.ptiles, ",")
	ptiles := make([]float64, 0, len(ptilestrs))
	for _, s := range ptilestrs {
		f, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
		if err != nil {
			c.Error(err)
		}
		if f < 0 || f > 100 {
			c.Error(fmt.Errorf("%.2f is not in the range [0,100]", f))
		}
		ptiles = append(ptiles, f/100.0)
	}
	sort.Float64s(ptiles)

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

	//print off the percentiles
	var j int
	for i := 0; i < len(results) && j < len(ptiles); i++ {
		if int(ptiles[j]*float64(genParams.n))-1 <= i {
			pcent := ptiles[j] * 100.0
			bars := strings.Repeat("|", int(pcent/5.0))
			fmt.Printf("%7.2f [% -20s]: %s\n", pcent, bars, results[i])
			j++
		}
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
