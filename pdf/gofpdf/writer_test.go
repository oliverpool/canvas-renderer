package gofpdf

import (
	"testing"

	"github.com/oliverpool/canvas-renderer/renderertest"
)

func TestWriter(t *testing.T) {
	err := renderertest.RenderPreview(Writer)
	if err != nil {
		t.Error(err)
	}
}
