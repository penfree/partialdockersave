# partialdockersave
This is a alternative of `docker save` which you can only save layers not exists in specified image.

* What's the purpose of this tool

Sometimes, we need to use `docker save` to dump the image and then transfer it to an enviroment without internet. We could have met a problem that the result filesize is always too big. 

In fact, we may already have an old image in the dest server which is much the same with new image. This tool help you to save only the layers that doesn't exist in the old images.

## Install

```
go get github.com/penfree/partialdockersave
```

## Usage

```
NAME:
   Save docker image without layers in `exclude`

USAGE:
   partialdockersave [global options] command [command options] [arguments...]

COMMANDS:
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --image value, -i value    The image to save
   --exclude value, -e value  The existing image that does not need to export
   --output value, -o value   The output tar.gz file (default: "image.tgz")
   --help, -h                 show help
   --version, -v              print the version
```

## Example

```bash
# multiple -i & -e can be specified to include or exclude
partialdockersave -i image1 -i image2 -e imagetoexclude -o result.tgz
```

## How it works

`docker save` can give us a tar file with one directory for each layer of the image. The key is to know which directory is not needed. 

* How is the directory name generated?

We can find from https://github.com/moby/moby/blob/master/image/tarexport/save.go that the directory name is encoded from layer meta data. So we can do the same thing from infomation got from `docker image inspect`, then we can filter them when saving image.

Unlike other tools like `undocker`, `dive`, they should do a real save to get the tar file of all images, and then extract manifest from the tar file.  This tool brings no more io operation than normal `docker save`.

## Future works

Other than raw images, we can do some work to analyze images in helm chart and compute the increment diff of two helm chart.
This can be done by implement `ImageLike` interface for helm charts.
