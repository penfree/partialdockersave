package main

// a real docker image
type ImageObject string

// get all images
func (i ImageObject) GetImages() []string {
	return []string{string(i)}
}

type ImageList []string

func (il ImageList) GetImages() []string {
	return ([]string)(il)
}
