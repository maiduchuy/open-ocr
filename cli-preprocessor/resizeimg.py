import numpy as np
import sys
import cv2

def main():
  imgIn = sys.argv[1]
  imgOut = sys.argv[2]

  img = cv2.imread(imgIn, 0)
  res = cv2.resize(img, None, fx=3, fy=3, interpolation=cv2.INTER_CUBIC)
  cv2.imwrite(imgOut,res)

if __name__ == "__main__":
    main()