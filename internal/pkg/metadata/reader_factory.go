package metadata

// ExifReaderFactory creates ExifReader instances.
type ExifReaderFactory interface {
	NewReader() (ExifReader, error)
}

type defaultExifReaderFactory struct{}

func (f *defaultExifReaderFactory) NewReader() (ExifReader, error) {
	reader, err := NewExiftoolClient()
	if err != nil {
		return NewLegacyExifReader(), err
	}
	return reader, nil
}

// NewExifReaderFactory returns the default ExifReaderFactory.
func NewExifReaderFactory() ExifReaderFactory {
	return &defaultExifReaderFactory{}
}
