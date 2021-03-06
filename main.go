package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"

	"github.com/Grarak/GoYTFetcher/api"
	"github.com/Grarak/GoYTFetcher/database"
	"github.com/Grarak/GoYTFetcher/logger"
	"github.com/Grarak/GoYTFetcher/miniserver"
	"github.com/Grarak/GoYTFetcher/utils"
)

var indexDir string

func clientHandler(client *miniserver.Client) miniserver.Response {
	args := strings.Split(client.Url, "/")[1:]
	if len(args) >= 3 && args[0] == "api" {
		return api.GetResponse(args[1], args[2], args[3:], client)
	}

	if !utils.StringIsEmpty(indexDir) {
		if client.Url != "/" && utils.FileExists(indexDir+client.Url) {
			return client.ResponseFile(indexDir + client.Url)
		}
		if utils.FileExists(indexDir + "/index.html") {
			return client.ResponseFile(indexDir + "/index.html")
		}
	}

	response := client.ResponseBody("Not found")
	response.SetStatusCode(http.StatusNotFound)
	return response
}

func main() {
	logger.Init()

	if _, err := exec.LookPath(utils.YOUTUBE_DL); err != nil {
		logger.E(utils.YOUTUBE_DL + " is not installed!")
		return
	}

	ffmpeg, err := exec.LookPath(utils.FFMPEG)
	if err != nil {
		logger.E(utils.FFMPEG + " is not installed!")
		return
	}

	codecs, err := utils.ExecuteCmd(ffmpeg, "-codecs")
	if err != nil || !strings.Contains(codecs, "libvorbis") {
		logger.E(utils.FFMPEG + " vorbis is not enabled")
		return
	}

	var port int
	var ytKey string
	flag.IntVar(&port, "p", 6713, "Which port to use")
	flag.StringVar(&ytKey, "yt", "", "Youtube Api key")
	flag.StringVar(&indexDir, "i", "", "Directory with index.html")
	flag.Parse()

	utils.Panic(utils.MkDir(utils.DATABASE))
	utils.Panic(utils.MkDir(utils.YOUTUBE_DIR))

	databaseInstance := database.GetDatabase(utils.GenerateRandom(16), ytKey)

	server := miniserver.NewServer(port)

	c := make(chan os.Signal, 1)
	cleanup := make(chan bool)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			logger.I(fmt.Sprintf("Captured %s, killing...", sig))
			server.StopListening()

			if err := databaseInstance.Close(); err != nil {
				logger.E(fmt.Sprintf("Failed to close database %s", err))
			}

			cleanup <- true
		}
	}()

	logger.I("Starting server on port " + strconv.Itoa(port))
	go server.StartListening(clientHandler)

	<-cleanup
}
