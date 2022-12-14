package sdfs

import (
	"fmt"
	"mp4/api"

	"sort"
)

const LATEST_VERSION = 1 // flag of the latest version of the file

type FileVersion struct {
	ConcatName string
	Seq        *api.Sequence
	Id         *api.WriteId
}

// check if two version has the same write id
func (v FileVersion) HasSameId(fileVersion FileVersion) bool {
	return v.Id.Ip == fileVersion.Id.Ip && v.Id.Port == fileVersion.Id.Port && v.Id.CreateTime.AsTime() == fileVersion.Id.CreateTime.AsTime()
}

type FileVersions []FileVersion

func (m FileVersions) Len() int {
	return len(m)
}

func (m FileVersions) Less(i, j int) bool {
	return m[i].Seq.Less(m[j].Seq)
}

func (m FileVersions) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

type FileTable map[string]FileVersions

func NewFileTable() *FileTable {
	return &FileTable{}
}

func (ft FileTable) Get(filename string, version int) (FileVersion, bool) {
	// check if the file exists or version is valid
	if !ft.Contains(filename) || len(ft[filename]) < version {
		return FileVersion{}, false
	}

	sort.Sort(ft[filename])
	return ft[filename][len(ft[filename])-version], true
}

func (ft FileTable) Insert(filename string, fileVersion FileVersion) error {
	if !ft.Contains(filename) {
		ft[filename] = make(FileVersions, 0)
	}

	// check if the write id is already in the file table
	for _, v := range ft[filename] {
		if v.HasSameId(fileVersion) {
			return fmt.Errorf("write id %v already exists in the file table", fileVersion.Id)
		}
		if v.Seq.Equal(fileVersion.Seq) {
			return fmt.Errorf("sequence %v already exists in the file table", fileVersion.Seq)
		}
	}

	// no duplicate seq/write id, insert the new version
	ft[filename] = append(ft[filename], fileVersion)
	return nil
}

func (ft FileTable) Contains(filename string) bool {
	_, ok := ft[filename]
	return ok
}

func (ft FileTable) Delete(filename string) {
	delete(ft, filename)
}

func (ft FileTable) NumVersions(filename string) int {
	return ft[filename].Len()
}

func (ft FileTable) GetVersions(filename string) FileVersions {
	if !ft.Contains(filename) {
		return FileVersions{}
	}
	return ft[filename]
}

func (ft FileTable) GetLatestVersion(filename string) (FileVersion, bool) {
	return ft.Get(filename, LATEST_VERSION)
}

func (ft FileTable) GetStoredFiles() []string {
	filenames := make([]string, 0)

	for file := range ft {
		filenames = append(filenames, file)
	}

	return filenames
}
