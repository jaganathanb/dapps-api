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
	"sync"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/jaganathanb/dapps-api/common"
	"github.com/jaganathanb/dapps-api/config"
	"github.com/jaganathanb/dapps-api/data/models"
	"github.com/jaganathanb/dapps-api/pkg/logging"
	"github.com/jaganathanb/dapps-api/pkg/s2t"
)

type GstDetail struct {
	Gst          models.Gst
	Returns      []models.GstStatus
	ErrorMessage string
}

type DashboardDetail struct {
	Landed       bool
	ErrorMessage string
}

type GstScrapper struct {
	logger        logging.Logger
	cfg           *config.Config
	speechService *s2t.DAppsSpeechToText
}

var gstScrapper *GstScrapper
var gstScrapperOnce sync.Once

func NewGstScrapper(cfg *config.Config) *GstScrapper {
	gstScrapperOnce.Do(func() {
		gstScrapper = &GstScrapper{
			logger:        logging.NewLogger(cfg),
			cfg:           cfg,
			speechService: s2t.NewDAppsSpeechToText(cfg),
		}
	})

	return gstScrapper
}

func (s *GstScrapper) ScrapGstPortal(gstins []string) (*common.SafeChannel[GstDetail], error) {
	quit := common.NewSafeChannel[GstDetail]()

	l := launcher.New().Headless(true).Devtools(false).Leakless(false)
	browser := rod.New().ControlURL(l.MustLaunch()).MustConnect().SlowMotion(time.Second * 1)
	page, err := browser.Page(proto.TargetCreateTarget{URL: s.cfg.Server.Gst.BaseUrl})

	if err != nil {
		s.logger.Errorf("Something went wrong!. Check internet connection too! - %s", err.Error())

		quit.SafeClose()

		return nil, err
	}

	defer common.RecoverFromPanic(func(err any) {
		s.logger.Errorf("Recovered from panic: ", err)

		quit.SafeClose()

		page.MustClose()
		browser.MustClose()
		l.Cleanup()
	})

	page.MustWindowMaximize()

	landed := common.NewSafeChannel[DashboardDetail]()
	done := common.NewSafeChannel[GstDetail]()

	go s.listenOnCaptchaEvents(page, landed)()
	go s.inputCaptch(page, landed)

	go s.listenOnGstReturnsEvents(page, done)()
	go s.searchAllGsts(landed, gstins, page, done, browser, l, quit)

	return quit, nil
}

func (s *GstScrapper) listenOnGstReturnsEvents(page *rod.Page, done *common.SafeChannel[GstDetail]) func() {
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

			if !done.IsClosed() {
				done.C <- gstDetail
			}
		}
	}, func(e *proto.NetworkLoadingFailed) {
		if e.Type == proto.NetworkResourceTypeFetch || e.Type == proto.NetworkResourceTypeXHR || e.Type == proto.NetworkResourceTypeMedia || e.Type == proto.NetworkResourceTypeDocument {

			gstDetail.ErrorMessage = e.ErrorText

			if !done.IsClosed() {
				done.C <- gstDetail
			}
		}
	})
}

func (s *GstScrapper) processCaptcha(requestId proto.NetworkRequestID, page *rod.Page, landed *common.SafeChannel[DashboardDetail]) {
	defer common.RecoverFromPanic(func(err any) {
		s.logger.Errorf("Error recovered!. %s", err)

		if !landed.IsClosed() {
			landed.C <- DashboardDetail{}
		}
	})

	m := proto.NetworkGetResponseBody{RequestID: requestId}
	r, err := m.Call(page)

	if err == nil {
		err = createAudioFile(r.Body)

		if err != nil {
			s.logger.Errorf("Error while processing the Speech to text. Error - %s", err.Error())

			if !landed.IsClosed() {
				landed.C <- DashboardDetail{}
			}
		} else {
			code, err := s.speechService.SpeechToText("test.mp3")
			if err == nil {
				s.landIntoDashboard(page, code, landed)
			} else {
				s.logger.Errorf("Error while processing the Speech to text. The code is - %s", code)

				if !landed.IsClosed() {
					landed.C <- DashboardDetail{}
				}
			}
		}
	} else {
		s.logger.Errorf("Error in TCP, Could not get the response for Captch audio. - %s", err.Error())

		if !landed.IsClosed() {
			landed.C <- DashboardDetail{}
		}
	}
}

func (s *GstScrapper) landIntoDashboard(page *rod.Page, code string, landed *common.SafeChannel[DashboardDetail]) {
	defer common.RecoverFromPanic(func(err any) {
		s.logger.Errorf("Error recovered!. %s", err)

		if !landed.IsClosed() {
			landed.C <- DashboardDetail{}
		}
	})

	var numericRegex = regexp.MustCompile(`[^\p{N} ]+`)
	captcha := page.MustElement("#captcha")

	captcha.MustSelectAllText().MustType(input.Backspace).MustInput(numericRegex.ReplaceAllString(code, ""))

	page.MustElement("[type=submit]").MustClick()

	err := rod.Try(func() {
		page.Timeout(time.Duration(15 * time.Second)).MustElement(".dp-widgt")
	})

	if errors.Is(err, context.DeadlineExceeded) {
		s.logger.Errorf("Timeout while looking for dashboard widget - Reason: %s", err.Error())

		err := rod.Try(func() {
			page.Timeout(time.Duration(15 * time.Second)).MustElement("#submitpwd")
		})

		if err == nil {
			if !landed.IsClosed() {
				landed.C <- DashboardDetail{
					ErrorMessage: fmt.Sprintf("NOTIFICATION|GST credential needs to be changed for the GST user id %s", s.cfg.Server.Gst.Username),
				}

				return
			}
		}

		var errMsg *rod.Element
		err = rod.Try(func() {
			errMsg = page.Timeout(time.Duration(15 * time.Second)).MustElement("span.err")
		})

		if err == nil {
			msg, _ := errMsg.Text()

			if strings.Contains(msg, "Enter valid Letters shown") {
				go s.inputCaptch(page, landed)

				return
			}
		}

		err = rod.Try(func() {
			errMsg = page.Timeout(time.Duration(15 * time.Second)).MustElement(".err")
		})

		if err == nil {
			msg, _ := errMsg.Text()

			if strings.Contains(msg, "Invalid Username or Password. Please try again.") {
				if !landed.IsClosed() {
					landed.C <- DashboardDetail{
						ErrorMessage: "NOTIFICATION|GST username or password is invalid. Please update GST credential and try again.",
					}
				}
			} else {
				if !landed.IsClosed() {
					landed.C <- DashboardDetail{}
				}
			}
		} else {
			s.logger.Errorf("Something went wrong while looking for dashboard widget - Reason: %s", err.Error())

			if !landed.IsClosed() {
				landed.C <- DashboardDetail{}
			}
		}
	} else if err != nil {
		s.logger.Errorf("Something went wrong while looking for dashboard widget - Reason: %s", err.Error())

		if !landed.IsClosed() {
			landed.C <- DashboardDetail{}
		}
	} else {
		s.logger.Infof("Successfully landed into dashboard widget")

		if !landed.IsClosed() {
			landed.C <- DashboardDetail{
				Landed: true,
			}
		}
	}
}

func (s *GstScrapper) listenOnCaptchaEvents(page *rod.Page, landed *common.SafeChannel[DashboardDetail]) func() {
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
				s.logger.Errorf("Maximum retry reached for getting Captcha code")

				if !landed.IsClosed() {
					landed.C <- DashboardDetail{}
				}
			}

			captchaRetry += 1
			s.processCaptcha(e.RequestID, page, landed)
		}
	}, func(e *proto.NetworkLoadingFailed) {
		s.logger.Errorf("NetworkLoadingFailed: %s", e.ErrorText)

		if e.Type == proto.NetworkResourceTypeFetch || e.Type == proto.NetworkResourceTypeXHR || e.Type == proto.NetworkResourceTypeMedia || e.Type == proto.NetworkResourceTypeDocument {
			if !landed.IsClosed() {
				landed.C <- DashboardDetail{
					ErrorMessage: e.ErrorText,
				}
			}
		}
	})
}

func (s *GstScrapper) inputCaptch(page *rod.Page, landed *common.SafeChannel[DashboardDetail]) {
	defer common.RecoverFromPanic(func(err any) {
		s.logger.Errorf("Error recovered!. %s", err)

		if !landed.IsClosed() {
			landed.C <- DashboardDetail{}
		}
	})

	username := page.MustElement("#username")
	password := page.MustElement("#user_pass")

	username.MustSelectAllText().MustType(input.Backspace).MustInput(s.cfg.Server.Gst.Username)

	page.MustElement("#imgCaptcha").MustWaitVisible()
	password.MustSelectAllText().MustType(input.Backspace).MustInput(s.cfg.Server.Gst.Password)

	page.MustElement("i.fa.fa-volume-up").MustParent().MustClick()

	page.MustWaitRequestIdle()()
}

func (s *GstScrapper) searchAllGsts(landed *common.SafeChannel[DashboardDetail], gstins []string, page *rod.Page, done *common.SafeChannel[GstDetail], browser *rod.Browser, l *launcher.Launcher, quit *common.SafeChannel[GstDetail]) {
	defer common.RecoverFromPanic(func(err any) {
		s.logger.Errorf("Error recovered!. %s", err)

		quit.SafeClose()
	})

	dashboardDetail := DashboardDetail{}

	if !landed.IsClosed() {
		dashboardDetail = <-landed.C
	}

	if dashboardDetail.Landed {
		landed.SafeClose()

		for _, gstin := range gstins {
			s.searchGst(page, gstin, done)

			if !done.IsClosed() {
				data := <-done.C

				if !quit.IsClosed() {
					quit.C <- data
				}
			}
		}
	} else {
		quit.C <- GstDetail{
			ErrorMessage: dashboardDetail.ErrorMessage,
		}
	}

	done.SafeClose()
	landed.SafeClose()

	page.MustClose()
	browser.MustClose()
	l.Cleanup()

	quit.SafeClose()
}

func (s *GstScrapper) searchGst(page *rod.Page, gstin string, done *common.SafeChannel[GstDetail]) {
	defer common.RecoverFromPanic(func(err any) {
		s.logger.Errorf("Error recovered!. %s", err)

		if !done.IsClosed() {
			done.C <- GstDetail{
				ErrorMessage: err.(string),
			}
		}
	})

	page.MustElementX("//*[@id=\"main\"]/ul/li[5]/a").MustClick()

	page.MustElementX("//*[@id=\"main\"]/ul/li[5]/ul[1]/li[1]/a").MustClick()

	page.MustElement("#for_gstin").MustInput(gstin).MustType(input.Enter)

	page.MustElement("#filingTable").MustClick()
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

func createAudioFile(fileContent string) error {
	dec, err := base64.StdEncoding.DecodeString(fileContent)
	if err != nil {
		return err
	}

	f, err := os.Create("test.mp3")
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.Write(dec); err != nil {
		return err
	}
	if err := f.Sync(); err != nil {
		return err
	}

	return nil
}
