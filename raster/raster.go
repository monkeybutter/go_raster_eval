package raster

// #include "gdal.h"
// #include "cpl_string.h"
// #cgo LDFLAGS: -lgdal
// char**
// get_open_options(int level)
// {
//	  char **papszOptions = NULL;
//	  papszOptions = CSLSetNameValue(papszOptions, "OVERVIEW_LEVEL", CPLSPrintf("%d", level));
//	  return papszOptions;
// }
import "C"

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"reflect"
	"unsafe"
)

const SIZE_OF_UINT16 = 2

func GetRaster(band string) (*image.Gray16, error) {
	C.GDALAllRegister()

	path := fmt.Sprintf("/g/data3/fr5/prl900/LS8_test/LC81390452014295LGN00_%s.TIF", band)
	filePathCStr := C.CString(path)
	defer C.free(unsafe.Pointer(filePathCStr))

	//Landsat Overviews test
	opt := C.get_open_options(3)
	hSrcDS := C.GDALOpenEx(filePathCStr, C.GA_ReadOnly, nil, opt, nil)
	if hSrcDS == nil {
		return nil, fmt.Errorf("GDAL Dataset is null %v", path)
	}
	defer C.GDALClose(hSrcDS)

	hBand := C.GDALGetRasterBand(hSrcDS, 1)
	if hBand == nil {
		return nil, fmt.Errorf("Null Band returned for granule %v", path)
	}

	nXSize := C.GDALGetRasterBandXSize(hBand)
	nYSize := C.GDALGetRasterBandYSize(hBand)
	fmt.Println("Size", nXSize, nYSize)
	canvas := make([]uint16, int(nXSize*nYSize))
	C.GDALRasterIO(hBand, C.GF_Read, 0, 0, nXSize, nYSize, unsafe.Pointer(&canvas[0]), nXSize, nYSize, C.GDT_UInt16, 0, 0)
	// Get the slice header
	header := *(*reflect.SliceHeader)(unsafe.Pointer(&canvas))
	// The length and capacity of the slice are different.
	header.Len *= SIZE_OF_UINT16
	header.Cap *= SIZE_OF_UINT16

	return &image.Gray16{Pix: *(*[]byte)(unsafe.Pointer(&header)), Stride: int(nXSize * SIZE_OF_UINT16), Rect: image.Rect(0, 0, int(nXSize), int(nYSize))}, nil
}

func SaveRaster(path string, img image.Image) error {
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	err = png.Encode(out, img)
	if err != nil {
		return err
	}
	return nil
}

/*
func main() {
	img, err := GetLS8Raster("./LC81390452014295LGN00_B1.TIF")
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(len(img.Pix))
	err = SaveRaster("./out.png", img)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Done")

}*/
