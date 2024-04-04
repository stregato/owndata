package main

// /*
// #include "cfunc.h"
// #include <stdlib.h>
// */
// import "C"
// import (
// 	"encoding/json"

// 	"github.com/stregato/mio/lib/core"
// )

// func cResult(v any, err error) C.Result {
// 	var res []byte

// 	if err != nil {
// 		return C.Result{nil, C.CString(err.Error())}
// 	}
// 	if v == nil {
// 		return C.Result{nil, nil}
// 	}

// 	res, err = json.Marshal(v)
// 	if err == nil {
// 		return C.Result{C.CString(string(res)), nil}
// 	}
// 	return C.Result{nil, C.CString(err.Error())}
// }

// func cInput(err error, i *C.char, v any) error {
// 	if err != nil {
// 		return err
// 	}
// 	data := C.GoString(i)
// 	return json.Unmarshal([]byte(data), v)
// }

// func cUnmarshal(i *C.char, v any) error {
// 	data := C.GoString(i)
// 	err := json.Unmarshal([]byte(data), v)
// 	if core.IsErr(err, "cannot unmarshal %s: %v", data) {
// 		return err
// 	}
// 	return nil
// }

// //export OpenSafeC
// func OpenSafeC(url, creator C.)
