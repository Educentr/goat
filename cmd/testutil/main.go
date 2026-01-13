package main

import (
	"fmt"
	"os"

	"github.com/Educentr/goat/testutil"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "generate-env":
		generateEnv()
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: testutil <command>")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  generate-env [input] [output]  Convert TREE.conf to .env format")
	fmt.Println("                                 Default: tests/etc/onlineconf/TREE.conf -> onlineconf.env")
	fmt.Println("  help                           Show this help")
}

func generateEnv() {
	inputPath := "tests/etc/onlineconf/TREE.conf"
	outputPath := "tests/etc/onlineconf/onlineconf.env"

	if len(os.Args) > 2 {
		inputPath = os.Args[2]
	}
	if len(os.Args) > 3 {
		outputPath = os.Args[3]
	}

	envVars, err := testutil.ParseOnlineConfFile(inputPath)
	if err != nil {
		fmt.Printf("Error parsing %s: %v\n", inputPath, err)
		os.Exit(1)
	}

	if err := testutil.WriteEnvFile(outputPath, envVars); err != nil {
		fmt.Printf("Error writing %s: %v\n", outputPath, err)
		os.Exit(1)
	}

	fmt.Printf("Generated %s with %d variables\n", outputPath, len(envVars))
}
