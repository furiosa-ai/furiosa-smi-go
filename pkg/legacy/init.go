package legacy

import "github.com/furiosa-ai/furiosa-smi-go/pkg/legacy/binding"

func Init() error {
	if ret := binding.FuriosaSmiInit(); ret != binding.FuriosaSmiReturnCodeOk {
		return toError(ret)
	}

	return nil
}
