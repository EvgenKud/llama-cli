package cli

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-skynet/LocalAI/core/backend"
	"github.com/go-skynet/LocalAI/core/config"
	"github.com/go-skynet/LocalAI/pkg/model"
)

type TranscriptCMD struct {
	Filename string `arg:""`

	Backend           string `short:"b" default:"whisper" help:"Backend to run the transcription model"`
	Model             string `short:"m" required:"" help:"Model name to run the TTS"`
	Language          string `short:"l" help:"Language of the audio file"`
	Threads           int    `short:"t" default:"1" help:"Number of threads used for parallel computation"`
	ModelsPath        string `env:"LOCALAI_MODELS_PATH,MODELS_PATH" type:"path" default:"${basepath}/models" help:"Path containing models used for inferencing" group:"storage"`
	BackendAssetsPath string `env:"LOCALAI_BACKEND_ASSETS_PATH,BACKEND_ASSETS_PATH" type:"path" default:"/tmp/localai/backend_data" help:"Path used to extract libraries that are required by some of the backends in runtime" group:"storage"`
}

func (t *TranscriptCMD) Run(ctx *Context) error {
	opts := &config.ApplicationConfig{
		ModelPath:         t.ModelsPath,
		Context:           context.Background(),
		AssetsDestination: t.BackendAssetsPath,
	}

	cl := config.NewBackendConfigLoader()
	ml := model.NewModelLoader(opts.ModelPath)
	if err := cl.LoadBackendConfigsFromPath(t.ModelsPath); err != nil {
		return err
	}

	c, exists := cl.GetBackendConfig(t.Model)
	if !exists {
		return errors.New("model not found")
	}

	c.Threads = &t.Threads

	defer ml.StopAllGRPC()

	tr, err := backend.ModelTranscription(t.Filename, t.Language, ml, c, opts)
	if err != nil {
		return err
	}
	for _, segment := range tr.Segments {
		fmt.Println(segment.Start.String(), "-", segment.Text)
	}
	return nil
}