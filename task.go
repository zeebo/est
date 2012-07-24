package main

import (
	"fmt"
	"time"
)

type Task struct {
	Name string

	Estimate    time.Duration
	Actual      time.Duration
	Annotations []Annotation
}

type Annotation struct {
	When          time.Time
	EstimateDelta time.Duration
	ActualDelta   time.Duration
}

func (t *Task) Apply(ann Annotation) {
	t.Estimate += ann.EstimateDelta
	t.Actual += ann.ActualDelta
	t.Annotations = append(t.Annotations, ann)
}

func (t Task) String() string {
	var ratio float64
	if t.Estimate != 0 {
		ratio = float64(t.Actual) / float64(t.Estimate)
	}

	return fmt.Sprintf("%s: %s/%s (%.2f)",
		t.Name,
		t.Actual,
		t.Estimate,
		ratio,
	)
}
