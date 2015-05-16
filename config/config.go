package config

type Configer interface {
	ConfigStruct() interface{}
	Init(interface{}) error
}
