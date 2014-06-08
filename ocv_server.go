package main

//#cgo CFLAGS: -I/usr/local/include/opencv -I/usr/local/include
//#cgo LDFLAGS: /usr/local/lib/libopencv_calib3d.dylib /usr/local/lib/libopencv_contrib.dylib /usr/local/lib/libopencv_core.dylib /usr/local/lib/libopencv_features2d.dylib /usr/local/lib/libopencv_flann.dylib /usr/local/lib/libopencv_gpu.dylib /usr/local/lib/libopencv_highgui.dylib /usr/local/lib/libopencv_imgproc.dylib /usr/local/lib/libopencv_legacy.dylib /usr/local/lib/libopencv_ml.dylib /usr/local/lib/libopencv_nonfree.dylib /usr/local/lib/libopencv_objdetect.dylib /usr/local/lib/libopencv_photo.dylib /usr/local/lib/libopencv_stitching.dylib /usr/local/lib/libopencv_video.dylib /usr/local/lib/libopencv_videostab.dylib
//#include <cv.h>
//#include <highgui.h>
import "C"
import "unsafe"
import (
	"bytes"
	"code.google.com/p/go.net/websocket"
	"encoding/base64"
	"fmt"
	"github.com/gorilla/mux"
	"image"
	"image/color"
	"image/png"
	"io"
	"net/http"
	"time"
)

var capture *C.CvCapture
var cv_image *C.IplImage
var grad_image *C.IplImage
var vis_image *C.IplImage

const width, height = 160, 120

func ToBase64(img *image.RGBA) string {
	imgBuf := new(bytes.Buffer)
	imgEncoder := base64.NewEncoder(base64.StdEncoding, imgBuf)
	png.Encode(imgEncoder, img)
	imgEncoder.Close()
	return imgBuf.String()
}

func UIHandler(ws *websocket.Conn) {

	image := image.NewRGBA(image.Rect(0, 0, width, height))

	for {
		//var msg string
		//websocket.Message.Receive(ws, &msg)

		time.Sleep(100 * time.Millisecond)

		for i := 0; i < height; i++ {
			for j := 0; j < width; j++ {

				var value C.CvScalar
				value = C.cvGet2D(unsafe.Pointer(vis_image), C.int(i), C.int(j))

				color := color.RGBA{
					uint8(value.val[2]),
					uint8(value.val[1]),
					uint8(value.val[0]),
					255}

				image.Set(j, i, color)
			}
		}

		drawMsg := fmt.Sprintf("DRAW %s", ToBase64(image))

		io.WriteString(ws, drawMsg)
	}
}

func main() {

	router := mux.NewRouter()
	router.Handle("/ws", websocket.Handler(UIHandler))
	staticHandler := http.FileServer(http.Dir("."))
	router.PathPrefix("/").Handler(staticHandler)
	go http.ListenAndServe("localhost:1234", router)

	// now show webcam images here
	windowTitle := C.CString("img")
	defer C.free(unsafe.Pointer(windowTitle))
	C.cvNamedWindow(windowTitle, 1)

	capture = C.cvCaptureFromCAM(0)
	grad_image = C.cvCreateImage(C.cvSize(640, 480), 8, 3)
	vis_image = C.cvCreateImage(C.cvSize(width, height), 8, 3)

	for {
		cv_image = C.cvQueryFrame(capture)

		C.cvSmooth(unsafe.Pointer(cv_image), unsafe.Pointer(cv_image), C.CV_GAUSSIAN, 5, 0, 0, 0)
		C.cvLaplace(unsafe.Pointer(cv_image), unsafe.Pointer(grad_image), 5)

		C.cvResize(unsafe.Pointer(grad_image), unsafe.Pointer(vis_image), C.CV_INTER_AREA)

		C.cvShowImage(windowTitle, unsafe.Pointer(grad_image))

		key := C.cvWaitKey(1)
		if key == 27 {
			break
		}
	}

	C.cvWaitKey(0)

}
