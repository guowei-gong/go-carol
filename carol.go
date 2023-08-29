package carol

import (
	"bufio"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// ExerciseBook 练习册用于记录已经完成的练习题, 它的内容格式为 `- Y-m-d: foldName` 如下
// - 2023-08-29: data_structure
// - 2023-08-29: concurrency/channel, concurrency/mutex
const ExerciseBook = "carol.md"

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

	// 解析题目完成记录文件
	done, err := getDone()
	if err != nil {
		return nil, err
	}

HERE:
	for _, d := range done {
		for _, e := range existing {
			if d.Name == e.Name {
				continue HERE
			}
		}
	}

	for i := range existing {
		for j := range done {
			if existing[i].Name == done[j].Name {
				existing[i].TimesDone = done[j].TimesDone
				existing[i].LastDone = done[j].LastDone
			}
		}
	}

	return existing, nil
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

// getDone 从 Carol 中返回已完成的练习题
func getDone() ([]Practice, error) {
	var result []Practice

	f, err := os.Open(ExerciseBook)
	if err != nil {
		return nil, err
	}

	bookLineRE := regexp.MustCompile(`^\s*\*\s*([0-9]{4}\-[0-9]{2}\-[0-9]{2}):\s*(.+)$`)
	comaRE := regexp.MustCompile(`\s*,\s*`)

	practices := make(map[string]Practice)

	s := bufio.NewScanner(f)
	for s.Scan() {
		lineParts := bookLineRE.FindStringSubmatch(s.Text())
		if lineParts == nil {
			continue
		}

		date, practiceStr := lineParts[1], lineParts[2]
		doneOn, err := time.Parse("2006-01-02", date)
		if err != nil {
			return nil, err
		}

		for _, name := range comaRE.Split(practiceStr, -1) {
			if name == "" {
				continue
			}

			name = strings.TrimSpace(name)

			if practice, ok := practices[name]; ok {
				practice.TimesDone++
				if doneOn.After(practice.LastDone) {
					practice.LastDone = doneOn
				}
				practices[name] = practice
			} else {
				practice.Name = name
				practice.TimesDone = 1
				practice.LastDone = doneOn
				practices[name] = practice
			}
		}
	}
	if s.Err() != nil {
		return nil, s.Err()
	}

	for name := range practices {
		result = append(result, practices[name])
	}

	return result, nil
}
