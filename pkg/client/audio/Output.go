package audio

import (
	"encoding/binary"
	"fmt"
	"io"

	t "github.com/F4c3hugg3r/Go-Chat-Server/pkg/shared"
	"github.com/ebitengine/oto/v3"
	"github.com/pion/webrtc/v4"
	opusDec "gopkg.in/hraban/opus.v2"
)

const Channels = 1
const Samplerate = 48000
const FrameSizeMs = 20 //ms

type SpeakerOutput struct {
	OtoContext *oto.Context
}

func NewSpeakerOutput(logChannel chan t.Log) (*SpeakerOutput, error) {
	speakerOutput := &SpeakerOutput{}

	otoCtx, ready, err := oto.NewContext(&oto.NewContextOptions{
		SampleRate:   Samplerate,
		ChannelCount: Channels,
		Format:       oto.FormatSignedInt16LE,
		BufferSize:   0,
	})
	if err != nil {
		logChannel <- t.Log{Text: fmt.Sprintf("SpeakerOutput: oto.NewContext error: %v", err)}
		return nil, err
	}

	speakerOutput.OtoContext = otoCtx
	<-ready
	logChannel <- t.Log{Text: fmt.Sprintf("SpeakerOutput: oto.Context bereit err = %v", err)}

	return speakerOutput, nil
}

// Reader für empfangene RTP/Opus-Pakete
type opusRTPReader struct {
	track   *webrtc.TrackRemote
	decoder *opusDec.Decoder
	pcmBuf  []int16
	byteBuf []byte
	offset  int
}

// möglicherweise im Server Opus zu PCM umwandeln um Crosscompiling zu vereinfachen
func NewOpusRTPReader(track *webrtc.TrackRemote, decoder *opusDec.Decoder) *opusRTPReader {
	frameSize := Samplerate * FrameSizeMs * Channels / 1000
	return &opusRTPReader{
		track:   track,
		decoder: decoder,
		pcmBuf:  make([]int16, frameSize),
		byteBuf: make([]byte, frameSize*2),
		offset:  0,
	}
}

func (r *opusRTPReader) Read(p []byte) (int, error) {
	// Wenn byteBuf leer, neues RTP-Paket lesen und dekodieren
	if r.offset == 0 {
		pkt, _, err := r.track.ReadRTP()
		if err != nil {
			return 0, io.EOF
		}

		n, err := r.decoder.Decode(pkt.Payload, r.pcmBuf)
		if err != nil {
			return 0, io.EOF
		}

		// uints in bytes umwandeln (jeweils ein int -> 2 bytes)
		for i := 0; i < n*Channels; i++ {
			binary.LittleEndian.PutUint16(r.byteBuf[i*2:], uint16(r.pcmBuf[i]))
		}

		// länge des bytesBuf anpassen
		r.byteBuf = r.byteBuf[:n*Channels*2]
	}

	// daten in oto buffer kopieren
	n := copy(p, r.byteBuf[r.offset:])

	// offset anpassen / zurücksetzen
	r.offset += n
	if r.offset >= len(r.byteBuf) {
		r.offset = 0
	}

	return n, nil
}
