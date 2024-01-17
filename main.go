package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"net"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

var pwd, _ = os.Getwd()
var Fapp fyne.App
var logTabItem *container.TabItem
var logLabel *widget.TextGrid
var logtxt = ""
var tabContainer *container.AppTabs

func main() {
	err := os.Setenv("FYNE_FONT", "C:\\Windows\\Fonts\\STSONG.TTF")
	if err != nil {
		return
	}
	Fapp := app.New()
	w := Fapp.NewWindow("plink-go-wrapper")
	w.Resize(fyne.Size{
		Width:  700,
		Height: 500,
	})
	tabContainer = container.NewAppTabs(FormSetting(), TabLog())
	w.SetContent(tabContainer)

	w.ShowAndRun()
}

func FormSetting() *container.TabItem {

	inputSSHServer := widget.NewEntry()
	inputSSHServer.SetPlaceHolder("SSH Server IP")

	inputSSHServerPort := widget.NewEntry()
	inputSSHServerPort.SetPlaceHolder("SSH Server Port")

	inputSSHServerUser := widget.NewEntry()
	inputSSHServerUser.SetPlaceHolder("SSH Server User Name")

	inputSSHServerPassword := widget.NewPasswordEntry()
	inputSSHServerPassword.SetPlaceHolder("SSH Server Password")

	inputSocksPort := widget.NewEntry()
	inputSocksPort.SetPlaceHolder("Local Socks 5 port")

	inputAutoPing := widget.NewEntry()
	inputAutoPing.SetPlaceHolder("auto ping a domain to check server status,leave blank for google.com")

	checkboxAutoConnect := widget.NewCheck("Automatically connect on startup", func(b bool) {
		fmt.Println("checkboxAutoConnect", b)
	})
	checkboxEnableCompress := widget.NewCheck("Enable Compress, a little faster", func(b bool) {
		fmt.Println("checkboxEnableCompress", b)
	})

	config, err := ConfigRead("default")
	if err != nil {
		fmt.Printf("Error reading config: %v\n", err)
	} else {
		inputSocksPort.Text = config.SocksPort
		inputSSHServer.Text = config.SSHServer
		inputSSHServerPort.Text = config.SSHServerPort
		inputSSHServerUser.Text = config.SSHServerUser
		inputSSHServerPassword.Text = config.SSHServerPassword
		inputAutoPing.Text = config.AutoPingSite
		checkboxAutoConnect.Checked = config.AutoConnectOnStartup
		checkboxEnableCompress.Checked = config.EnableCompress

	}

	fmt.Printf("Socks Port: %s\n", config.SocksPort)
	fmt.Printf("SSH Server: %s\n", config.SSHServer)
	fmt.Printf("SSH Server Port: %d\n", config.SSHServerPort)
	fmt.Printf("SSH Server User: %s\n", config.SSHServerUser)
	fmt.Printf("SSH Server Password: %s\n", config.SSHServerPassword)
	fmt.Printf("Auto Connect on Startup: %t\n", config.AutoConnectOnStartup)
	fmt.Printf("Enable Compress: %t\n", config.EnableCompress)

	buttonSave := widget.NewButton("Save & Run", func() {
		updateLogLabel("buttonSave clicked")
		config.SocksPort = inputSocksPort.Text
		config.SSHServer = inputSSHServer.Text
		config.SSHServerPort = inputSSHServerPort.Text
		config.SSHServerUser = inputSSHServerUser.Text
		config.SSHServerPassword = inputSSHServerPassword.Text
		config.AutoPingSite = inputAutoPing.Text
		config.AutoConnectOnStartup = checkboxAutoConnect.Checked
		config.EnableCompress = checkboxEnableCompress.Checked
		if nil == ConfigSave("default", config) {
			updateLogLabel("config Save success")
			go RunPlink(config)
			tabContainer.SelectIndex(1)

		}

	})

	return container.NewTabItem("setting", widget.NewForm(
		widget.NewFormItem("desc", widget.NewLabel("SETTING Form....")),
		widget.NewFormItem("SSH Server", inputSSHServer),
		widget.NewFormItem("SSH Port", inputSSHServerPort),
		widget.NewFormItem("SSH User", inputSSHServerUser),
		widget.NewFormItem("SSH Password", inputSSHServerPassword),
		widget.NewFormItem("Socks Port", inputSocksPort),
		widget.NewFormItem("Auto Ping", inputAutoPing),
		widget.NewFormItem("Auto Connect on Startup", checkboxAutoConnect),
		widget.NewFormItem("Enable Compress", checkboxEnableCompress),
		widget.NewFormItem("", buttonSave),
	))
}
func TabLog() *container.TabItem {
	//logLabel = widget.NewLabel("LOG.....")
	//logLabel.Text = ""
	logLabel = widget.NewTextGrid()

	logTabItem = container.NewTabItem("log", container.NewVBox(logLabel))
	return logTabItem
}

func updateLogLabel(text string) {
	const maxLines = 25

	// 添加新的日志行
	logtxt = logtxt + "\nlog:" + text

	// 分割文本为行
	lines := strings.Split(logtxt, "\n")

	// 检查是否超过最大行数
	if len(lines) > maxLines {
		// 删除最开始的一行
		lines = lines[1:]
	}

	// 重新组合文本
	logtxt = strings.Join(lines, "\n")

	// 设置文本到 logLabel
	logLabel.SetText(logtxt)

	fmt.Println(logtxt)
}

type SessionConfig struct {
	SocksPort            string `json:"socks_port"`
	SSHServer            string `json:"ssh_server"`
	SSHServerPort        string `json:"ssh_server_port"`
	SSHServerUser        string `json:"ssh_server_user"`
	SSHServerPassword    string `json:"ssh_server_password"`
	AutoPingSite         string `json:"auto_ping_site"`
	AutoConnectOnStartup bool   `json:"auto_connect_on_startup"`
	EnableCompress       bool   `json:"enable_compress"`
}

func ConfigRead(sessionName string) (*SessionConfig, error) {
	// 读取文件内容

	fileContent, err := os.ReadFile(pwd + "/session/" + sessionName + ".json")
	if err != nil {
		return nil, fmt.Errorf("无法读取配置文件: %v", err)
	}

	// 解析 JSON 数据
	var config SessionConfig
	err = json.Unmarshal(fileContent, &config)
	if err != nil {
		return nil, fmt.Errorf("解析配置文件出错: %v", err)
	}

	return &config, nil
}

func ConfigSave(sessionName string, config *SessionConfig) error {
	// 将配置转换为 JSON 格式
	configJSON, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		return fmt.Errorf("无法转换配置为 JSON 格式: %v", err)
	}

	// 写入文件
	filePath := pwd + "/session/" + sessionName + ".json"
	err = os.WriteFile(filePath, configJSON, 0644)
	if err != nil {
		return fmt.Errorf("无法写入配置文件: %v", err)
	}

	return nil
}

func RunPlink(config *SessionConfig) {

	cmd := exec.Command("taskkill", "/F", "/IM", "plink.exe")

	// 启动命令并等待执行完成
	err := cmd.Run()

	time.Sleep(1 * time.Second)

	portAvailable := CheckPort(config.SocksPort)
	if !portAvailable {
		updateLogLabel("Socks Port already in use.")
		return
	} else {
		updateLogLabel("Socks Port available.")
	}

	updateLogLabel("RUN:plink.exe")

	cmd = exec.Command(pwd+"/plink.exe", "-N", "-C", config.SSHServerUser+"@"+config.SSHServer+":"+config.SSHServerPort, "-pw", config.SSHServerPassword, "-D", "127.0.0.1:"+config.SocksPort)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Printf("Unable to access standard output pipe: %v\n", err)
		return
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		fmt.Printf("Unable to access standard error pipe: %v\n", err)
		return
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		fmt.Printf("Unable to access standard in pipe: %v\n", err)
		return
	}

	// 使用Scanner读取标准输出
	go func(scanner *bufio.Scanner) {
		for scanner.Scan() {
			line := scanner.Text()
			updateLogLabel("run:" + line)

			// 检测输出中是否包含"wrong"，并做相应处理
			if strings.Contains(line, "Access denied") {
				updateLogLabel("run:wrong password")
				cmd.Process.Kill()
				return
			}
		}
	}(bufio.NewScanner(stdoutPipe))

	// 使用Scanner读取标准错误
	go func(scanner *bufio.Scanner) {
		for scanner.Scan() {
			line := scanner.Text()
			updateLogLabel("err:" + line)

			// 检测输出中是否包含"wrong"，并做相应处理
			if strings.Contains(line, "Access denied") {
				updateLogLabel("err:wrong password")
				cmd.Process.Kill()
				return
			}

			if strings.Contains(line, "Store key in cache") {
				updateLogLabel("err:need auto store key")
				// 自动输入 "y"
				//cmd.Stdin = strings.NewReader("y\n")
				fmt.Fprint(stdin, "y\n")
				fmt.Fprint(stdin, "\n")
				updateLogLabel("err:auto store key finished")

				//cmd.Process.Kill()
				//return
			}
			if strings.EqualFold(line, "connection.") {
				updateLogLabel("err:need auto refresh key")
				// 自动输入 "y"
				//cmd.Stdin = strings.NewReader("y\n")
				fmt.Fprint(stdin, "y\n")
				fmt.Fprint(stdin, "\n")
				updateLogLabel("err:auto store key finished")

				//cmd.Process.Kill()
				//return
			}
		}
	}(bufio.NewScanner(stderrPipe))

	// 启动命令
	err = cmd.Start()
	if err != nil {
		fmt.Printf("启动命令出错: %v\n", err)
		return
	}

	fmt.Printf("plink.exe的PID: %d\n", cmd.Process.Pid)

	// 等待命令执行完毕
	err = cmd.Wait()
	if err != nil {
		updateLogLabel(fmt.Sprintf("RUN:plink.exe. Command execution error: %v\n", err))
	} else {
		updateLogLabel("plink.exe has exited.")
	}
}

func CheckPort(port string) bool {
	l, err := net.Listen("tcp", net.JoinHostPort("127.0.0.1", port))
	if err != nil {
		return false
	}
	defer l.Close()
	return true
}
