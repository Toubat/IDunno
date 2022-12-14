package ring

import (
	"math"
	"mp4/api"
	"os"

	"mp4/logger"
	"net"
	"sort"
	"sync"
	"time"

	"github.com/jedib0t/go-pretty/table"
	"github.com/jedib0t/go-pretty/text"
)

var DROP_PROB float64 = 0.00

// const HOSTNAME = "fa22-cs425-2401.cs.illinois.edu:8889"

const DNS_ADDR = "fa22-cs425-2401.cs.illinois.edu:8889"

// const DNS_ADDR = "localhost:8889"

const MAX_SUCCESSORS int = 6                                                 // 4 should be enough for completeness
const ACK_MESSAGE string = "PONG"                                            // ack message
const PING_TIMEOUT time.Duration = time.Duration(700) * time.Millisecond     // 500 milliseconds
const READ_TIMEOUT time.Duration = time.Duration(700) * time.Millisecond     // 500 milliseconds
const WRITE_TIMEOUT time.Duration = time.Duration(700) * time.Millisecond    // 500 milliseconds
const EXPIRATION_TIME time.Duration = time.Duration(6000) * time.Millisecond // 5000 milliseconds
const INTERVAL time.Duration = time.Duration(1450) * time.Millisecond        // 800 milliseconds

// Action flag for process being added/deleted from the membership list
type MemAction int

const (
	MEMBER_DELETE MemAction = iota
	MEMBER_INSERT
	MEMBER_LEAVED
)

// Pool contains a list of processes to be deleted in the future
type ExpirationPool map[string]time.Time

// Callback function (additional callback other that the ring itself, i.e. SDFS related actions) for inserting/deleting process in membership list
type OnMemberUpdate func(process *api.Process, action MemAction)

/* RingServer
 * - Implements RingServerService interface
 * - Implements RingServerEvent interface
 */
type RingServer struct {
	*net.UDPConn      // udp connection
	*api.Process      // current process
	MembershipList    // RingServer.process must be in the list
	ExpirationPool    // list of processes to be deleted
	OnMemberUpdate    // callback function when membership list is updated
	RingServerService // service interface
	RingServerEvent   // event interface
	sync.Mutex        // lock for concurrent access
}

func NewRingServer(conn *net.UDPConn, ip string, port int32) *RingServer {
	process := &api.Process{
		Ip:     ip,
		Port:   port,
		Status: api.Status_Alive,
	}
	return &RingServer{
		UDPConn:        conn,
		Process:        process, // initially only contain self
		MembershipList: MembershipList{},
		ExpirationPool: make(ExpirationPool, 0),
	}
}

func (server *RingServer) SetMemberUpdateCallback(callback OnMemberUpdate) {
	server.OnMemberUpdate = callback
}

/**
 * The server handler to recycle the expiration pool and initiate the ring stabilization machanism
 * 1. Check if there any process that has passed its timeout time and delete it
 * 2. Initialize the ring stabilization with NotifyMemberUpdate(),
 *    which is a callback that handles the SDFS file re-replication and leader election
 */
func (server *RingServer) RecyclePool(currTime time.Time) {
	server.Lock()
	defer server.Unlock()

	deletedAddresses := make([]string, 0)
	for address, expiredTime := range server.ExpirationPool {
		if currTime.After(expiredTime) {
			// remove process from membership list
			processIndex := -1
			for i, process := range server.MembershipList {
				if address == process.Address() {
					processIndex = i
					break
				}
			}

			if processIndex == -1 {
				logger.Error("Trying to delete a process that does not exist: " + address)
				continue
			}

			// remove process from membership list
			deletedProcess := server.MembershipList[processIndex]
			logger.Delete(deletedProcess)
			deletedAddresses = append(deletedAddresses, address)
			server.MembershipList = append(server.MembershipList[:processIndex], server.MembershipList[processIndex+1:]...)
			server.NotifyMemberUpdate(deletedProcess, MEMBER_DELETE)
		}
	}

	// remove deleted processes from expiration pool
	for _, address := range deletedAddresses {
		delete(server.ExpirationPool, address)
	}
}

/* Find 4 successors of current process
 *
 * @return []*api.Process: list of (at most) 4 successors
 */
func (server *RingServer) Successors() []*api.Process {
	server.Lock()
	defer server.Unlock()

	// sort list by timestamp
	sort.Sort(server.MembershipList)

	// find index of current process
	index := server.FindProcessIndex(server.Process)

	// return MAX_SUCCESSORS successors
	successors := make([]*api.Process, 0)
	for i := 0; i < int(math.Min(float64(MAX_SUCCESSORS), float64(server.MembershipList.Len()))); i++ {
		process := server.MembershipList[(index+i+1)%server.MembershipList.Len()]

		// skip self if number of processes is no greater than MAX_SUCCESSORS + 1
		if api.IsSameProcess(process, server.Process) {
			break
		}
		successors = append(successors, process)
	}

	return successors
}

func (server *RingServer) FindProcessIndex(process *api.Process) int {
	for i, p := range server.MembershipList {
		if api.IsSameProcess(p, process) {
			return i
		}
	}

	return -1
}

func (server *RingServer) GetMembershipList() []*api.Process {
	server.Lock()
	defer server.Unlock()

	// copy list
	membershipList := make([]*api.Process, len(server.MembershipList))
	copy(membershipList, server.MembershipList)

	return membershipList
}

/**
 * A delegate(a group of actions) that initiates the SDFS file re-replication and leader election
 * Both callbacks get called are asynchronous.
 *
 */
func (server *RingServer) NotifyMemberUpdate(process *api.Process, action MemAction) {
	go server.UpdateLeader()
	go server.OnMemberUpdate(process, action)
}

func (server *RingServer) ListMembers() {
	server.Lock()
	defer server.Unlock()

	sort.Sort(server.MembershipList)

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{
		"Machine Address",
		"Join Time",
		"Status",
	})

	for _, process := range server.MembershipList {
		t.AppendRow(table.Row{
			process.Address(),
			process.JoinTime.AsTime().Format("2006-01-02 15:04:05"),
			process.Status.String(),
		})
	}

	t.AppendFooter(table.Row{"Total Machines", len(server.MembershipList)})
	t.SetAutoIndex(true)
	t.SetStyle(table.StyleLight)
	t.Style().Format.Header = text.FormatTitle
	t.Style().Format.Footer = text.FormatTitle
	t.Render()
}

func (server *RingServer) ListSelf() {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{
		"Machine Address",
		"Join Time",
		"Status",
	})

	t.AppendRow(table.Row{
		server.Address(),
		server.JoinTime.AsTime().Format("2006-01-02 15:04:05"),
		server.Status.String(),
	})

	t.SetAutoIndex(true)
	t.SetStyle(table.StyleLight)
	t.Style().Format.Header = text.FormatTitle
	t.Render()
}

func SetNetworkDropRate(prob float64) {
	DROP_PROB = prob
}
