package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	_ "image/gif"   // Import the GIF decoder
	_ "image/jpeg"  // Import the JPEG decoder
	_ "image/png"   // Import the PNG decoder
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Flier struct {
	ID       string   `json:"id"`
	Design   Design   `json:"design"`
	Language string   `json:"language"`
	URL      string   `json:"url"`
	Data     Data     `json:"data"`
}

type Design struct {
	TemplateID  string     `json:"templateId"`
	Resolution  Resolution `json:"resolution"`
	Type        string     `json:"type"`
	Tags        []string   `json:"tags"`
	FileFormat  string     `json:"fileFormat"`
	Orientation string     `json:"orientation"`
}

type Resolution struct {
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Unit   string `json:"unit"`
}

type Data struct {
	ImageBase64 string `json:"imageBase64"`
}

type Metadata struct {
	Version     string   `json:"version"`
	LastUpdated string   `json:"lastUpdated"`
	TotalFliers int      `json:"totalFliers"`
	URL         string   `json:"url"`
	Schema      Schema   `json:"schema"`
}

type Schema struct {
	Format   string `json:"format"`
	Encoding string `json:"encoding"`
	Filetype string `json:"filetype"`
}

type FliersData struct {
	Fliers   []Flier  `json:"fliers"`
	Metadata Metadata `json:"metadata"`
}

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

func main() {
	// Folder containing image files
	imageFolder := "images/" // Replace with your folder path
	folderURL := "path/to/fliers"   // Replace with the URL path

	// Read all files in the directory
	files, err := os.ReadDir(imageFolder)
	if err != nil {
		fmt.Println("Error reading directory:", err)
		return
	}

	var fliers []Flier
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

		// Read the image file
		imageBytes, err := os.ReadFile(imagePath)
		if err != nil {
			fmt.Printf("Error reading image %s: %v\n", file.Name(), err)
			continue
		}

		// Convert image to Base64
		imageBase64 := base64.StdEncoding.EncodeToString(imageBytes)
		imageData := fmt.Sprintf("data:image/%s;base64,%s", strings.ToLower(fileExt), imageBase64)

		// Add flier details
		flier := Flier{
			ID:       strings.TrimSuffix(file.Name(), filepath.Ext(file.Name())), // Use file name as ID without extension
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
			Data:     Data{ImageBase64: imageData},
		}

		// Append flier to the list
		fliers = append(fliers, flier)
	}

	// Create metadata
	metadata := Metadata{
		Version:     "1.0",
		LastUpdated: time.Now().Format(time.RFC3339),
		TotalFliers: len(fliers),
		URL:         folderURL,
		Schema: Schema{
			Format:   "JSON",
			Encoding: "UTF-8",
			Filetype: "text",
		},
	}

	// Create the JSON structure
	fliersData := FliersData{
		Fliers:   fliers,
		Metadata: metadata,
	}

	// Convert to JSON
	jsonData, err := json.MarshalIndent(fliersData, "", "  ")
	if err != nil {
		fmt.Println("Error encoding JSON:", err)
		return
	}

	// Print or save JSON output
	fmt.Println(string(jsonData))

	// Optionally, save the JSON to a file
	err = os.WriteFile("output.json", jsonData, 0644)
	if err != nil {
		fmt.Println("Error writing JSON file:", err)
		return
	}

	fmt.Println("JSON saved successfully as output.json")
}
