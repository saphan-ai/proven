//go:build unix

package cli

import (
	"fmt"
	"os"
	"reflect"
)

// linkCount is how many directory names a file has.
// On Linux and Darwin the status value carries that number in the exported
// field Nlink (uint64 on Linux, uint16 on Darwin). The field is read by
// name. Package syscall is not imported: a production file under internal/
// must not import it.
func linkCount(path string) (uint64, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return 0, err
	}
	sys := info.Sys()
	if sys == nil {
		return 0, fmt.Errorf("file status has no link count")
	}
	v := reflect.ValueOf(sys)
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return 0, fmt.Errorf("file status has no link count")
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return 0, fmt.Errorf("file status has no link count")
	}
	field := v.FieldByName("Nlink")
	if !field.IsValid() || !field.CanUint() {
		return 0, fmt.Errorf("file status has no link count")
	}
	return field.Uint(), nil
}
