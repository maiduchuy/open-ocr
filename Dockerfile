FROM ubuntu

RUN apt-get update

RUN apt-get -q -y install build-essential
RUN apt-get -q -y install cmake git libgtk2.0-dev pkg-config libavcodec-dev libavformat-dev libswscale-dev wget
RUN apt-get -q -y install python-dev python-numpy libtbb2 libtbb-dev libjpeg-dev libpng-dev libtiff-dev libjasper-dev libdc1394-22-dev
RUN apt-get -q -y install golang

RUN wget https://bootstrap.pypa.io/get-pip.py
RUN python get-pip.py
RUN pip install numpy

RUN cd / && git clone https://github.com/opencv/opencv.git
RUN cd /opencv && mkdir build && cd build && cmake -D CMAKE_BUILD_TYPE=RELEASE \
-D CMAKE_INSTALL_PREFIX=/usr/local \
-D INSTALL_C_EXAMPLES=OFF \
-D INSTALL_PYTHON_EXAMPLES=ON \
-D BUILD_EXAMPLES=ON ..

RUN cd /opencv/build && make -j2 && make install

ENV GOPATH /opt/go
RUN mkdir -p $GOPATH


RUN go get -u -v github.com/maiduchuy/open-ocr

# build open-ocr-preprocessor binary and copy it to /usr/bin
RUN cd $GOPATH/src/github.com/maiduchuy/open-ocr/cli-preprocessor && go build -v -o open-ocr-preprocessor && cp open-ocr-preprocessor /usr/bin