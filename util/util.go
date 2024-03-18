package util

import (
	"github.com/chenjie199234/im/config"

	"github.com/chenjie199234/Corelib/container/list"
)

var worker *list.BlockList[func()]

func Init() {
	worker = list.NewBlockList[func()]()
	workernum := config.GetUnicastWorker()
	if workernum == 0 {
		workernum = 100
	}
	for i := 0; i < int(workernum); i++ {
		go func() {
			for {
				f, ok := worker.Pop()
				if !ok {
					//closed
					return
				}
				f()
			}
		}()
	}
}
func AddWork(f func()) {
	worker.Push(f)
}
