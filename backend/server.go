package backend

import (
	"context"
	"fmt"
	"mp4/api"
	"mp4/logger"
	"mp4/sdfs"
	"net/http"
	"strconv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	STATUS_JSON_WORKER         = "jw"
	STATUS_JSON_JOBS           = "jj"
	STATUS_JSON_JOB_ID         = "jij"
	STATUS_JSON_COMPLETED_JOBS = "jcj"
)

const DNS_ADDR = "fa22-cs425-2401.cs.illinois.edu:8889"

type IDunnoStatServer struct {
	Port int
}

func NewIdunnoStatServer(port int) *IDunnoStatServer {
	return &IDunnoStatServer{
		Port: port,
	}
}
func (s *IDunnoStatServer) Serve() {

	logger.StatServerInfo("Starting stat server on port " + strconv.Itoa(s.Port))

	http.HandleFunc("/worker", WorkerHandler)
	http.HandleFunc("/jobs", JobsHandler)
	http.HandleFunc("/completed-jobs", CompletedJobsHandler)

	http.ListenAndServe(":"+strconv.Itoa(s.Port), nil)
}

func WorkerHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Headers", "Access-Control-Allow-Headers, Origin,Accept, X-Requested-With, Content-Type, Access-Control-Request-Method, Access-Control-Request-Headers")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	RequestsHandler(w, STATUS_JSON_WORKER, "")
}

func JobsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Headers", "Access-Control-Allow-Headers, Origin,Accept, X-Requested-With, Content-Type, Access-Control-Request-Method, Access-Control-Request-Headers")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	
	query := r.URL.Query()
	id := query.Get("id")

	if id == "" {
		RequestsHandler(w, STATUS_JSON_JOBS, id)
	} else {
		RequestsHandler(w, STATUS_JSON_JOB_ID, id)
	}
}

func CompletedJobsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Headers", "Access-Control-Allow-Headers, Origin,Accept, X-Requested-With, Content-Type, Access-Control-Request-Method, Access-Control-Request-Headers")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	RequestsHandler(w, STATUS_JSON_COMPLETED_JOBS, "")
}

func LookupLeader() (string, error) {
	conn, err := grpc.Dial(DNS_ADDR, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return "", err
	}
	defer conn.Close()

	DNSClient := api.NewDNSServiceClient(conn)
	res, err := DNSClient.Lookup(context.Background(), &api.LookupLeaderRequest{})
	if err != nil {
		return "", err
	}

	return res.GetAddress(), nil
}

func RequestsHandler(w http.ResponseWriter, which string, payload string) {
	coordinatorAddr, err := LookupLeader()
	if err != nil {
		logger.StatServerError("Failed to lookup leader")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Failed to lookup leader")
		return
	}

	conn, err := grpc.Dial(coordinatorAddr, sdfs.GRPC_OPTIONS...)
	if err != nil {
		logger.StatServerError("Failed to connect to coordinator")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Failed to connect to coordinator")
	}
	defer conn.Close()

	client := api.NewCoordinatorServiceClient(conn)

	res, err := client.IDunnoStatus(context.Background(), &api.IDunnoStatusRequest{
		Which:   which,
		Payload: payload,
	})

	if err != nil || res.Message == "" {
		logger.StatServerError("Failed to get status from coordinator")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Failed to get status from coordinator")
	}

	fmt.Fprint(w, res.Message)
}
