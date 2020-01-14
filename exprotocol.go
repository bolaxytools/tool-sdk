package sdk

const (
	prefix  byte = 0x7E // 自定义数据前缀标记 '~'
	version byte = 0x01
)

/**
	 业务分类
                +-------------------------------------------------------
     	        |   DataDisplay   |    数据展示，string(output)进行展示
*/
const (
	DataDisplay byte = iota
)

/**
	 协议打包
                +------------------------------------------
     input      |  数据
	 category   |  协议分类
*/
func WrapData(category byte, input []byte) []byte {
	rt := make([]byte, len(input)+4)
	rt[0] = prefix
	rt[1] = version
	rt[2] = category
	rt[3] = byte(len(input) + 4)
	copy(rt[4:], input)

	return rt
}

/**
	 协议解包
                +------------------------------------------
     input      |  待解码数据
     version    |  协议版本号
	 category   |  协议分类
	 out        |  数据,若不是自定义data，返回空数据
*/
func UnWrapData(input []byte) (version byte, category byte, out []byte) {
	if len(input) < 4 || input[0] != prefix {
		return
	}

	return input[1], input[2], input[4:input[3]]
}
