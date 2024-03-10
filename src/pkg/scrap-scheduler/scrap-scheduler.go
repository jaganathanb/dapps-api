package scrap_scheduler

import (
	"fmt"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
)

func ScheduleCronJob() {
	// create a scheduler
	s, _ := gocron.NewScheduler()

	job, _ := s.NewJob(
		gocron.DailyJob(
			1,
			gocron.NewAtTimes(
				gocron.NewAtTime(10, 30, 0),
			),
		),
		gocron.NewTask(
			func(a, b string) {
				fmt.Printf("Job run with param %s, %s", a, b)
			},
			"a",
			"b",
		),
		gocron.WithEventListeners(
			gocron.AfterJobRuns(
				func(jobID uuid.UUID, jobName string) {
					fmt.Println(jobID)
				},
			),
		),
	)

	// start the scheduler
	s.Start()

	job.RunNow()

	// block until you are ready to shut down
	select {
	case <-time.After(time.Minute):
	}

	fmt.Println("Shutting down...")
	_ = s.Shutdown()
}
