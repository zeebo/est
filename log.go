package main

import (
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"os"
	"sort"
	"text/template"
	"time"
)

func init() {
	cmd := &command{
		short: "displays info for estimates",
		long:  "afsdf",
		usage: "log [-today|-week|-lastweek] [-template=template] [-json|-xml|-cal] [regex]",

		needsBackend: true,

		flags: flag.NewFlagSet("help", flag.ExitOnError),
		run:   log,
	}

	cmd.flags.BoolVar(&logParams.logToday, "today", false, "show estimates with changes today")
	cmd.flags.BoolVar(&logParams.logWeek, "week", false, "show estimates with changes this week")
	cmd.flags.BoolVar(&logParams.logLastWeek, "lastweek", false, "show estimates with changes last week")
	cmd.flags.StringVar(&logParams.logTemplate, "template", "", "use this template when displaying tasks")
	cmd.flags.BoolVar(&logParams.logJson, "json", false, "show estimates in json format")
	cmd.flags.BoolVar(&logParams.logXML, "xml", false, "show estimates in xml format")
	cmd.flags.BoolVar(&logParams.logCal, "cal", false, "show estimates caldav format")

	commands["log"] = cmd
}

type logParamsType struct {
	logToday    bool
	logWeek     bool
	logLastWeek bool
	logTemplate string
	logJson     bool
	logXML      bool
	logCal      bool
}

var logParams logParamsType

type minTime time.Time
type maxTime time.Time

func (m *minTime) update(t time.Time) {
	if t.Before(m.time()) {
		*m = minTime(t)
	}
}

func (m minTime) time() time.Time {
	return time.Time(m)
}

func (m *maxTime) update(t time.Time) {
	if t.After(m.time()) {
		*m = maxTime(t)
	}
}

func (m maxTime) time() time.Time {
	return time.Time(m)
}

func log(c *command) {
	args := c.flags.Args()
	if len(args) > 1 {
		c.Usage(1)
	}

	//set the regex for the search
	regex := ""
	if len(args) == 1 {
		regex = args[0]
	}

	//get the date range
	low, now := maxTime{}, time.Now()
	high := minTime(now)

	//get the time corresponding to the start of today
	startToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	if logParams.logToday {
		low.update(startToday)
		high.update(now.AddDate(0, 0, 1))
	}

	//roll the days back until we hit monday for the start of the week
	startWeek := startToday
	for startWeek.Weekday() != time.Monday {
		startWeek = startWeek.AddDate(0, 0, -1)
	}

	if logParams.logWeek {
		low.update(startWeek)
		high.update(startWeek.AddDate(0, 0, 7))
	}

	//roll the days back again for last week
	startWeek = startWeek.AddDate(0, 0, -1)
	for startWeek.Weekday() != time.Monday {
		startWeek = startWeek.AddDate(0, 0, -1)
	}

	if logParams.logLastWeek {
		low.update(startWeek)
		high.update(startWeek.AddDate(0, 0, 7))
	}

	tasks, err := defaultBackend.Find(regex, low.time(), high.time())
	if err != nil {
		c.Error(err)
	}

	//sort them by which one has the most recent annotation, with the most
	//recent on top
	sort.Sort(sortedTasks(tasks))

	switch {
	default:
		if logParams.logTemplate == "" {
			logParams.logTemplate = defaultLogTemplate
		}
		t, err := template.New("").Parse(logParams.logTemplate)
		if err != nil {
			c.Error(err)
		}
		for _, task := range tasks {
			t.Execute(os.Stdout, task)
		}
	case logParams.logJson:
		b, err := json.MarshalIndent(tasks, "", "\t")
		if err != nil {
			c.Error(err)
		}
		fmt.Printf("%s\n", b)
	case logParams.logXML:
		type Tasks struct {
			Task []*Task
		}
		t := Tasks{tasks}
		b, err := xml.MarshalIndent(t, "", "\t")
		if err != nil {
			c.Error(err)
		}
		fmt.Printf("%s\n", b)
	case logParams.logCal:
		fmt.Println("not implemented yet")
	}
}

type sortedTasks []*Task

func (t sortedTasks) Len() int { return len(t) }
func (t sortedTasks) Less(i, j int) bool {
	ianno, janno := t[i].Annotations, t[j].Annotations
	switch {
	case len(ianno) == 0 && len(janno) == 0:
		return i < j //no ordering so leave them in place
	case len(ianno) == 0:
		return false //we have nothing on i, but something on j, so put j earlier
	case len(janno) == 0:
		return true //we have nothing on j, but something on i, so put i earlier
	}
	return ianno[len(ianno)-1].When.After(janno[len(janno)-1].When)
}
func (t sortedTasks) Swap(i, j int) { t[i], t[j] = t[j], t[i] }

const defaultLogTemplate = `{{.}}
{{range .Annotations}}{{$.Name | printf "% -20s"}}{{.When.Format "2006-01-02 15:04:05" | printf "% -20s"}}{{if .EstimateDelta}}Estimate: {{.EstimateDelta}}{{end}}{{if .ActualDelta}}Actual:   {{.ActualDelta}}{{end}}
{{end}}`
