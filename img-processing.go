package ocrworker

import (
  "fmt"
  "os"

  "github.com/couchbaselabs/logg"
)

type ImageProcessing struct {
}

func (s ImageProcessing) preprocess(ocrRequest *OcrRequest) error {

  // write bytes to a temp file

  tmpFileNameInput, err := createTempFileName()
  tmpFileNameInput = fmt.Sprintf("%s.png", tmpFileNameInput)
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
    "DetectText on %s -> %s with %s",
    tmpFileNameInput,
    tmpFileNameOutput,
    5558,
  )

  return nil
}
