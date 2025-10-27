package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func main() {
	application := app.New()
	window := application.NewWindow("Go Port Scanner")
	window.Resize(fyne.NewSize(520, 400))

	hostEntry := widget.NewEntry()
	hostEntry.SetPlaceHolder("例如: scanme.nmap.org")

	startPortEntry := widget.NewEntry()
	startPortEntry.SetPlaceHolder("起始端口 (例如 1)")

	endPortEntry := widget.NewEntry()
	endPortEntry.SetPlaceHolder("结束端口 (例如 1024)")

	results := widget.NewMultiLineEntry()
	results.Wrapping = fyne.TextWrapWord
	results.SetPlaceHolder("扫描结果将显示在这里……")
	results.Disable()

	statusLabel := widget.NewLabel("就绪")

	scanButton := widget.NewButton("开始扫描", nil)

	form := widget.NewForm(
		widget.NewFormItem("目标主机", hostEntry),
		widget.NewFormItem("起始端口", startPortEntry),
		widget.NewFormItem("结束端口", endPortEntry),
	)

	content := container.NewBorder(form, container.NewVBox(statusLabel, scanButton), nil, nil, results)
	window.SetContent(content)

	scanButton.OnTapped = func() {
		host := strings.TrimSpace(hostEntry.Text)
		if host == "" {
			dialog.ShowError(fmt.Errorf("目标主机不能为空"), window)
			return
		}

		startPort, err := strconv.Atoi(strings.TrimSpace(startPortEntry.Text))
		if err != nil || startPort < 1 || startPort > 65535 {
			dialog.ShowError(fmt.Errorf("请输入有效的起始端口 (1-65535)"), window)
			return
		}

		endPort, err := strconv.Atoi(strings.TrimSpace(endPortEntry.Text))
		if err != nil || endPort < startPort || endPort > 65535 {
			dialog.ShowError(fmt.Errorf("请输入有效的结束端口 (必须 >= 起始端口，最大 65535)"), window)
			return
		}

		results.SetText("")
		scanButton.Disable()
		statusLabel.SetText("扫描中……")

		go func(host string, start, end int) {
			defer func() {
				application.Driver().RunOnMain(func() {
					scanButton.Enable()
					statusLabel.SetText("扫描完成")
				})
			}()

			startTime := time.Now()
			openPorts := make([]int, 0)

			for port := start; port <= end; port++ {
				address := fmt.Sprintf("%s:%d", host, port)
				conn, err := net.DialTimeout("tcp", address, time.Second)
				if err == nil {
					conn.Close()
					openPorts = append(openPorts, port)
					application.Driver().RunOnMain(func() {
						results.SetText(results.Text + fmt.Sprintf("端口 %d: 开放\n", port))
					})
				} else {
					application.Driver().RunOnMain(func() {
						results.SetText(results.Text + fmt.Sprintf("端口 %d: 关闭\n", port))
					})
				}
			}

			duration := time.Since(startTime)
			summary := fmt.Sprintf("\n扫描完成！耗时: %s\n开放端口数量: %d", duration.Round(time.Millisecond), len(openPorts))
			application.Driver().RunOnMain(func() {
				results.SetText(results.Text + summary)
			})
		}(host, startPort, endPort)
	}

	window.ShowAndRun()
}
