package s2t

import (
	"context"
	"log"
	"os"

	"github.com/AssemblyAI/assemblyai-go-sdk"
)

func SpeechToText(filePath string) (string, error) {
	apiKey := "c11ce14411ae432393eac94001f5cf4d"

	ctx := context.Background()

	client := assemblyai.NewClient(apiKey)

	f, err := os.Open(filePath)
	if err != nil {
		log.Fatal("Couldn't open audio file:", err)

		return "", err
	}
	defer f.Close()

	transcript, err := client.Transcripts.TranscribeFromReader(ctx, f, nil)
	if err != nil {
		log.Fatal("Something bad happened:", err)

		return "", err
	}

	return *transcript.Text, nil
}
