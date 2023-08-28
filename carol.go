package carol

import (
	"bufio"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type Practice struct {
	Name      string
	LastDone  time.Time
	TimesDone int
	Level     string
	Topics    []string
}

// Get 获取所有的编程题, 练习的统计信息
func Get() ([]Practice, error) {
	existing, err := getExisting()
	if err != nil {
		return nil, err
	}

	// TODO: 解析题目完成记录文件
}

// getExisting 获取已存在的所有编程题
func getExisting() ([]Practice, error) {
	cmd := exec.Command("go", "list", "-f", "{{.Dir}}", "./...")

	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	var result []Practice

	for _, line := range strings.Split(string(out), "\n") {
		name := strings.TrimPrefix(line, cwd)
		name = strings.TrimPrefix(name, "/")

		if name == "" || strings.HasSuffix(name, "cmd") {
			continue
		}

		l, t, err := parsePractice(name)
		if err != nil {
			return nil, err
		}

		result = append(result, Practice{Name: name, Level: l, Topics: uniq(t)})
	}
	return result, err
}

// uniq 移除重复的练习主题
func uniq(topics []string) []string {
	seen := make(map[string]bool)

	var unique []string

	for _, t := range topics {
		if _, ok := seen[t]; !ok {
			seen[t] = true
			unique = append(unique, t)
		}
	}

	return unique
}

func parsePractice(name string) (level string, topics []string, err error) {
	fn := func(path string, d fs.DirEntry, err error) error {
		if filepath.Ext(path) == ".go" {
			f, err := os.Open(path)
			if err != nil {
				return err
			}
			defer f.Close()

			s := bufio.NewScanner(f)
			for s.Scan() {
				line := s.Text()

				if strings.HasPrefix(line, "// Level:") {
					level = grepLevel(s.Text())
				}

				if strings.HasPrefix(line, "// Topics:") {
					topics = append(topics, grepTopics(s.Text())...)
				}
			}

			if err := s.Err(); err != nil {
				return err
			}
		}
		return nil
	}

	absPath, err := filepath.Abs(name)
	if err != nil {
		return "", nil, err
	}

	err = filepath.WalkDir(absPath, fn)
	if err != nil {
		return "", nil, err
	}

	return level, topics, err
}

func grepLevel(line string) string {
	_, level, _ := strings.Cut(line, ":")
	return strings.TrimSpace(level)
}

func grepTopics(line string) []string {
	_, topicsStr, _ := strings.Cut(line, ":")
	topics := strings.Split(topicsStr, ",")

	for i := range topics {
		topics[i] = strings.TrimSpace(topics[i])
	}

	return topics
}
