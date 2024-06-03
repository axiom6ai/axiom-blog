package globalInit

import (
	"axiom-blog/pkg/util"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

var App = &app{}

type app struct {
	Name    string
	Version string
	Date    time.Time

	// 项目根目录
	RootDir string

	// 启动时间
	LaunchTime time.Time
	Uptime     time.Duration

	Year int

	Domain string
	Desc   map[string]string

	Build struct {
		GitCommitLog string
		BuildTime    string
		GitRelease   string
		GoVersion    string
		GinVersion   string
	}

	Env string
}

func init() {
	App.Version = "V1.0"
	App.LaunchTime = time.Now()
	App.Year = time.Now().Year()

	// 默认在项目根目录运行程序
	App.RootDir = "."

	// 用来处理单元测试或当前目录不在根目录，获取项目根目录
	if !viper.InConfig("http.port") {
		App.RootDir = inferRootDir()
	}

	fileInfo, err := os.Stat(os.Args[0])
	if err != nil {
		panic(fmt.Sprintf("获取程序信息失败:%v", err))
	}

	App.Date = fileInfo.ModTime()
	App.Build.GoVersion = runtime.Version()
	App.Build.GinVersion = gin.Version
}

// inferRootDir 递归推导项目根目录
func inferRootDir() string {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	var infer func(d string) string
	infer = func(d string) string {
		if d == "/" {
			panic(fmt.Sprintf("请确保在项目根目录或子目录下运行程序，当前在:%v", cwd))
		}

		if util.Exist(d + "/config") {
			return d
		}

		return infer(filepath.Dir(d))
	}

	return infer(cwd)
}

func (a *app) FillBuildInfo(gitCommitLog, buildTime, gitRelease string) {
	App.Name = conf.Name
	App.Domain = fmt.Sprintf(conf.Host, conf.Port)
	App.Desc = map[string]string{
		"desc": conf.Desc.Desc,
		"key":  conf.Desc.Keyword,
	}

	a.Build.GitCommitLog = gitCommitLog
	a.Build.BuildTime = buildTime

	pos := strings.Index(gitRelease, "/")
	if pos >= -1 {
		a.Build.GitRelease = gitRelease[pos+1:]
	}
}

// SetFrameMode debug\test\release
func (a *app) SetFrameMode(mode string) {
	gin.SetMode(mode)
	App.Env = gin.Mode()
}
