package appIcon

import (
	"archive/zip"
	"bytes"
	"image"
	_ "image/jpeg"
	"image/png"
	"io"
	"os"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/webp"

	"github.com/disintegration/imaging"
)

type iconSpec struct {
	Size int
	Name string
}

type imageEntry struct {
	Spec    iconSpec
	Img     image.Image
	ZipPath string
}

func loadImage(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	return img, err
}

func resizeImg(src image.Image, size int) image.Image {
	return imaging.Resize(src, size, size, imaging.Lanczos)
}

func encodePNG(img image.Image) []byte {
	var buf bytes.Buffer
	png.Encode(&buf, img)
	return buf.Bytes()
}

func createZip(zipPath string, entries []imageEntry, extraFiles map[string][]byte, extraText map[string]string) error {
	f, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer f.Close()

	zw := zip.NewWriter(f)
	defer zw.Close()

	for _, e := range entries {
		w, err := zw.Create(e.ZipPath)
		if err != nil {
			return err
		}
		if _, err := w.Write(encodePNG(e.Img)); err != nil {
			return err
		}
	}

	for name, data := range extraFiles {
		w, err := zw.Create(name)
		if err != nil {
			return err
		}
		if _, err := w.Write(data); err != nil {
			return err
		}
	}

	for name, text := range extraText {
		w, err := zw.Create(name)
		if err != nil {
			return err
		}
		if _, err := io.WriteString(w, text); err != nil {
			return err
		}
	}
	return nil
}
