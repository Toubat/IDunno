import { MoonIcon, SunIcon } from "@chakra-ui/icons";
import { useColorMode, Text, Image, Flex, IconButton, useColorModeValue } from "@chakra-ui/react";
import logoLight from "../assets/logo-light.svg";
import logoDark from "../assets/logo-dark.svg";

export const AppBar: React.FC = () => {
  const { colorMode, toggleColorMode } = useColorMode();
  const logoColor = useColorModeValue("gray.700", "gray.50");
  const appbarBgColor = useColorModeValue("white", "gray.700");
  const appbarBorderColor = useColorModeValue("gray.200", "gray.700");
  const iconColor = useColorModeValue("gray.500", "gray.100");
  const logoSvg = useColorModeValue(logoLight, logoDark);

  return (
    <Flex
      bg={appbarBgColor}
      borderColor={appbarBorderColor}
      className="border-b-2"
      w="full"
      h={16}
      align="center"
      px={5}
      justify="space-between"
      pos="fixed"
      top={0}
      zIndex={999999}
    >
      <Flex>
        <Image className="w-9 h-9 mt-[0.3rem] mr-2" src={logoSvg} alt="Dan Abramov" />
        <Text
          className="mt-[0.13rem]"
          fontSize="2xl"
          color={logoColor}
          fontWeight="semibold"
          fontFamily="sans-serif"
        >
          IDunno
        </Text>
      </Flex>
      <IconButton
        onClick={toggleColorMode}
        aria-label="Toggle Theme"
        icon={
          colorMode === "light" ? <MoonIcon color={iconColor} /> : <SunIcon color={iconColor} />
        }
      />
    </Flex>
  );
};
