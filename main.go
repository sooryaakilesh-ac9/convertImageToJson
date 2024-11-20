package main

import (
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg" // Import the JPEG decoder
	_ "image/png"  // Import the PNG decoder
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Flyer
type Flyer struct {
	ID       string `json:"id"`
	Design   Design `json:"design"`
	Language string `json:"lang"`
	URL      string `json:"url"`
}

// Design object within flyer
type Design struct {
	TemplateID  string     `json:"templateId"`
	Resolution  Resolution `json:"resolution"`
	Type        string     `json:"type"`
	Tags        []string   `json:"tags"`
	FileFormat  string     `json:"fileFormat"`
	Orientation string     `json:"orientation"`
}

// Resolution object within flyer
type Resolution struct {
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Unit   string `json:"unit"`
}

// TODO create a separate json file imaageMetadata.json
type Metadata struct {
	Version     string `json:"version"`
	LastUpdated string `json:"lastUpdated"`
	TotalFlyers int    `json:"totalFlyers"`
	URL         string `json:"url"`
	Schema      Schema `json:"schema"`
}

// Schema object within flyer
type Schema struct {
	Format   string `json:"format"`
	Encoding string `json:"encoding"`
	Filetype string `json:"filetype"`
}

type FlyersData struct {
	Flyers []Flyer `json:"Flyers"`
}

// gets the resolution and orientation of the image
func getImageResolutionAndOrientation(imagePath string) (int, int, string, string, error) {
	// Open the image file
	file, err := os.Open(imagePath)
	if err != nil {
		return 0, 0, "", "", err
	}
	defer file.Close()

	// Decode the image
	img, _, err := image.Decode(file)
	if err != nil {
		return 0, 0, "", "", err
	}

	// Get the width and height of the image
	width := img.Bounds().Dx()
	height := img.Bounds().Dy()

	// Determine file format
	fileExt := strings.ToUpper(strings.TrimPrefix(filepath.Ext(imagePath), "."))
	if fileExt == "JPG" {
		fileExt = "JPEG"
	}

	// Determine orientation
	var orientation string
	if width > height {
		orientation = "landscape"
	} else {
		orientation = "portrait"
	}

	return width, height, fileExt, orientation, nil
}

func processBatch(files []os.DirEntry, imageFolder, folderURL string, wg *sync.WaitGroup, results chan<- Flyer) {
	defer wg.Done()

	for _, file := range files {
		// Only process files (ignore directories)
		if file.IsDir() {
			continue
		}

		// Construct full path for the file
		imagePath := filepath.Join(imageFolder, file.Name())

		// Get image resolution, format, and orientation
		width, height, fileExt, orientation, err := getImageResolutionAndOrientation(imagePath)
		if err != nil {
			fmt.Printf("Error processing image %s: %v\n", file.Name(), err)
			continue
		}

		// Create Flyer object
		Flyer := Flyer{
			ID: strings.TrimSuffix(file.Name(), filepath.Ext(file.Name())), // Use file name as ID without extension
			Design: Design{
				TemplateID:  "template1",
				Resolution:  Resolution{Width: width, Height: height, Unit: "px"},
				Type:        "image",
				Tags:        []string{""},
				FileFormat:  fileExt,
				Orientation: orientation,
			},
			Language: "en-US",
			URL:      filepath.ToSlash(imagePath),
		}

		// Send the Flyer to the channel
		results <- Flyer
	}
}

func batchProcess(files []os.DirEntry, imageFolder, folderURL string, batchSize int) ([]Flyer, error) {
	var wg sync.WaitGroup
	results := make(chan Flyer, len(files))
	var flyers []Flyer

	// Process files in batches
	for i := 0; i < len(files); i += batchSize {
		// Determine the batch slice
		end := i + batchSize
		if end > len(files) {
			end = len(files)
		}
		wg.Add(1)
		go processBatch(files[i:end], imageFolder, folderURL, &wg, results)
	}

	// Wait for all goroutines to finish
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect all results from the channel
	for flyer := range results {
		flyers = append(flyers, flyer)
	}

	return flyers, nil
}

func main() {
	imageFolder := "images/"
	folderURL := "path/to/Flyers"
	batchSize := 10

	// Read all files in the directory
	files, err := os.ReadDir(imageFolder)
	if err != nil {
		fmt.Println("Error reading directory:", err)
		return
	}

	// Process files in batches
	flyers, err := batchProcess(files, imageFolder, folderURL, batchSize)
	if err != nil {
		fmt.Println("Error processing files:", err)
		return
	}

	// Create metadata
	metadata := Metadata{
		Version:     "1.0",
		LastUpdated: time.Now().Format(time.RFC3339),
		TotalFlyers: len(flyers),
		URL:         folderURL,
		Schema: Schema{
			Format:   "JSON",
			Encoding: "UTF-8",
			Filetype: "text",
		},
	}

	// Create the JSON structure
	FlyersData := FlyersData{
		Flyers: flyers,
	}

	// Convert flyer to JSON
	jsonData, err := json.MarshalIndent(FlyersData, "", "  ")
	if err != nil {
		fmt.Println("Error encoding JSON:", err)
		return
	}

	// Convert metda data to json
	metaJsonData, err := json.MarshalIndent(metadata, "", " ")
	if err != nil {
		fmt.Println("Error encoding JSON:", err)
		return
	}

	// Print or save JSON output
	fmt.Println(string(jsonData))
	fmt.Println(string(metaJsonData))

	// Save the flyer JSON to a file
	if err = os.WriteFile("flyers.json", jsonData, 0644); err != nil {
		fmt.Println("Error writing flyer JSON file", err)
		return
	}

	// Save the metadata to JSON file
	if err := os.WriteFile("imagesMetadata.json", metaJsonData, 0644); err != nil {
		fmt.Println("Error writing metadata JSON file", err)
		return
	}

	fmt.Println("JSON saved successfully as output.json")
}
