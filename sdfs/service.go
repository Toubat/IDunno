package sdfs

import (
	"context"
	"fmt"
	"mp4/api"
	"mp4/logger"
	"mp4/utils"

	"strconv"
)

func (server *SDFSServer) Read(ctx context.Context, req *api.ReadRequest) (*api.ReadResponse, error) {
	server.Lock()
	// logger.Get(req.GetFilename(), int(req.GetVersion()))

	fv, ok := server.FileTable.Get(req.GetFilename(), int(req.GetVersion()))
	if !ok {
		// version not found
		server.Unlock()
		logger.Info("Trying to read a non-existing version of file " + req.GetFilename() + " with version " + strconv.Itoa(int(req.GetVersion())))
		return &api.ReadResponse{Status: api.ResponseStatus_ERROR}, fmt.Errorf("version not found")
	}

	if data, ok := server.FileCache.Get(utils.DataKey(fv.ConcatName)); ok {
		// cache hit
		server.Unlock()
		// logger.Info("File " + req.GetFilename() + " with version " + strconv.Itoa(int(req.GetVersion())) + " found in cache")
		return &api.ReadResponse{Status: api.ResponseStatus_OK, Data: data, Seq: fv.Seq}, nil
	}
	server.Unlock()

	// read from local file system if cache miss
	data, err := server.ReadSDFSFile(fv.ConcatName)
	if err != nil {
		logger.Error("Failed to read file " + fv.ConcatName + ": " + err.Error())
		return &api.ReadResponse{Status: api.ResponseStatus_ERROR}, err
	}

	// logger.Info("Read " + strconv.Itoa(len(data)) + " bytes from SDFS")
	return &api.ReadResponse{Status: api.ResponseStatus_OK, Data: data, Seq: fv.Seq}, nil
}

func (server *SDFSServer) Write(ctx context.Context, req *api.WriteRequest) (*api.WriteResponse, error) {
	logger.Put(req.GetFilename())

	server.Lock()
	concatFileName := utils.ConcatFilename(req.GetFilename(), req.GetSeq())

	// insert file into virtual file table
	err := server.FileTable.Insert(req.GetFilename(), FileVersion{
		ConcatName: concatFileName,
		Seq:        req.GetSeq(),
		Id:         req.GetWriteId(),
	})

	// duplicated write/seq id, ignore and return immediately
	if err != nil {
		server.Unlock()
		logger.Error(fmt.Sprintf("Duplicated write/seq id %v", req.GetWriteId()))
		return &api.WriteResponse{Status: api.ResponseStatus_OK}, nil
	}

	// insert file data into cache if file size is smaller than 10 MB
	if len(req.GetData()) <= 10*utils.MegaByte {
		server.FileCache.Put(utils.DataKey(concatFileName), req.GetData())
	}
	server.Unlock()

	// write to local file system
	err = server.WriteSDFSFile(concatFileName, req.GetData())
	if err != nil {
		logger.Error("Fail to write file " + req.GetFilename() + ": " + err.Error())
		return &api.WriteResponse{Status: api.ResponseStatus_ERROR}, err
	}

	// logger.Info("Written " + strconv.Itoa(len(req.GetData())) + " bytes into SDFS")
	// logger.Info(fmt.Sprintf("Current latest version: %v", server.FileTable.NumVersions(req.GetFilename())))
	return &api.WriteResponse{Status: api.ResponseStatus_OK}, nil
}

func (server *SDFSServer) Delete(ctx context.Context, req *api.DeleteRequest) (*api.DeleteResponse, error) {
	logger.Remove(req.GetFilename())

	server.Lock()
	// check if file exists in file table
	if !server.FileTable.Contains(req.GetFilename()) {
		logger.Error("Trying to delete a non-existing file " + req.GetFilename())
		server.Unlock()
		return &api.DeleteResponse{Status: api.ResponseStatus_OK}, nil
	}

	// add file into delete pool for soft deletion
	for _, fv := range server.FileTable.GetVersions(req.GetFilename()) {
		logger.Info(fmt.Sprintf("Adding file %v with version %v into delete pool", fv.ConcatName, fv.Seq))
		server.DeletePool.Push(fv.ConcatName)
	}

	// remove key from file table
	logger.Info("Removing file " + req.GetFilename() + " from file table")
	server.FileTable.Delete(req.GetFilename())
	server.Unlock()

	return &api.DeleteResponse{Status: api.ResponseStatus_OK}, nil
}

func (server *SDFSServer) Lookup(ctx context.Context, req *api.LookupRequest) (*api.LookupResponse, error) {
	logger.Lookup(req.GetFilename())

	server.Lock()
	defer server.Unlock()

	if server.FileTable.Contains(req.GetFilename()) {
		logger.Info("File " + req.GetFilename() + " found in file table")
		return &api.LookupResponse{Status: api.ResponseStatus_OK, Ip: server.Ring.Ip, Port: server.Ring.Port}, nil
	}

	logger.Info("File " + req.GetFilename() + " not found in file table")
	return &api.LookupResponse{Status: api.ResponseStatus_ERROR}, fmt.Errorf("file %v not found", req.GetFilename())
}

func (server *SDFSServer) BulkLookup(ctx context.Context, req *api.BulkLookupRequest) (*api.BulkLookupResponse, error) {
	missingFiles := make([]string, 0)

	server.Lock()
	for _, filename := range req.GetFilenames() {
		logger.Lookup(filename)

		if !server.FileTable.Contains(filename) {
			logger.Info("File " + filename + " not found in file table")
			missingFiles = append(missingFiles, filename)
		} else {
			logger.Info("File " + filename + " found in file table")
		}
	}
	server.Unlock()

	return &api.BulkLookupResponse{
		Ip:           server.Ring.Process.Ip,
		Port:         server.Ring.Process.Port,
		MissingFiles: missingFiles,
	}, nil
}

func (server *SDFSServer) FetchSequence(ctx context.Context, req *api.FetchSequenceRequest) (*api.FetchSequenceResponse, error) {
	server.Lock()
	defer server.Unlock()

	// unable to process routing request if ring is not converged
	stable := server.Signal.Len() == 0 && len(server.Ring.ExpirationPool) == 0
	if !stable {
		return &api.FetchSequenceResponse{Status: api.ResponseStatus_NOT_CONVERGED}, nil
	}

	// increment sequence number
	server.SeqCounter += 1
	// get sequence struct
	seq := &api.Sequence{
		Time:  server.Ring.GetJoinTime(),
		Count: int32(server.SeqCounter),
	}

	return &api.FetchSequenceResponse{Seq: seq}, nil
}
