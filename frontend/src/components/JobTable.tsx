import { createColumnHelper } from "@tanstack/react-table";
import {
  Text,
  Badge,
  Progress,
  Stack,
  StatLabel,
  Stat,
  StatNumber,
  StatHelpText,
  StatArrow,
} from "@chakra-ui/react";
import { DataTable } from "./DataTable";
import { useRecoilRefresher_UNSTABLE, useRecoilValue } from "recoil";
import { Job, jobsState } from "../states/jobs";
import { toDecimalPlace } from "../shared";
import { useEffect, useMemo, useState } from "react";

const jobHelper = createColumnHelper<Job>();

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
  jobHelper.accessor("totalQueries", {
    cell: (info) => (
      <Badge textTransform="capitalize" colorScheme="blue" textAlign="center">
        {info.getValue()}
      </Badge>
    ),
    header: "Total Queries",
    meta: { align: "center" },
  }),
  jobHelper.accessor("completedQueries", {
    cell: (info) => (
      <Badge textTransform="capitalize" colorScheme="cyan" textAlign="center">
        {info.getValue()}
      </Badge>
    ),
    header: "Completed Queries",
    meta: { align: "center" },
  }),
  jobHelper.accessor("totalQueryTime", {
    cell: (info) => (
      <Badge textTransform="lowercase"> {toDecimalPlace(info.getValue(), 2)} sec</Badge>
    ),
    header: "Query Time",
  }),
  jobHelper.accessor("runningVMs", {
    cell: (info) => (
      <Badge textTransform="capitalize" colorScheme="green" textAlign="center">
        {info.getValue()}
      </Badge>
    ),
    header: "Machines",
    meta: { align: "center" },
  }),
  jobHelper.accessor("progress", {
    cell: (info) => <Progress hasStripe value={info.getValue()} />,
    header: "Progress",
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
  jobHelper.accessor("timeLeft", {
    cell: (info) => (
      <Badge textTransform="lowercase" colorScheme="pink">
        {toDecimalPlace(info.getValue(), 2)} sec
      </Badge>
    ),
    header: "Time Left",
  }),
];

export const JobTable: React.FC = () => {
  const [prevDiff, setPrevDiff] = useState(100);
  const { jobs, relativeQPSDifference } = useRecoilValue(jobsState);
  const refresh = useRecoilRefresher_UNSTABLE(jobsState);

  const sortedJobs = useMemo(() => {
    return [...jobs].sort((a, b) => {
      return a.id < b.id ? 1 : -1;
    });
  }, [jobs]);

  useEffect(() => {
    return () => {
      setPrevDiff(relativeQPSDifference);
    };
  }, [relativeQPSDifference]);

  return (
    <DataTable
      columns={jobColumns}
      data={sortedJobs}
      refresh={refresh}
      render={() => {
        return (
          <Stat maxW="200">
            <StatLabel>Relative QPS Difference</StatLabel>
            <StatNumber>{toDecimalPlace(relativeQPSDifference, 2)}%</StatNumber>
            <StatHelpText>
              <StatArrow
                type={relativeQPSDifference > prevDiff ? "increase" : "decrease"}
                color={relativeQPSDifference > prevDiff ? "red.500" : "green.500"}
              />
              {toDecimalPlace(Math.abs((100 * (relativeQPSDifference - prevDiff)) / prevDiff), 2)}%
            </StatHelpText>
          </Stat>
        );
      }}
    />
  );
};
