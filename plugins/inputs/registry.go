package inputs

var Inputs = map[string]interface{}{}

func Add(name string, creator interface{}) {
	Inputs[name] = creator
}
