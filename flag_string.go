// Code generated by "stringer -type Flag"; DO NOT EDIT

package git

import "fmt"

const _Flag_name = "MustExitMustPanicWarnVerboseLocalOnly"

var _Flag_index = [...]uint8{0, 8, 17, 21, 28, 37}

func (i Flag) String() string {
	i -= 7
	if i < 0 || i >= Flag(len(_Flag_index)-1) {
		return fmt.Sprintf("Flag(%d)", i+7)
	}
	return _Flag_name[_Flag_index[i]:_Flag_index[i+1]]
}
