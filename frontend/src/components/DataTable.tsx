import { useEffect, useMemo, useState } from "react";
import {
  Table,
  Thead,
  Tbody,
  Tr,
  Th,
  Td,
  chakra,
  useColorModeValue,
  Card,
  Flex,
  Text,
  Stack,
  Menu,
  MenuButton,
  MenuList,
  MenuItem,
  Button,
  Box,
} from "@chakra-ui/react";
import { ChevronDownIcon, TriangleDownIcon, TriangleUpIcon } from "@chakra-ui/icons";
import {
  useReactTable,
  flexRender,
  getCoreRowModel,
  ColumnDef,
  SortingState,
  getSortedRowModel,
} from "@tanstack/react-table";

export type DataTableProps<Data extends object> = {
  data: Data[];
  columns: ColumnDef<Data, any>[];
  refresh?: () => void;
  render?: () => React.ReactNode;
};

export function DataTable<Data extends object>({
  data,
  columns,
  refresh = () => {},
  render,
}: DataTableProps<Data>) {
  const [sorting, setSorting] = useState<SortingState>([]);
  const theadBgColor = useColorModeValue("gray.200", "gray.600");
  const theadTextColor = useColorModeValue("gray.700", "gray.50");
  const theadBorderColor = useColorModeValue("gray.200", "gray.600");

  const table = useReactTable({
    columns,
    data,
    getCoreRowModel: getCoreRowModel(),
    onSortingChange: setSorting,
    getSortedRowModel: getSortedRowModel(),
    state: { sorting },
  });

  const [page, setPage] = useState<number>(0);
  const [pageSize, setPageSize] = useState<number>(10);

  const maxPages = useMemo(() => Math.ceil(table.getRowModel().rows.length / pageSize), [pageSize]);
  const pageRows = useMemo(() => {
    const start = page * pageSize;
    const end = start + pageSize;
    return table.getRowModel().rows.slice(start, end);
  }, [pageSize, page, data]);

  useEffect(() => {
    const interval = setInterval(() => {
      refresh();
    }, 1000);

    return () => {
      clearInterval(interval);
    };
  }, []);

  return (
    <Card w="full" variant="outline" overflowY="hidden">
      {render && (
        <Box px={4} pt={4} pb={2}>
          {render()}
        </Box>
      )}
      <Box overflowX="auto">
        <Table variant="striped" colorScheme="gray">
          <Thead className="border-b-[2px]" borderColor={theadBorderColor} bg={theadBgColor}>
            {table.getHeaderGroups().map((headerGroup) => (
              <Tr key={headerGroup.id}>
                {headerGroup.headers.map((header) => {
                  // see https://tanstack.com/table/v8/docs/api/core/column-def#meta to type this correctly
                  const meta: any = header.column.columnDef.meta;
                  return (
                    <Th
                      key={header.id}
                      onClick={header.column.getToggleSortingHandler()}
                      isNumeric={meta?.isNumeric}
                      color={theadTextColor}
                    >
                      {flexRender(header.column.columnDef.header, header.getContext())}

                      <chakra.span pl="4">
                        {header.column.getIsSorted() ? (
                          header.column.getIsSorted() === "desc" ? (
                            <TriangleDownIcon aria-label="sorted descending" />
                          ) : (
                            <TriangleUpIcon aria-label="sorted ascending" />
                          )
                        ) : null}
                      </chakra.span>
                    </Th>
                  );
                })}
              </Tr>
            ))}
          </Thead>
          <Tbody>
            {pageRows.map((row) => (
              <Tr key={row.id}>
                {row.getVisibleCells().map((cell) => {
                  // see https://tanstack.com/table/v8/docs/api/core/column-def#meta to type this correctly
                  const meta: any = cell.column.columnDef.meta;
                  return (
                    <Td key={cell.id} textAlign={meta?.align}>
                      {flexRender(cell.column.columnDef.cell, cell.getContext())}
                    </Td>
                  );
                })}
              </Tr>
            ))}
          </Tbody>
        </Table>
      </Box>
      <Flex p={4} align="center" justify="space-between">
        <Text>
          Show {page * pageSize + 1} to {Math.min(page * pageSize + pageSize, data.length)} results
          | Total {data.length}
        </Text>
        <Stack direction="row">
          <Menu>
            <MenuButton as={Button} rightIcon={<ChevronDownIcon />} size="sm">
              {pageSize}
            </MenuButton>
            <MenuList>
              {[10, 25, 50, 100, 200, 500, 1000].map((size) => (
                <MenuItem key={size} onClick={() => setPageSize(size)}>
                  {size}
                </MenuItem>
              ))}
            </MenuList>
          </Menu>
          <Button
            variant="outline"
            size="sm"
            onClick={() => setPage(page - 1)}
            disabled={page === 0}
          >
            Previous
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={() => setPage(page + 1)}
            disabled={page === maxPages - 1}
          >
            Next
          </Button>
        </Stack>
      </Flex>
    </Card>
  );
}
