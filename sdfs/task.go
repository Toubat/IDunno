package sdfs

import "mp4/api"

type SDFSTaskType string

const (
	SDFS_PUT    SDFSTaskType = "PUT"
	SDFS_GET    SDFSTaskType = "GET"
	SDFS_DELETE SDFSTaskType = "DELETE"
	SDFS_LIST   SDFSTaskType = "LIST"
	SDFS_STORE  SDFSTaskType = "STORE"
)

type SDFSTask interface {
	GetType() SDFSTaskType
	GetSDFSFile() string
}

type SDFSPutTask struct {
	LocalFile string
	SDFSFile  string
	Data      []byte
	WriteId   *api.WriteId
	SDFSTask
}

type SDFSGetTask struct {
	LocalFile string
	SDFSFile  string
	Version   int32
}

type SDFSDeleteTask struct {
	SDFSFile string
}

type SDFSListTask struct {
	SDFSFile string
}

type SDFSStoreTask struct {
}

// GetType implementation for SDFSTask
func (t SDFSPutTask) GetType() SDFSTaskType {
	return SDFS_PUT
}

func (t SDFSGetTask) GetType() SDFSTaskType {
	return SDFS_GET
}

func (t SDFSDeleteTask) GetType() SDFSTaskType {
	return SDFS_DELETE
}

func (t SDFSListTask) GetType() SDFSTaskType {
	return SDFS_LIST
}

func (t SDFSStoreTask) GetType() SDFSTaskType {
	return SDFS_STORE
}

// GetSDFSFile implementation for SDFSTask
func (t SDFSPutTask) GetSDFSFile() string {
	return t.SDFSFile
}

func (t SDFSGetTask) GetSDFSFile() string {
	return t.SDFSFile
}

func (t SDFSListTask) GetSDFSFile() string {
	return t.SDFSFile
}

func (t SDFSDeleteTask) GetSDFSFile() string {
	return t.SDFSFile
}

func (t SDFSStoreTask) GetSDFSFile() string {
	return ""
}
