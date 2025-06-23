package tgutils

import (
	"regexp"
	"strings"
)

// Splitter интерфейс для различных стратегий разделения
type Splitter interface {
	Split(text string) []string
}

// SimpleSplitter - разделение по пробелам
type SimpleSplitter struct{}

func (s SimpleSplitter) Split(text string) []string {
	return strings.Fields(text)
}

// RegexSplitter - разделение по регулярному выражению
type RegexSplitter struct {
	pattern string
}

func (r RegexSplitter) Split(text string) []string {
	re := regexp.MustCompile(r.pattern)
	return re.FindAllString(text, -1)
}

// CommandAdapter - адаптер для обработки команд
type CommandAdapter struct {
	splitter Splitter
}

func NewCommandAdapter(splitter Splitter) *CommandAdapter {
	return &CommandAdapter{splitter: splitter}
}

// ParseCommand разбирает команду и возвращает название и аргументы
func (ca *CommandAdapter) ParseCommand(text string) (string, []string) {
	if !strings.HasPrefix(text, "/") {
		return "", nil
	}

	parts := ca.splitter.Split(text)
	if len(parts) == 0 {
		return "", nil
	}

	command := strings.TrimPrefix(parts[0], "/")
	var args []string
	if len(parts) > 1 {
		args = parts[1:]
	}

	return command, args
}
