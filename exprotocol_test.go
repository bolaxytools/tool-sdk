package sdk

import (
	"fmt"
	"testing"
)

func TestWrapUnwrap(t *testing.T) {
	str := "Go 爱好者 "
	wrapData := WrapData(DataDisplay, []byte(str))

	fmt.Printf("TestWrapData WrapData str length:%d, wrap data lenth:%d\n", len(str), len(wrapData))

	ver, category, data := UnWrapData(wrapData)
	fmt.Printf("TestWrapData UnWrapData version:%d, category:%d, data:%s\n", ver, category, string(data))
}
