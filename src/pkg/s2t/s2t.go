package s2t

import (
	"context"
	"os"
	"sync"

	"github.com/AssemblyAI/assemblyai-go-sdk"
	"github.com/jaganathanb/dapps-api/config"
	"github.com/jaganathanb/dapps-api/pkg/logging"
)

type DAppsSpeechToText struct {
	logger logging.Logger
	cfg    *config.Config
}

var speechService *DAppsSpeechToText
var speechServiceOnce sync.Once

func NewDAppsSpeechToText(cfg *config.Config) *DAppsSpeechToText {
	speechServiceOnce.Do(func() {
		speechService = &DAppsSpeechToText{
			logger: logging.NewLogger(cfg),
			cfg:    cfg,
		}
	})

	return speechService
}

func (s *DAppsSpeechToText) SpeechToText(filePath string) (string, error) {
	apiKey := "c11ce14411ae432393eac94001f5cf4d"

	ctx := context.Background()

	client := assemblyai.NewClient(apiKey)

	f, err := os.Open(filePath)
	if err != nil {
		s.logger.Errorf("Couldn't open audio file:", err)

		return "", err
	}
	defer f.Close()

	transcript, err := client.Transcripts.TranscribeFromReader(ctx, f, nil)
	if err != nil {
		s.logger.Errorf("Something bad happened:", err)

		return "", err
	}

	return *transcript.Text, nil
}
