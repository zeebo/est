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
	Name string

	Estimate    time.Duration
	Actual      time.Duration
	Annotations []Annotation `xml:"Annotation"`
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

func (a Annotation) String() string {
	if a.EstimateDelta > 0 {
		return fmt.Sprint(a.When, " Estimate:", a.EstimateDelta)
	}
	return fmt.Sprint(a.When, " Actual:", a.ActualDelta)
}

func (t *Task) Apply(ann Annotation) {
	t.Estimate += ann.EstimateDelta
	t.Actual += ann.ActualDelta
	t.Annotations = append(t.Annotations, ann)
}

func (t Task) String() string {
	return fmt.Sprintf("%s: %s/%s (%.2f)",
		t.Name,
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
