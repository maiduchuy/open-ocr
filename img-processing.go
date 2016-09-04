package ocrworker

import (
  "fmt"
  "io/ioutil"
  "os"
  "os/exec"

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
    "Image Processing on %s -> %s",
    tmpFileNameInput,
    tmpFileNameOutput,
  )

  out, err := exec.Command(
    "python",
    "/opt/go/src/github.com/maiduchuy/image-processing/resizeimg.py",
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
