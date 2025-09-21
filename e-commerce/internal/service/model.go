package service

import(
	"sync"
)
// userStore struct which will have mutex & users map to store username & passwd
type UserStore struct{
	Mu sync.Mutex
	User map[string]string
}

type User struct {
	Username string
	Password string
}