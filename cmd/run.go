// Copyright © 2018 Takeshi Ihara <afrontier829@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/user"
	"regexp"
	"strings"

	"github.com/joho/godotenv"
	"github.com/nlopes/slack"
	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run kikanbo bot on your slack",
	Long:  `Run kikanbo bot on your slack.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file")
		}

		token := os.Getenv("KIKANBO_TOKEN")
		if token == "" {
			log.Fatal("Not Found: ENV['KIKANBO_TOKEN]")
		}
		memberID := os.Getenv("BOT_MEMBER_ID")
		if memberID == "" {
			log.Fatal("Not Found: ENV['BOT_MEMBER_ID']")
		}
		devicesFileDir := os.Getenv("DEVICES_FILE_DIR")
		if devicesFileDir == "" {
			usr, _ := user.Current()
			devicesFileDir = usr.HomeDir + "/.kikanbo"
		}
		devicesFilePath := os.Getenv("DEVICES_FILE_PATH")
		if devicesFilePath == "" {
			devicesFilePath = devicesFileDir + "/devices.json"
		}

		api := slack.New(token)
		os.Exit(run(api, memberID, devicesFileDir, devicesFilePath))
	},
}

func run(api *slack.Client, memberID string, fileDir string, filePath string) int {
	rtm := api.NewRTM()
	go rtm.ManageConnection()

	for {
		fmt.Println(rtm.IncomingEvents)

		select {
		case msg := <-rtm.IncomingEvents:
			switch ev := msg.Data.(type) {
			case *slack.MessageEvent:
				fmt.Printf("Message: %v\n", ev)
				if strings.Contains(ev.Text, "<@"+memberID+">") {
					preProcessing(fileDir, filePath)

					r := regexp.MustCompile(`<@(..*)>\s(..*)`)
					keywords := r.FindAllStringSubmatch(ev.Text, -1)
					keyword := ""
					if len(keywords) >= 1 && len(keywords[0]) >= 2 {
						keyword = keywords[0][2]
					}

					connectedDevices := connectedDevices()
					disconnectedDevices := disconnectedDevices(filePath, connectedDevices)

					params := slack.PostMessageParameters{
						Username:  "kikanbo",
						IconEmoji: ":ramen:",
					}
					if len(attachmentFields(connectedDevices, keyword)) != 0 {
						connected := slack.Attachment{
							AuthorName:    "Connected",
							AuthorSubname: ":bulb:",
							Color:         "#7CD197",
							Fields:        attachmentFields(connectedDevices, keyword),
						}
						params.Attachments = append(params.Attachments, connected)
					}
					if len(attachmentFields(disconnectedDevices, keyword)) != 0 {
						disconnected := slack.Attachment{
							AuthorName:    "Disconnected",
							AuthorSubname: ":electric_plug:",
							Color:         "#F35A00",
							Fields:        attachmentFields(disconnectedDevices, keyword),
						}
						params.Attachments = append(params.Attachments, disconnected)
					}
					_, _, err := api.PostMessage(ev.Channel, "", params)
					if err != nil {
						fmt.Printf("%s\n", err)
					}
					postProcessing(filePath, append(connectedDevices, disconnectedDevices...))
				}
			case *slack.PresenceChangeEvent:
				fmt.Printf("Presence Change: %v\n", ev)
			case *slack.RTMError:
				fmt.Printf("Error: %s\n", ev.Error())
			case *slack.InvalidAuthEvent:
				fmt.Printf("Invalid credentials")
				return 1
			}
		}
	}
}

// Device is a representation of a device
type Device struct {
	Name         string `json:"name"`
	OS           string `json:"os"`
	SerialNumber string `json:"serial_number"`
}

func preProcessing(fileDir string, filePath string) {
	{
		_, err := os.Stat(fileDir)
		if err != nil {
			if err := os.MkdirAll(fileDir, 0777); err != nil {
				fmt.Println(err)
			}
		}
	}

	{
		_, err := os.Stat(filePath)
		if err != nil {
			ioutil.WriteFile(filePath, []byte("[]"), os.ModePerm)
		}
	}
}

func connectedDevices() []Device {
	out, err := exec.Command("instruments", "-s", "devices").Output()
	if err != nil {
		fmt.Println("Command Exec Error:", err)
	}

	r := regexp.MustCompile(`(..*)\s\((..*)\)\s\[(\w+)\]`)
	devices := make([]Device, 0)

	for _, x := range strings.Split(string(out), "\n") {
		if strings.LastIndex(x, "(Simulator)") == -1 && r.MatchString(x) {
			device := r.FindAllStringSubmatch(x, -1)[0]
			devices = append(
				devices,
				Device{
					Name:         device[1],
					OS:           device[2],
					SerialNumber: device[3],
				},
			)
		}
	}

	return devices
}

func disconnectedDevices(filePath string, connectedDevices []Device) []Device {
	raw, err := ioutil.ReadFile(filePath)
	devices := make([]Device, 0)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	json.Unmarshal(raw, &devices)

	for _, x := range connectedDevices {
		xs := make([]Device, 0)
		for _, e := range devices {
			if x.SerialNumber != e.SerialNumber {
				xs = append(
					xs,
					e,
				)
			}
		}
		devices = xs
	}

	return devices
}

func attachmentFields(devices []Device, keyword string) []slack.AttachmentField {
	fields := make([]slack.AttachmentField, 0)
	for _, x := range devices {
		if keyword == "" || strings.Contains(x.Name, keyword) || strings.Contains("iOS "+x.OS, keyword) {
			fields = append(
				fields,
				slack.AttachmentField{
					Title: x.Name,
					Value: "iOS " + x.OS,
					Short: true,
				},
			)
		}
	}
	return fields
}

func postProcessing(filePath string, devices []Device) {
	outputJSON, err := json.Marshal(devices)
	if err != nil {
		panic(err)
	}
	ioutil.WriteFile(filePath, outputJSON, os.ModePerm)
}

func init() {
	rootCmd.AddCommand(runCmd)
}
