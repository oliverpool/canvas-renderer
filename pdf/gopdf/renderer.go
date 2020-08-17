package gopdf

import (
	"bytes"
	"fmt"
	"image"
	"image/png"

	"github.com/signintech/gopdf"
	"github.com/tdewolff/canvas"
)

type PDF struct {
	*gopdf.GoPdf
	width, height float64
}

// NewPDF creates a portable document format renderer.
func New(width, height float64) PDF {
	pdf := &gopdf.GoPdf{}
	size := *gopdf.PageSizeA4
	size.W = width
	size.H = height
	pdf.Start(gopdf.Config{
		PageSize: size, // todo use actual w,h (https://github.com/signintech/gopdf/issues/140)
	})
	pdf.AddPage()
	return PDF{
		GoPdf:  pdf,
		width:  width,
		height: height,
	}
}

func (pdf PDF) Size() (float64, float64) {
	return pdf.width, pdf.height
}

func (pdf PDF) RenderPath(path *canvas.Path, style canvas.Style, m canvas.Matrix) {
}

func (pdf PDF) RenderText(text *canvas.Text, m canvas.Matrix) {
}

func (pdf PDF) applyMatrix(m canvas.Matrix) {
	tx, ty, theta, sx, sy, phi := m.Decompose()
	pdf.SetX(tx)
	pdf.SetX(ty)
	pdf.Rotate(-theta, 0, 0)
	_ = sx
	_ = sy
	pdf.Rotate(-phi, 0, 0)
}

func (pdf PDF) RenderImage(img image.Image, m canvas.Matrix) {
	tx, ty, theta, sx, sy, phi := m.Decompose()
	fmt.Println(tx, ty, theta, sx, sy, phi)
	// pdf.Rotate(-theta, 0, pdf.height)
	// pdf.Rotate(-phi, tx*sx, pdf.height-ty*sy)

	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	ih, _ := gopdf.ImageHolderByBytes(buf.Bytes())
	size := img.Bounds().Size()
	rect := &gopdf.Rect{
		W: float64(size.X) * sx,
		H: float64(size.Y) * sy,
	}
	fmt.Println(pdf.height)
	topLeft := m.Dot(canvas.Point{0, float64(size.Y)})
	bottomLeft := m.Dot(canvas.Point{0, 0})

	_ = pdf.ImageByHolder(ih, 0, pdf.height-0-rect.H, rect)
}
