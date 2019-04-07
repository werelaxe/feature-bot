package main

import (
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"
)

func buildMembersMessage(info *Info) string {
	var users []string
	for _, user := range info.Users {
		users = append(users, user.Name)
	}
	return fmt.Sprintf("Люди, которые могут быть особенными:\n%v.", strings.Join(users, "\n"))
}

func getSuffix(cnt int) string {
	count := fmt.Sprint(cnt)
	if len(count) >= 2 {
		if count[len(count)-2] == '1' {
			return ""
		}
	}
	if strings.HasSuffix(count, "2") || strings.HasSuffix(count, "3") || strings.HasSuffix(count, "4") {
		return "а"
	}
	return ""
}

func formatName(name string, maxLen int) string {
	return name + strings.Repeat(" ", maxLen-utf8.RuneCountInString(name))
}

func buildStatMessage(info *Info, sorted bool) string {
	maxLen := -1
	for _, user := range info.Users {
		if utf8.RuneCountInString(user.Name) > maxLen {
			maxLen = utf8.RuneCountInString(user.Name)
		}
	}
	var users []User
	for _, user := range info.Users {
		users = append(users, user)
	}
	if sorted {
		sort.Slice(users, func(i, j int) bool {
			return users[i].Count > users[j].Count
		})
	}
	var statLines []string
	for _, user := range users {
		statLines = append(statLines, fmt.Sprintf("%v : %v раз%v", formatName(user.Name, maxLen), user.Count, getSuffix(user.Count)))
	}
	stat := strings.Join(statLines, "\n")
	return fmt.Sprintf("Статистика по человекам с особенностью %v:\n`%v`", info.Feature, stat)
}
