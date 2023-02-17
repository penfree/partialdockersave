package main

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"io"
	"os"
	"strings"

	"github.com/cloudflare/cfssl/log"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

// functions to operate image
type ImageLike interface {
	GetImages() []string
}

// save image to file and compress them
func rawSaveImage(ctx context.Context, images []string, excludeLayers map[string]bool, filename string, client *client.Client) error {
	reader, err := client.ImageSave(ctx, images)
	if err != nil {
		log.Errorf("image save failed %v", err)
	}
	// read the tar
	tr := tar.NewReader(reader)
	// output stream
	out, err := os.Create(filename)
	if err != nil {
		log.Errorf("Error writing archive: %v", err)
	}
	defer reader.Close()
	defer out.Close()
	// compress the output tar
	gzipWriter := gzip.NewWriter(out)
	defer gzipWriter.Close()
	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	// record skipped file size and skipped layers
	skipSize := 0
	keepSize := 0
	skipedLayers := make(map[string]int)

	// read original file, filter excluded layers and then write to output stream
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		} else if err != nil {
			log.Errorf("%v", err)
		}
		// directory is the layer id
		directory := strings.Split(hdr.Name, "/")[0]
		if _, ok := excludeLayers[directory]; ok {
			if size, ok := skipedLayers[directory]; ok {
				skipedLayers[directory] = size + int(hdr.Size)
			} else {
				skipedLayers[directory] = int(hdr.Size)
			}
			skipSize += int(hdr.Size)
			continue
		}
		keepSize += int(hdr.Size)
		tarWriter.WriteHeader(hdr)
		io.Copy(tarWriter, tr)
	}
	log.Infof("%d bytes saved to %s, %d layers (%d bytes) skipped", keepSize, filename, len(skipedLayers), skipSize)
	return nil
}

// save images
func SaveImage(ctx context.Context, image ImageLike, excludeImage ImageLike, filename string, client *client.Client) error {
	// get all exists images
	imageToSave := image.GetImages()
	// get all exclude Images
	imageToExclude := excludeImage.GetImages()

	// pull them all
	for _, img := range imageToSave {
		log.Infof("Pulling image %s", img)
		client.ImagePull(ctx, img, types.ImagePullOptions{})
	}
	for _, img := range imageToExclude {
		log.Infof("Pulling image %s", img)
		client.ImagePull(ctx, img, types.ImagePullOptions{})
	}

	// collect existing layers
	existLayers := make(map[string]bool)
	for _, img := range imageToExclude {
		layers, err := GetLayerIds(ctx, img, client)
		if err != nil {
			log.Errorf("Get layers of %s failed, %v", img, err)
		}
		for _, l := range layers {
			existLayers[l] = true
		}
	}

	// save image to file
	return rawSaveImage(ctx, imageToSave, existLayers, filename, client)
}
