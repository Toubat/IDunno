package main

import (
	"fmt"
	"mp4/api"
	"mp4/logger"
	"mp4/utils"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type Worker struct {
	JobId          string
	CurrBatchInput *api.BatchInput
	LastQueryTime  *timestamppb.Timestamp
	Process        *api.Process
}

func (w *Worker) Reset() {
	w.JobId = utils.EMPTY_STRING
	w.CurrBatchInput = nil
	w.LastQueryTime = api.CurrentTimestamp()
}

func (w *Worker) Idle() bool {
	return w.JobId == utils.EMPTY_STRING
}

type WorkerList []*Worker

// implement sort interface
func (wl WorkerList) Len() int {
	return len(wl)
}

func (wl WorkerList) Less(i, j int) bool {
	return wl[i].LastQueryTime.AsTime().After(wl[j].LastQueryTime.AsTime())
}

func (wl WorkerList) Swap(i, j int) {
	wl[i], wl[j] = wl[j], wl[i]
}

// address -> worker
type ResourceManager map[string]*Worker

func NewResourceManager() *ResourceManager {
	return &ResourceManager{}
}

func (rm *ResourceManager) GetWorker(workerAddr string) *Worker {
	return (*rm)[workerAddr]
}

func (rm ResourceManager) AddWorker(process *api.Process) {
	rm[process.Address()] = &Worker{
		JobId:          utils.EMPTY_STRING,
		CurrBatchInput: nil,
		LastQueryTime:  api.CurrentTimestamp(),
		Process:        process,
	}
}

func (rm ResourceManager) RemoveWorker(process *api.Process) *Worker {
	worker, ok := rm[process.Address()]

	if !ok {
		logger.Error(fmt.Sprintf("trying to delete worker %v, but worker not found.", process.Address()))
		return nil
	}

	logger.Info(fmt.Sprintf("Removing worker %v", process.Address()))
	delete(rm, process.Address())
	return worker
}

func (rm ResourceManager) GetIdleWorkers() []*Worker {
	idleWorkers := make([]*Worker, 0)
	for _, worker := range rm {
		if worker.Idle() {
			idleWorkers = append(idleWorkers, worker)
		}
	}
	return idleWorkers
}

func (rm ResourceManager) GetWorkersById(jobId string) []*Worker {
	workers := make([]*Worker, 0)

	for _, worker := range rm {
		if worker.JobId == jobId {
			workers = append(workers, worker)
		}
	}

	return workers
}

func (rm ResourceManager) GetWorkerCountById(jobId string) int {
	return len(rm.GetWorkersById(jobId))
}

func (rm ResourceManager) Len() int {
	return len(rm)
}
