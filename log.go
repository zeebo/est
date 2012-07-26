package main

import (
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"sort"
	"text/template"
	"time"
)

func init() {
	cmd := &command{
		short: "displays info for estimates",
		long:  "afsdf",
		usage: "log [-today|-week|-lastweek] [-template=template] [-json|-xml|-cal|-cmds] [regex]",

		needsBackend: true,

		flags: flag.NewFlagSet("log", flag.ExitOnError),
		run:   log,
	}

	cmd.flags.BoolVar(&logParams.logToday, "today", false, "show estimates with changes today")
	cmd.flags.BoolVar(&logParams.logWeek, "week", false, "show estimates with changes this week")
	cmd.flags.BoolVar(&logParams.logLastWeek, "lastweek", false, "show estimates with changes last week")
	cmd.flags.StringVar(&logParams.logTemplate, "template", "", "use this template when displaying tasks")
	cmd.flags.BoolVar(&logParams.logJson, "json", false, "show estimates in json format")
	cmd.flags.BoolVar(&logParams.logXML, "xml", false, "show estimates in xml format")
	cmd.flags.BoolVar(&logParams.logCal, "cal", false, "show estimates caldav format")
	cmd.flags.BoolVar(&logParams.logCmds, "cmds", false, "show estimates in the commands to make them")

	commands["log"] = cmd
}

var logParams struct {
	logToday    bool
	logWeek     bool
	logLastWeek bool
	logTemplate string
	logJson     bool
	logXML      bool
	logCal      bool
	logCmds     bool
}

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
	case logParams.logCmds:
		logParams.logTemplate = cmdTemplate
		fallthrough
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
			fmt.Println("")
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
		err := logPrintCal(tasks)
		if err != nil {
			c.Error(err)
		}
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

var defaultLogTemplate = `{{.}}
{{range .Annotations}}{{$.Name | printf "% -20s"}}{{.}}
{{end}}`

var cmdTemplate = `est new {{.Name}}
{{range .Annotations}}est {{if .EstimateDelta}}add-est{{else}}add{{end}} {{$.Name}} -when="` + whenTemplateString + `" {{if .EstimateDelta}}{{.EstimateDelta}}{{else}}{{.ActualDelta}}{{end}}
{{end}}`

const logCalendarTemplate = `BEGIN:VEVENT
CREATED:{{.Now.Format "20060102T150405Z"}}
LAST-MODIFIED:{{.Now.Format "20060102T150405Z"}}
UID:{{.UID}}
SUMMARY:{{.Name}}
DTSTART;TZID=America/New_York:{{.Start.Format "20060102T150405"}}
DTEND;TZID=America/New_York:{{.End.Format "20060102T150405"}}
END:VEVENT
`

func logPrintCal(tasks []*Task) (err error) {
	//print the header
	fmt.Println(`BEGIN:VCALENDAR
PRODID:-//Mozilla.org/NONSGML Mozilla Calendar V1.1//EN
VERSION:2.0
BEGIN:VTIMEZONE
TZID:America/New_York
X-LIC-LOCATION:America/New_York
BEGIN:DAYLIGHT
TZOFFSETFROM:-0500
TZOFFSETTO:-0400
TZNAME:EDT
DTSTART:19700308T020000
RRULE:FREQ=YEARLY;BYDAY=2SU;BYMONTH=3
END:DAYLIGHT
BEGIN:STANDARD
TZOFFSETFROM:-0400
TZOFFSETTO:-0500
TZNAME:EST
DTSTART:19701101T020000
RRULE:FREQ=YEARLY;BYDAY=1SU;BYMONTH=11
END:STANDARD
END:VTIMEZONE`)

	//be sure the print the footer
	defer fmt.Println(`END:VCALENDAR`)

	//parse our template
	t, err := template.New("").Parse(logCalendarTemplate)
	if err != nil {
		return
	}

	//seed the random number generator
	rand.Seed(time.Now().UnixNano())

	//define our type for the template
	type calendarEntry struct {
		Now        time.Time
		Name       string
		UID        string
		Start, End time.Time
	}

	//grab our location:
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		return
	}

	//loop over nonzero annotations
	for _, task := range tasks {
		for _, anno := range task.Annotations {
			if anno.ActualDelta == 0 {
				continue
			}

			entry := calendarEntry{
				Now:   time.Now(),
				UID:   randUID(),
				Name:  fmt.Sprintf("%s (%s)", task.Name, task.Estimate),
				Start: anno.When.Add(-1 * anno.ActualDelta).In(loc),
				End:   anno.When.In(loc),
			}

			t.Execute(os.Stdout, entry)
		}
	}
	return
}

func randUID() string {
	// be6a8aa0-56d0-4b4c-bd76-a35b237e9efb
	bytes := make([]byte, 16)
	for i := range bytes {
		bytes[i] = byte(rand.Intn(256))
	}
	return fmt.Sprintf("%x-%x-%x-%x-%x", bytes[0:4], bytes[4:6], bytes[6:8], bytes[8:10], bytes[10:16])
}
