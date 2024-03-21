package util

import (
	"errors"
	"sort"
	"strings"
)

func FormChatKey(self, target, targetType string) (chatkey string) {
	if targetType == "user" {
		strs := []string{self, target}
		sort.Strings(strs)
		chatkey = strings.Join(strs, "-")
	} else {
		chatkey = target
	}
	return
}
func ParseChatKey(self, chatkey string) (target string, targetType string, e error) {
	if strings.Contains(chatkey, self) {
		targetType = "user"
		pieces := strings.Split(chatkey, "-")
		if len(pieces) != 2 {
			//this is impossible
			e = errors.New("chat key format wrong")
		} else if pieces[0] == self {
			target = pieces[1]
		} else if pieces[1] == self {
			target = pieces[0]
		} else {
			//this is impossible
			e = errors.New("chat key owner wrong")
		}
	} else {
		target = chatkey
		targetType = "group"
	}
	return
}
