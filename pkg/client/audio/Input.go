package audio

import (
	"fmt"
	"io"
	"sync"

	t "github.com/F4c3hugg3r/Go-Chat-Server/pkg/shared"
	"github.com/pion/mediadevices"
	"github.com/pion/mediadevices/pkg/codec/opus"
	"github.com/pion/mediadevices/pkg/prop"
	"github.com/pion/mediadevices/pkg/wave"
)

type MicrophoneInput struct {
	MediaStream        mediadevices.MediaStream
	OriginalAudioTrack mediadevices.Track
	SilentAudioTrack   mediadevices.Track
	CodecSelector      *mediadevices.CodecSelector
}

func NewMicrophoneInput(logChannel chan t.Log) (*MicrophoneInput, error) {
	input := &MicrophoneInput{}
	var err error

	opusParams, err := opus.NewParams()
	if err != nil {
		logChannel <- t.Log{Text: fmt.Sprintf("WebRTC: Fehler bei opus.NewParams: %v", err)}
		return nil, err
	}

	input.CodecSelector = mediadevices.NewCodecSelector(mediadevices.WithAudioEncoders(&opusParams))

	input.MediaStream, err = mediadevices.GetUserMedia(mediadevices.MediaStreamConstraints{
		Audio: func(c *mediadevices.MediaTrackConstraints) {
			c.SampleRate = prop.Int(Samplerate)
			c.ChannelCount = prop.Int(Channels)
		},
		Codec: input.CodecSelector,
	})
	if err != nil {
		logChannel <- t.Log{Text: fmt.Sprintf("Input: Fehler bei GetUserMedia: %v", err)}
		return nil, err
	}

	logChannel <- t.Log{Text: "Input: MediaStream erhalten"}

	for _, track := range input.MediaStream.GetTracks() {
		logChannel <- t.Log{Text: fmt.Sprintf("Input: Füge Track hinzu: ID=%s, Kind=%s", track.ID(), track.Kind())}
		track.OnEnded(func(err error) {
			logChannel <- t.Log{Text: fmt.Sprintf("Input: Track (ID: %s) beendet mit Fehler: %v", track.ID(), err)}
		})

		input.OriginalAudioTrack = track
	}

	source := &SilenceSource{closed: false, mu: &sync.RWMutex{}}
	input.SilentAudioTrack = mediadevices.NewAudioTrack(source, input.CodecSelector)

	return input, nil
}

// SilenceSource implementiert die mediadevices.AudioSource Schnittstelle
type SilenceSource struct {
	mu     *sync.RWMutex
	closed bool
}

func (s *SilenceSource) Read() (wave.Audio, func(), error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil, nil, io.EOF
	}

	samples := Samplerate / 50 // 20ms bei 48000Hz = 960 Samples
	// Interleaved nur für Stereo audio benötigt
	chunk := wave.NewInt16NonInterleaved(wave.ChunkInfo{
		Len:          samples,
		Channels:     Channels,
		SamplingRate: Samplerate,
	})

	// Fülle mit Nullwerten (Stille)
	for ch := 0; ch < Channels; ch++ {
		for i := 0; i < samples; i++ {
			chunk.Data[ch][i] = 0
		}
	}

	return chunk, func() {}, nil
}

func (s *SilenceSource) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.closed = !s.closed
	return nil
}

func (s *SilenceSource) ID() string {
	return "silence-audio-source"
}
