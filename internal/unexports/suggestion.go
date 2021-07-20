package unexports

import (
	"strings"
)

// suggester 模糊匹配提示器
type suggester struct {
	key                                   string
	i, j                                  int
	suggestionA, suggestionB, suggestionC string
}

// newSuggester 创建提示器
func newSuggester(key string) *suggester {
	return &suggester{key: key}
}

// AddItem 添加匹配条目
func (s *suggester) AddItem(item string) {
	if fuzzyMatch(item, s.key, "/") {
		if s.j%3 == 0 {
			s.suggestionA = item
		} else if s.j%3 == 1 {
			s.suggestionB = item
		} else {
			s.suggestionC = item
		}
		s.j++
	} else if fuzzyMatch(item, s.key, ".") {
		if s.i%2 == 0 {
			s.suggestionB = item
		} else {
			s.suggestionC = item
		}
		s.i++
	}
}

// Suggestions 获取提示内容
func (s *suggester) Suggestions() []string {
	return []string{s.suggestionA, s.suggestionB, s.suggestionC}
}

// fuzzyMatch 模糊匹配,用于提供suggestion
func fuzzyMatch(target, source, token string) bool {
	if len(target) == 0 || len(source) == 0 || len(token) == 0 {
		return false
	}

	keywords := strings.Split(source, token)
	keyword := keywords[len(keywords)-1]
	return strings.Contains(target, keyword)
}
