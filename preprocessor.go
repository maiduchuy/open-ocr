package ocrworker

const PREPROCESSOR_IMGPROC = "img-proc"
const PREPROCESSOR_CONVERTPDF = "convert-pdf"

type Preprocessor interface {
	preprocess(ocrRequest *OcrRequest) error
}
