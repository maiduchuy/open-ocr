package ocrworker

const PREPROCESSOR_IDENTITY = "identity"
const PREPROCESSOR_STROKE_WIDTH_TRANSFORM = "stroke-width-transform"
const PREPROCESSOR_IMGPROC = "img-proc"

type Preprocessor interface {
	preprocess(ocrRequest *OcrRequest) error
}

type IdentityPreprocessor struct {
}

func (i IdentityPreprocessor) preprocess(ocrRequest *OcrRequest) error {
	return nil
}
