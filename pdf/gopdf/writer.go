package gopdf

import (
	"fmt"
	"io"

	"github.com/tdewolff/canvas"
)

// Writer writes the canvas as a PDF file.
func Writer(w io.Writer, c *canvas.Canvas) error {
	pdf := New(c.W, c.H)
	fmt.Println(c.W, c.H)
	c.Render(pdf)
	return pdf.Write(w)
}
