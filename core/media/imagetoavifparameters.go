package media

import "time"

type ImageToAVIFParameters struct {
	InputPath string
	Outputs   []ImageToAVIFOutput
	Timeout   time.Duration
}

type ImageToAVIFOutput struct {
	Width  int
	Height int
	Path   string
}
