package api

import (
	"math"
	"strings"
	"time"
)

const QUERY_TIME_LIMIT = 6

// second per query (global)
func (j *Job) SecondPerQuery() float64 {
	if j.CompletedQueries == 0 {
		return 1
	}

	return j.TotalQueryTime().Seconds() / float64(j.CompletedQueries)
}

// second per query (local average)
func (j *Job) QueryProcessingTime() float64 {
	if j.CompletedQueries == 0 {
		return 1
	}

	totalSeconds := 0.0
	completedQueries := 0
	for _, state := range j.BatchStates {
		if state.Status == BatchStatus_Completed {
			totalSeconds += math.Min(state.ReceiveTime.AsTime().Sub(state.QueryTime.AsTime()).Seconds(), QUERY_TIME_LIMIT)
			completedQueries++
		}
	}

	return totalSeconds / float64(completedQueries)
}

func (j *Job) TotalQueryTime() time.Duration {
	if j.FinishTime == nil {
		return time.Since(j.StartTime.AsTime())
	}
	return j.FinishTime.AsTime().Sub(j.StartTime.AsTime())
}

func (j *Job) GetExpectedQPS(resource int) float64 {
	if resource == 0 {
		return 0
	}

	remainQueryCount := float64(j.TotalQueries - j.CompletedQueries)
	totalElapsedTime := j.TotalQueryTime().Seconds() + remainQueryCount*j.SecondPerQuery()/float64(resource)
	return float64(j.TotalQueries) / totalElapsedTime
}

func (j *Job) GetQPS(lastSeconds float64) float64 {
	if lastSeconds == 0 {
		return 0
	}

	totalSeconds := time.Since(j.StartTime.AsTime()).Seconds()
	lastSeconds = math.Min(lastSeconds, totalSeconds)
	queryCount := 0

	for _, state := range j.BatchStates {
		if state.QueryTime != nil && time.Since(state.QueryTime.AsTime()).Seconds() <= lastSeconds {
			queryCount++
		}
	}

	return float64(queryCount) / lastSeconds
}

func (j *Job) MeasureQPS() {
	j.QueryRates = append(j.QueryRates, float32(j.GetQPS(10)))
}

func (j *Job) MeasureQueryProcessTime() {
	j.QueryProcessTimes = append(j.QueryProcessTimes, float32(j.QueryProcessingTime()))
}

func (j *Job) GetExpectedTimeLeft(resource int) float64 {
	if resource == 0 {
		return math.MaxFloat64
	}

	remainQueryCount := float64(j.TotalQueries - j.CompletedQueries)
	return remainQueryCount / j.GetQPS(10)
}

func (j *Job) FetchBatchInput() *BatchInput {
	for i := range j.BatchStates {
		if j.BatchStates[i].Status == BatchStatus_Available {
			j.BatchStates[i].Status = BatchStatus_InProgress
			j.BatchStates[i].QueryTime = CurrentTimestamp()
			return j.BatchStates[i].BatchInput
		}
	}
	return nil
}

func (j *Job) GetCompletedBatchCount() int {
	count := 0
	for _, state := range j.BatchStates {
		if state.Status == BatchStatus_Completed {
			count++
		}
	}
	return count
}

func (j *Job) GetResults() ([]*EvalResult, float32) {
	metricSum := float32(0)
	evalResults := make([]*EvalResult, 0)

	for _, state := range j.BatchStates {
		batchOutput := state.BatchOutput

		if batchOutput == nil {
			continue
		}

		metricSum += batchOutput.Metric
		evalResults = append(evalResults, batchOutput.Results...)
	}

	for i := range evalResults {
		input := strings.Split(evalResults[i].GetInput(), "/")

		if len(input) > 0 {
			evalResults[i].Input = input[len(input)-1]
		}
	}

	return evalResults, metricSum / float32(len(evalResults))
}
