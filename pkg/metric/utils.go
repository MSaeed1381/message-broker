package metric

type Status bool

const (
	SUCCESS Status = true
	FAILURE Status = false
)

type Method string

const (
	Publish   Method = "Publish"
	Subscribe Method = "Subscribe"
	Fetch     Method = "Fetch"
)

func StatusToStr(status Status) string {
	if status == SUCCESS {
		return "SUCCESS"
	}
	return "FAILURE"
}

func MethodToStr(method Method) string {
	switch method {
	case Publish:
		return "Publish"
	case Subscribe:
		return "Subscribe"
	case Fetch:
		return "Fetch"
	default:
		return "Undefined"
	}
}
