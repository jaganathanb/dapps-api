package scrap_scheduler

import (
	"sync"

	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
	"github.com/jaganathanb/dapps-api/config"
	"github.com/jaganathanb/dapps-api/pkg/logging"
)

type DAppsJobScheduler struct {
	logger       logging.Logger
	cfg          *config.Config
	scheduler    gocron.Scheduler
	shutdownChan <-chan string
}

type scrapper func(userId int)

var jobScheduler *DAppsJobScheduler
var jobSchedulerOnce sync.Once

func NewDAppsJobScheduler(cfg *config.Config) *DAppsJobScheduler {
	jobSchedulerOnce.Do(func() {
		s, _ := gocron.NewScheduler()

		jobScheduler = &DAppsJobScheduler{
			logger:    logging.NewLogger(cfg),
			cfg:       cfg,
			scheduler: s,
		}

		go jobScheduler.start()
	})

	return jobScheduler
}

func (s *DAppsJobScheduler) start() {
	jobScheduler.scheduler.Start()

	// block until you are ready to shut down
	select {}
}

func (s *DAppsJobScheduler) RemoveJobs(tag string) {
	s.scheduler.RemoveByTags(tag)
}

func (s *DAppsJobScheduler) AddJob(crontab string, tag string, cb scrapper, userId int) (string, error) {
	job, err := s.scheduler.NewJob(
		gocron.CronJob(crontab, false),
		gocron.NewTask(cb, userId),
		gocron.WithTags(tag),
		gocron.WithEventListeners(gocron.BeforeJobRuns(func(jobID uuid.UUID, jobName string) {
			s.logger.Infof("Job %s started with name %s", jobID, jobName)
		})),
	)

	if err != nil {
		return "", err
	}

	return job.ID().String(), err
}
