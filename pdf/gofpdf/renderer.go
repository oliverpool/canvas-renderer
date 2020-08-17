package gopdf

import (
	"bytes"
	"crypto/md5"
	"image"
	"image/png"
	"io"

	"github.com/phpdave11/gofpdf"
	"github.com/tdewolff/canvas"
)

const ptPerMm = 72 / 25.4

type PDF struct {
	*gofpdf.Fpdf
	width, height float64
}

// NewPDF creates a portable document format renderer.
func New(width, height float64) PDF {
	fpdf := gofpdf.NewCustom(&gofpdf.InitType{
		UnitStr: "mm",
		Size:    gofpdf.SizeType{Wd: width, Ht: height},
		// FontDirStr: example.FontDir(),
	})
	fpdf.SetMargins(0, 0, 0)
	fpdf.SetAutoPageBreak(false, 0)
	fpdf.AddPage()
	return PDF{
		Fpdf:   fpdf,
		width:  width,
		height: height,
	}
}

func (pdf PDF) Size() (float64, float64) {
	return pdf.width, pdf.height
}

func (pdf PDF) RenderPath(path *canvas.Path, style canvas.Style, m canvas.Matrix) {
	// TODO
}

func (pdf PDF) RenderText(text *canvas.Text, m canvas.Matrix) {
	canvas.RenderTextAsPath(pdf, text, m)
}

func (pdf PDF) transformBegin(m canvas.Matrix) {
	pdf.TransformBegin()
	m = canvas.Identity.Scale(ptPerMm, ptPerMm).Mul(m)
	m = m.Scale(1/ptPerMm, 1/ptPerMm)
	pdf.Transform(gofpdf.TransformMatrix{
		m[0][0], m[1][0], m[0][1], m[1][1], m[0][2], m[1][2],
	})
}

func (pdf PDF) RenderImage(img image.Image, m canvas.Matrix) {
	pdf.transformBegin(m)
	defer pdf.TransformEnd()

	var buf bytes.Buffer
	hash := md5.New()
	mr := io.MultiWriter(&buf, hash)

	_ = png.Encode(mr, img)
	imgName := string(hash.Sum(nil))

	opt := gofpdf.ImageOptions{
		ImageType:             "PNG",
		ReadDpi:               false,
		AllowNegativePosition: true,
	}
	pdf.RegisterImageOptionsReader(imgName, opt, &buf)
	size := img.Bounds().Size()
	pdf.ImageOptions(imgName, 0, pdf.height-float64(size.Y), float64(size.X), float64(size.Y), false, opt, 0, "")
}
