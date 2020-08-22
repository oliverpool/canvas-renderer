package gofpdf

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"strconv"

	"github.com/oliverpool/canvas-renderer/renderertest"
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
	fpdf.SetCellMargin(0)
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

func (pdf PDF) setFillColor(c color.RGBA) {
	pdf.setColor(pdf.SetFillColor, c)
}
func (pdf PDF) setStrokeColor(c color.RGBA) {
	pdf.setColor(pdf.SetDrawColor, c)
}
func (pdf PDF) setTextColor(c color.RGBA) {
	pdf.setColor(pdf.SetTextColor, c)
}
func (pdf PDF) setColor(cb func(r, g, b int), c color.RGBA) {
	a := float64(c.A) / 255.0
	adjusted := func(v uint8) int {
		return int(float64(v) / a)
	}
	cb(adjusted(c.R), adjusted(c.G), adjusted(c.B))
	pdf.SetAlpha(a, "")
}
func (pdf PDF) setLineWidth(w float64) {
	pdf.SetLineWidth(w)
}
func (pdf PDF) setLineCap(capper canvas.Capper) {
	var capStyle string
	switch capper.(type) {
	case canvas.ButtCapper:
		capStyle = "butt"
	case canvas.RoundCapper:
		capStyle = "round"
	case canvas.SquareCapper:
		capStyle = "square"
	default:
		panic("PDF: line cap not supported")
	}
	pdf.SetLineCapStyle(capStyle)
}
func (pdf PDF) setLineJoin(joiner canvas.Joiner) {
	var joinStyle string
	switch joiner.(type) {
	case canvas.BevelJoiner:
		joinStyle = "bevel"
	case canvas.RoundJoiner:
		joinStyle = "round"
	case canvas.MiterJoiner:
		joinStyle = "miter"
	default:
		panic("PDF: line cap not supported")
	}
	pdf.SetLineJoinStyle(joinStyle)
}
func (pdf PDF) setDashes(offset float64, dashes []float64) {
	pdf.SetDashPattern(dashes, offset)
}

// RenderPath is adapted from https://github.com/tdewolff/canvas/blob/master/pdf/renderer.go
// Copyright (c) 2015 Taco de Wolff
// MIT License
func (pdf PDF) RenderPath(path *canvas.Path, style canvas.Style, m canvas.Matrix) {
	fill := style.FillColor.A != 0
	stroke := style.StrokeColor.A != 0 && 0.0 < style.StrokeWidth
	differentAlpha := fill && stroke && style.FillColor.A != style.StrokeColor.A

	// PDFs don't support the arcs joiner, miter joiner (not clipped), or miter joiner (clipped) with non-bevel fallback
	strokeUnsupported := false
	if _, ok := style.StrokeJoiner.(canvas.ArcsJoiner); ok {
		strokeUnsupported = true
	} else if miter, ok := style.StrokeJoiner.(canvas.MiterJoiner); ok {
		if math.IsNaN(miter.Limit) {
			strokeUnsupported = true
		} else if _, ok := miter.GapJoiner.(canvas.BevelJoiner); !ok {
			strokeUnsupported = true
		}
	}

	// PDFs don't support connecting first and last dashes if path is closed, so we move the start of the path if this is the case
	// TODO
	//if style.DashesClose {
	//	strokeUnsupported = true
	//}

	closed := false
	data := path.Transform(canvas.Identity.Scale(ptPerMm, ptPerMm).Mul(m)).ToPDF()
	// data := path.Transform(m).ToPDF()
	if 1 < len(data) && data[len(data)-1] == 'h' {
		data = data[:len(data)-2]
		closed = true
	}

	if !stroke || !strokeUnsupported {
		if fill && !stroke {
			pdf.setFillColor(style.FillColor)
			pdf.RawWriteStr(" ")
			pdf.RawWriteStr(" ")
			pdf.RawWriteStr(data)
			pdf.RawWriteStr(" f")
			if style.FillRule == canvas.EvenOdd {
				pdf.RawWriteStr("*")
			}
		} else if !fill && stroke {
			pdf.setStrokeColor(style.StrokeColor)
			pdf.setLineWidth(style.StrokeWidth)
			pdf.setLineCap(style.StrokeCapper)
			pdf.setLineJoin(style.StrokeJoiner)
			pdf.setDashes(style.DashOffset, style.Dashes)
			pdf.RawWriteStr(" ")
			pdf.RawWriteStr(data)
			if closed {
				pdf.RawWriteStr(" s")
			} else {
				pdf.RawWriteStr(" S")
			}
			if style.FillRule == canvas.EvenOdd {
				pdf.RawWriteStr("*")
			}
		} else if fill && stroke {
			if !differentAlpha {
				pdf.setFillColor(style.FillColor)
				pdf.setStrokeColor(style.StrokeColor)
				pdf.setLineWidth(style.StrokeWidth)
				pdf.setLineCap(style.StrokeCapper)
				pdf.setLineJoin(style.StrokeJoiner)
				pdf.setDashes(style.DashOffset, style.Dashes)
				pdf.RawWriteStr(" ")
				pdf.RawWriteStr(data)
				if closed {
					pdf.RawWriteStr(" b")
				} else {
					pdf.RawWriteStr(" B")
				}
				if style.FillRule == canvas.EvenOdd {
					pdf.RawWriteStr("*")
				}
			} else {
				pdf.setFillColor(style.FillColor)
				pdf.RawWriteStr(" ")
				pdf.RawWriteStr(data)
				pdf.RawWriteStr(" f")
				if style.FillRule == canvas.EvenOdd {
					pdf.RawWriteStr("*")
				}

				pdf.setStrokeColor(style.StrokeColor)
				pdf.setLineWidth(style.StrokeWidth)
				pdf.setLineCap(style.StrokeCapper)
				pdf.setLineJoin(style.StrokeJoiner)
				pdf.setDashes(style.DashOffset, style.Dashes)
				pdf.RawWriteStr(" ")
				pdf.RawWriteStr(data)
				if closed {
					pdf.RawWriteStr(" s")
				} else {
					pdf.RawWriteStr(" S")
				}
				if style.FillRule == canvas.EvenOdd {
					pdf.RawWriteStr("*")
				}
			}
		}
	} else {
		// stroke && strokeUnsupported
		if fill {
			pdf.setFillColor(style.FillColor)
			pdf.RawWriteStr(" ")
			pdf.RawWriteStr(data)
			pdf.RawWriteStr(" f")
			if style.FillRule == canvas.EvenOdd {
				pdf.RawWriteStr("*")
			}
		}

		// stroke settings unsupported by PDF, draw stroke explicitly
		if 0 < len(style.Dashes) {
			path = path.Dash(style.DashOffset, style.Dashes...)
		}
		path = path.Stroke(style.StrokeWidth, style.StrokeCapper, style.StrokeJoiner)

		pdf.setFillColor(style.StrokeColor)
		pdf.RawWriteStr(" ")
		pdf.RawWriteStr(path.ToPDF())
		pdf.RawWriteStr(" f")
		if style.FillRule == canvas.EvenOdd {
			pdf.RawWriteStr("*")
		}
	}
}

func (pdf PDF) RenderText(text *canvas.Text, m canvas.Matrix) {
	// canvas.RenderTextAsPath(pdf, text, m)

	text.WalkSpans(func(y, dx float64, span canvas.TextSpan) {
		words := span.Words()
		if len(words) == 0 {
			return
		}
		face := span.Face

		// ensure font of the same name but different style won't be mixed
		faceName := face.Name() + "-" + strconv.Itoa(int(face.Style))
		styleStr := ""

		_, buf := face.Font.Raw()
		pdf.AddUTF8FontFromBytes(faceName, styleStr, buf)
		pdf.SetFont(faceName, styleStr, 16)
		pdf.SetFontUnitSize(face.Size * face.Scale)
		pdf.setTextColor(face.Color)

		if span.GlyphSpacing > 0 {
			// not implemented in gofpdf
			// r.w.SetTextCharSpace(span.GlyphSpacing)
			err := fmt.Errorf("GlyphSpacing is not supported")
			fmt.Println(err)
			// pdf.SetError(err)
			// return
		}

		if 0.0 < face.FauxBold {
			pdf.SetTextRenderingMode(2)
			pdf.SetLineWidth(face.FauxBold * 2.0)
		} else {
			pdf.SetTextRenderingMode(0)
		}

		/* // Hack to sonmehow adjust the position of the text
		metrics := face.Metrics()
		dy := metrics.Descent / 5. // random math
		/*/
		dy := 0.
		//*/

		pdf.transformBegin(m.Translate(dx, y+dy).Shear(face.FauxItalic, 0.0))
		pdf.SetXY(0, pdf.height)
		for _, w := range words {
			width := pdf.GetStringWidth(w)
			pdf.CellFormat(width+span.WordSpacing, 0, w, "", 0, "A", false, 0, "")
		}
		pdf.TransformEnd()
	})

	text.RenderDecoration(pdf, m)
}

func adjustMatrix(m canvas.Matrix) canvas.Matrix {
	return canvas.Identity.Scale(ptPerMm, ptPerMm).Mul(m).Scale(1/ptPerMm, 1/ptPerMm)
}

func (pdf PDF) transformBegin(m canvas.Matrix) PDF {
	pdf.TransformBegin()
	am := adjustMatrix(m)
	pdf.Transform(gofpdf.TransformMatrix{
		am[0][0], am[1][0], am[0][1], am[1][1], am[0][2], am[1][2],
	})
	return pdf
}

func (pdf PDF) RenderImage(img image.Image, m canvas.Matrix) {
	defer pdf.transformBegin(m).TransformEnd()

	size := img.Bounds().Size()

	switch i := img.(type) {
	case renderertest.JPEGImage:
		pdf.renderImage(bytes.NewBuffer(i.JPEGBytes()), "JPG", size, m)
	case renderertest.PNGImage:
		pdf.renderImage(bytes.NewBuffer(i.PNGBytes()), "PNG", size, m)
	default:
		var buf bytes.Buffer
		_ = png.Encode(&buf, img)
		pdf.renderImage(&buf, "PNG", size, m)
	}
}

func (pdf PDF) renderImage(buf *bytes.Buffer, imgType string, size image.Point, m canvas.Matrix) {
	hash := md5.New()
	hash.Write(buf.Bytes())
	imgName := string(hash.Sum(nil))

	opt := gofpdf.ImageOptions{
		ImageType:             imgType,
		ReadDpi:               false,
		AllowNegativePosition: true,
	}
	pdf.RegisterImageOptionsReader(imgName, opt, buf)
	pdf.ImageOptions(imgName, 0, pdf.height-float64(size.Y), float64(size.X), float64(size.Y), false, opt, 0, "")
}
