import { createColumnHelper } from "@tanstack/react-table";
import { Text, Badge, Stack, Select, Flex, IconButton } from "@chakra-ui/react";
import { DataTable } from "./DataTable";
import { useRecoilRefresher_UNSTABLE, useRecoilValue } from "recoil";
import { Batch, batchesState, jobIdsState, jobInfoState } from "../states";
import { useState } from "react";
import { MdRefresh } from "react-icons/md";

const batchHelper = createColumnHelper<Batch>();

const batchColumns = [
  batchHelper.accessor("batchInput", {
    cell: (info) => (
      <Text decoration="underline" fontWeight="medium">
        {info.getValue()}
      </Text>
    ),
    header: "Inputs",
  }),
  batchHelper.accessor("batchOutput", {
    cell: (info) => <Badge>{info.getValue().slice(0, 55)}</Badge>,
    header: "Outputs",
  }),
];

export const InfoTable: React.FC = () => {
  const [selectedJobId, setSelectedJobId] = useState<string | null>(null);

  const batches = useRecoilValue(batchesState(selectedJobId || ""));
  const jobIds = useRecoilValue(jobIdsState);
  const refresh = useRecoilRefresher_UNSTABLE(jobInfoState(selectedJobId || ""));

  return (
    <Stack direction="column" w="full">
      <Flex gap={2}>
        <Select
          variant="filled"
          placeholder="Select jobs"
          onChange={(e) => setSelectedJobId(e.target.value)}
        >
          {jobIds.map((jobId) => {
            return (
              <option key={jobId} value={jobId} onChange={() => setSelectedJobId(jobId)}>
                {jobId}
              </option>
            );
          })}
        </Select>
        <IconButton onClick={refresh} aria-label="refresh" icon={<MdRefresh />} />
      </Flex>
      {selectedJobId !== null && <DataTable columns={batchColumns} data={batches} />}
    </Stack>
  );
};
