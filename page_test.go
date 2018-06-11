package poppler

import (
	"image/color"
	"image/png"
	"os"
	"testing"
)

func TestPage_Thumbnail(t *testing.T) {
	doc, err := Open("sample/test.pdf")
	if err != nil {
		t.Error(err)
	}
	page := doc.GetPage(0)
	img := page.WriteToPNGStream(&RenderOptions{
		FillColor: color.RGBA{255, 255, 255, 255},
		NoAA:      true,
		Scale:     300 / 72,
	})
	out, err := os.Create("sample/test.png")

	if err != nil {
		t.Error(err)
	}

	err = png.Encode(out, img)

	if err != nil {
		t.Error(err)
	}

}
