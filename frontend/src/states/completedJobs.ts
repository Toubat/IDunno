import axios from "axios";
import { selector } from "recoil";
import { STAT_SERVER_ADDRESS } from "../shared";

export type CompletedJob = {
  batchSize: number;
  id: string;
  modelType: string;
  qps: number;
  totalQueries: number;
  totalQueryTime: number;
};

export const completedJobsState = selector<CompletedJob[]>({
  key: "completedJobsState",
  get: async () => {
    const response = await axios.get(`${STAT_SERVER_ADDRESS}/completed-jobs`);
    return response.data;
  },
});
