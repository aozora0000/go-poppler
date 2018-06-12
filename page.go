package poppler

/*
#cgo pkg-config: poppler-glib cairo
#include <stdlib.h>
#include <poppler.h>
#include <glib.h>
#include <cairo.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <cairo.h>

static unsigned char getbyte(unsigned char *buf, int idx) {
	return buf[idx];
}

 typedef struct cairo_wrap_vector {
 	char *buf;
 	size_t cap;
 	size_t len;
 	size_t size;
 } cairo_wrap_vector;

 void cairo_wrap_vector_free(cairo_wrap_vector *v) {
 	if (v->buf) {
 		free(v->buf);
 	}
 	free(v);
 }

 void cairo_wrap_vector_init(cairo_wrap_vector *v, int size) {
 	v->len = 0;
 	v->cap = 0;
 	v->size = size;
 	size_t cap = size;
 	char *buf = malloc(sizeof(char) * cap);
 	if (buf) {
 		v->buf = buf;
 		v->cap = cap;
 	}
 }

 cairo_wrap_vector *cairo_wrap_vector_new(int size) {
 	cairo_wrap_vector *v = malloc(sizeof(cairo_wrap_vector));
 	cairo_wrap_vector_init(v, size);
 	return v;
 }

 static int cairo_wrap_vector_grow(cairo_wrap_vector *v, size_t len) {
 	int status = 0;
 	if (v->cap <= v->len+len) {
 		unsigned int cap;
 		char* buf;
 		if (v->cap == 0) {
 			cap = v->size + len;
 			buf = malloc(sizeof(char) * cap);
 		} else {
 			cap = (v->cap*2) + len;
 			buf = realloc(v->buf, sizeof(char) * cap);
 		}
 		if (buf) {
 			v->buf = buf;
 			v->cap = cap;
 		} else {
 			status = 1;
 		}
 	}
 	return status;
 }

 static int cairo_wrap_vector_append(cairo_wrap_vector *v, const unsigned char *data, size_t len) {
 	int status = cairo_wrap_vector_grow(v, len);
 	if (status == 0) {
 		memcpy(&v->buf[v->len], data, len);
 		v->len += len;
 	}
 	return status;
 }

 cairo_status_t cairo_wrap_writer(void *closure, const unsigned char *data, unsigned int length) {
 	cairo_status_t status;
 	cairo_wrap_vector *v = (cairo_wrap_vector*)closure;
 	switch (cairo_wrap_vector_append(v, data, (size_t) length)) {
 	case 0:
 		status = CAIRO_STATUS_SUCCESS;
 		break;
 	case 1:
 		status = CAIRO_STATUS_NO_MEMORY;
 		break;
 	default:
 		status = CAIRO_STATUS_WRITE_ERROR;
 		break;
 	}
 	return status;
 }

 cairo_status_t cairo_wrap_write_surface_to_vector(cairo_surface_t *surface, cairo_wrap_vector *vec)
 {
 	cairo_status_t status = cairo_surface_write_to_png_stream(surface, cairo_wrap_writer, vec);
 	return status;
 }
*/
import "C"
import (
	"image"
	"image/color"
	"unsafe"
	"github.com/ungerik/go-cairo"
)

//import "fmt"

type Page struct {
	p *C.struct__PopplerPage
}
type RenderOptions struct {
	FillColor color.RGBA
	NoAA      bool
	Scale     float64
	MemorySize int
}

func (p *Page) Text() string {
	return C.GoString(C.poppler_page_get_text(p.p))
}

func (p *Page) TextAttributes() (results []TextAttributes) {
	a := C.poppler_page_get_text_attributes(p.p)
	defer C.poppler_page_free_text_attributes(a)
	var attr *C.PopplerTextAttributes
	results = make([]TextAttributes, 0)
	el := C.g_list_first(a)
	for el != nil {
		attr = (*C.PopplerTextAttributes)(el.data)
		fn := *attr.font_name
		result := TextAttributes{
			FontName:     toString(&fn),
			FontSize:     float64(attr.font_size),
			IsUnderlined: toBool(attr.is_underlined),
			StartIndex:   int(attr.start_index),
			EndIndex:     int(attr.end_index),
			Color: Color{
				R: int(attr.color.red),
				G: int(attr.color.green),
				B: int(attr.color.blue),
			},
		}
		results = append(results, result)
		el = el.next
	}
	return
}

func (p *Page) Size() (width, height float64) {
	var w, h C.double
	C.poppler_page_get_size(p.p, &w, &h)
	return float64(w), float64(h)
}

func (p *Page) Index() int {
	return int(C.poppler_page_get_index(p.p))
}

func (p *Page) Label() string {
	return toString(C.poppler_page_get_label(p.p))
}

func (p *Page) Duration() float64 {
	return float64(C.poppler_page_get_duration(p.p))
}

func (p *Page) Images() (results []Image) {
	l := C.poppler_page_get_image_mapping(p.p)
	defer C.poppler_page_free_image_mapping(l)
	results = make([]Image, 0)
	var im *C.PopplerImageMapping
	for el := C.g_list_first(l); el != nil; el = el.next {
		im = (*C.PopplerImageMapping)(el.data)
		result := Image{
			Id: int(im.image_id),
			Area: Rectangle{
				X1: float64(im.area.x1),
				Y1: float64(im.area.y1),
				X2: float64(im.area.x2),
				Y2: float64(im.area.y2),
			},
			p: p.p,
		}
		results = append(results, result)
	}
	return
}

func (p *Page) TextLayout() (layouts []Rectangle) {
	var rect *C.PopplerRectangle
	var n C.guint
	if toBool(C.poppler_page_get_text_layout(p.p, &rect, &n)) {
		defer C.g_free((C.gpointer)(rect))
		layouts = make([]Rectangle, int(n))
		r := (*[1 << 30]C.PopplerRectangle)(unsafe.Pointer(rect))[:n:n]
		for i := 0; i < int(n); i++ {
			layouts[i] = Rectangle{
				X1: float64(r[i].x1),
				Y1: float64(r[i].y1),
				X2: float64(r[i].x2),
				Y2: float64(r[i].y2),
			}
		}
	}
	return
}

func (p *Page) TextLayoutAndAttrs() (result []TextEl) {
	text := p.Text()
	attrs := p.TextAttributes()
	layout := p.TextLayout()
	result = make([]TextEl, len(layout))
	attrsRef := make([]*TextAttributes, len(attrs))
	for i, a := range attrs {
		attr := a
		attrsRef[i] = &attr
	}
	i := 0
	for _, t := range text {
		var a *TextAttributes
		for _, a = range attrsRef {
			if i >= a.StartIndex && i <= a.EndIndex {
				break
			}
		}
		result[i] = TextEl{
			Text:  string(t),
			Attrs: a,
			Rect:  layout[i],
		}
		i++
	}
	return
}

func (p *Page) GetSize() (int, int) {
	width := C.double(0)
	height := C.double(0)
	C.poppler_page_get_size(p.p, &width, &height)
	return int(width), int(height)
}

func (p *Page) WriteToPNGStream(opts *RenderOptions) ([]byte, cairo.Status) {
	width, height := p.GetSize()
	size := 64
	if opts != nil {
		width = int(opts.Scale * float64(width))
		height = int(opts.Scale * float64(height))
		size = opts.MemorySize
	}
	surface := C.cairo_image_surface_create(C.CAIRO_FORMAT_ARGB32, C.int(width), C.int(height))
	defer C.cairo_surface_destroy(surface)

	ctx := C.cairo_create(surface)
	defer C.cairo_destroy(ctx)

	ow, oh := p.Size()
	fw := float64(width)
	fh := float64(height)
	sw, sh := float64(fw/ow), float64(fh/oh)
	C.cairo_scale(ctx, C.double(sw), C.double(sh))

	fillColor := color.RGBA{255, 255, 255, 255}
	if opts != nil {
		fillColor = opts.FillColor
	}
	C.cairo_set_source_rgba(ctx, C.double(float64(fillColor.R)/float64(255)),
		C.double(float64(fillColor.G)/float64(255)),
		C.double(float64(fillColor.B)/float64(255)),
		C.double(float64(fillColor.A)/float64(255)))
	C.cairo_rectangle(ctx, 0, 0, C.double(width), C.double(height))
	C.cairo_fill(ctx)

	if opts != nil && opts.NoAA {
		C.cairo_set_antialias(ctx, C.CAIRO_ANTIALIAS_NONE)
	}

	C.poppler_page_render_for_printing(p.p, ctx)

	vec := C.cairo_wrap_vector_new(C.int(size))
	defer C.cairo_wrap_vector_free(vec)

	status := cairo.Status(C.cairo_wrap_write_surface_to_vector(surface, vec))
	buf := C.GoBytes(unsafe.Pointer(vec.buf), C.int(vec.len))

	return buf, status
}

func (p *Page) TransferPNGImage(opts *RenderOptions) image.Image {
	width, height := p.GetSize()
	if opts != nil {
		width = int(opts.Scale * float64(width))
		height = int(opts.Scale * float64(height))
	}
	surface := C.cairo_image_surface_create(C.CAIRO_FORMAT_ARGB32, C.int(width), C.int(height))
	defer C.cairo_surface_destroy(surface)

	ctx := C.cairo_create(surface)
	defer C.cairo_destroy(ctx)

	ow, oh := p.Size()
	fw := float64(width)
	fh := float64(height)
	sw, sh := float64(fw/ow), float64(fh/oh)
	C.cairo_scale(ctx, C.double(sw), C.double(sh))

	fillColor := color.RGBA{255, 255, 255, 255}
	if opts != nil {
		fillColor = opts.FillColor
	}
	C.cairo_set_source_rgba(ctx, C.double(float64(fillColor.R)/float64(255)),
		C.double(float64(fillColor.G)/float64(255)),
		C.double(float64(fillColor.B)/float64(255)),
		C.double(float64(fillColor.A)/float64(255)))
	C.cairo_rectangle(ctx, 0, 0, C.double(width), C.double(height))
	C.cairo_fill(ctx)

	if opts != nil && opts.NoAA {
		C.cairo_set_antialias(ctx, C.CAIRO_ANTIALIAS_NONE)
	}

	C.poppler_page_render_for_printing(p.p, ctx)
	data := C.cairo_image_surface_get_data(surface)
	nrgba := image.NewNRGBA(image.Rect(0, 0, width, height))
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			nrgba.SetNRGBA(x, y, color.NRGBA{
				R: uint8(C.getbyte(data, C.int(x*4+4*y*width+2))),
				G: uint8(C.getbyte(data, C.int(x*4+4*y*width+1))),
				B: uint8(C.getbyte(data, C.int(x*4+4*y*width+0))),
				A: uint8(C.getbyte(data, C.int(x*4+4*y*width+3))),
			})
		}
	}

	return nrgba
}
