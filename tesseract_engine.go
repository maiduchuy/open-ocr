package ocrworker

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"

	"github.com/couchbaselabs/logg"
)

// This variant of the TesseractEngine calls tesseract via exec
type TesseractEngine struct {
}

type TesseractEngineArgs struct {
	configVars  map[string]string `json:"config_vars"`
	pageSegMode string            `json:"psm"`
	lang        string            `json:"lang"`
}

func NewTesseractEngineArgs(ocrRequest OcrRequest) (*TesseractEngineArgs, error) {

	engineArgs := &TesseractEngineArgs{}

	if ocrRequest.EngineArgs == nil {
		return engineArgs, nil
	}

	// config vars
	configVarsMapInterfaceOrig := ocrRequest.EngineArgs["config_vars"]

	if configVarsMapInterfaceOrig != nil {

		logg.LogTo("OCR_TESSERACT", "got configVarsMap: %v type: %T", configVarsMapInterfaceOrig, configVarsMapInterfaceOrig)

		configVarsMapInterface := configVarsMapInterfaceOrig.(map[string]interface{})

		configVarsMap := make(map[string]string)
		for k, v := range configVarsMapInterface {
			v, ok := v.(string)
			if !ok {
				return nil, fmt.Errorf("Could not convert configVar into string: %v", v)
			}
			configVarsMap[k] = v
		}

		engineArgs.configVars = configVarsMap

	}

	// page seg mode
	pageSegMode := ocrRequest.EngineArgs["psm"]
	if pageSegMode != nil {
		pageSegModeStr, ok := pageSegMode.(string)
		if !ok {
			return nil, fmt.Errorf("Could not convert psm into string: %v", pageSegMode)
		}
		engineArgs.pageSegMode = pageSegModeStr
	}

	// language
	lang := ocrRequest.EngineArgs["lang"]
	if lang != nil {
		langStr, ok := lang.(string)
		if !ok {
			return nil, fmt.Errorf("Could not convert lang into string: %v", lang)
		}
		engineArgs.lang = langStr
	}

	return engineArgs, nil

}

// return a slice that can be passed to tesseract binary as command line
// args, eg, ["-c", "tessedit_char_whitelist=0123456789", "-c", "foo=bar"]
func (t TesseractEngineArgs) Export() []string {
	result := []string{}
	for k, v := range t.configVars {
		result = append(result, "-c")
		keyValArg := fmt.Sprintf("%s=%s", k, v)
		result = append(result, keyValArg)
	}
	if t.pageSegMode != "" {
		result = append(result, "-psm")
		result = append(result, t.pageSegMode)
	}
	if t.lang != "" {
		result = append(result, "-l")
		result = append(result, t.lang)
	}

	return result
}

func (t TesseractEngine) ProcessRequest(ocrRequest OcrRequest) (OcrResult, error) {

	tmpDir, err := func() (string, error) {
		if ocrRequest.ImgUrl != "" {
			return t.tmpFileFromImageUrl(ocrRequest.ImgUrl)
		} else {
			return t.tmpFilesFromImageFiles(ocrRequest.ImgFiles, ocrRequest.Name)
		}

	}()

	if err != nil {
		logg.LogTo("OCR_TESSERACT", "error getting tmpDir")
		return OcrResult{}, err
	}

	defer os.RemoveAll(tmpDir)

	engineArgs, err := NewTesseractEngineArgs(ocrRequest)
	if err != nil {
		logg.LogTo("OCR_TESSERACT", "error getting engineArgs")
		return OcrResult{}, err
	}

	ocrResult, err := t.processImageFile(tmpDir, *engineArgs, ocrRequest.Name)

	return ocrResult, err

}

// func (t TesseractEngine) tmpFileFromImageBytes(ImgFiles [][]byte, name string) (string, error) {
// 	tmpFileName := filepath.Join(os.TempDir(), name)
// 	logg.LogTo("OCR_TESSERACT", "Test: %s", tmpFileName)

// 	// we have to write the contents of the image url to a temp
// 	// file, because the leptonica lib can't seem to handle byte arrays
// 	err := saveBytesToFileName(ImgFiles, tmpFileName)
// 	if err != nil {
// 		return "", err
// 	}

// 	return tmpFileName, nil

// }

func (t TesseractEngine) tmpFilesFromImageFiles(ImgFiles [][]byte, name string) (string, error) {
	tmpDir, _ := ioutil.TempDir(os.TempDir(), "pages_")
	for index, element := range ImgFiles {
		// index is the index where we are
		// element is the element from someSlice for where we are
		tmpFileName := filepath.Join(tmpDir, "temppdf_"+fmt.Sprintf("%03d", index))
		logg.LogTo("OCR_TESSERACT", "Test: %s", tmpFileName)

		// we have to write the contents of the image url to a temp
		// file, because the leptonica lib can't seem to handle byte arrays
		err := saveBytesToFileName(element, tmpFileName)
		if err != nil {
			return "", err
		}
	}

	return tmpDir, nil
}

func (t TesseractEngine) tmpFileFromImageUrl(imgUrl string) (string, error) {

	tmpFileName := func(url string) string {
		baseName := filepath.Base(url)
		extension := filepath.Ext(baseName)
		baseNameNoExt := url[0 : len(baseName)-len(extension)]
		return filepath.Join(os.TempDir(), baseNameNoExt)
	}(imgUrl)
	// we have to write the contents of the image url to a temp
	// file, because the leptonica lib can't seem to handle byte arrays
	err := saveUrlContentToFileName(imgUrl, tmpFileName)
	if err != nil {
		return "", err
	}

	return tmpFileName, nil

}

func (t TesseractEngine) processImageFile(tmpDirIn string, engineArgs TesseractEngineArgs, name string) (OcrResult, error) {

	tmpDirOut, _ := ioutil.TempDir(os.TempDir(), "pages_")
	defer os.RemoveAll(tmpDirOut)

	// possible file extensions
	// fileExtensions := []string{"pdf"}

	var combinedArgs []string
	var compressedArgs []string

	// build args array
	cflags := engineArgs.Export()

	err_walk := filepath.Walk(tmpDirIn, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			logg.LogFatal("Error running command: %s.", err)
		}
		logg.LogTo("OCR_TESSERACT", "Path is: %s. Name is: %s.", path, f.Name())
		matched, _ := regexp.MatchString("^temppdf_[0-9]{3}$", f.Name())
		if matched {
			tmpFileOut := filepath.Join(tmpDirOut, f.Name())
			cmdArgs := []string{path, tmpFileOut, "pdf"}
			cmdArgs = append(cmdArgs, cflags...)
			logg.LogTo("OCR_TESSERACT", "cmdArgs: %v", cmdArgs)
			// exec tesseract
			cmd := exec.Command("tesseract", cmdArgs...)
			output, err_exec := cmd.CombinedOutput()
			if err_exec != nil {
				logg.LogTo("OCR_TESSERACT", "Error exec tesseract: %v %v", err_exec, string(output))
				return err_exec
			}
			combinedArgs = append(combinedArgs, tmpFileOut+".pdf")
		}
		return nil
	})
	if err_walk != nil {
		logg.LogFatal("Error running command: %s.", err_walk)
	}

	tmpOutCombinedPdf, err := createTempFileName()
	tmpOutCombinedPdf = fmt.Sprintf("%s.pdf", tmpOutCombinedPdf)
	if err != nil {
		return OcrResult{}, err
	}
	defer os.Remove(tmpOutCombinedPdf)

	tmpOutCompressedPdf, err := createTempFileName()
	tmpOutCompressedPdf = fmt.Sprintf("%s.pdf", tmpOutCompressedPdf)
	if err != nil {
		return OcrResult{}, err
	}
	defer os.Remove(tmpOutCompressedPdf)

	// allFileOut := filepath.Join(tmpDirOut, "*.pdf")

	combinedArgs = append(combinedArgs, "cat", "output", tmpOutCombinedPdf)
	logg.LogTo("OCR_TESSERACT", "combinedArgs: %v", combinedArgs)
	out_pdftk, err_pdftk := exec.Command("pdftk", combinedArgs...).CombinedOutput()
	if err_pdftk != nil {
		logg.LogFatal("Error running command: %s.  out: %s", err_pdftk, out_pdftk)
	}
	logg.LogTo("OCR_TESSERACT", "output: %v", string(out_pdftk))

	compressedArgs = append(
		compressedArgs,
		"-sDEVICE=pdfwrite",
		"-dCompatibilityLevel=1.4",
		"-dPDFSETTINGS=/screen",
		"-dNOPAUSE",
		"-dBATCH",
		"-dQUIET",
		"-sOutputFile="+tmpOutCompressedPdf,
		tmpOutCombinedPdf,
	)

	out_qpdf, err_qpdf := exec.Command("gs", compressedArgs...).CombinedOutput()
	if err_qpdf != nil {
		logg.LogFatal("Error running command: %s.  out: %s", err_qpdf, out_qpdf)
	}

	outBytes, err := ioutil.ReadFile(tmpOutCompressedPdf)
	// outBytes, outFile, err := findAndReadOutfile(tmpOutCompressedPdf, fileExtensions)

	// delete output file when we are done
	// defer os.Remove(outFile)

	if err != nil {
		logg.LogTo("OCR_TESSERACT", "Error getting data from out file: %v", err)
		return OcrResult{}, err
	}

	return OcrResult{
		Text:         string(outBytes),
		BaseFileName: fmt.Sprintf("%v_ocrd", name),
	}, nil

}

// func findOutfile(outfileBaseName string, fileExtensions []string) (string, error) {

// 	for _, fileExtension := range fileExtensions {

// 		outFile := fmt.Sprintf("%v.%v", outfileBaseName, fileExtension)
// 		logg.LogTo("OCR_TESSERACT", "checking if exists: %v", outFile)

// 		if _, err := os.Stat(outFile); err == nil {
// 			return outFile, nil
// 		}

// 	}

// 	return "", fmt.Errorf("Could not find outfile.  Basename: %v Extensions: %v", outfileBaseName, fileExtensions)

// }

// func findAndReadOutfile(outfileBaseName string, fileExtensions []string) ([]byte, string, error) {

// 	outfile, err := findOutfile(outfileBaseName, fileExtensions)
// 	if err != nil {
// 		return nil, "", err
// 	}
// 	outBytes, err := ioutil.ReadFile(outfile)
// 	if err != nil {
// 		return nil, "", err
// 	}
// 	return outBytes, outfile, nil

// }
