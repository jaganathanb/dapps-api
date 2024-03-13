package gst_scrapper

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/jaganathanb/dapps-api/config"
	"github.com/jaganathanb/dapps-api/pkg/logging"
	"github.com/jaganathanb/dapps-api/pkg/s2t"
)

func ScrapGstPortal(gstin string, logger logging.Logger, cfg *config.Config) (<-chan bool, <-chan interface{}, <-chan interface{}) {
	quit := make(chan bool)
	gstCh := make(chan interface{})
	returnsCh := make(chan interface{})

	defer func() {
		if err := recover(); err != nil {
			logger.Error(logging.Category(logging.ExternalService), logging.SubCategory(logging.IO), "Panic occured!", err.(map[logging.ExtraKey]interface{}))
			quit <- true
		}
	}()

	l := launcher.New().Headless(false).Devtools(true).Leakless(false)

	browser := rod.New().ControlURL(l.MustLaunch()).MustConnect()

	page := browser.MustPage(cfg.Server.Gst.BaseUrl).MustWindowMaximize()

	var captchaRequestId proto.NetworkRequestID
	var gstRequestId proto.NetworkRequestID
	var returnsRequestId proto.NetworkRequestID

	go page.EachEvent(func(e *proto.NetworkRequestWillBeSent) {
		fmt.Printf("Request: %s %s\n", e.Request.Method, e.Request.URL)

		switch true {
		case strings.Contains(e.Request.URL, "/audiocaptcha") && e.Request.Method != "OPTIONS":
			captchaRequestId = e.RequestID
			break
		case strings.Contains(e.Request.URL, "/api/search/tp") && e.Request.Method != "OPTIONS":
			gstRequestId = e.RequestID
			break
		case strings.Contains(e.Request.URL, "/api/search/taxpayerReturnDetails") && e.Request.Method != "OPTIONS":
			returnsRequestId = e.RequestID
			break
		}

	}, func(e *proto.NetworkLoadingFinished) {
		if e.RequestID == captchaRequestId {
			processCaptcha(e.RequestID, page, logger, quit)
		}

		if e.RequestID == gstRequestId {
			processGst(e.RequestID, page, gstCh)
		}

		if e.RequestID == returnsRequestId {
			processReturns(e, page, gstCh, browser, l, quit)
		}
	})()

	username := page.MustElement("#username")
	password := page.MustElement("#user_pass")

	username.MustInput(cfg.Server.Gst.Username)
	time.Sleep(3 * time.Second)

	page.MustElement("#imgCaptcha").MustWaitVisible()
	password.MustInput(cfg.Server.Gst.Password)
	time.Sleep(5 * time.Second)

	page.MustElement("i.fa.fa-volume-up").MustParent().MustClick().MustWaitStable()

	page.MustWaitRequestIdle()()

	return quit, gstCh, returnsCh
}

func processReturns(e *proto.NetworkLoadingFinished, page *rod.Page, gstCh chan interface{}, browser *rod.Browser, l *launcher.Launcher, quit chan bool) {
	m := proto.NetworkGetResponseBody{RequestID: e.RequestID}
	r, err := m.Call(page)

	var val any
	if err == nil {
		json.Unmarshal([]byte(r.Body), &val)
		gstCh <- val
	}

	page.MustClose()
	browser.MustClose()
	l.Cleanup()

	quit <- true
}

func processGst(requestId proto.NetworkRequestID, page *rod.Page, gstCh chan interface{}) {
	m := proto.NetworkGetResponseBody{RequestID: requestId}
	r, err := m.Call(page)

	var val any
	if err == nil {
		json.Unmarshal([]byte(r.Body), &val)
		gstCh <- val
	}
}

func processCaptcha(requestId proto.NetworkRequestID, page *rod.Page, logger logging.Logger, quit chan bool) {
	m := proto.NetworkGetResponseBody{RequestID: requestId}
	r, err := m.Call(page)

	if err == nil {
		dec, err := base64.StdEncoding.DecodeString(r.Body)
		if err != nil {
			panic(err)
		}

		f, err := os.Create("test.mp3")
		if err != nil {
			panic(err)
		}
		defer f.Close()

		if _, err := f.Write(dec); err != nil {
			panic(err)
		}
		if err := f.Sync(); err != nil {
			panic(err)
		}

		code, err := s2t.SpeechToText("test.mp3")
		if err == nil {
			var numericRegex = regexp.MustCompile(`[^\p{N} ]+`)
			captcha := page.MustElement("#captcha")

			captcha.MustInput(numericRegex.ReplaceAllString(code, ""))

			page.MustElement("[type=submit]").MustClick()

			dialog, cancel := page.WithCancel()

			go func() {
				time.Sleep(time.Duration(10 * time.Second))
				cancel()
			}()

			dialog.MustElementR("a", "/Remind me later/i").MustClick()

			page.MustElementX("//*[@id=\"main\"]/ul/li[5]/a").MustClick()

			page.MustElementX("//*[@id=\"main\"]/ul/li[5]/ul[1]/li[1]/a").MustClick()

			page.MustElement("#for_gstin").MustInput("33BVFPV4346B1ZP").MustType(input.Enter)

			page.MustElementX("//*[@id=\"filingTable\"]").MustClick().MustScreenshot()

			logger.Info(logging.General, logging.SubCategory(logging.IO), "All done!", nil)
		}
	} else {
		logger.Error(logging.Category(logging.ExternalService), logging.SubCategory(logging.IO), "Error in TCP!", nil)
		quit <- true
	}
}
