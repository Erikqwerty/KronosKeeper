package TaskRunner

import (
	"fmt"

	"github.com/Erikqwerty/KronosKeeper/internal/pkg/config"
	"github.com/robfig/cron"
)

type TaskRunner struct {
	cron *cron.Cron
}

// NewTaskRunner создает новый экземпляр TaskRunner
func NewTaskRunner() *TaskRunner {
	return &TaskRunner{
		cron: cron.New(),
	}
}

func (tr *TaskRunner) AddTask(spec string, cmd func()) error {
	err := tr.cron.AddFunc(spec, cmd)
	return err
}

func (tr *TaskRunner) ListTasks() []*cron.Entry {
	return tr.cron.Entries()
}

func (tr *TaskRunner) AddTaskS(units *[]config.BackupUnit) {

}

// Start запускает планировщик cron
func (tr *TaskRunner) Start() {
	tr.cron.Start()
}

// Stop останавливает планировщик cron
func (tr *TaskRunner) Stop() {
	tr.cron.Stop()
}

func (tr *TaskRunner) Restart() {
	tr.Stop()
	tr.cron = cron.New()

	for _, entry := range tr.cron.Entries() {
		tr.cron.Schedule(entry.Schedule, entry.Job)
	}

	tr.Start()

	fmt.Println("Планировщик cron перезапущен")
}
