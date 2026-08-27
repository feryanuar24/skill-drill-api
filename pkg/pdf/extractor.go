package pdf

import (
	"fmt"
	"io"

	pdf "github.com/ledongthuc/pdf"
)

func ExtractText(reader io.ReaderAt, size int64) (string, error) {
	r, err := pdf.NewReader(reader, size)
	if err != nil {
		return "", fmt.Errorf("open pdf: %w", err)
	}

	extracted, err := r.GetPlainText()
	if err != nil {
		return "", fmt.Errorf("extract pdf text: %w", err)
	}

	text, err := io.ReadAll(extracted)
	if err != nil {
		return "", fmt.Errorf("read extracted pdf text: %w", err)
	}

	return string(text), nil
}
