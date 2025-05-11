package agent

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/dimakirio/calculatorv1/internal/models"
	"github.com/dimakirio/calculatorv1/pkg/config"
	"github.com/dimakirio/calculatorv1/pkg/logger"
)

type Agent struct {
	log *logger.Logger
	cfg *config.Config
}

func NewAgent(log *logger.Logger, cfg *config.Config) *Agent {
	return &Agent{log: log, cfg: cfg}
}

func (a *Agent) Start() {
	for i := 0; i < a.cfg.ComputingPower; i++ {
		go a.worker()
	}
}

func (a *Agent) worker() {
	for {
		task := a.getTask()
		if task != nil {
			result := a.calculate(task)
			a.sendResult(task.ID, result)
		}
		time.Sleep(time.Second)
	}
}

func (a *Agent) getTask() *models.Task {
	resp, err := http.Get("http://localhost:8080/internal/task")
	if err != nil {
		a.log.Error("Failed to get task: " + err.Error())
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	var task models.Task
	if err := json.NewDecoder(resp.Body).Decode(&task); err != nil {
		a.log.Error("Failed to decode task: " + err.Error())
		return nil
	}

	return &task
}

func (a *Agent) calculate(task *models.Task) float64 {
	// Логика вычисления задачи
	switch task.Operation {
	case "+":
		return task.Arg1 + task.Arg2
	case "-":
		return task.Arg1 - task.Arg2
	case "*":
		return task.Arg1 * task.Arg2
	case "/":
		return task.Arg1 / task.Arg2
	default:
		return 0
	}
}

func (a *Agent) sendResult(taskID string, result float64) {
	data := map[string]interface{}{
		"id":     taskID,
		"result": result,
	}
	jsonData, _ := json.Marshal(data)

	resp, err := http.Post("http://localhost:8080/internal/task", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		a.log.Error("Failed to send result: " + err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		a.log.Error("Failed to send result, status code: " + resp.Status)
	}
}
