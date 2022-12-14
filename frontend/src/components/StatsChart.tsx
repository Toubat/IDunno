import { ColorMode, Flex, IconButton, Select, Stack, useColorMode } from "@chakra-ui/react";
import { useEffect, useState } from "react";
import { useRecoilRefresher_UNSTABLE, useRecoilValue } from "recoil";
import { jobIdsState, jobInfoState, queryProcessTimesState, queryRatesState } from "../states";
// @ts-ignore
import ReactApexChart from "react-apexcharts";
import { ApexOptions } from "apexcharts";
import { MdRefresh } from "react-icons/md";
import { median, percentile } from "stats-lite";

const getChartConfig = (
  queryRates: number[],
  jobId: string,
  colorMode: ColorMode = "light",
  title: string
) => {
  const upper = colorMode === "light" ? "#4299E1" : "#38B2AC";
  const lower = colorMode === "light" ? "#90CDF4" : "#81E6D9";
  const textColor = colorMode === "light" ? "#2D3748" : "#EDF2F7";

  const min = Math.min(...queryRates);
  const max = Math.max(...queryRates);
  const medianValue = median(queryRates);
  const q1 = percentile(queryRates, 0.25);
  const q3 = percentile(queryRates, 0.75);

  const options: ApexOptions = {
    chart: {
      type: "boxPlot",
    },
    title: {
      text: `${title} Box Plot for job ${jobId}`,
      align: "left",
      style: {
        color: textColor,
      },
    },
    plotOptions: {
      boxPlot: {
        colors: {
          upper,
          lower,
        },
      },

      bar: {
        horizontal: true,
      },
    },
    theme: {
      mode: colorMode,
    },
  };

  return {
    series: [
      {
        type: "boxPlot",
        data: [
          {
            x: title,
            y: [min, q1, medianValue, q3, max],
          },
        ],
        style: {},
      },
    ],
    options,
  };
};

export const StatsChart = () => {
  const { colorMode } = useColorMode();
  const [selectedJobId, setSelectedJobId] = useState<string | null>(null);
  const jobIds = useRecoilValue(jobIdsState);
  const queryRates = useRecoilValue(queryRatesState(selectedJobId || ""));
  const queryProcessTimes = useRecoilValue(queryProcessTimesState(selectedJobId || ""));
  const refresh = useRecoilRefresher_UNSTABLE(jobInfoState(selectedJobId || ""));

  const queryRateConfig = getChartConfig(
    queryRates.filter((rate) => rate > 0.1),
    selectedJobId || "",
    colorMode,
    "Query Rate"
  );

  const queryProcessTimeConfig = getChartConfig(
    queryProcessTimes.filter((t) => t != 1),
    selectedJobId || "",
    colorMode,
    "Query Process Time"
  );

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
      <ReactApexChart
        options={queryRateConfig.options}
        series={queryRateConfig.series}
        type="boxPlot"
        height={225}
      />
      <ReactApexChart
        options={queryProcessTimeConfig.options}
        series={queryProcessTimeConfig.series}
        type="boxPlot"
        height={225}
      />
    </Stack>
  );
};
