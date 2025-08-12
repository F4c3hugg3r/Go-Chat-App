package legacy

// IN WEBRTC
// func (p *Peer) MuteMic() error {
// 	p.mu.Lock()
// 	defer p.mu.Unlock()

// 	switch p.isMuted {
// 	case false:
// 		err := p.audioTransceiver.Sender().ReplaceTrack(p.chatClient.MicrophoneInput.SilentAudioTrack)
// 		if err != nil {
// 			p.logChannel <- t.Log{Text: fmt.Sprintf("Failed to replace track: %v", err)}
// 			return err
// 		}

// 		p.isMuted = true
// 		p.logChannel <- t.Log{Text: "Microphone muted"}

// 	case true:
// 		err := p.audioTransceiver.Sender().ReplaceTrack(p.chatClient.MicrophoneInput.OriginalAudioTrack)
// 		if err != nil {
// 			p.logChannel <- t.Log{Text: fmt.Sprintf("Failed to replace track: %v", err)}
// 			return err
// 		}

// 		p.isMuted = false
// 		p.logChannel <- t.Log{Text: "Microphone unmuted"}
// 	}

// 	return nil
// }

// func (p *Peer) AddTracksToPeerConnection(logChannel chan t.Log) error {
// 	var err error
// 	p.audioTransceiver, err = p.peerConn.AddTransceiverFromTrack(p.chatClient.MicrophoneInput.OriginalAudioTrack,
// 		webrtc.RTPTransceiverInit{
// 			Direction: webrtc.RTPTransceiverDirectionSendrecv,
// 		},
// 	)
// 	if err != nil {
// 		logChannel <- t.Log{Text: fmt.Sprintf("WebRTC: Fehler bei AddTransceiverFromTrack: %v", err)}
// 		return err
// 	}

// 	return nil
// }

// func (p *Peer) InitWebRTCAPI() error {
// 	p.logChannel <- t.Log{Text: "WebRTC: InitWebRTCAPI gestartet"}

// 	mediaEngine := webrtc.MediaEngine{}
// 	p.chatClient.MicrophoneInput.CodecSelector.Populate(&mediaEngine)
// 	// einheitliche API Instanz damit alle Peers die gleichen Codecs nutzen
// 	p.api = webrtc.NewAPI(webrtc.WithMediaEngine(&mediaEngine))

// 	p.logChannel <- t.Log{Text: "WebRTC: InitWebRTCAPI abgeschlossen"}

// 	return nil
// }

// INPUT

// type MicrophoneInput struct {
// 	MediaStream        mediadevices.MediaStream
// 	OriginalAudioTrack mediadevices.Track
// 	SilentAudioTrack   mediadevices.Track
// 	CodecSelector      *mediadevices.CodecSelector
// }

// func NewMicrophoneInput(logChannel chan t.Log) (*MicrophoneInput, error) {
// 	input := &MicrophoneInput{}
// 	var err error

// 	opusParams, err := opus.NewParams()
// 	if err != nil {
// 		logChannel <- t.Log{Text: fmt.Sprintf("WebRTC: Fehler bei opus.NewParams: %v", err)}
// 		return nil, err
// 	}

// 	input.CodecSelector = mediadevices.NewCodecSelector(mediadevices.WithAudioEncoders(&opusParams))

// 	input.MediaStream, err = mediadevices.GetUserMedia(mediadevices.MediaStreamConstraints{
// 		Audio: func(c *mediadevices.MediaTrackConstraints) {
// 			c.SampleRate = prop.Int(Samplerate)
// 			c.ChannelCount = prop.Int(Channels)
// 		},
// 		Codec: input.CodecSelector,
// 	})
// 	if err != nil {
// 		logChannel <- t.Log{Text: fmt.Sprintf("Input: Fehler bei GetUserMedia: %v", err)}
// 		return nil, err
// 	}

// 	logChannel <- t.Log{Text: "Input: MediaStream erhalten"}

// 	for _, track := range input.MediaStream.GetTracks() {
// 		logChannel <- t.Log{Text: fmt.Sprintf("Input: Füge Track hinzu: ID=%s, Kind=%s", track.ID(), track.Kind())}
// 		track.OnEnded(func(err error) {
// 			logChannel <- t.Log{Text: fmt.Sprintf("Input: Track (ID: %s) beendet mit Fehler: %v", track.ID(), err)}
// 		})

// 		input.OriginalAudioTrack = track
// 	}

// 	source := &SilenceSource{closed: false, mu: &sync.RWMutex{}}
// 	input.SilentAudioTrack = mediadevices.NewAudioTrack(source, input.CodecSelector)

// 	return input, nil
// }

// // SilenceSource implementiert die mediadevices.AudioSource Schnittstelle
// type SilenceSource struct {
// 	mu     *sync.RWMutex
// 	closed bool
// }

// func (s *SilenceSource) Read() (wave.Audio, func(), error) {
// 	s.mu.Lock()
// 	defer s.mu.Unlock()

// 	if s.closed {
// 		return nil, nil, io.EOF
// 	}

// 	samples := Samplerate / 50 // 20ms bei 48000Hz = 960 Samples
// 	// Interleaved nur für Stereo audio benötigt
// 	chunk := wave.NewInt16NonInterleaved(wave.ChunkInfo{
// 		Len:          samples,
// 		Channels:     Channels,
// 		SamplingRate: Samplerate,
// 	})

// 	// Fülle mit Nullwerten (Stille)
// 	for ch := 0; ch < Channels; ch++ {
// 		for i := 0; i < samples; i++ {
// 			chunk.Data[ch][i] = 0
// 		}
// 	}

// 	return chunk, func() {}, nil
// }

// func (s *SilenceSource) Close() error {
// 	s.mu.Lock()
// 	defer s.mu.Unlock()

// 	s.closed = !s.closed
// 	return nil
// }

// func (s *SilenceSource) ID() string {
// 	return "silence-audio-source"
// }
