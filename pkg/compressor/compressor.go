package compressor

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type CompressionStats struct {
	OriginalSize     int64
	CompressedSize   int64
	CompressionRatio float64
	TimeTaken        time.Duration
}

// FileCompressor handles file compression operations
type FileCompressor struct {
	compressionLevel int
}

// NewFileCompressor creates a new file compressor with specified compression level
func NewFileCompressor(level int) *FileCompressor {
	if level < 1 || level > 9 {
		level = 6 // Default compression level
	}
	return &FileCompressor{compressionLevel: level}
}

// PrintStats prints compression statistics
func PrintStats(stats *CompressionStats, operation string) {
	fmt.Printf("\n=== %s Statistics ===\n", operation)
	fmt.Printf("Original Size:    %d bytes (%.2f KB)\n",
		stats.OriginalSize, float64(stats.OriginalSize)/1024)
	fmt.Printf("Processed Size:   %d bytes (%.2f KB)\n",
		stats.CompressedSize, float64(stats.CompressedSize)/1024)

	if operation == "Compression" {
		fmt.Printf("Compression Ratio: %.2f%%\n", stats.CompressionRatio)
		savings := float64(stats.OriginalSize-stats.CompressedSize) / float64(stats.OriginalSize) * 100
		fmt.Printf("Space Saved:      %.2f%%\n", savings)
	} else {
		fmt.Printf("Expansion Ratio:  %.2f%%\n", stats.CompressionRatio)
	}

	fmt.Printf("Time Taken:       %v\n", stats.TimeTaken)
}

// PrintUsage prints usage information
func PrintUsage() {
	fmt.Println("Go File Compressor")
	fmt.Println("Usage:")
	fmt.Println("  go run main.go compress <input_file> <output_file> [compression_level]")
	fmt.Println("  go run main.go decompress <input_file> <output_file>")
	fmt.Println("  go run main.go compress-dir <input_directory> <output_directory> [compression_level]")
	fmt.Println()
	fmt.Println("Parameters:")
	fmt.Println("  compression_level: 1-9 (1=fastest, 9=best compression, default=6)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  go run main.go compress document.txt document.txt.gz")
	fmt.Println("  go run main.go compress document.txt document.txt.gz 9")
	fmt.Println("  go run main.go decompress document.txt.gz document_restored.txt")
	fmt.Println("  go run main.go compress-dir ./documents ./compressed_docs")
}

// CompressFile compresses a single file using gzip compression
func (fc *FileCompressor) CompressFile(inputPath, outputPath string) (*CompressionStats, error) {
	startTime := time.Now()

	// Open input file
	inputFile, err := os.Open(inputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open input file: %v", err)
	}
	defer inputFile.Close()

	// Get input file info
	inputInfo, err := inputFile.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get input file info: %v", err)
	}

	// Create output file
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create output file: %v", err)
	}
	defer outputFile.Close()

	// Create gzip writer with specified compression level
	gzWriter, err := gzip.NewWriterLevel(outputFile, fc.compressionLevel)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip writer: %v", err)
	}
	defer gzWriter.Close()

	// Set gzip header
	gzWriter.Name = filepath.Base(inputPath)
	gzWriter.ModTime = inputInfo.ModTime()

	// Copy and compress data
	compressedSize, err := io.Copy(gzWriter, inputFile)
	if err != nil {
		return nil, fmt.Errorf("failed to compress file: %v", err)
	}

	// Close gzip writer to flush remaining data
	if err := gzWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close gzip writer: %v", err)
	}

	// Calculate compression ratio
	originalSize := inputInfo.Size()
	compressionRatio := float64(compressedSize) / float64(originalSize) * 100

	stats := &CompressionStats{
		OriginalSize:     originalSize,
		CompressedSize:   compressedSize,
		CompressionRatio: compressionRatio,
		TimeTaken:        time.Since(startTime),
	}

	return stats, nil
}

// DecompressFile decompresses a gzip file
func (fc *FileCompressor) DecompressFile(inputPath, outputPath string) (*CompressionStats, error) {
	startTime := time.Now()

	// Open compressed file
	inputFile, err := os.Open(inputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open compressed file: %v", err)
	}
	defer inputFile.Close()

	// Get compressed file info
	inputInfo, err := inputFile.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get compressed file info: %v", err)
	}

	// Create gzip reader
	// Create gzip reader that decompresses data as it reads
	gzReader, err := gzip.NewReader(inputFile)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %v", err)
	}
	defer gzReader.Close()

	// Create output file
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create output file: %v", err)
	}
	defer outputFile.Close()

	// Decompress data
	decompressedSize, err := io.Copy(outputFile, gzReader)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress file: %v", err)
	}

	// Restore original modification time if available
	if !gzReader.ModTime.IsZero() {
		os.Chtimes(outputPath, time.Now(), gzReader.ModTime)
	}

	stats := &CompressionStats{
		OriginalSize:     inputInfo.Size(),
		CompressedSize:   decompressedSize,
		CompressionRatio: float64(decompressedSize) / float64(inputInfo.Size()) * 100,
		TimeTaken:        time.Since(startTime),
	}

	return stats, nil
}

// CompressDirectory compresses multiple files in a directory
func (fc *FileCompressor) CompressDirectory(inputDir, outputDir string) error {
	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	// Walk through the input directory
	return filepath.Walk(inputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Calculate relative path
		relPath, err := filepath.Rel(inputDir, path)
		if err != nil {
			return err
		}

		// Create output path
		outputPath := filepath.Join(outputDir, relPath+".gz")

		// Create output directory structure
		outputDirPath := filepath.Dir(outputPath)
		if err := os.MkdirAll(outputDirPath, 0755); err != nil {
			return err
		}

		// Compress file
		stats, err := fc.CompressFile(path, outputPath)
		if err != nil {
			fmt.Printf("Error compressing %s: %v\n", path, err)
			return nil // Continue with other files
		}

		fmt.Printf("Compressed: %s -> %s (%.2f%% of original)\n",
			path, outputPath, stats.CompressionRatio)

		return nil
	})
}

// Decompressing a folder with compressed files
func (fc *FileCompressor) DecompressDirectory(inputDir, outputDir string) error {

	// if outputDir doesn't exist create it
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("Failed to create output directory:  %v\n", err)
	}

	// Walk throught the input directory
	return filepath.Walk(inputDir, func(path string, info os.FileInfo, err error) error {

		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Calculate relative path
		relPath, err := filepath.Rel(inputDir, path)
		if err != nil {
			return err
		}

		// Create output path
		outputPath := filepath.Join(outputDir, strings.Split(relPath, ".gz")[0])

		// fmt.Printf("outputPath %v:\n",outputPath)
		// fmt.Printf("relPath %v:\n",relPath)
		// fmt.Printf("asdsadasd %v:\n",strings.Split(relPath, ".gz"))

		// Create output directory structure
		outputDirPath := filepath.Dir(outputPath)
		if err := os.MkdirAll(outputDirPath, 0755); err != nil {
			return err
		}

		stats, err := fc.DecompressFile(path, outputPath)
		if err != nil {
			fmt.Printf("Error decompressing %s: %v\n", path, err)
			return nil // Continue with other files
		}

		fmt.Printf("De-Compressed: %s -> %s (%.2f%% of original)\n",
			path, outputPath, stats.CompressionRatio)

		return nil

	})
}
