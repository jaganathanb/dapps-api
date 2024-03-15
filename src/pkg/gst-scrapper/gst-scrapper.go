package gst_scrapper

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
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
	"github.com/jaganathanb/dapps-api/data/models"
	"github.com/jaganathanb/dapps-api/pkg/logging"
	"github.com/jaganathanb/dapps-api/pkg/s2t"
)

type GstDetail struct {
	Gst     models.Gst
	Returns []models.GstStatus
}

func ScrapGstPortal(gstins []string, logger logging.Logger, cfg *config.Config) <-chan GstDetail {
	quit := make(chan GstDetail)

	defer func() {
		if err := recover(); err != nil {
			logger.Error(logging.Category(logging.ExternalService), logging.SubCategory(logging.IO), "Panic occured!", err.(map[logging.ExtraKey]interface{}))
			quit <- GstDetail{}
		}
	}()

	l := launcher.New().Headless(true).Devtools(false).Leakless(false)
	browser := rod.New().ControlURL(l.MustLaunch()).MustConnect().SlowMotion(time.Second * 1)
	page := browser.MustPage(cfg.Server.Gst.BaseUrl).MustWindowMaximize()

	landed := make(chan bool)
	go listenOnCaptchaEvents(page, logger, cfg, quit, landed)()
	go inputCaptch(page, cfg)

	done := make(chan GstDetail)
	go listenOnGstReturnsEvents(page, done)()
	go searchAllGsts(landed, gstins, page, done, browser, l, quit)

	return quit
}

func listenOnGstReturnsEvents(page *rod.Page, done chan GstDetail) func() {
	var gstRequestId proto.NetworkRequestID
	var returnsRequestId proto.NetworkRequestID
	var gstDetail GstDetail

	return page.EachEvent(func(e *proto.NetworkRequestWillBeSent) {
		fmt.Printf("Request: %s %s\n", e.Request.Method, e.Request.URL)

		switch true {
		case strings.Contains(e.Request.URL, "/api/search/tp") && e.Request.Method != "OPTIONS":
			gstRequestId = e.RequestID
			break
		case strings.Contains(e.Request.URL, "/api/search/taxpayerReturnDetails") && e.Request.Method != "OPTIONS":
			returnsRequestId = e.RequestID
			break
		}

	}, func(e *proto.NetworkLoadingFinished) {
		if e.RequestID == gstRequestId {
			gstDetail.Gst = *processGst(e.RequestID, page)
		}

		if e.RequestID == returnsRequestId {
			gstDetail.Returns = processReturns(e.RequestID, page)

			done <- gstDetail
		}
	})

}

func listenOnCaptchaEvents(page *rod.Page, logger logging.Logger, cfg *config.Config, quit chan GstDetail, landed chan bool) func() {
	var captchaRequestId proto.NetworkRequestID
	captchaRetry := 0

	return page.EachEvent(func(e *proto.NetworkRequestWillBeSent) {
		fmt.Printf("Request: %s %s\n", e.Request.Method, e.Request.URL)

		switch true {
		case strings.Contains(e.Request.URL, "/audiocaptcha") && e.Request.Method != "OPTIONS":
			captchaRequestId = e.RequestID
			break
		}

	}, func(e *proto.NetworkLoadingFinished) {
		if e.RequestID == captchaRequestId {
			if captchaRetry == 3 {
				quit <- GstDetail{}
			}

			captchaRetry += 1
			processCaptcha(e.RequestID, page, cfg, logger, quit, landed)
		}
	})
}

func searchAllGsts(landed chan bool, gstins []string, page *rod.Page, done chan GstDetail, browser *rod.Browser, l *launcher.Launcher, quit chan GstDetail) {
	<-landed

	for _, gstin := range gstins {
		searchGst(page, gstin)

		data := <-done
		quit <- data
	}
	close(done)

	page.MustClose()
	browser.MustClose()
	l.Cleanup()

	close(quit)
}

func inputCaptch(page *rod.Page, cfg *config.Config) {
	username := page.MustElement("#username")
	password := page.MustElement("#user_pass")

	username.MustSelectAllText().MustType(input.Backspace).MustInput(cfg.Server.Gst.Username)

	page.MustElement("#imgCaptcha").MustWaitVisible()
	password.MustSelectAllText().MustType(input.Backspace).MustInput(cfg.Server.Gst.Password)

	page.MustElement("i.fa.fa-volume-up").MustParent().MustClick()

	page.MustWaitRequestIdle()()
}

func processReturns(requestId proto.NetworkRequestID, page *rod.Page) []models.GstStatus {
	m := proto.NetworkGetResponseBody{RequestID: requestId}
	r, err := m.Call(page)

	var val map[string][][]models.GstStatus
	if err == nil {
		json.Unmarshal([]byte(r.Body), &val)
		return val["filingStatus"][0]
	}

	return nil
}

func processGst(requestId proto.NetworkRequestID, page *rod.Page) *models.Gst {
	m := proto.NetworkGetResponseBody{RequestID: requestId}
	r, err := m.Call(page)

	var val models.Gst
	if err == nil {
		json.Unmarshal([]byte(r.Body), &val)
		return &val
	}

	return nil
}

func processCaptcha(requestId proto.NetworkRequestID, page *rod.Page, cfg *config.Config, logger logging.Logger, quit chan GstDetail, landed chan bool) {
	m := proto.NetworkGetResponseBody{RequestID: requestId}
	r, err := m.Call(page)

	if err == nil {
		createAudioFile(r.Body)

		code, err := s2t.SpeechToText("test.mp3")
		if err == nil {
			landIntoDashboard(page, code, cfg, landed)
		} else {
			quit <- GstDetail{}
		}

		logger.Info(logging.General, logging.SubCategory(logging.IO), "Landed into Dashboard!", nil)
	} else {
		logger.Error(logging.Category(logging.ExternalService), logging.SubCategory(logging.IO), "Error in TCP!", nil)
		quit <- GstDetail{}
	}
}

func landIntoDashboard(page *rod.Page, code string, cfg *config.Config, landed chan bool) {
	var numericRegex = regexp.MustCompile(`[^\p{N} ]+`)
	captcha := page.MustElement("#captcha")

	captcha.MustSelectAllText().MustType(input.Backspace).MustInput(numericRegex.ReplaceAllString(code, ""))

	page.MustElement("[type=submit]").MustClick()

	err := rod.Try(func() {
		page.Timeout(time.Duration(10 * time.Second)).MustElement(".dp-widgt")
	})

	if errors.Is(err, context.DeadlineExceeded) {
		fmt.Println("timeout error")
		if captcha.MustParent().MustParent().MustHas(".err") {
			go inputCaptch(page, cfg)
		}
	} else if err != nil {
		fmt.Println("other types of error")
	} else {
		landed <- true
	}
}

func createAudioFile(fileContent string) {
	dec, err := base64.StdEncoding.DecodeString(fileContent)
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
}

func searchGst(page *rod.Page, gstin string) {
	page.MustElementX("//*[@id=\"main\"]/ul/li[5]/a").MustClick()

	page.MustElementX("//*[@id=\"main\"]/ul/li[5]/ul[1]/li[1]/a").MustClick()

	page.MustElement("#for_gstin").MustInput(gstin).MustType(input.Enter)

	page.MustElement("#filingTable").MustClick()
}
