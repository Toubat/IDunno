import { createColumnHelper } from "@tanstack/react-table";
import { Text, Badge } from "@chakra-ui/react";
import { DataTable } from "./DataTable";
import { useRecoilRefresher_UNSTABLE, useRecoilValue } from "recoil";
import { CompletedJob, completedJobsState } from "../states/completedJobs";
import { toDecimalPlace } from "../shared";
import { useMemo } from "react";

const jobHelper = createColumnHelper<CompletedJob>();

export const jobColumns = [
  jobHelper.accessor("id", {
    cell: (info) => <Text fontWeight="medium">{info.getValue()}</Text>,
    header: "Job Id",
  }),
  jobHelper.accessor("modelType", {
    cell: (info) => <Badge>{info.getValue()}</Badge>,
    header: "Model Type",
  }),
  jobHelper.accessor("batchSize", {
    cell: (info) => (
      <Badge textTransform="capitalize" colorScheme="purple" px={2} textAlign="center">
        {info.getValue()}
      </Badge>
    ),
    header: "Batch",
    meta: { align: "center" },
  }),
  jobHelper.accessor("qps", {
    cell: (info) => (
      <Badge textTransform="lowercase" colorScheme="yellow">
        {toDecimalPlace(info.getValue(), 2)}
      </Badge>
    ),
    header: "Query Rate",
    meta: { align: "center" },
  }),
  jobHelper.accessor("totalQueries", {
    cell: (info) => (
      <Badge textTransform="capitalize" colorScheme="blue" textAlign="center">
        {info.getValue()}
      </Badge>
    ),
    header: "Total Queries",
    meta: { align: "center" },
  }),

  jobHelper.accessor("totalQueryTime", {
    cell: (info) => (
      <Badge textTransform="lowercase"> {toDecimalPlace(info.getValue(), 2)} sec</Badge>
    ),
    header: "Query Time",
  }),
];

export const CompletedJobTable: React.FC = () => {
  const completedJobs = useRecoilValue(completedJobsState);
  const refresh = useRecoilRefresher_UNSTABLE(completedJobsState);

  const sortedJobs = useMemo(() => {
    return [...completedJobs].sort((a, b) => {
      return a.id < b.id ? 1 : -1;
    });
  }, [completedJobs]);

  return <DataTable columns={jobColumns} data={sortedJobs} refresh={refresh} />;
};
