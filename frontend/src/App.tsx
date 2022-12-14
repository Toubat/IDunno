import {
  Box,
  Flex,
  useColorModeValue,
  Container,
  Card,
  CardHeader,
  Heading,
  CardBody,
  Tabs,
  TabList,
  Tab,
  ButtonGroup,
  IconButton,
  Tooltip,
  CircularProgress,
  Stack,
} from "@chakra-ui/react";
import { Suspense, useMemo } from "react";
import { ErrorBoundary } from "react-error-boundary";
import { IconType } from "react-icons";
import { FaServer, FaTasks } from "react-icons/fa";
import { AiOutlineFileDone } from "react-icons/ai";
import { ImStatsBars } from "react-icons/im";
import { GrDocumentText } from "react-icons/gr";
import { useRecoilState } from "recoil";
import { AppBar } from "./components/AppBar";
import { CompletedJobTable } from "./components/CompletedJobTable";
import { ErrorFallback } from "./components/ErrorFallback";
import { InfoTable } from "./components/InfoTable";
import { JobTable } from "./components/JobTable";
import { WorkerTable } from "./components/WorkerTable";
import { tabIndexState } from "./states";
import { StatsChart } from "./components/StatsChart";
import "./index.css";

export interface TabInfo {
  name: string;
  Icon: IconType;
  Component: React.FC<any>;
}

const tabs: TabInfo[] = [
  { name: "Jobs", Icon: FaTasks, Component: JobTable },
  { name: "Workers", Icon: FaServer, Component: WorkerTable },
  { name: "Completed", Icon: AiOutlineFileDone, Component: CompletedJobTable },
  { name: "Output", Icon: GrDocumentText, Component: InfoTable },
  { name: "Stats", Icon: ImStatsBars, Component: StatsChart },
];

function App() {
  const appBgColor = useColorModeValue("gray.100", "gray.800");
  const dashboardBgColor = useColorModeValue("gray.50", "gray.900");
  const dividerBorderColor = useColorModeValue("gray.200", "gray.700");
  const tabColor = useColorModeValue("gray.600", "gray.300");
  const iconActiveColor = useColorModeValue("gray.200", "gray.700");

  const [tabIndex, setTabIndex] = useRecoilState(tabIndexState);

  const TableComponent = useMemo(() => {
    return tabs[tabIndex].Component;
  }, [tabIndex]);

  return (
    <Box bg={appBgColor} minH="100vh">
      <AppBar />
      <Box h={16} />
      <Container maxW="9xl" pb={4}>
        <Card bg={dashboardBgColor} m={2} my={4} size="md">
          <CardHeader display="flex" justifyContent="space-between" alignItems="center">
            <Heading size="md">Dashboard</Heading>
            <Stack direction="row">
              <ButtonGroup isAttached>
                {tabs.map((tab, idx) => (
                  <Tooltip key={tab.name} hasArrow label={tab.name}>
                    <IconButton
                      onClick={() => setTabIndex(idx)}
                      size="md"
                      variant="outline"
                      aria-label={tab.name}
                      icon={<tab.Icon />}
                      bg={tabIndex === idx ? iconActiveColor : "transparent"}
                    ></IconButton>
                  </Tooltip>
                ))}
              </ButtonGroup>
            </Stack>
          </CardHeader>
          <CardBody>
            <Flex gap={4}>
              <Tabs
                size="md"
                variant="soft-rounded"
                orientation="vertical"
                onChange={setTabIndex}
                index={tabIndex}
                display={{ base: "none", lg: "block" }}
              >
                <TabList display={"flex"}>
                  {tabs.map((tab) => (
                    <Tab key={tab.name} rounded="md" color={tabColor}>
                      {tab.name}
                    </Tab>
                  ))}
                </TabList>
              </Tabs>
              <Box
                className="border-l-[1px]"
                borderColor={dividerBorderColor}
                display={{ base: "none", lg: "block" }}
              />
              <ErrorBoundary
                FallbackComponent={ErrorFallback}
                onReset={() => {
                  window.location.reload();
                }}
              >
                <Suspense
                  fallback={
                    <Flex className="border" w="full" minH="30" justify="center" align="center">
                      <CircularProgress my={4} size="16" isIndeterminate color="blue.500" />
                    </Flex>
                  }
                >
                  <TableComponent />
                </Suspense>
              </ErrorBoundary>
            </Flex>
          </CardBody>
        </Card>
      </Container>
    </Box>
  );
}

export default App;
