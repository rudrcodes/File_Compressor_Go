package main

import (
	"fmt"
	"os"

	comp "github.com/rudrcodes/File_Compressor_Go/pkg/compressor"
)

//Commands to run various functions :

/*
* # Compress a single file
* go run main.go compress document.txt document.txt.gz

* # Compress with maximum compression
* go run main.go compress document.txt document.txt.gz 9

* # Decompress a file
* go run main.go decompress document.txt.gz restored.txt

* # Compress entire directory
* go run main.go compress-dir ./documents ./compressed_docs

 */
// CompressionStats holds statistics about the compression operation

func main() {

	// first get the option from the user what it has to do and then the filenames based on what the user has to do

	// var option string

	var command string
	var inputPath string
	var outputPath string
	fmt.Println("What to do today?")
	fmt.Println("** OPTIONS **")
	fmt.Println("1) Compress File")
	fmt.Println("2) Compress Folder")
	fmt.Println("3) De-Compress File")
	fmt.Println("4) De-Compress Folder")

	//ask the user to input an option

	m := make(map[string]string)
	m["1"] = "compress"
	m["2"] = "compress-dir"
	m["3"] = "decompress"
	m["4"] = "decompress-dir"

	fmt.Scanln(&command)

	var inputType string

	if command == "1" || command == "3" {
		//take file names
		inputType = "File"
	} else {
		//take folder names
		inputType = "Folder"

	}

	fmt.Printf("Enter %s names : \n", inputType)
	fmt.Printf("Input %s  : \n", inputType)
	fmt.Scanln(&inputPath)
	fmt.Printf("Output %s  : \n", inputType)
	fmt.Scanln(&outputPath)

	// fmt.Println("os.args : ", os.Args)
	// if len(os.Args) < 4 {
	// 	PrintUsage()
	// 	os.Exit(1)
	// }

	// command := os.Args[1]
	// inputPath := os.Args[2]
	// outputPath := os.Args[3]

	// Default compression level

	fmt.Println("Command:", command)
	fmt.Println("inputPath:", inputPath)
	fmt.Println("outputPath:", outputPath)
	compressionLevel := 6

	// Parse compression level if provided
	if len(os.Args) > 4 && (command == "compress" || command == "compress-dir") {
		if level := os.Args[4]; level != "" {
			if l := int(level[0] - '0'); l >= 1 && l <= 9 {
				compressionLevel = l
			}
		}
	}

	// creates a FileCompressor struct
	compressor := comp.NewFileCompressor(compressionLevel)

	switch m[command] {
	case "compress":
		fmt.Printf("Compressing %s to %s (level %d)...\n", inputPath, outputPath, compressionLevel)

		stats, err := compressor.CompressFile(inputPath, outputPath)
		if err != nil {
			fmt.Printf("Error: %v\n", err)

			// This function immediately terminates your program and returns an exit code to the operating system.
			os.Exit(1)
		}

		fmt.Printf("✓ Compression completed successfully!\n")

		comp.PrintStats(stats, "Compression")

	case "decompress":
		fmt.Printf("Decompressing %s to %s...\n", inputPath, outputPath)

		stats, err := compressor.DecompressFile(inputPath, outputPath)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✓ Decompression completed successfully!\n")
		comp.PrintStats(stats, "Decompression")

	case "compress-dir":
		fmt.Printf("Compressing directory %s to %s (level %d)...\n", inputPath, outputPath, compressionLevel)

		err := compressor.CompressDirectory(inputPath, outputPath)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✓ Directory compression completed successfully!\n")

	case "decompress-dir":
		fmt.Printf("Decompressing Directory %s to %s...\n: ", inputPath, outputPath)
		err := compressor.DecompressDirectory(inputPath, outputPath)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("✓ Decompression completed successfully!\n")
		// stats, err := compressor.DecompressDirectory(inputPath, outputPath)
		// if err != nil {
		// 	fmt.Printf("Error: %v\n", err)
		// 	os.Exit(1)
		// }

		// fmt.Println("✓ Decompression completed successfully!\n")
		// PrintStats(stats, "Decompression")

	default:
		fmt.Printf("Unknown command: %s\n", command)
		comp.PrintUsage()

		os.Exit(1)
	}
}
