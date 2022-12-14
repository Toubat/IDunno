import axios from "axios";
import { selector } from "recoil";
import { STAT_SERVER_ADDRESS } from "../shared";

export type Worker = {
  address: string;
  joinTime: string;
  runningJob: string;
  lastQueryTime: string;
};

export const workersState = selector<Worker[]>({
  key: "workersState",
  get: async () => {
    const response = await axios.get(`${STAT_SERVER_ADDRESS}/worker`);
    return response.data;
  },
});
