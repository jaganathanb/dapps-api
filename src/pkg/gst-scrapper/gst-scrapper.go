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
	ShouldRetry  bool
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

func (s *GstScrapper) ScrapGstReturnsDetail(gsts []models.Gst, useCredentialFromSettings bool) (*common.SafeChannel[GstDetail], error) {
	quit := common.NewSafeChannel[GstDetail]()

	l := launcher.New().Headless(false).Devtools(false).Leakless(false)
	browser := rod.New().ControlURL(l.MustLaunch()).MustConnect().SlowMotion(time.Second * 1).Trace(s.cfg.Server.RunMode == "debug")

	go s.runGstProcesses(gsts, browser, l, useCredentialFromSettings, quit)

	defer common.RecoverFromPanic(func(err any) {
		s.logger.Errorf("Recovered from panic: ", err)

		quit.SafeClose()

		browser.MustClose()
		l.Cleanup()
	})

	return quit, nil
}

func (s *GstScrapper) runGstProcesses(gsts []models.Gst, browser *rod.Browser, launcher *launcher.Launcher, useCredentialFromSettings bool, quit *common.SafeChannel[GstDetail]) {
	pool := rod.NewBrowserPool(len(gsts))

	create := func() *rod.Browser {
		return browser.MustIncognito()
	}

	scrapJob := func(gst models.Gst) GstDetail {
		browser := pool.Get(create)
		defer pool.Put(browser)

		page := browser.MustPage(s.cfg.Server.Gst.BaseUrl).MustWaitLoad()

		page.MustWindowMaximize()

		dashboard := s.login(page, gst, useCredentialFromSettings)

		gstDetail := s.getGstReturnsDetail(page, dashboard)

		page.Close()

		return gstDetail
	}

	wg := sync.WaitGroup{}
	for _, gst := range gsts {
		wg.Add(1)
		go func(g models.Gst) {
			defer wg.Done()
			detail := scrapJob(g)

			quit.C <- detail
		}(gst)
	}
	wg.Wait()

	pool.Cleanup(func(p *rod.Browser) { p.MustClose() })

	launcher.Cleanup()

	quit.SafeClose()
}

func (s *GstScrapper) getGstReturnsDetail(page *rod.Page, dashboard DashboardDetail) GstDetail {
	if !dashboard.Landed {
		return GstDetail{ErrorMessage: dashboard.ErrorMessage}
	} else {
		gstData, _, err := s.extractResponseFromHttpRequest(page, "auth/profile/detail", ".dp-widgt > a.tp-pfl-lnk", "/View Profile /")

		if err == nil {
			var gst models.Gst
			json.Unmarshal([]byte(gstData.Body), &gst)

			page.MustElementR(".nav > .menuList > a.dropdown-toggle", "/Services/").MustClick() // Click Services
			page.MustElementR(".smenu > .has-sub > a", "/Returns/").MustHover()                 // Hover Returns

			returns, _, err := s.extractResponseFromHttpRequest(page, "/returns/auth/api/returnstatus", "ul.isubmenu.ret.post > li > a", "/Track Return Status/")

			if err == nil {
				var statuses []models.GstStatus
				json.Unmarshal([]byte(returns.Body), &statuses)

				s.logger.Infof("The returns & gst are : %v --- %v", statuses, gst)

				return GstDetail{Gst: gst, Returns: statuses}
			} else {
				return GstDetail{ErrorMessage: "Not able to extract response from Gst API calls"}
			}
		} else {
			return GstDetail{ErrorMessage: "Not able to extract response from Gst API calls"}
		}
	}
}

func (s *GstScrapper) extractResponseFromHttpRequest(page *rod.Page, url string, selector string, regexString string) (*proto.NetworkGetResponseBodyResult, string, error) {
	var requestId proto.NetworkRequestID

	var response *proto.NetworkGetResponseBodyResult

	wait := page.EachEvent(func(e *proto.NetworkRequestWillBeSent) {
		s.logger.Infof("Request: %s %s\n", e.Request.Method, e.Request.URL)

		switch true {
		case strings.Contains(e.Request.URL, url) && e.Request.Method != "OPTIONS":
			requestId = e.RequestID
			break
		}
	}, func(e *proto.NetworkLoadingFinished) (stop bool) {
		if e.RequestID == requestId {
			m := proto.NetworkGetResponseBody{RequestID: requestId}
			res, _ := m.Call(page)

			response = res

			return true
		}

		return false
	}, func(e *proto.NetworkLoadingFailed) (stop bool) {
		if e.Type == proto.NetworkResourceTypeFetch || e.Type == proto.NetworkResourceTypeMedia || e.Type == proto.NetworkResourceTypeXHR {
			response = nil
			return true
		}

		return false
	})

	var ele *rod.Element
	err := rod.Try(func() {
		ele = page.MustElementR(selector, regexString)
	})

	if err != nil {
		return nil, string(requestId), err
	} else {
		ele.MustClick()
	}

	wait()

	return response, string(requestId), nil
}

func (s *GstScrapper) login(page *rod.Page, gst models.Gst, useCredentialFromSettings bool) DashboardDetail {
	var dashboard DashboardDetail
	for v := range 3 {
		s.logger.Infof("Try logging in as %d time", v)

		err := s.setUsernamePassword(page, gst, useCredentialFromSettings)

		if err != nil {
			dashboard.ErrorMessage = fmt.Sprintf("Something went wrong while setting username password. The error is: %s", err.Error())
			break
		}

		data, id, err := s.extractResponseFromHttpRequest(page, "/audiocaptcha", "i.fa.fa-volume-up", "/.*/")

		if err == nil && data != nil {
			dashboard = s.setCaptchaAndLogin(page, data, id)

			if dashboard.Landed {
				break
			} else {
				if !dashboard.ShouldRetry {
					break
				}
			}
		} else {
			s.logger.Errorf("Something went wrong while processing captcha code. Trying again %v time", v)
		}
	}

	return dashboard
}

func (s *GstScrapper) setCaptchaAndLogin(page *rod.Page, audioData *proto.NetworkGetResponseBodyResult, id string) DashboardDetail {
	fileName := fmt.Sprintf("%s.mp3", id)
	defer func(fn string) {
		os.Remove(fn)
	}(fileName)

	err := createAudioFile(audioData.Body, id)

	if err == nil {
		code, err := s.speechService.SpeechToText(fileName)
		if err == nil {
			var numericRegex = regexp.MustCompile(`[^\p{N} ]+`)

			err = rod.Try(func() {
				captcha := page.MustElement("#captcha")

				captcha.MustSelectAllText().MustType(input.Backspace).MustInput(numericRegex.ReplaceAllString(code, ""))

				page.MustElement("[type=submit]").MustClick()
			})

			if err != nil {
				return DashboardDetail{ErrorMessage: fmt.Sprintf("Something went wrong while setting captcha code and clicking login button. The error is: %s", err.Error())}
			}

			err := rod.Try(func() {
				page.Timeout(time.Duration(10 * time.Second)).MustElement("body.modal-open")
			})

			if err != nil {
				return s.checkDashboardPage(page)
			} else {
				found, ele, _ := page.HasR("#adhrtableV div.modal-footer > a", "/Remind me later/")
				if found {
					ele.MustClick()

					return s.checkDashboardPage(page)
				} else {
					found, _, _ := page.HasR("#confirmDlg div.modal-footer > a", "/FILE AMENDMENT/")
					if found {
						return DashboardDetail{ErrorMessage: "Bank account is not linked with GSTIN"}
					}
				}

				return DashboardDetail{ErrorMessage: "There is unknow dialog preventing the process to get gst details"}
			}
		} else {
			s.logger.Errorf("Error while processing the Speech to text. The code is - %s", code)
			return DashboardDetail{ShouldRetry: true}
		}
	}

	s.logger.Errorf("Error while processing the Speech to text. Error - %s", err.Error())
	return DashboardDetail{
		ShouldRetry: true,
	}
}

func (s *GstScrapper) checkDashboardPage(page *rod.Page) DashboardDetail {
	err := rod.Try(func() {
		page.Timeout(time.Duration(5 * time.Second)).MustElement(".dp-widgt")
	})

	if errors.Is(err, context.DeadlineExceeded) {
		s.logger.Errorf("Timeout while looking for dashboard widget - Reason: %s", err.Error())

		err := rod.Try(func() {
			page.Timeout(time.Duration(5 * time.Second)).MustElement("#submitpwd")
		})

		if err == nil {
			return DashboardDetail{
				ErrorMessage: fmt.Sprintf("NOTIFICATION|GST credential needs to be changed for the GST user id %s", s.cfg.Server.Gst.Username),
			}
		}

		var errMsg *rod.Element
		err = rod.Try(func() {
			errMsg = page.Timeout(time.Duration(5 * time.Second)).MustElement("span.err").CancelTimeout()
		})

		if err == nil {
			msg, _ := errMsg.Text()

			if strings.Contains(msg, "Enter valid Letters shown") {
				return DashboardDetail{ShouldRetry: true}
			}
		}

		err = rod.Try(func() {
			errMsg = page.Timeout(time.Duration(5 * time.Second)).MustElement("div.alert-danger").CancelTimeout()
		})

		if err == nil {
			msg, _ := errMsg.Text()

			if strings.Contains(msg, "Invalid Username or Password. Please try again.") {
				return DashboardDetail{
					ErrorMessage: "NOTIFICATION|GST username or password is invalid. Please update GST credential and try again.",
				}
			}
		} else {
			return DashboardDetail{ErrorMessage: fmt.Sprintf("Something went wrong while looking for dashboard widget - Reason: %s", err.Error())}
		}
	} else if err != nil {
		return DashboardDetail{ErrorMessage: fmt.Sprintf("Something went wrong while looking for dashboard widget - Reason: %s", err.Error())}
	} else {
		s.logger.Infof("Successfully landed into dashboard widget")
		return DashboardDetail{
			Landed: true,
		}
	}

	return DashboardDetail{ErrorMessage: "Something went wrong. It is neither timeout nor element not found."}
}

func (s *GstScrapper) setUsernamePassword(page *rod.Page, gst models.Gst, useCredentialFromSettings bool) error {
	var usernameEl *rod.Element
	var passwordEl *rod.Element

	err := rod.Try(func() {
		usernameEl = page.Timeout(time.Duration(5 * time.Second)).MustElement("#username").CancelTimeout()
		passwordEl = page.Timeout(time.Duration(5 * time.Second)).MustElement("#user_pass").CancelTimeout()
	})

	if err != nil {
		return err
	}

	username := gst.Username
	password := gst.Password

	if useCredentialFromSettings {
		username = s.cfg.Server.Gst.Username
		password = s.cfg.Server.Gst.Password
	}

	usernameEl.MustSelectAllText().MustType(input.Backspace).MustInput(username)

	err = rod.Try(func() {
		page.Timeout(time.Duration(5 * time.Second)).MustElement("#imgCaptcha").MustWaitVisible()
	})

	if err != nil {
		return err
	}

	passwordEl.MustSelectAllText().MustType(input.Backspace).MustInput(password)

	return nil
}

func createAudioFile(fileContent string, id string) error {
	dec, err := base64.StdEncoding.DecodeString(fileContent)
	if err != nil {
		return err
	}

	f, err := os.Create(fmt.Sprintf("%s.mp3", id))
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
