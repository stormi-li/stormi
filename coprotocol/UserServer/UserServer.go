package UserServer

// 协议码
const (
	Insert = iota
	Update
)

// 数据传输结构体
type UserServerDto struct {
	Id       int
	UserName string
}

