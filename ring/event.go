package ring

import (
	"mp4/api"
	"mp4/logger"
	"net"
	"time"
)

type RingServerEvent interface {
	OnPing(remoteAddr *net.UDPAddr, processes []*api.Process)
	OnJoin(remoteAddr *net.UDPAddr, process *api.Process)
	OnAck(process *api.Process)
	OnLeave(process *api.Process)
	OnFailure(process *api.Process)
}

/*
 * Update membership list with processes in ping message
 *
 * @param address: address of process that sent ping message
 * @param processes: list of processes in ping message
 */
func (server *RingServer) OnPing(remoteAddr *net.UDPAddr, processes []*api.Process) {
	// logger.Info("Ping received...")

	server.Lock()
	for _, process := range processes {
		// skip self process
		if api.IsSameProcess(process, server.Process) {
			continue
		}

		processIndex := server.FindProcessIndex(process)
		if processIndex == -1 {
			// skip dead process
			if process.Status != api.Status_Alive {
				continue
			}

			// process is alive and not in current membership list
			// meaning a new process, add process to current membership list
			logger.Join(process)
			server.MembershipList = append(server.MembershipList, process)
			server.NotifyMemberUpdate(process, MEMBER_INSERT)
		} else {
			currProcess := server.MembershipList[processIndex]

			// check if process is more recent than current process
			if !process.LastUpdateTime.AsTime().After(currProcess.LastUpdateTime.AsTime()) {
				continue
			}

			// update its status and last update time in current membership list
			currProcess.LastUpdateTime = process.LastUpdateTime
			currProcess.Status = process.Status
			logger.Update(process)

			// update deleted expiration pool
			_, ok := server.ExpirationPool[process.Address()]
			if currProcess.Status == api.Status_Alive {
				if ok {
					delete(server.ExpirationPool, process.Address())
				}
			} else {
				if !ok {
					expiredTime := api.CurrentTimestamp().AsTime().Add(EXPIRATION_TIME)
					server.ExpirationPool[process.Address()] = expiredTime
				}
			}
		}
	}
	server.Unlock()

	// finish updating membership list, send ack to sender
	server.Ack(remoteAddr)
}

/*
 * Update process status in membership list
 *
 * @param process: process to update
 */
func (server *RingServer) OnAck(process *api.Process) {
	server.Lock()
	defer server.Unlock()

	addr := process.Address()
	processIndex := server.FindProcessIndex(process)
	if processIndex == -1 {
		logger.Error("Cannot update process " + addr + " in membership list: process does not exist")
		return
	}

	// update process lastUpdateTime in membership list
	currentTimestamp := api.CurrentTimestamp()
	server.MembershipList[processIndex].LastUpdateTime = currentTimestamp
	logger.Update(process)
}

/*
 * Add process to membership list
 */
func (server *RingServer) OnJoin(remoteAddr *net.UDPAddr, process *api.Process) {
	server.Lock()
	defer server.Unlock()

	// check existence of process in membership list
	addr := process.Address()
	processIndex := server.FindProcessIndex(process)
	if processIndex != -1 {
		logger.Error("Trying to add process " + addr + " to membership list: process already exists")
		return
	}

	// insert process to membership list
	process.LastUpdateTime = api.CurrentTimestamp()
	process.JoinTime = api.CurrentTimestamp()
	process.Status = api.Status_Alive

	// marshal the message
	res, err := api.MarshalMeta(api.MessageType_Join, &api.Metadata_Join{
		Join: &api.JoinMessage{
			Process: process,
		},
	})
	if err != nil {
		logger.Error("Failed to marshal join message: " + err.Error())
		return
	}

	// send the newly joined process info to new node
	server.UDPConn.SetWriteDeadline(time.Now().Add(WRITE_TIMEOUT))
	_, err = server.UDPConn.WriteToUDP(res, remoteAddr)
	if err != nil {
		logger.Error("Failed to send join message: " + err.Error())
		return
	}

	// add process to membership list
	logger.Join(process)
	server.MembershipList = append(server.MembershipList, process)
	server.NotifyMemberUpdate(process, MEMBER_INSERT)
}

/*
 * Remove process from membership list
 */
func (server *RingServer) OnLeave(process *api.Process) {
	server.Lock()
	defer server.Unlock()

	addr := process.Address()
	processIndex := server.FindProcessIndex(process)
	if processIndex == -1 {
		logger.Error("Cannot update process " + addr + " in membership list: process does not exist")
		return
	}

	// update process status in membership list
	server.MembershipList[processIndex].Status = api.Status_Leaved
	server.ExpirationPool[addr] = api.CurrentTimestamp().AsTime().Add(EXPIRATION_TIME)
	logger.Leave(process)
}

/*
 * Add a process to expiration pool and set its state to time out if it has failed, not delete it immediately
 *
 * @param ip: ip address of process to remove
 * @param port: port of process to remove
 */
func (server *RingServer) OnFailure(process *api.Process) {
	server.Lock()
	defer server.Unlock()

	addr := process.Address()
	processIndex := server.FindProcessIndex(process)
	if processIndex == -1 {
		logger.Error("Cannot remove process from membership list: " + addr + " does not exist")
		return
	}

	// update process status in membership list
	server.MembershipList[processIndex].Status = api.Status_Timeout
	server.MembershipList[processIndex].LastUpdateTime = api.CurrentTimestamp()

	// add process to expiration pool
	if _, ok := server.ExpirationPool[addr]; !ok {
		server.ExpirationPool[addr] = api.CurrentTimestamp().AsTime().Add(EXPIRATION_TIME)
	}
	logger.Failure(process)
}
