package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
)

var binDir string

func TestMain(m *testing.M) {
	// 编译 pbrew 二进制到临时目录（从项目根目录构建）
	var err error
	binDir, err = os.MkdirTemp("", "pbrew-e2e-*")
	if err != nil {
		panic(err)
	}
	binPath := filepath.Join(binDir, "pbrew")

	// 获取项目根目录（e2e/ 的上级）
	projectRoot, err := filepath.Abs(filepath.Join(".."))
	if err != nil {
		os.RemoveAll(binDir)
		panic(err)
	}

	cmd := exec.Command("go", "build", "-o", binPath, ".")
	cmd.Dir = projectRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		os.RemoveAll(binDir)
		panic(err)
	}

	// 将二进制目录加入 PATH，使 testscript 能找到 pbrew
	origPath := os.Getenv("PATH")
	os.Setenv("PATH", binDir+string(os.PathListSeparator)+origPath)

	code := m.Run()
	os.RemoveAll(binDir)
	os.Exit(code)
}

func TestScripts(t *testing.T) {
	marketplaceDir, err := filepath.Abs(filepath.Join("testdata", "AgentComponentMarketPlace"))
	if err != nil {
		t.Fatal(err)
	}

	testscript.Run(t, testscript.Params{
		Dir: filepath.Join("testdata", "script"),
		Setup: func(env *testscript.Env) error {
			env.Vars = append(env.Vars,
				"MARKETPLACE="+marketplaceDir,
				"HOME="+filepath.Join(env.WorkDir, "home"),
			)
			return nil
		},
		Cmds: map[string]func(*testscript.TestScript, bool, []string){
			"gitinit":   gitinitCmd,
			"writefile": writefileCmd,
		},
	})
}

// gitinit 创建一个伪 git 仓库目录并切换工作目录，使 FindProjectRoot 能正确返回项目根目录。
// 用法：gitinit
// 该命令在当前工作目录下创建 project/ 子目录，内含 .git/HEAD 文件，并将工作目录切换到 project/。
func gitinitCmd(ts *testscript.TestScript, neg bool, args []string) {
	if neg {
		ts.Fatalf("gitinit does not support negation")
	}
	work := ts.Getenv("WORK")
	dir := filepath.Join(work, "project")
	ts.Check(os.MkdirAll(filepath.Join(dir, ".git"), 0755))
	ts.Check(os.WriteFile(filepath.Join(dir, ".git", "HEAD"), []byte("ref: refs/heads/main\n"), 0644))
	ts.Check(os.Chdir(dir))
}

// writefile 将内容写入指定文件。用法：writefile <path> <content...>
// 所有空格后的参数合并为一个字符串，适合写入单行 JSON 等内容。
// 如需写入多行内容，请使用 testscript 的 txtar 文件语法创建文件后 cp。
func writefileCmd(ts *testscript.TestScript, neg bool, args []string) {
	if neg {
		ts.Fatalf("writefile does not support negation")
	}
	if len(args) < 2 {
		ts.Fatalf("usage: writefile <path> <content...>")
	}
	path := args[0]
	content := strings.Join(args[1:], " ")
	ts.Check(os.MkdirAll(filepath.Dir(path), 0755))
	ts.Check(os.WriteFile(path, []byte(content), 0644))
}
