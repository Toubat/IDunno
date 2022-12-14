package sdfs

import "mp4/api"

type SDFSTaskResult interface {
	GetType() SDFSTaskType
	GetStatus() api.ResponseStatus
}

type SDFSPutTaskResult struct {
	Status api.ResponseStatus
}

type SDFSGetTaskResult struct {
	Status api.ResponseStatus
	Seq    *api.Sequence
	Data   []byte
}

type SDFSDeleteTaskResult struct {
	Status api.ResponseStatus
}

type SDFSListTaskResult struct {
	Status api.ResponseStatus
	Ip     string
	Port   int32
}

type SDFSListTaskResults struct {
	Status  api.ResponseStatus
	Results []SDFSListTaskResult
}

// GetType implementation for SDFSTaskResult
func (r SDFSPutTaskResult) GetType() SDFSTaskType {
	return SDFS_PUT
}

func (r SDFSGetTaskResult) GetType() SDFSTaskType {
	return SDFS_GET
}

func (r SDFSDeleteTaskResult) GetType() SDFSTaskType {
	return SDFS_DELETE
}

func (r SDFSListTaskResult) GetType() SDFSTaskType {
	return SDFS_LIST
}

func (r SDFSListTaskResults) GetType() SDFSTaskType {
	return SDFS_LIST
}

// GetStatus implementation for SDFSTaskResult
func (r SDFSPutTaskResult) GetStatus() api.ResponseStatus {
	return r.Status
}

func (r SDFSGetTaskResult) GetStatus() api.ResponseStatus {
	return r.Status
}

func (r SDFSDeleteTaskResult) GetStatus() api.ResponseStatus {
	return r.Status
}

func (r SDFSListTaskResult) GetStatus() api.ResponseStatus {
	return r.Status
}

func (r SDFSListTaskResults) GetStatus() api.ResponseStatus {
	return r.Status
}
