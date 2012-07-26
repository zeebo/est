package main

import (
	"fmt"
	"time"
)

type startLog struct {
	Name string
	When time.Time
}

type Task struct {
	Name    string
	logName string

	Estimate     time.Duration
	Actual       time.Duration
	Annotations  []Annotation `xml:"Annotation"`
	matchedAnnos []Annotation
}

type Annotation struct {
	When          time.Time
	EstimateDelta time.Duration
	ActualDelta   time.Duration
}

func (a Annotation) Negate() Annotation {
	return Annotation{
		When:          a.When,
		EstimateDelta: -1 * a.EstimateDelta,
		ActualDelta:   -1 * a.ActualDelta,
	}
}

func (a Annotation) DeltaString() string {
	if a.EstimateDelta != 0 {
		return fmt.Sprintf("Estimate: %s", a.EstimateDelta)
	}
	return fmt.Sprintf("Actual: %s", a.ActualDelta)
}

func (a Annotation) WhenString() string {
	return fmt.Sprint(a.When.Local().Format(timeFormat))
}

func (a Annotation) CommandName() string {
	if a.EstimateDelta > 0 {
		return "add-est"
	}
	return "add"
}

func (a Annotation) Command() string {
	return fmt.Sprintf(`est %s -when="%s" %s`,
		a.CommandName(),
		a.WhenString(),
		a.Delta(),
	)
}

func (a Annotation) Delta() string {
	if a.EstimateDelta != 0 {
		return fmt.Sprint(a.EstimateDelta)
	}
	return fmt.Sprint(a.ActualDelta)
}

func (a Annotation) String() string {
	format := fmt.Sprintf("%% -%ds%%s", timeFormatLen)
	return fmt.Sprintf(format, a.WhenString(), a.DeltaString())
}

func (t *Task) Apply(ann Annotation) {
	t.Estimate += ann.EstimateDelta
	t.Actual += ann.ActualDelta
	t.Annotations = append(t.Annotations, ann)
}

func (t *Task) setupTemplate(width int, low, high time.Time) {
	format := fmt.Sprintf("%% -%ds", width)
	t.logName = fmt.Sprintf(format, t.Name)

	for _, a := range t.Annotations {
		if a.When.Before(high) && a.When.After(low) {
			t.matchedAnnos = append(t.matchedAnnos, a)
		}
	}
}

func (t Task) MatchedAnnotations() []Annotation {
	return t.matchedAnnos
}

func (t Task) MatchedEstimate() (x time.Duration) {
	for _, a := range t.matchedAnnos {
		x += a.EstimateDelta
	}
	return
}

func (t Task) MatchedActual() (x time.Duration) {
	for _, a := range t.matchedAnnos {
		x += a.ActualDelta
	}
	return
}

func (t Task) String() string {
	return fmt.Sprintf("%s: %s / %s (%0.2f)",
		t.Name,
		t.Actual,
		t.Estimate,
		t.Ratio(),
	)
}

func (t Task) LogName() string {
	return t.logName
}

func (t Task) Pretty() string {
	return fmt.Sprintf("\033[1m%s%s / %s (%0.2f)\033[0m",
		t.logName,
		t.Actual,
		t.Estimate,
		t.Ratio(),
	)
}

func (t Task) Ratio() (ratio float64) {
	if t.Estimate != 0 {
		ratio = float64(t.Actual) / float64(t.Estimate)
	}
	return
}
