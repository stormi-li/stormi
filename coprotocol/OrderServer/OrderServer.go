package OrderServer

// 方法
const (
	Func1 = iota
	Func2
)

// 数据传输结构体
type OrderServerDto struct {
	Code    int
	Message string
	Data    string
}

