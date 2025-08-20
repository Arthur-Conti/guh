package timer

import (
	"fmt"
	"time"

	fl "github.com/Arthur-Conti/guh/libs/fast_logger"
)

type TimerGrade string

var (
	HorribleGrade TimerGrade = "horrible"
	OkGrade       TimerGrade = "ok"
	GoodGrade     TimerGrade = "good"
	AmazingGrade  TimerGrade = "amazing"
)

var TimerGradeMap = map[int]TimerGrade{
	3: HorribleGrade,
	2: OkGrade,
	1: GoodGrade,
	0: AmazingGrade,
}

type Timer struct {
	NowTime time.Time
}

func NewTimer() *Timer {
	return &Timer{}
}

func (t *Timer) Start() {
	t.NowTime = time.Now()
}

func (t *Timer) End() time.Duration {
	return time.Since(t.NowTime)
}

func (t *Timer) EndAndPrint(message string) {
	fmt.Println(message, time.Since(t.NowTime))
}

func (t *Timer) ClassifyTime(duration time.Duration) int {
	if duration > time.Minute {
		return 3
	} else if duration > time.Second {
		return 2
	} else if duration > time.Millisecond {
		return 1
	} else {
		return 0
	}
}

func (t *Timer) TestFunction(function func()) {
	t.Start()
	function()
	duration := t.End()
	grade := TimerGradeMap[t.ClassifyTime(duration)]
	fl.Logf("Function was classified as %v grade. It took: %v\n", grade, duration)
}
