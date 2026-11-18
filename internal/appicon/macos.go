package appIcon

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"path/filepath"
	"sort"
)

var macosIconSpecs = []iconSpec{
	{16, "icon_16x16"},
	{32, "icon_16x16@2x"},
	{32, "icon_32x32"},
	{64, "icon_32x32@2x"},
	{128, "icon_128x128"},
	{256, "icon_128x128@2x"},
	{256, "icon_256x256"},
	{512, "icon_256x256@2x"},
	{512, "icon_512x512"},
	{1024, "icon_512x512@2x"},
}

type icnsEntry struct {
	OSType string
	Size   int
	PNG    []byte
}

var icnsTypeMap = map[int]string{
	16:   "icp4",
	32:   "icp5",
	64:   "icp6",
	128:  "icp7",
	256:  "icp8",
	512:  "icp9",
	1024: "ic10",
}

func encodeICNS(images []icnsEntry) ([]byte, error) {
	var body bytes.Buffer
	for _, img := range images {
		if imgSize, ok := icnsTypeMap[img.Size]; ok {
			body.WriteString(imgSize)
			entryLen := 8 + len(img.PNG)
			buf := make([]byte, 4)
			binary.BigEndian.PutUint32(buf, uint32(entryLen))
			body.Write(buf)
			body.Write(img.PNG)
		}
	}

	var out bytes.Buffer
	out.WriteString("icns")
	totalLen := 8 + body.Len()
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(totalLen))
	out.Write(buf)
	out.Write(body.Bytes())
	return out.Bytes(), nil
}

func generateMacos(src image.Image, outDir, appName string) error {
	zipPath := filepath.Join(outDir, "macos.zip")

	var icnsImgs []icnsEntry
	for _, s := range macosIconSpecs {
		img := resizeImg(src, s.Size)
		icnsImgs = append(icnsImgs, icnsEntry{Size: s.Size, PNG: encodePNG(img)})
	}
	sort.Slice(icnsImgs, func(i, j int) bool { return icnsImgs[i].Size < icnsImgs[j].Size })

	icnsData, err := encodeICNS(icnsImgs)
	if err != nil {
		return err
	}

	if err := createZip(zipPath, nil, map[string][]byte{appName + ".icns": icnsData}, nil); err != nil {
		return err
	}
	fmt.Printf("    → %s\n", zipPath)
	return nil
}
