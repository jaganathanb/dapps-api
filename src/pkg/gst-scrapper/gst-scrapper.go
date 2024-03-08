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
	"github.com/jaganathanb/dapps-api/pkg/logging"
	"github.com/jaganathanb/dapps-api/pkg/s2t"
)

const (
	siteUrl = "https://services.gst.gov.in/services/login"
)

func ScrapGstPortal(logger logging.Logger) (<-chan bool, <-chan interface{}, <-chan interface{}) {
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
	//defer l.Cleanup()

	browser := rod.New().ControlURL(l.MustLaunch()).MustConnect()
	//defer browser.MustClose()

	page := browser.MustPage(siteUrl).MustWindowMaximize()
	//defer page.MustClose()

	var captchaRequestId proto.NetworkRequestID
	var gstRequestId proto.NetworkRequestID
	var returnsRequestId proto.NetworkRequestID
	go page.EachEvent(func(e *proto.NetworkRequestWillBeSent) {
		fmt.Printf("Request: %s %s\n", e.Request.Method, e.Request.URL)
		if strings.Contains(e.Request.URL, "/audiocaptcha") && e.Request.Method != "OPTIONS" {
			captchaRequestId = e.RequestID
		}

		if strings.Contains(e.Request.URL, "/api/search/tp") && e.Request.Method != "OPTIONS" {
			gstRequestId = e.RequestID
		}

		if strings.Contains(e.Request.URL, "/api/search/taxpayerReturnDetails") && e.Request.Method != "OPTIONS" {
			returnsRequestId = e.RequestID
		}
	}, func(e *proto.NetworkLoadingFinished) {
		if e.RequestID == captchaRequestId {
			m := proto.NetworkGetResponseBody{RequestID: e.RequestID}
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
						time.Sleep(time.Duration(10 * time.Second)) // cancel after 10 secs
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

		if e.RequestID == gstRequestId {
			m := proto.NetworkGetResponseBody{RequestID: e.RequestID}
			r, err := m.Call(page)

			var val any
			if err == nil {
				json.Unmarshal([]byte(r.Body), &val)
				gstCh <- val
			}
		}

		if e.RequestID == returnsRequestId {
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
	})()

	username := page.MustElement("#username")
	password := page.MustElement("#user_pass")
	username.MustInput("Cryptic333")
	time.Sleep(5 * time.Second)
	password.MustInput("SAabc*963")

	page.MustElement("i.fa.fa-volume-up").MustParent().MustClick()

	page.MustWaitRequestIdle()()

	return quit, gstCh, returnsCh
}
