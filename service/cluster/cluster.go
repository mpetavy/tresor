package cluster

import (
	"strings"
	"sync"
)

var (
	master   sync.Mutex
	mutexMap map[string]*sync.Mutex
)

type lockid struct {
	id string
}

func init() {
	mutexMap = make(map[string]*sync.Mutex)
}

func STORAGE(storage string) lockid {
	return lockid{"STORAGE-" + strings.ToUpper(storage)}
}

func STORAGE_UID(storage string, uid string) lockid {
	return lockid{"STORAGE_UID-" + strings.ToUpper(storage) + "-" + strings.ToUpper(uid)}
}

func STORAGE_VOLUME(storage string, volume string) lockid {
	return lockid{"STORAGE_VOLUME-" + strings.ToUpper(storage) + "-" + strings.ToUpper(volume)}
}

func Lock(id lockid) {
	master.Lock()

	k := strings.ToUpper(id.id)

	m, ok := mutexMap[k]
	if !ok {
		m = new(sync.Mutex)
		mutexMap[k] = m
	}

	master.Unlock()

	m.Lock()
}

func Unlock(id lockid) {
	master.Lock()

	k := strings.ToUpper(id.id)

	m, ok := mutexMap[k]
	if !ok {
		return
	}

	master.Unlock()

	m.Unlock()
}
