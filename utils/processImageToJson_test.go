package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

// createTestImage creates a valid test image file with the specified dimensions
func createTestImage(path string, width, height int, format string) error {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	var buf bytes.Buffer
	var err error

	switch format {
	case "JPEG", "JPG":
		err = jpeg.Encode(&buf, img, nil)
	case "PNG":
		err = png.Encode(&buf, img)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	if err != nil {
		return err
	}

	return os.WriteFile(path, buf.Bytes(), 0644)
}

// TestFlyer tests the Flyer struct JSON marshaling/unmarshaling
func TestFlyer(t *testing.T) {
	flyer := Flyer{
		ID: "test123",
		Design: Design{
			TemplateID: "template1",
			Resolution: Resolution{
				Width:  1920,
				Height: 1080,
				Unit:   "px",
			},
			Type:        "image",
			Tags:        []string{"test"},
			FileFormat:  "JPEG",
			Orientation: "landscape",
		},
		Language: "en-US",
		URL:      "http://example.com/test.jpg",
	}

	jsonData, err := json.Marshal(flyer)
	if err != nil {
		t.Errorf("Failed to marshal Flyer: %v", err)
	}

	var unmarshaledFlyer Flyer
	err = json.Unmarshal(jsonData, &unmarshaledFlyer)
	if err != nil {
		t.Errorf("Failed to unmarshal Flyer: %v", err)
	}

	if !reflect.DeepEqual(flyer, unmarshaledFlyer) {
		t.Errorf("Unmarshaled Flyer does not match original:\nExpected: %+v\nGot: %+v", flyer, unmarshaledFlyer)
	}
}

// TestGetImageResolutionAndOrientation tests image processing functionality
func TestGetImageResolutionAndOrientation(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test-images")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name           string
		createTestFile func(string) string
		wantWidth      int
		wantHeight     int
		wantFormat     string
		wantOrient     string
		wantErr        bool
	}{
		{
			name: "valid_landscape_jpeg",
			createTestFile: func(dir string) string {
				path := filepath.Join(dir, "landscape.jpg")
				err := createTestImage(path, 1920, 1080, "JPEG")
				if err != nil {
					t.Fatalf("Failed to create test image: %v", err)
				}
				return path
			},
			wantWidth:  1920,
			wantHeight: 1080,
			wantFormat: "JPEG",
			wantOrient: "landscape",
			wantErr:    false,
		},
		{
			name: "valid_portrait_jpeg",
			createTestFile: func(dir string) string {
				path := filepath.Join(dir, "portrait.jpg")
				err := createTestImage(path, 1080, 1920, "JPEG")
				if err != nil {
					t.Fatalf("Failed to create test image: %v", err)
				}
				return path
			},
			wantWidth:  1080,
			wantHeight: 1920,
			wantFormat: "JPEG",
			wantOrient: "portrait",
			wantErr:    false,
		},
		{
			name: "invalid_image",
			createTestFile: func(dir string) string {
				path := filepath.Join(dir, "invalid.jpg")
				err := os.WriteFile(path, []byte("invalid image data"), 0644)
				if err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
				return path
			},
			wantErr: true,
		},
		{
			name: "non_existent_file",
			createTestFile: func(dir string) string {
				return filepath.Join(dir, "nonexistent.jpg")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.createTestFile(tempDir)
			width, height, format, orient, err := getImageResolutionAndOrientation(path)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if width != tt.wantWidth {
				t.Errorf("Wrong width: got %d, want %d", width, tt.wantWidth)
			}
			if height != tt.wantHeight {
				t.Errorf("Wrong height: got %d, want %d", height, tt.wantHeight)
			}
			if format != tt.wantFormat {
				t.Errorf("Wrong format: got %s, want %s", format, tt.wantFormat)
			}
			if orient != tt.wantOrient {
				t.Errorf("Wrong orientation: got %s, want %s", orient, tt.wantOrient)
			}
		})
	}
}

// TestBatchProcess tests the batch processing functionality
func TestBatchProcess(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test-batch")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create valid test images
	testFiles := []struct {
		name   string
		width  int
		height int
		format string
	}{
		{"test1.jpg", 1920, 1080, "JPEG"},
		{"test2.jpg", 1080, 1920, "JPEG"},
		{"test3.jpg", 800, 600, "JPEG"},
	}

	for _, tf := range testFiles {
		path := filepath.Join(tempDir, tf.name)
		err := createTestImage(path, tf.width, tf.height, tf.format)
		if err != nil {
			t.Fatalf("Failed to create test image %s: %v", tf.name, err)
		}
	}

	files, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("Failed to read directory: %v", err)
	}

	tests := []struct {
		name      string
		batchSize int
		wantLen   int
		wantErr   bool
	}{
		{
			name:      "normal_batch",
			batchSize: 2,
			wantLen:   3,
			wantErr:   false,
		},
		{
			name:      "single_item_batch",
			batchSize: 1,
			wantLen:   3,
			wantErr:   false,
		},
		{
			name:      "batch_larger_than_input",
			batchSize: 10,
			wantLen:   3,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flyers, err := BatchProcess(files, tempDir, "http://example.com", tt.batchSize)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(flyers) != tt.wantLen {
				t.Errorf("Wrong number of processed items: got %d, want %d", len(flyers), tt.wantLen)
			}

			// Verify flyer properties
			for _, flyer := range flyers {
				if flyer.Design.FileFormat != "JPEG" {
					t.Errorf("Wrong file format: got %s, want JPEG", flyer.Design.FileFormat)
				}
				if flyer.Design.Resolution.Unit != "px" {
					t.Errorf("Wrong resolution unit: got %s, want px", flyer.Design.Resolution.Unit)
				}
			}
		})
	}
}

// TestReadAndForward tests the main processing function
func TestReadAndForward(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test-read-forward")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create valid test images
	testFiles := []struct {
		name   string
		width  int
		height int
		format string
	}{
		{"test1.jpg", 1920, 1080, "JPEG"},
		{"test2.jpg", 1080, 1920, "JPEG"},
	}

	for _, tf := range testFiles {
		path := filepath.Join(tempDir, tf.name)
		err := createTestImage(path, tf.width, tf.height, tf.format)
		if err != nil {
			t.Fatalf("Failed to create test image %s: %v", tf.name, err)
		}
	}

	tests := []struct {
		name      string
		imageDir  string
		folderURL string
		batchSize int
		wantErr   bool
	}{
		{
			name:      "valid_directory",
			imageDir:  tempDir,
			folderURL: "http://example.com",
			batchSize: 2,
			wantErr:   false,
		},
		{
			name:      "invalid_directory",
			imageDir:  "nonexistent",
			folderURL: "http://example.com",
			batchSize: 2,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up any existing output files
			os.Remove("flyers.json")
			os.Remove("imagesMetadata.json")

			err := ReadAndForward(tt.imageDir, tt.folderURL, tt.batchSize)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Verify output files were created
			if _, err := os.Stat("flyers.json"); err != nil {
				t.Errorf("flyers.json not created: %v", err)
			}

			if _, err := os.Stat("imagesMetadata.json"); err != nil {
				t.Errorf("imagesMetadata.json not created: %v", err)
			}

			// Verify JSON content
			flyersData, err := os.ReadFile("flyers.json")
			if err != nil {
				t.Errorf("Failed to read flyers.json: %v", err)
			}

			var flyersStruct FlyersData
			if err := json.Unmarshal(flyersData, &flyersStruct); err != nil {
				t.Errorf("Failed to parse flyers.json: %v", err)
			}

			if len(flyersStruct.Flyers) != len(testFiles) {
				t.Errorf("Wrong number of flyers in output: got %d, want %d",
					len(flyersStruct.Flyers), len(testFiles))
			}

			// Clean up generated files
			os.Remove("flyers.json")
			os.Remove("imagesMetadata.json")
		})
	}
}

// TestMetadata tests the metadata structure and JSON handling
func TestMetadata(t *testing.T) {
	metadata := Metadata{
		Version:     "1.0",
		LastUpdated: time.Now().Format(time.RFC3339),
		TotalFlyers: 10,
		URL:         "http://example.com",
		Schema: Schema{
			Format:   "JSON",
			Encoding: "UTF-8",
			Filetype: "text",
		},
	}

	jsonData, err := json.Marshal(metadata)
	if err != nil {
		t.Errorf("Failed to marshal Metadata: %v", err)
	}

	var unmarshaledMetadata Metadata
	err = json.Unmarshal(jsonData, &unmarshaledMetadata)
	if err != nil {
		t.Errorf("Failed to unmarshal Metadata: %v", err)
	}

	if metadata.Version != unmarshaledMetadata.Version {
		t.Errorf("Version mismatch: got %s, want %s", unmarshaledMetadata.Version, metadata.Version)
	}
	if metadata.TotalFlyers != unmarshaledMetadata.TotalFlyers {
		t.Errorf("TotalFlyers mismatch: got %d, want %d", unmarshaledMetadata.TotalFlyers, metadata.TotalFlyers)
	}
	if metadata.URL != unmarshaledMetadata.URL {
		t.Errorf("URL mismatch: got %s, want %s", unmarshaledMetadata.URL, metadata.URL)
	}
	if !reflect.DeepEqual(metadata.Schema, unmarshaledMetadata.Schema) {
		t.Errorf("Schema mismatch: got %+v, want %+v", unmarshaledMetadata.Schema, metadata.Schema)
	}
}
