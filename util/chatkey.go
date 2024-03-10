package util

import (
	"errors"
	"sort"
	"strings"
)

func FormChatKey(sender, target, targetType string) (chatkey string) {
	if targetType == "user" {
		strs := []string{sender, target}
		sort.Strings(strs)
		chatkey = strings.Join(strs, "-")
	} else {
		chatkey = target
	}
	return
}
func ParseChatKey(sender, chatkey string) (target string, targetType string, e error) {
	if strings.Contains(chatkey, sender) {
		targetType = "user"
		pieces := strings.Split(chatkey, "-")
		if len(pieces) != 2 {
			//this is impossible
			e = errors.New("chat key format wrong")
		} else if pieces[0] == sender {
			target = pieces[1]
		} else if pieces[1] == sender {
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
