package poppler

import (
	"image/color"
	"image/png"
	"os"
	"testing"
	"image"
	"bytes"
	)

// Converted implements image.Image, so you can
// pretend that it is the converted image.
type Converted struct {
	Img image.Image
	Mod color.Model
}

// We return the new color model...
func (c *Converted) ColorModel() color.Model{
	return c.Mod
}

// ... but the original bounds
func (c *Converted) Bounds() image.Rectangle{
	return c.Img.Bounds()
}

// At forwards the call to the original image and
// then asks the color model to convert it.
func (c *Converted) At(x, y int) color.Color{
	return c.Mod.Convert(c.Img.At(x,y))
}


func TestPage_WriteToPNGStream(t *testing.T) {
	doc, err := Open("sample/test.pdf")
	if err != nil {
		t.Error(err)
	}
	page := doc.GetPage(0)
	b, _ := page.WriteToPNGStream(&RenderOptions{
		FillColor: color.RGBA{255, 255, 255, 255},
		NoAA:      true,
		Scale:     5.554,
		MemorySize: 64,
	})
	out, err := os.Create("sample/stream.png")

	if err != nil {
		t.Error(err)
	}
	img, _, _ := image.Decode(bytes.NewReader(b))
	gr := &Converted{img, color.GrayModel}

	err = png.Encode(out, gr)

	if err != nil {
		t.Error(err)
	}

}
