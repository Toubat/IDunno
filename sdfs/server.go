package sdfs

import (
	"context"
	"fmt"
	"mp4/api"
	"mp4/logger"
	"mp4/ring"
	"mp4/utils"

	"os"
	"strconv"
	"sync"
	"time"

	"google.golang.org/grpc"
)

const MAX_RETRY = 5

type SignalEvent struct {
	EventType ring.MemAction
	*api.Process
}

type SDFSServer struct {
	FileTable             *FileTable                 // file table
	FileCache             *utils.LFUCache            // file cache
	Ring                  *ring.RingServer           // ring server
	HashRing              *HashRing                  // ring for consistent hashing
	Signal                *utils.Queue[*SignalEvent] // signal queue
	DeletePool            *utils.Queue[string]       // delete pool
	SeqCounter            int                        // sequence counter
	sync.Mutex                                       // lock for concurrent access
	api.SDFSServiceServer                            // service interface
}

func NewSDFSServer() *SDFSServer {
	return &SDFSServer{
		FileTable:  NewFileTable(),
		FileCache:  utils.NewLFUCache(100 * utils.MegaByte),
		Signal:     utils.NewQueue[*SignalEvent](),
		DeletePool: utils.NewQueue[string](),
		HashRing:   NewHashRing(),
		SeqCounter: 0,
	}
}

func (server *SDFSServer) Cron() {
	for {
		time.Sleep(200 * time.Millisecond)

		server.Recycle()
		server.Converge()
	}
}

func (server *SDFSServer) OnMemberUpdate(eventType ring.MemAction, process *api.Process) {
	server.Lock()
	defer server.Unlock()

	server.Signal.Push(&SignalEvent{eventType, process})
}

// Read hashed SDFS file from local disk
func (server *SDFSServer) ReadSDFSFile(filename string) ([]byte, error) {
	dir := strconv.Itoa(int(server.Ring.Process.GetPort()))

	// open local file
	file, err := os.Open(dir + "/" + filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, err
	}

	data := make([]byte, info.Size())
	_, err = file.Read(data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// Write data to SDFS file
func (server *SDFSServer) WriteSDFSFile(filename string, data []byte) error {
	dir := strconv.Itoa(int(server.Ring.Process.GetPort()))

	// check if dir exist
	_, err := os.Stat(dir)
	if err != nil {
		// create dir if not exist
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}
	}

	// create local file
	file, err := os.Create(dir + "/" + filename)
	if err != nil {
		return err
	}

	// writing to local file system
	if _, err = file.Write(data); err != nil {
		return err
	}

	return nil
}

// Delete SDFS file
func (server *SDFSServer) DeleteSDFSFile(filename string) error {
	dir := strconv.Itoa(int(server.Ring.Process.GetPort()))

	// check if dir exist
	if _, err := os.Stat(dir); err != nil {
		return err
	}

	// delete local file
	if err := os.Remove(dir + "/" + filename); err != nil {
		return err
	}

	return nil
}

// Delete all files in SDFS stored in local disk
func (server *SDFSServer) ClearSDFSFiles() {
	dir := strconv.Itoa(int(server.Ring.Process.GetPort()))

	// check if dir exist
	if _, err := os.Stat(dir); err != nil {
		return
	}

	// list all files in dir
	files, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	// delete all files in dir
	for _, file := range files {
		os.Remove(dir + "/" + file.Name())
	}
}

func (server *SDFSServer) Recycle() {
	if !server.DeletePool.Empty() {
		server.Lock()
		fileName := server.DeletePool.Top()
		logger.Info(fmt.Sprintf("Deleting file %s...", fileName))
		server.DeletePool.Pop()
		server.Unlock()

		if err := server.DeleteSDFSFile(fileName); err != nil {
			logger.Error(fmt.Sprintf("Failed to delete file %s: %v", fileName, err))
		}
	}

	stable := server.Signal.Len() == 0 && len(server.Ring.ExpirationPool) == 0
	if !stable {
		logger.Info("SDFS is not stable yet, stop file table scan...")
		return
	}

	server.Lock()
	defer server.Unlock()

	// refresh again just to make sure that convergence is ran after recycling
	server.HashRing.Refresh(server.Ring.MembershipList)
	for _, file := range server.FileTable.GetStoredFiles() {
		shouldDelete := true
		replicas := server.HashRing.FindReplicas(file, REPLICA_COUNT)
		for _, replica := range replicas {
			if api.IsSameProcess(replica, server.Ring.Process) {
				shouldDelete = false
				break
			}
		}

		if shouldDelete {
			for _, fv := range server.FileTable.GetVersions(file) {
				logger.Info(fmt.Sprintf("Adding file %v with version %v into delete pool", fv.ConcatName, fv.Seq))
				server.DeletePool.Push(fv.ConcatName)
			}

			// remove key from file table
			logger.Info("Removing file " + file + " from file table")
			server.FileTable.Delete(file)
		}
	}
}

func (server *SDFSServer) Converge() {
	// important to not adding lock here, otherwise it will pose a deadlock
	if !server.Signal.Empty() && server.Ring.MembershipList.Len() > 0 {
		if len(server.Ring.ExpirationPool) != 0 {
			logger.Info("Waiting for expiration pool to be empty...")
			return
		}

		logger.Info("Converging...")
		wg := sync.WaitGroup{}

		server.Lock()
		prevMainFiles := make([]string, 0)
		for _, file := range server.FileTable.GetStoredFiles() {
			if api.IsSameProcess(server.HashRing.GetRouteProcess(file), server.Ring.Process) {
				prevMainFiles = append(prevMainFiles, file)
			}
		}

		// important to refresh hash ring after getting prevMainFiles
		server.HashRing.Refresh(server.Ring.MembershipList)

		currMainFile := make([]string, 0)
		for _, file := range server.FileTable.GetStoredFiles() {
			if api.IsSameProcess(server.HashRing.GetRouteProcess(file), server.Ring.Process) {
				currMainFile = append(currMainFile, file)
			}
		}

		// check 1 predecessor
		predecessor := server.HashRing.FindPredecessor(server.Ring.Process)

		// check 3 successors
		successors := server.HashRing.FindSuccessors(server.Ring.Process, 3)
		server.Unlock()

		if predecessor != nil {
			logger.Info(fmt.Sprintf("Transfering files to predecessor %s...", predecessor.Address()))
			wg.Add(1)

			// file that is actually been routed to predecessor
			routedFiles := make([]string, 0)
			for _, file := range prevMainFiles {
				if api.IsSameProcess(server.HashRing.GetRouteProcess(file), predecessor) {
					routedFiles = append(routedFiles, file)
				}
			}
			go func(p *api.Process, files []string) {
				defer wg.Done()
				server.TransferFiles(p, files)
			}(predecessor, routedFiles)
		}

		logger.Info(fmt.Sprintf("Transfering files to %d successors...", len(successors)))
		for _, successor := range successors {
			wg.Add(1)

			go func(p *api.Process, files []string) {
				defer wg.Done()
				server.TransferFiles(p, files)
			}(successor, currMainFile)
		}

		wg.Wait()
		server.Lock()
		server.Signal.Clear()
		server.Unlock()

		logger.Info(server.HashRing.DebugValue())
		logger.Info("Server converged!")
	}
}

func (server *SDFSServer) TransferFiles(p *api.Process, files []string) {
	// dial to replica
	conn, err := grpc.Dial(p.Address(), GRPC_OPTIONS...)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to dial to %v while trying to converge: %v", p.Address(), err))
		return
	}
	defer conn.Close()

	// create client
	client := api.NewSDFSServiceClient(conn)

	res, err := client.BulkLookup(context.Background(), &api.BulkLookupRequest{
		Filenames: files,
	})
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to send request to %v while trying to converge: %v", p.Address(), err))
		return
	}

	missingFiles := res.GetMissingFiles()
	if len(missingFiles) == 0 {
		logger.Info(fmt.Sprintf("No missing files for %s", p.Address()))
		return
	}

	numTransfered := 0
	for _, file := range missingFiles {
		versions := server.FileTable.GetVersions(file)
		if len(versions) == 0 {
			logger.Error(fmt.Sprintf("Expected to find version for file %s, but found none", file))
			continue
		}

		idx := 0
		for idx < len(versions) {
			version := versions[idx]
			idx += 1

			data, err := server.ReadSDFSFile(version.ConcatName)
			if err != nil {
				logger.Error(fmt.Sprintf("Failed to read file %s while trying to converge: %v", version.ConcatName, err))
				continue
			}

			req := &api.WriteRequest{
				Filename: file,
				Data:     data,
				WriteId:  version.Id,
				Seq:      version.Seq,
			}

			retry := 0
			for retry < MAX_RETRY {
				_, err := client.Write(context.Background(), req)
				if err != nil {
					logger.Error(fmt.Sprintf("Failed to send request to %v while trying to converge: %v", p.Address(), err))
					retry++
				}
				numTransfered++
				break
			}

			if retry == MAX_RETRY {
				logger.Error(fmt.Sprintf("Failed to transfer file %v to %v while trying to converge: %v", file, p.Address(), err))
			} else {
				logger.Write(len(data))
			}
		}
	}

	logger.Info(fmt.Sprintf("Successfully transferred %d files to %v", numTransfered, p.Address()))
}
