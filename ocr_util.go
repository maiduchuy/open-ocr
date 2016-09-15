package ocrworker

import (
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"os/exec"

	"github.com/nu7hatch/gouuid"
	"github.com/couchbaselabs/logg"
)

func saveUrlContentToFileName(url, tmpFileName string) error {

	// TODO: current impl uses more memory than it needs to

	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(tmpFileName, bodyBytes, 0600)

}

func saveBytesToFileName(bytes []byte, tmpFileName string) error {
	return ioutil.WriteFile(tmpFileName, bytes, 0600)
}

func url2bytes(url string) ([]byte, error) {

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return bodyBytes, nil

}

func createTempFileName() (string, error) {
	tempDir := os.TempDir()
	uuidRaw, err := uuid.NewV4()
	if err != nil {
		return "", err
	}
	uuidStr := uuidRaw.String()
	return filepath.Join(tempDir, uuidStr), nil
}

func createTempDir() (string, error) {
	tempDir := os.TempDir()
	return ioutil.TempDir(tempDir, "pages_")
}

func checkOCR(bytes []byte, ocrRequest *OcrRequest) {
	tmpFileName, err := createTempFileName()
	if err != nil {
		return
	}

	tmpFileName = tmpFileName + ".pdf"
	err = ioutil.WriteFile(tmpFileName, bytes, os.ModeTemporary)
	cmdArgs := []string{tmpFileName}

	cmd := exec.Command("pdffonts", cmdArgs...)
	output, err_exec := cmd.CombinedOutput()
	if err_exec != nil {
		logg.LogTo("OCR_TESSERACT", "Error exec pdffonts: %v %v", err_exec, string(output))
		return
	}
	logg.LogTo("OCR_TESSERACT", "OCR checking result: %v", string(output))
}
