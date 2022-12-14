import { createColumnHelper } from "@tanstack/react-table";
import { formatDate } from "../shared";
import { Worker, workersState } from "../states";
import { Text, Badge } from "@chakra-ui/react";
import { DataTable } from "./DataTable";
import { useRecoilRefresher_UNSTABLE, useRecoilValue } from "recoil";
import { useMemo } from "react";

const workerHelper = createColumnHelper<Worker>();

const workerColumns = [
  workerHelper.accessor("address", {
    cell: (info) => <Text fontWeight="medium">{info.getValue()}</Text>,
    header: "Worker Address",
  }),
  workerHelper.accessor("joinTime", {
    cell: (info) => {
      const date = new Date(info.getValue());
      return formatDate(date);
    },
    header: "Join Time",
  }),
  workerHelper.accessor("runningJob", {
    cell: (info) => {
      const job = info.getValue();

      return job === "" ? (
        <Badge colorScheme="blue">No Job Available</Badge>
      ) : (
        <Badge colorScheme="purple">{job}</Badge>
      );
    },
    header: "Running Job Id",
  }),
  workerHelper.accessor("runningJob", {
    cell: (info) => {
      const idle = info.getValue() === "";
      return <Badge colorScheme={idle ? "yellow" : "green"}>{idle ? "Idle" : "Active"}</Badge>;
    },
    header: "Status",
    id: "statu∫s",
  }),
  workerHelper.accessor("lastQueryTime", {
    cell: (info) => {
      const date = new Date(info.getValue());
      return formatDate(date);
    },
    header: "Last Query Time",
  }),
];

export const WorkerTable: React.FC = () => {
  const workers = useRecoilValue(workersState);
  const refresh = useRecoilRefresher_UNSTABLE(workersState);

  const sortedWorkers = useMemo(() => {
    return [...workers].sort((a, b) => {
      return new Date(a.joinTime).getTime() - new Date(b.joinTime).getTime();
    });
  }, [workers]);

  return <DataTable columns={workerColumns} data={sortedWorkers} refresh={refresh} />;
};
