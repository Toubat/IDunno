package ralloc_test

import (
	"math"
	"mp4/api"
	"mp4/ralloc"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func Test_Ralloc_OneJob(t *testing.T) {
	assert := assert.New(t)

	numVM := 10
	job := &api.Job{
		TotalQueries:     100,
		CompletedQueries: 0,
		StartTime:        timestamppb.New(time.Now()),
	}
	qps := ralloc.JobToQPS([]*api.Job{job}, numVM)
	alloc, diff := ralloc.GlobalFairTimeRalloc(1, numVM, qps)

	assert.Equal(float64(0), diff, "should have 0 min absolute difference")
	assert.Equal(numVM, alloc[0], "should assign all VM resources")
}

func Test_Ralloc_TwoJobs(t *testing.T) {
	assert := assert.New(t)

	jobs := []*api.Job{{
		TotalQueries:     10000,
		CompletedQueries: 2,
		StartTime:        timestamppb.New(time.Now().Add(-1 * time.Second)),
	}, {
		TotalQueries:     10000,
		CompletedQueries: 2,
		StartTime:        timestamppb.New(time.Now().Add(-2 * time.Second)),
	}}

	// basic test (9 VMs)
	qps := ralloc.JobToQPS(jobs, 9)
	alloc, diff := ralloc.GlobalFairTimeRalloc(2, 9, qps)
	assert.LessOrEqual(diff, 0.1, "should be within 10%")
	assert.Equal(3, alloc[0])
	assert.Equal(6, alloc[1])

	// medium test (12 VMs)
	qps = ralloc.JobToQPS(jobs, 12)
	alloc, diff = ralloc.GlobalFairTimeRalloc(2, 12, qps)
	assert.LessOrEqual(diff, 0.1, "should be within 10%")
	assert.Equal(4, alloc[0])
	assert.Equal(8, alloc[1])

	// large test (120 VMs)
	qps = ralloc.JobToQPS(jobs, 120)
	alloc, diff = ralloc.GlobalFairTimeRalloc(2, 120, qps)
	assert.LessOrEqual(diff, 0.1, "should be within 10%")
	assert.Equal(40, alloc[0])
	assert.Equal(80, alloc[1])
}

func Test_Ralloc_UnbalancedCompletedQueries(t *testing.T) {
	assert := assert.New(t)

	jobs := []*api.Job{{
		TotalQueries:     10000,
		CompletedQueries: 9000,
		StartTime:        timestamppb.New(time.Now().Add(-9000 * time.Second)),
	}, {
		TotalQueries:     10000,
		CompletedQueries: 1,
		StartTime:        timestamppb.New(time.Now().Add(-1 * time.Second)),
	}}

	qps := ralloc.JobToQPS(jobs, 100)
	alloc, diff := ralloc.GlobalFairTimeRalloc(2, 100, qps)
	assert.LessOrEqual(diff, 0.1, "should be within 10%")
	assert.Equal(99, alloc[0])
	assert.Equal(1, alloc[1])

	jobs = []*api.Job{{
		TotalQueries:     10000,
		CompletedQueries: 4000,
		StartTime:        timestamppb.New(time.Now().Add(-1000 * time.Second)),
	}, {
		TotalQueries:     10000,
		CompletedQueries: 1,
		StartTime:        timestamppb.New(time.Now().Add(-1 * time.Second)),
	}}

	qps = ralloc.JobToQPS(jobs, 100)
	alloc, diff = ralloc.GlobalFairTimeRalloc(2, 100, qps)
	assert.LessOrEqual(diff, 0.1, "should be within 10%")
	assert.Equal(90, alloc[0])
	assert.Equal(10, alloc[1])
}

func Test_Ralloc_ThreeJobs(t *testing.T) {
	assert := assert.New(t)

	jobs := []*api.Job{{
		TotalQueries:     10000,
		CompletedQueries: 6,
		StartTime:        timestamppb.New(time.Now().Add(-2 * time.Second)),
	}, {
		TotalQueries:     10000,
		CompletedQueries: 6,
		StartTime:        timestamppb.New(time.Now().Add(-3 * time.Second)),
	}, {
		TotalQueries:     10000,
		CompletedQueries: 6,
		StartTime:        timestamppb.New(time.Now().Add(-6 * time.Second)),
	}}

	qps := ralloc.JobToQPS(jobs, 120)
	alloc, diff := ralloc.GlobalFairTimeRalloc(3, 120, qps)
	assert.LessOrEqual(diff, 0.1, "should be within 10%")
	assert.LessOrEqual(math.Abs(float64(20-alloc[0])), 10.)
	assert.LessOrEqual(math.Abs(float64(40-alloc[1])), 10.)
	assert.LessOrEqual(math.Abs(float64(60-alloc[2])), 10.)
}
