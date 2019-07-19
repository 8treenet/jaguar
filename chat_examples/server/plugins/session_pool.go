package plugins

import "sync"

func init() {
	sessionPool = make(map[uint32]*Session)
}

var (
	mutex       sync.Mutex
	sessionPool map[uint32]*Session
)

func GetSession(uid uint32) *Session {
	defer mutex.Unlock()
	mutex.Lock()
	return sessionPool[uid]
}

func SetSession(uid uint32, s *Session) {
	defer mutex.Unlock()
	mutex.Lock()
	sessionPool[uid] = s
}

func RemoveSession(uid uint32) {
	defer mutex.Unlock()
	mutex.Lock()
	delete(sessionPool, uid)
}

func AllSession() (result []*Session) {
	defer mutex.Unlock()
	mutex.Lock()

	for _, session := range sessionPool {
		result = append(result, session)
	}
	return
}
