package executor

import (
	"bytes"
	"log"
	"net/http"
	"time"

	"github.com/aarthikrao/timeMachine/models/jobmodels"
	"go.uber.org/zap"
)

const defaultRequestContentType = "application/json"

type executor struct {
	logger *zap.Logger
}

func NewJobExecutor() Executor {
	devLogger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal("unable to get executor", zap.Error(err))
	}
	return &executor{

		logger: devLogger,
	}
}

func (exe *executor) Run(job jobmodels.Job) error {
	if len(job.Route) == 0 {
		exe.logger.Debug("return: no route")
		return nil
	}

	triggerTime := time.UnixMilli(job.TriggerMS)
	time.AfterFunc(time.Until(triggerTime), func() {
		exe.makeHttpRequest(job)
	})
	return nil

}

func (exe *executor) Delete(jobId string) error {
	return nil
}

func (exe *executor) makeHttpRequest(job jobmodels.Job) {
	exe.logger.Debug("request recieved to execute", zap.Any("job", job))
	_, err := http.Post(job.Route, defaultRequestContentType, bytes.NewReader(job.Meta))
	if err != nil {
		exe.logger.Error("error while posting job content",
			zap.Any("job", job),
			zap.Error(err))
		return
	}
	exe.logger.Debug("done with makeing post call", zap.Any("job", job))
}
