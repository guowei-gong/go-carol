package carol

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"text/tabwriter"
	"time"
)

// ExerciseBook 练习册用于记录已经完成的练习题, 它的内容格式为 `- Y-m-d: foldName` 如下
// - 2023-08-29: data_structure
// - 2023-08-29: concurrency/channel, concurrency/mutex
const ExerciseBook = "exercise_book.md"

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

// Print 在标准输出中打印与练习题的统计有关的表格, 显示信息如下.
// 1. 练习题名称
// 2. 练习的次数
// 3. 练习题的难度
// 4. 练习题所属主题
// 5. 上一次完成练习时间, 以天为单位
func Print(practices []Practice, lastDoneDaysAgo int, column int, level string) {
	const format = "%v\t%v\t%5v\t%v\t%v\n"

	// 头部
	tw := new(tabwriter.Writer).Init(os.Stdout, 0, 8, 2, ' ', 0)
	fmt.Fprintf(tw, format, "Name", "Last done", "Done", "Level", "Topics")
	fmt.Fprintf(tw, format, "----", "---------", "----", "-----", "------")

	// 正文
	var totalCount int
	var practicesCount int

	sortPractices(practices, &column)
	for _, p := range practices {
		if !show(p, lastDoneDaysAgo) {
			continue
		}
		if level != "" && p.Level != level {
			continue
		}

		practicesCount++
		totalCount += p.TimesDone

		fmt.Fprintf(tw, format, p.Name, humanize(p.LastDone), fmt.Sprintf("%dx", p.TimesDone), p.Level, strings.Join(p.Topics, ", "))
	}
	// 尾部
	fmt.Fprintf(tw, format, "----", "", "----", "", "")
	fmt.Fprintf(tw, format, practicesCount, "", totalCount, "", "")

	tw.Flush()
}

type customSort struct {
	practices []Practice
	less      func(x, y Practice) bool
}

func (x customSort) Len() int           { return len(x.practices) }
func (x customSort) Less(i, j int) bool { return x.less(x.practices[i], x.practices[j]) }
func (x customSort) Swap(i, j int)      { x.practices[i], x.practices[j] = x.practices[j], x.practices[i] }

// 根据 column 对结果集进行排序, 次要排序使用练习题名称
func sortPractices(practices []Practice, column *int) {
	sort.Sort(customSort{practices, func(x, y Practice) bool {
		switch *column {
		case 1:
			if x.Name != y.Name {
				return x.Name < y.Name
			}
		case 2:
			if x.LastDone != y.LastDone {
				return x.LastDone.After(y.LastDone)
			}
		case 3:
			if x.TimesDone != y.TimesDone {
				return x.TimesDone > y.TimesDone
			}
		default:

		}
		if x.Name != y.Name {
			return x.Name < y.Name
		}
		return false
	}})
}

// show 根据时间过滤练习题
func show(p Practice, lastDoneDaysAgo int) bool {
	if lastDoneDaysAgo < 0 {
		return true
	}
	t := time.Now().Add(-time.Hour * 24 * time.Duration(lastDoneDaysAgo+1))
	return p.LastDone.After(t)
}

// humanize 格式化为符合人类阅读习惯的时间
func humanize(lastDone time.Time) string {
	if lastDone.IsZero() {
		return "never"
	}
	daysAgo := int(time.Since(lastDone).Hours() / 24)
	w := "day"
	if daysAgo != 1 {
		w += "s"
	}
	return fmt.Sprintf("%d %s ago", daysAgo, w)
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
