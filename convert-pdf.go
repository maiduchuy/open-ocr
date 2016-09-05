package ocrworker

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/couchbaselabs/logg"
)

type ConvertPdf struct {
}

func (c ConvertPdf) preprocess(ocrRequest *OcrRequest) error {

	// write bytes to a temp file

	tmpFileNameInput, err := createTempFileName()
	tmpFileNameInput = fmt.Sprintf("%s.pdf", tmpFileNameInput)
	if err != nil {
		return err
	}
	defer os.Remove(tmpFileNameInput)

	tmpFileNameOutput, err := createTempFileName()
	tmpFileNameOutput = fmt.Sprintf("%s.png", tmpFileNameOutput)
	if err != nil {
		return err
	}
	defer os.Remove(tmpFileNameOutput)

	err = saveBytesToFileName(ocrRequest.ImgBytes, tmpFileNameInput)
	if err != nil {
		return err
	}

	logg.LogTo(
		"PREPROCESSOR_WORKER",
		"Convert PDF on %s -> %s",
		tmpFileNameInput,
		tmpFileNameOutput,
	)

	tmpDir := createTempDir()
	defer os.RemoveAll(tmpDir)

	tmpDirFiles := fmt.Sprintf("%s/%s_%s.pdf", tmpDir, ocrRequest.Name, "%03d")
	logg.LogTo(
		"PREPROCESSOR_WORKER",
		"Temp dir is: %s",
		tmpDir,
	)
	logg.LogTo(
		"PREPROCESSOR_WORKER",
		"Temp dir files is: %s",
		tmpDirFiles,
	)

	out_pdftk, err_pdftk := exec.Command(
		"pdftk",
		tmpFileNameInput,
		"burst",
		"dont_ask",
		"output",
		tmpDirFiles,
	).CombinedOutput()
	if err != nil {
		logg.LogFatal("Error running command: %s.  out: %s", err_pdftk, out_pdftk)
	}
	logg.LogTo("PREPROCESSOR_WORKER", "output: %v", string(out_pdftk))

	err_walk := filepath.Walk(tmpDir, func(path string, f os.FileInfo, err error) error {
		out_imagemagick, err_imagemagick := exec.Command(
			"convert",
			"-density",
			"300",
			"-depth",
			"8",
			"-alpha",
			"Off",
			path,
			tmpFileNameOutput,
		).CombinedOutput()
		if err != nil {
			logg.LogFatal("Error running command: %s.  out: %s", err_imagemagick, out_imagemagick)
		}
		logg.LogTo("PREPROCESSOR_WORKER", "output: %v", string(out_imagemagick))
		return nil
	})
	if err_walk != nil {
		logg.LogFatal("Error running command: %s.", err_walk)
	}

	// read bytes from output file into ocrRequest.ImgBytes
	resultBytes, err := ioutil.ReadFile(tmpFileNameOutput)
	if err != nil {
		return err
	}

	ocrRequest.ImgBytes = resultBytes

	return nil
}
