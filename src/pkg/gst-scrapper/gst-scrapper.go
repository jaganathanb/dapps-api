package gst_scrapper

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"regexp"
	"strings"
	"sync"
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

type GstScrapper struct {
	logger logging.Logger
	cfg    *config.Config
}

var gstScrapper *GstScrapper
var gstScrapperOnce sync.Once

func NewGstScrapper(cfg *config.Config) *GstScrapper {
	gstScrapperOnce.Do(func() {
		gstScrapper = &GstScrapper{
			logger: logging.NewLogger(cfg),
			cfg:    cfg,
		}
	})

	return gstScrapper
}

func (s *GstScrapper) ScrapGstPortal(gstins []string) <-chan GstDetail {
	quit := make(chan GstDetail)

	defer func() {
		if err := recover(); err != nil {
			s.logger.Error(logging.Category(logging.ExternalService), logging.SubCategory(logging.IO), "Panic occured!", err.(map[logging.ExtraKey]interface{}))
			quit <- GstDetail{}
		}
	}()

	l := launcher.New().Headless(true).Devtools(false).Leakless(false)
	browser := rod.New().ControlURL(l.MustLaunch()).MustConnect().SlowMotion(time.Second * 1)
	page := browser.MustPage(s.cfg.Server.Gst.BaseUrl).MustWindowMaximize()

	landed := make(chan bool)
	go s.listenOnCaptchaEvents(page, quit, landed)()
	go s.inputCaptch(page)

	done := make(chan GstDetail)
	go s.listenOnGstReturnsEvents(page, done)()
	go searchAllGsts(landed, gstins, page, done, browser, l, quit)

	return quit
}

func (s *GstScrapper) listenOnGstReturnsEvents(page *rod.Page, done chan GstDetail) func() {
	var gstRequestId proto.NetworkRequestID
	var returnsRequestId proto.NetworkRequestID
	var gstDetail GstDetail

	return page.EachEvent(func(e *proto.NetworkRequestWillBeSent) {
		s.logger.Infof("Request: %s %s\n", e.Request.Method, e.Request.URL)

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

func (s *GstScrapper) processCaptcha(requestId proto.NetworkRequestID, page *rod.Page, quit chan GstDetail, landed chan bool) {
	m := proto.NetworkGetResponseBody{RequestID: requestId}
	r, err := m.Call(page)

	if err == nil {
		createAudioFile(r.Body)

		code, err := s2t.SpeechToText("test.mp3")
		if err == nil {
			s.landIntoDashboard(page, code, landed)
		} else {
			quit <- GstDetail{}
		}

		s.logger.Info(logging.General, logging.SubCategory(logging.IO), "Landed into Dashboard!", nil)
	} else {
		s.logger.Error(logging.Category(logging.ExternalService), logging.SubCategory(logging.IO), "Error in TCP!", nil)
		quit <- GstDetail{}
	}
}

func (s *GstScrapper) landIntoDashboard(page *rod.Page, code string, landed chan bool) {
	var numericRegex = regexp.MustCompile(`[^\p{N} ]+`)
	captcha := page.MustElement("#captcha")

	captcha.MustSelectAllText().MustType(input.Backspace).MustInput(numericRegex.ReplaceAllString(code, ""))

	page.MustElement("[type=submit]").MustClick()

	err := rod.Try(func() {
		page.Timeout(time.Duration(10 * time.Second)).MustElement(".dp-widgt")
	})

	if errors.Is(err, context.DeadlineExceeded) {
		s.logger.Error(logging.Internal, logging.Api, "Timeout while looking for dashboard widget", nil)
		if captcha.MustParent().MustParent().MustHas(".err") {
			go s.inputCaptch(page)
		}
	} else if err != nil {
		s.logger.Error(logging.Internal, logging.Api, "Something went wrong while looking for dashboard widget", nil)
	} else {
		s.logger.Info(logging.Internal, logging.Api, "Successfully landed into dashboard widget", nil)
		landed <- true
	}
}

func (s *GstScrapper) listenOnCaptchaEvents(page *rod.Page, quit chan GstDetail, landed chan bool) func() {
	var captchaRequestId proto.NetworkRequestID
	captchaRetry := 0

	return page.EachEvent(func(e *proto.NetworkRequestWillBeSent) {
		s.logger.Infof("Request: %s %s\n", e.Request.Method, e.Request.URL)

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
			s.processCaptcha(e.RequestID, page, quit, landed)
		}
	})
}

func (s *GstScrapper) inputCaptch(page *rod.Page) {
	username := page.MustElement("#username")
	password := page.MustElement("#user_pass")

	username.MustSelectAllText().MustType(input.Backspace).MustInput(s.cfg.Server.Gst.Username)

	page.MustElement("#imgCaptcha").MustWaitVisible()
	password.MustSelectAllText().MustType(input.Backspace).MustInput(s.cfg.Server.Gst.Password)

	page.MustElement("i.fa.fa-volume-up").MustParent().MustClick()

	page.MustWaitRequestIdle()()
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
