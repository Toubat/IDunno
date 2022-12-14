import { Alert, AlertDescription, AlertIcon, AlertTitle, Button, Code } from "@chakra-ui/react";

export const ErrorFallback: React.FC<any> = ({ error, resetErrorBoundary }) => {
  return (
    <Alert
      status="error"
      variant="subtle"
      flexDirection="column"
      alignItems="center"
      justifyContent="center"
      textAlign="center"
      height="200px"
      w="100vw"
      h="100vh"
      maxH="400"
      overflowY="auto"
    >
      <AlertIcon boxSize="40px" mr={0} />
      <AlertTitle mt={4} mb={1} fontSize="lg">
        Something went wrong:
      </AlertTitle>
      <AlertDescription maxWidth="sm">
        <Code colorScheme="red">{error.message}</Code>
      </AlertDescription>
      <Button mt={4} size="sm" colorScheme="red" onClick={resetErrorBoundary}>
        Try Again
      </Button>
    </Alert>
  );
};
