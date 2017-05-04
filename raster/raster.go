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
	"unsafe"
)

const SIZE_OF_UINT16 = 2

type RasterType string

const (
	BOOL    = RasterType("BOOL")
	UINT8   = RasterType("UINT8")
	INT16   = RasterType("INT16")
	UINT16  = RasterType("UINT16")
	FLOAT32 = RasterType("FLOAT32")
)

type FlexRaster struct {
	RasterType
	Width, Height int
	Data          []float32
	NoData        float32
}

func GetRaster(band string) (*FlexRaster, error) {
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
        fmt.Println("Raster", nXSize, nYSize)
	nodata := float32(uint16(C.GDALGetRasterNoDataValue(hBand, nil)))
	canvas := make([]uint16, int(nXSize*nYSize))
	C.GDALRasterIO(hBand, C.GF_Read, 0, 0, nXSize, nYSize, unsafe.Pointer(&canvas[0]), nXSize, nYSize, C.GDT_UInt16, 0, 0)

	out := make([]float32, int(nXSize*nYSize))
	for i, value := range canvas {
		out[i] = float32(value)
	}

	return &FlexRaster{UINT16, int(nXSize), int(nYSize), out, nodata}, nil
}
