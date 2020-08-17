package renderertest

import (
	"testing"

	"github.com/tdewolff/canvas/pdf"
)

func TestPdfPreviewReference(t *testing.T) {
	err := RenderPreview(pdf.Writer)
	if err != nil {
		t.Error(err)
	}
}
