package file

import (
	"fmt"
	"github.com/google/uuid"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
	"io"
	"io/ioutil"
	"strings"
	"unicode"
)

type File struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Size  int64  `json:"size"`
	Bytes []byte `json:"file"`
}

type CreateFileDTO struct {
	Name   string `json:"name"`
	Size   int64  `json:"size"`
	Reader io.Reader
}

// checks if a rune (a Unicode code point) is a nonspacing mark
func isMn(r rune) bool {
	return unicode.Is(unicode.Mn, r) // Mn: nonspacing marks
}

/*
*
The NormalizeName method is responsible for normalizing the file name. It performs the following steps:

1. Replaces all spaces in the file name with underscores ("_").
2. Applies Unicode normalization to the file name using the transform.Chain function from the golang.org/x/text/transform package. Specifically, it:
  - Decomposes the string into its canonical decomposition form (NFD).
  - Removes any nonspacing marks (using transform.RemoveFunc(isMn)).
  - Recomposes the string into its canonical composition form (NFC).

This normalization process helps ensure that the file name is consistent and can be used in various file systems and applications without issues.
*/
func (d CreateFileDTO) NormalizeName() {
	d.Name = strings.ReplaceAll(d.Name, " ", "_")
	t := transform.Chain(norm.NFD, transform.RemoveFunc(isMn), norm.NFC)
	d.Name, _, _ = transform.String(t, d.Name)
}

func NewFile(dto CreateFileDTO) (*File, error) {
	bytes, err := ioutil.ReadAll(dto.Reader)
	if err != nil {
		return nil, fmt.Errorf("error reading file - err: %w", err)
	}

	id, err := uuid.NewUUID()
	if err != nil {
		return nil, fmt.Errorf("error generating uuid - err: %w", err)
	}

	return &File{
		ID:    id.String(),
		Name:  dto.Name,
		Size:  dto.Size,
		Bytes: bytes,
	}, nil

}
