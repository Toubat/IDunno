import axios from "axios";
import { selector } from "recoil";
import { STAT_SERVER_ADDRESS } from "../shared";
import { CompletedJob, completedJobsState } from "./completedJobs";

export type Job = {
  id: string;
  modelType: string;
  batchSize: number;
  totalQueries: number;
  completedQueries: number;
  totalQueryTime: number; // seconds
  runningVMs: number;
  progress: number;
  qps: number;
  timeLeft: number;
};

export type Jobs = {
  jobs: Job[];
  relativeQPSDifference: number;
};

export const jobsState = selector<Jobs>({
  key: "jobsState",
  get: async () => {
    const response = await axios.get(`${STAT_SERVER_ADDRESS}/jobs`);
    return response.data;
  },
});

export const jobIdsState = selector<string[]>({
  key: "jobIdsState",
  get: async ({ get }) => {
    const jobs = get(jobsState);
    const completedJobs = get(completedJobsState);

    const jobIds = jobs.jobs.map((job: Job) => job.id);
    const completedJobIds = completedJobs.map((job: CompletedJob) => job.id);

    return [...jobIds, ...completedJobIds];
  },
});
