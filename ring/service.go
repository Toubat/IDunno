package ring

import (
	"context"

	"mp4/api"
	"mp4/logger"
	"mp4/utils"

	"net"
	"sort"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type RingServerService interface {
	Cron()
	Listen() error
	Ping(process *api.Process) error
	Ack(remoteAddr *net.UDPAddr) error
	Join() error
	Leave() error
	LookupLeader() (string, error)
	UpdateLeader() error
}

/**
 * Leave the ring
 *
 */
func (server *RingServer) Leave() {
	server.Lock()
	server.Status = api.Status_Leaved
	server.MembershipList = make(MembershipList, 0)
	server.Unlock()

	go server.OnMemberUpdate(server.Process, MEMBER_LEAVED)
}

/**
 * Cron job running periodically, it has following main functionalities at each period:
 * 1) First recycle suspected process and initiate the ring stabilization, see stabalization details in NotifyMemberUpdate() callback
 * 2) ping the next 4 successors in the ring and put any failed process into expiration pool
 *
 * @return error: raise error if leave fails
 */
func (server *RingServer) Cron() {
	ticker := time.NewTicker(INTERVAL)

	for time := range ticker.C {
		// remove expired processes from membership list
		server.RecyclePool(time)

		// update last update time
		server.Lock()
		server.Process.LastUpdateTime = api.CurrentTimestamp()
		server.Unlock()

		// ping the next 4 successors in the ring
		successors := server.Successors()

		for _, successor := range successors {
			go func(process *api.Process) {
				logger.Ping(process)
				if err := server.Ping(process); err != nil {
					logger.Error(err.Error())
					server.OnFailure(process)
				}
			}(successor)
		}
	}
}

/**
 * Listen is responsible for handling incoming messages from other servers or introduce
 *
 * @return error: raise error if listen fails
 */
func (server *RingServer) Listen() error {
	for {
		// receive metadata
		buffer := make([]byte, 2048)
		n, remoteAddr, err := server.UDPConn.ReadFromUDP(buffer)
		if err != nil {
			logger.Error("Failed to receive data from address " + remoteAddr.String())
			return err
		}

		if server.Status == api.Status_Leaved {
			continue
		}

		// unmarshal metadata
		meta, err := api.UnmarshalMeta(buffer[:n])
		if err != nil {
			logger.Error("Failed to unmarshal message from address " + remoteAddr.String())
			return err
		}

		// handle metadata
		switch meta.GetType() {
		case api.MessageType_Ping:
			go server.OnPing(remoteAddr, meta.GetPing().GetProcesses())
		case api.MessageType_Join:
			go server.OnJoin(remoteAddr, meta.GetJoin().GetProcess())
		case api.MessageType_Leave:
			go server.OnLeave(meta.GetLeave().GetProcess())
		}
	}
}

/*
 * Ping process at ip:port by sending full membership list
 *
 * @param ip: ip address of process to ping
 * @param port: port of process to ping
 * @return error: raise error if ping fails
 */
func (server *RingServer) Ping(process *api.Process) error {
	// marshal ping metadata
	server.Lock()
	pingMeta, err := api.MarshalMeta(api.MessageType_Ping, &api.Metadata_Ping{
		Ping: &api.PingMessage{
			Processes: server.MembershipList,
		},
	})
	server.Unlock()

	if err != nil {
		logger.Error("Error marshalling Ping message")
		return err
	}

	addr := process.Address()
	// establish UDP connection
	// logger.Info("Connecting to " + addr + "...")
	conn, err := net.DialTimeout("udp", addr, PING_TIMEOUT)
	if err != nil {
		logger.Error("Timeout when dialing UDP connection to address " + addr)
		return err
	}
	// logger.Info("Successfully dialed UDP connection to address " + addr)
	defer conn.Close()

	// send ping metadata
	n, err := utils.WithDropProb(DROP_PROB, func() (int, error) {
		conn.SetWriteDeadline(time.Now().Add(WRITE_TIMEOUT))
		return conn.Write(pingMeta)
	})
	if err != nil {
		logger.Error("Failed to send data to address " + addr)
		return err
	}

	// receive ack metadata
	buffer := make([]byte, 1024)
	conn.SetReadDeadline(time.Now().Add(READ_TIMEOUT))
	n, err = conn.Read(buffer)
	if err != nil {
		logger.Error("Failed to receive data from address " + addr)
		return err
	}

	// unmarshal ack metadata
	ackMeta, err := api.UnmarshalMeta(buffer[:n])
	if err != nil {
		logger.Error("Failed to unmarshal ack message from address " + addr)
		return err
	}

	if ackMeta.GetType() != api.MessageType_Ack || ackMeta.GetAck().GetReceived() != ACK_MESSAGE {
		logger.Error("Received invalid ack message from address " + addr)
		return err
	}

	// ack callback
	server.OnAck(process)

	return nil
}

/*
 * Send ack to process at ip:port
 *
 * @param remoteAddr: ip:port of process to send ack to
 * @return error: raise error if ack fails
 */
func (server *RingServer) Ack(remoteAddr *net.UDPAddr) error {
	// marshal ack metadata
	ackMeta, err := api.MarshalMeta(api.MessageType_Ack, &api.Metadata_Ack{
		Ack: &api.AckMessage{
			Received: ACK_MESSAGE,
		},
	})
	if err != nil {
		logger.Error("Error marshalling Ack message")
		return err
	}

	// write ack metadata
	_, err = utils.WithDropProb(DROP_PROB, func() (int, error) {
		server.UDPConn.SetWriteDeadline(time.Now().Add(WRITE_TIMEOUT))
		return server.UDPConn.WriteToUDP(ackMeta, remoteAddr)
	})
	if err != nil {
		logger.Error("Failed to send data to address " + remoteAddr.String())
		return err
	}

	return nil
}

/*
 * Join ring by sending join message to introducer
 *
 * @return error: raise error if join fails
 */
func (server *RingServer) Join() error {
	if server.FindProcessIndex(server.Process) != -1 {
		logger.Info("Trying to join ring but process " + server.Process.Address() + " already in ring")
		return nil
	}

	// marshal join metadata
	joinMeta, err := api.MarshalMeta(api.MessageType_Join, &api.Metadata_Join{
		Join: &api.JoinMessage{
			Process: server.Process,
		},
	})
	if err != nil {
		logger.Error("Error marshalling Join message")
		return err
	}

	leaderAddr, err := server.LookupLeader()
	if err != nil {
		logger.Error("Failed to lookup leader address")
		return err
	}

	if leaderAddr == "" {
		// No leader found, starting new ring
		logger.Info("No leader found, starting new ring")
		server.Lock()
		server.Process.LastUpdateTime = api.CurrentTimestamp()
		server.Process.JoinTime = api.CurrentTimestamp()
		server.Process.Status = api.Status_Alive
		server.MembershipList = append(server.MembershipList, server.Process)
		logger.Join(server.Process)
		server.Unlock()

		server.NotifyMemberUpdate(server.Process, MEMBER_INSERT)
		return nil
	}

	// establish UDP connection
	conn, err := net.DialTimeout("udp", leaderAddr, PING_TIMEOUT)
	if err != nil {
		logger.Error("Timeout when dialing UDP connection to introducer")
		return err
	}
	defer conn.Close()

	// send join metadata
	_, err = utils.WithDropProb(DROP_PROB, func() (int, error) {
		conn.SetWriteDeadline(time.Now().Add(WRITE_TIMEOUT))
		return conn.Write(joinMeta)
	})
	if err != nil {
		logger.Error("Failed to send data to introducer: " + err.Error())
		return err
	}

	// receive join time metadata
	buffer := make([]byte, 1024)
	conn.SetReadDeadline(time.Now().Add(READ_TIMEOUT))
	n, err := conn.Read(buffer)
	if err != nil {
		logger.Error("Failed to receive data from introducer: " + err.Error())
		return err
	}

	// unmarshal process join metadata
	processMeta, err := api.UnmarshalMeta(buffer[:n])
	if processMeta.GetType() != api.MessageType_Join || processMeta.GetJoin().Process == nil {
		logger.Error("Received invalid join time message from introducer: " + leaderAddr)
		return err
	}

	// update process join time & last update time
	server.Lock()
	server.Process = processMeta.GetJoin().Process
	server.MembershipList = append(server.MembershipList, server.Process)
	logger.Join(server.Process)
	server.Unlock()

	// edge case: should not call NotifyMemberUpdate here, since we don't want to call
	// UpdateLeader for newly joined node (as it would cause error to DNS server)
	go server.OnMemberUpdate(server.Process, MEMBER_INSERT)

	return nil
}

func (server *RingServer) LookupLeader() (string, error) {
	// logger.Info("Looking up leader...")
	server.Lock()
	defer server.Unlock()

	// choose process with smallest join time if already in ring
	if len(server.MembershipList) > 0 {
		sort.Sort(server.MembershipList)
		return server.MembershipList[0].Address(), nil
	}

	// new process join, lookup leader in DNS table
	// logger.Info("Looking up leader in DNS table...")
	conn, err := grpc.Dial(DNS_ADDR, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Error("Failed to dial DNS server")
		return "", err
	}
	defer conn.Close()

	DNSClient := api.NewDNSServiceClient(conn)
	res, err := DNSClient.Lookup(context.Background(), &api.LookupLeaderRequest{})
	if err != nil {
		logger.Error("Failed to lookup leader " + err.Error())
		return "", err
	}

	// logger.Info("Found introducer address: " + res.GetAddress())
	return res.GetAddress(), nil
}

/*
 * Update leader in DNS table, only the process that has the earliest join time can elect itself to be the leader and update the DNS table
 *
 * @return error: raise error if update fails
 */
func (server *RingServer) UpdateLeader() error {
	// find process with smallest join time in ring
	server.Lock()
	sort.Sort(server.MembershipList)
	shouldUpdate := len(server.MembershipList) > 0 && api.IsSameProcess(server.Process, server.MembershipList[0])
	server.Unlock()

	// only process with smallest join time can be leader
	if !shouldUpdate {
		return nil
	}

	// update leader process in DNS table
	conn, err := grpc.Dial(DNS_ADDR, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Error("Failed to dial DNS server")
		return err
	}
	defer conn.Close()

	DNSClient := api.NewDNSServiceClient(conn)
	res, err := DNSClient.Update(context.Background(), &api.UpdateLeaderRequest{
		Leader: server.Process,
	})
	if err != nil || res.GetStatus() != api.ResponseStatus_OK {
		logger.Error("Failed to lookup leader")
		return err
	}

	// logger.Info("Updated leader address to " + server.Process.Address())
	return nil
}
