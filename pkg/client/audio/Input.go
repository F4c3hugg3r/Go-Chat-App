package audio

import (
	"fmt"
	"time"

	t "github.com/F4c3hugg3r/Go-Chat-Server/pkg/shared"
	p "github.com/gordonklaus/portaudio"
	"github.com/pion/mediadevices"
	"github.com/pion/mediadevices/pkg/codec/opus"
	"github.com/pion/webrtc/v4"
	"github.com/pion/webrtc/v4/pkg/media"
	opusEnc "gopkg.in/hraban/opus.v2"
)

type PortAudioMicInput struct {
	// SampleRate float64
	byteBuf []byte
	pcmBuf  []int16
	// Channels   int
	Stream        *p.Stream
	Muted         bool
	Track         *webrtc.TrackLocalStaticSample
	OpusEnc       *opusEnc.Encoder
	CodecSelector *mediadevices.CodecSelector
}

func InitializePortAudioMic(logChannel chan t.Log) (*PortAudioMicInput, error) {
	logChannel <- t.Log{Text: "[InitializePortAudio] Initialisiere PortAudio Mikrofon"}
	frameSize := Samplerate * FrameSizeMs * Channels / 1000

	err := p.Initialize()
	if err != nil {
		return nil, fmt.Errorf("PortAudio: Fehler bei Initialisierung: %w", err)
	}

	pcmBuf := make([]int16, frameSize)
	byteBuf := make([]byte, 1000)

	pm := &PortAudioMicInput{
		pcmBuf: pcmBuf,
		// SampleRate: mic.DefaultSampleRate,
		// Channels:   1,
		byteBuf: byteBuf,
	}

	opusParams, err := opus.NewParams()
	if err != nil {
		logChannel <- t.Log{Text: fmt.Sprintf("WebRTC: Fehler bei opus.NewParams: %v", err)}
		return nil, err
	}

	pm.CodecSelector = mediadevices.NewCodecSelector(mediadevices.WithAudioEncoders(&opusParams))

	logChannel <- t.Log{Text: "[InitializePortAudio] CodecSelector initialized"}

	pm.OpusEnc, err = opusEnc.NewEncoder(Samplerate, Channels, opusEnc.AppAudio)
	if err != nil {
		return nil, fmt.Errorf("Opus Encoder: Fehler beim Erstellen: %w", err)
	}

	// maybe zum Quality verbessern
	// pm.OpusEnc.SetBitrate(64000)
	// pm.OpusEnc.SetComplexity(10)

	stream, err := p.OpenDefaultStream(Channels, 0, Samplerate, len(pcmBuf), pm.SendToWebRTCCallback(logChannel))
	if err != nil {
		return nil, fmt.Errorf("PortAudio: Fehler beim Öffnen des DefaultStreams: %w", err)
	}

	pm.Stream = stream

	pm.Track, err = webrtc.NewTrackLocalStaticSample(
		webrtc.RTPCodecCapability{
			MimeType:  opusParams.RTPCodec().MimeType,
			Channels:  uint16(Channels),
			ClockRate: uint32(Samplerate),
		}, "audio", "microphone",
	)
	if err != nil {
		return nil, fmt.Errorf("WebRTC: Fehler beim Erstellen des Tracks: %w", err)
	}

	logChannel <- t.Log{Text: "[InitializePortAudio] PortAudio Mikrofon erfolgreich initialisiert"}
	return pm, nil
}

func (pm *PortAudioMicInput) SendToWebRTCCallback(logChannel chan t.Log) func(in []int16) {
	return func(in []int16) {
		pm.pcmBuf = in

		n, err := pm.OpusEnc.Encode(pm.pcmBuf, pm.byteBuf)
		if err != nil {
			logChannel <- t.Log{Text: fmt.Sprintf("[PortAudioMicInput] Opus encode error: %v", err)}
			return
		}

		err = pm.Track.WriteSample(media.Sample{Data: pm.byteBuf[:n], Duration: 20 * time.Millisecond})
		if err != nil {
			logChannel <- t.Log{Text: fmt.Sprintf("[PortAudioMicInput] WebRTC WriteSample error: %v", err)}
		}
	}
}
