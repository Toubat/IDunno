import axios from "axios";
import { selectorFamily } from "recoil";
import { STAT_SERVER_ADDRESS } from "../shared";

export type Batch = {
  batchInput: string;
  batchOutput: string;
};

export type JobInfo = {
  id: string;
  batches: Batch[];
  metric: number;
  queryRates: number[];
  queryProcessTimes: number[];
};

export const jobInfoState = selectorFamily<JobInfo, string>({
  key: "jobInfoState",
  get: (jobId: string) => async () => {
    if (jobId === "") return null;

    const response = await axios.get(`${STAT_SERVER_ADDRESS}/jobs?id=${jobId}`);
    return response.data;
  },
});

export const batchesState = selectorFamily<Batch[], string>({
  key: "batchesState",
  get:
    (jobId: string) =>
    async ({ get }) => {
      if (jobId === "") return [];

      const jobInfo = get(jobInfoState(jobId));
      return jobInfo.batches;
    },
});

export const queryRatesState = selectorFamily<number[], string>({
  key: "queryRatesState",
  get:
    (jobId: string) =>
    async ({ get }) => {
      if (jobId === "") return [];

      const jobInfo = get(jobInfoState(jobId));
      console.log(jobInfo.queryRates);
      return jobInfo.queryRates;
    },
});

export const queryProcessTimesState = selectorFamily<number[], string>({
  key: "queryProcessTimesState",
  get:
    (jobId: string) =>
    async ({ get }) => {
      if (jobId === "") return [];

      const jobInfo = get(jobInfoState(jobId));
      return jobInfo.queryProcessTimes;
    },
});
