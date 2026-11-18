package appIcon

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"path/filepath"
)

func encodeICO(images []struct {
	Size int
	PNG  []byte
}) ([]byte, error) {
	var out bytes.Buffer

	binary.Write(&out, binary.LittleEndian, uint16(0))
	binary.Write(&out, binary.LittleEndian, uint16(1))
	binary.Write(&out, binary.LittleEndian, uint16(len(images)))

	dataOffset := 6 + len(images)*16
	for _, img := range images {
		w := img.Size
		if w >= 256 {
			w = 0
		}
		h := img.Size
		if h >= 256 {
			h = 0
		}
		out.WriteByte(byte(w))
		out.WriteByte(byte(h))
		out.WriteByte(0)
		out.WriteByte(0)
		binary.Write(&out, binary.LittleEndian, uint16(1))
		binary.Write(&out, binary.LittleEndian, uint16(32))
		binary.Write(&out, binary.LittleEndian, uint32(len(img.PNG)))
		binary.Write(&out, binary.LittleEndian, uint32(dataOffset))
		dataOffset += len(img.PNG)
	}
	for _, img := range images {
		out.Write(img.PNG)
	}
	return out.Bytes(), nil
}

func generateWindows(src image.Image, outDir, appName string) error {
	zipPath := filepath.Join(outDir, "windows.zip")

	var icoImgs []struct {
		Size int
		PNG  []byte
	}
	for _, s := range []int{16, 24, 32, 48, 64, 128, 256} {
		icoImgs = append(icoImgs, struct {
			Size int
			PNG  []byte
		}{s, encodePNG(resizeImg(src, s))})
	}

	icoData, err := encodeICO(icoImgs)
	if err != nil {
		return err
	}

	if err := createZip(zipPath, nil, map[string][]byte{appName + ".ico": icoData}, nil); err != nil {
		return err
	}
	fmt.Printf("    → %s\n", zipPath)
	return nil
}
