package ralloc

import (
	"fmt"
	"math"
	"mp4/api"
)

// Ralloc is a resource allocation algorithm that allocates resources
// to each jobs in a fair-time fashion.
//
// Args:
//   - n: number of jobs
//   - m: number of resources
//   - qps: n x m+1 2D array of query per second for each job and resource
//
// Returns:
//   - 1D array of resource allocated for each job that sums up to m
//     (i.e. the total number of resources) while minimizing the
//     pair-wise difference of query per second between each job.
//   - min absolute pair-wise difference of qps for each job under
//     the optimal resource assignment.
func GlobalFairTimeRalloc(n int, m int, qps [][]float64) ([]int, float64) {
	if len(qps) != n || len(qps[0]) != m+1 {
		panic(fmt.Sprintf("Expected qps to be %v x %v matrix, but got %v x %v.", n, m+1, len(qps), len(qps[0])))
	}

	dp := make2d[float64](n, m+1)  // dp[i][j] = min absolute pair-wise difference of qps for each job given jobs[0...i] and j resources
	min := make2d[float64](n, m+1) // min[i][j] = min qps for each job given jobs[0...i] and j resources under optimal allocation dp[i][j]
	max := make2d[float64](n, m+1) // max[i][j] = max qps for each job given jobs[0...i] and j resources under optimal allocation dp[i][j]
	alloc := make2d[int](n, m+1)   // alloc[i][j] = optimal allocation for i-th job given jobs[0...i] and j resources

	// base case (only 1 job)
	for j := 0; j <= m; j++ {
		dp[0][j] = 0
		min[0][j] = qps[0][j]
		max[0][j] = qps[0][j]
		alloc[0][j] = j
	}

	// recurrence
	for i := 1; i < n; i++ {
		for j := 0; j <= m; j++ {
			dp[i][j] = math.MaxInt64
			resource := 0

			for k := 0; k <= j; k++ {
				maxDiff := math.Max(math.Max(math.Abs(qps[i][k]-min[i-1][j-k]), math.Abs(qps[i][k]-max[i-1][j-k])), dp[i-1][j-k])
				if maxDiff < dp[i][j] {
					dp[i][j] = maxDiff
					resource = k
				}
			}

			alloc[i][j] = resource
			min[i][j] = math.Min(qps[i][resource], min[i-1][j-resource])
			max[i][j] = math.Max(qps[i][resource], max[i-1][j-resource])
		}
	}

	// backtracking
	resources, curr := make([]int, n), m
	for i := n - 1; i >= 0; i-- {
		resources[i] = alloc[i][curr]
		curr -= resources[i]
	}
	optimalQPS := make([]float64, n)

	return resources, GetRelQPSDiff(optimalQPS)
}

func LocalFairTimeRalloc(jobs []*api.Job, totalResources int) ([]int, float64) {
	time := make([]float64, len(jobs)) // second per query (local average)
	rawAlloc := make([]float64, len(jobs))
	alloc := make([]int, len(jobs))

	totalTime := 0.0
	for i, job := range jobs {
		time[i] = job.QueryProcessingTime()
		totalTime += time[i]
	}

	// allocate resources to each job
	for i := 0; i < len(jobs); i++ {
		rawAlloc[i] = float64(totalResources) * time[i] / totalTime
	}

	// ceil up the allocation
	for i := 0; i < len(jobs); i++ {
		actualAlloc := int(math.Min(math.Round(rawAlloc[i]), float64(totalResources)))
		alloc[i] = actualAlloc
		totalResources -= actualAlloc
	}

	return alloc, GetRelQPSDiff(time)
}

func GetRelQPSDiff(qps []float64) float64 {
	if len(qps) <= 1 {
		return 0
	}

	maxErr := float64(math.MinInt64)
	eps := 1e-6

	for i := 0; i < len(qps); i++ {
		for j := i + 1; j < len(qps); j++ {
			diff := math.Abs(qps[i] - qps[j])
			err := diff / (math.Max(qps[i], qps[j]) + eps)
			maxErr = math.Max(maxErr, err)
		}
	}

	return maxErr
}

func JobToQPS(jobs []*api.Job, totalResources int) [][]float64 {
	qps := make2d[float64](len(jobs), totalResources+1)

	for i, job := range jobs {
		for j := 0; j <= totalResources; j++ {
			qps[i][j] = job.GetExpectedQPS(j)
		}
	}

	return qps
}

func make2d[T any](rows int, cols int) [][]T {
	arr := make([][]T, rows)
	for i := 0; i < rows; i++ {
		arr[i] = make([]T, cols)
	}
	return arr
}
