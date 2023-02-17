package main

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/image"
	v1 "github.com/docker/docker/image/v1"
	"github.com/docker/docker/layer"
	"github.com/opencontainers/go-digest"
)

// Convert result of `docker image inspect` to V1Image
func ConvertImageInspectToV1Image(ins types.ImageInspect) image.V1Image {
	created, _ := time.Parse(time.RFC3339, ins.Created)
	v1Img := image.V1Image{
		ID:              ins.ID,
		Parent:          ins.Parent,
		Comment:         ins.Comment,
		Created:         created,
		Container:       ins.Container,
		ContainerConfig: *ins.ContainerConfig,
		DockerVersion:   ins.DockerVersion,
		Author:          ins.Author,
		Config:          ins.Config,
		Architecture:    ins.Architecture,
		//Variant:         ins.Variant,
		OS:   ins.Os,
		Size: ins.Size,
	}
	return v1Img
}

// Get all the layer ids of an image, namely the directory name in tar file of `docker save`
func GetLayerIds(ctx context.Context, name string, client *client.Client) ([]string, error) {
	inspectInfo, _, err := client.ImageInspectWithRaw(ctx, name)
	if err != nil {
		fmt.Errorf("inspect image failed, %v", err)
	}
	var parent digest.Digest
	var layers []string
	img := ConvertImageInspectToV1Image(inspectInfo)
	for i := range inspectInfo.RootFS.Layers {
		v1Img := image.V1Image{
			// This is for backward compatibility used for
			// pre v1.9 docker.
			Created: time.Unix(0, 0),
		}
		// For the top layer, maybe due to ConvertImageInspectToV1Image, the digest is not same as real layer id
		// but it's still ok for this scenario, because it is not reasonable to compare two image with no diff
		if i == len(inspectInfo.RootFS.Layers)-1 {
			v1Img = img
		}
		rootFS := image.RootFS{
			Type: inspectInfo.RootFS.Type,
		}
		for _, diffId := range inspectInfo.RootFS.Layers {
			rootFS.DiffIDs = append(rootFS.DiffIDs, layer.DiffID(diffId))
		}

		rootFS.DiffIDs = rootFS.DiffIDs[:i+1]
		v1ID, err := v1.CreateID(v1Img, rootFS.ChainID(), parent)
		if err != nil {
			return layers, err
		}

		v1Img.ID = v1ID.Encoded()
		layers = append(layers, v1Img.ID)
		parent = v1ID
	}
	return layers, nil
}
