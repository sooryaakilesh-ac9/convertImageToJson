package main

import (
	_ "image/jpeg" // Import the JPEG decoder
	_ "image/png"  // Import the PNG decoder
	"toJson/utils"
)

func main() {
	imageFolder := "images/"
	folderURL := "path/to/Flyers"
	batchSize := 10

	if err := utils.ReadAndForward(imageFolder, folderURL, batchSize); err != nil {
		panic(err)
	}
}
