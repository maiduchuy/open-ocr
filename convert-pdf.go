package ocrworker

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

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
	logg.LogTo(
		"PREPROCESSOR_WORKER",
		"Temp dir is: %s",
		tmpDir,
	)

	out, err := exec.Command(
		"convert",
		"-density",
		"300",
		"-depth",
		"8",
		"-alpha",
		"Off",
		tmpFileNameInput,
		tmpFileNameOutput,
	).CombinedOutput()
	if err != nil {
		logg.LogFatal("Error running command: %s.  out: %s", err, out)
	}
	logg.LogTo("PREPROCESSOR_WORKER", "output: %v", string(out))

	// read bytes from output file into ocrRequest.ImgBytes
	resultBytes, err := ioutil.ReadFile(tmpFileNameOutput)
	if err != nil {
		return err
	}

	ocrRequest.ImgBytes = resultBytes

	return nil
}
