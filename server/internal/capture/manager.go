package capture

import (
	"m1k1o/neko/internal/types/codec"
	"errors"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"m1k1o/neko/internal/config"
	"m1k1o/neko/internal/types"
)

type CaptureManagerCtx struct {
	logger  zerolog.Logger
	desktop types.DesktopManager

	// sinks
	broadcast *BroacastManagerCtx
	audio     *StreamSinkManagerCtx
	video     *StreamSinkManagerCtx
	videosd   *StreamSinkManagerCtx
}

func New(desktop types.DesktopManager, config *config.Capture) *CaptureManagerCtx {
	logger := log.With().Str("module", "capture").Logger()
	// logger.Info().Msgf("config %s", config);

	return &CaptureManagerCtx{
		logger:  logger,
		desktop: desktop,

		// sinks
		broadcast: broadcastNew(func(url string) (string, error) {
			return NewBroadcastPipeline(config.AudioDevice, config.Display, config.BroadcastPipeline, url)
		}, config.BroadcastUrl),
		audio: streamSinkNew(config.AudioCodec, func() (string, error) {
			return NewAudioPipeline(config.AudioCodec, config.AudioDevice, config.AudioPipeline, config.AudioBitrate)
		}, "audio"),
		video: streamSinkNew(codec.H264(), func() (string, error) {
			// use screen fps as default
			fps := desktop.GetScreenSize().Rate
			// if max fps is set, cap it to that value
			if config.VideoMaxFPS > 0 && config.VideoMaxFPS < fps {
				fps = config.VideoMaxFPS
			}
			return NewVideoPipeline(codec.H264(), config.Display, config.VideoPipeline, fps, config.VideoBitrate, config.VideoHWEnc)
		}, "video"),
		videosd: streamSinkNew(codec.H264(), func() (string, error) {
			// use screen fps as default
			fps := desktop.GetScreenSize().Rate
			// if max fps is set, cap it to that value
			if config.VideoMaxFPS > 0 && config.VideoMaxFPS < fps {
				fps = config.VideoMaxFPS
			}
			return NewVideoPipeline(codec.H264(), config.Display, config.VideoPipeline, fps, 1200, config.VideoHWEnc)
		}, "videosd"),
	}
}

func (manager *CaptureManagerCtx) Start() {
	if manager.broadcast.Started() {
		if err := manager.broadcast.createPipeline(); err != nil {
			manager.logger.Panic().Err(err).Msg("unable to create broadcast pipeline")
		}
	}

	go func() {
		for {
			before, ok := <-manager.desktop.GetScreenSizeChangeChannel()
			if !ok {
				manager.logger.Info().Msg("screen size change channel was closed")
				return
			}

			if before {
				// before screen size change, we need to destroy all pipelines

				manager.logger.Info().Msg("before screen size change, we need to destroy all pipelines")

				if manager.video.Started() {
					manager.video.destroyPipeline()
				}

				if manager.videosd.Started() {
					manager.videosd.destroyPipeline()
				}

				if manager.broadcast.Started() {
					manager.broadcast.destroyPipeline()
				}
			} else {
				// after screen size change, we need to recreate all pipelines

				manager.logger.Info().Msg("after screen size change, we need to recreate all pipelines")

				if manager.video.Started() {
					err := manager.video.createPipeline()
					if err != nil && !errors.Is(err, types.ErrCapturePipelineAlreadyExists) {
						manager.logger.Panic().Err(err).Msg("unable to recreate video pipeline")
					}
				}

				if manager.videosd.Started() {
					err := manager.videosd.createPipeline()
					if err != nil && !errors.Is(err, types.ErrCapturePipelineAlreadyExists) {
						manager.logger.Panic().Err(err).Msg("unable to recreate videosd pipeline")
					}
				}

				if manager.broadcast.Started() {
					err := manager.broadcast.createPipeline()
					if err != nil && !errors.Is(err, types.ErrCapturePipelineAlreadyExists) {
						manager.logger.Panic().Err(err).Msg("unable to recreate broadcast pipeline")
					}
				}
			}
		}
	}()
}

func (manager *CaptureManagerCtx) Shutdown() error {
	manager.logger.Info().Msgf("shutdown")

	manager.broadcast.shutdown()

	manager.audio.shutdown()
	manager.video.shutdown()
	manager.videosd.shutdown()

	return nil
}

func (manager *CaptureManagerCtx) Broadcast() types.BroadcastManager {
	return manager.broadcast
}

func (manager *CaptureManagerCtx) Audio() types.StreamSinkManager {
	return manager.audio
}

func (manager *CaptureManagerCtx) Video() types.StreamSinkManager {
	return manager.video
}

func (manager *CaptureManagerCtx) Videosd() types.StreamSinkManager {
	return manager.videosd
}
