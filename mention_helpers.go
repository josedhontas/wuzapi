package main

import (
	"regexp"
	"unicode"
	"unicode/utf8"

	"go.mau.fi/whatsmeow/types"
)

var mentionToken = regexp.MustCompile(`(^|[^\pL\pN_@./])@([0-9]{7,})`)

func extractMentionJIDs(text string) []string {
	var mentions []string
	for _, match := range mentionToken.FindAllStringSubmatchIndex(text, -1) {
		end := match[5]
		if end < len(text) {
			next, _ := utf8.DecodeRuneInString(text[end:])
			if unicode.IsLetter(next) || unicode.IsDigit(next) || next == '_' || next == '@' || next == '.' || next == '/' {
				continue
			}
		}
		mentions = append(mentions, text[match[4]:match[5]]+"@"+types.DefaultUserServer)
	}
	return mentions
}

func mergeMentionJIDs(explicit []string, text string) []string {
	result := make([]string, 0, len(explicit))
	seen := make(map[string]struct{}, len(explicit))
	for _, jid := range explicit {
		if _, ok := seen[jid]; ok {
			continue
		}
		seen[jid] = struct{}{}
		result = append(result, jid)
	}
	for _, jid := range extractMentionJIDs(text) {
		if _, ok := seen[jid]; ok {
			continue
		}
		seen[jid] = struct{}{}
		result = append(result, jid)
	}
	return result
}
