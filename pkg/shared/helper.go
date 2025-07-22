package shared

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"io"
	"strings"

	"github.com/pion/webrtc/v4"
	opusDec "gopkg.in/hraban/opus.v2"
)

// Reader für empfangene RTP/Opus-Pakete
type opusRTPReader struct {
	track   *webrtc.TrackRemote
	decoder *opusDec.Decoder
	pcmBuf  []int16
	byteBuf []byte
	offset  int
}

func NewOpusRTPReader(track *webrtc.TrackRemote, decoder *opusDec.Decoder) *opusRTPReader {
	return &opusRTPReader{
		track:   track,
		decoder: decoder,
		pcmBuf:  make([]int16, 960),
		byteBuf: make([]byte, 960*2),
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

		for i := 0; i < n; i++ {
			binary.LittleEndian.PutUint16(r.byteBuf[i*2:], uint16(r.pcmBuf[i]))
		}

		r.byteBuf = r.byteBuf[:n*2]
	}

	n := copy(p, r.byteBuf[r.offset:])

	r.offset += n
	if r.offset >= len(r.byteBuf) {
		r.offset = 0
	}

	return n, nil
}

// generateSecureToken generates a token containing random chars
func GenerateSecureToken(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}

	return base64.RawURLEncoding.EncodeToString(b)
}

// DecodeToResponse decodes a responseBody to a Response struct
func DecodeToResponse(body []byte) (*Response, error) {
	response := &Response{}
	dec := json.NewDecoder(strings.NewReader(string(body)))

	err := dec.Decode(&response)
	if err != nil {
		return response, err
	}

	return response, nil
}

// DecodeToMessage decodes a responseBody to a Message struct
func DecodeToMessage(body []byte) (Message, error) {
	message := Message{}
	dec := json.NewDecoder(strings.NewReader(string(body)))
	err := dec.Decode(&message)

	if err != nil {
		return message, err
	}

	return message, nil
}

// DecodeToGroup decodes a responseBody to a Group struct
func DecodeToGroup(body []byte) (*JsonGroup, error) {
	group := &JsonGroup{}
	dec := json.NewDecoder(strings.NewReader(string(body)))

	err := dec.Decode(&group)
	if err != nil {
		return group, err
	}

	return group, nil
}

func DecodeStringToJsonGroup(jsonGroup string) (*JsonGroup, error) {
	var group *JsonGroup
	dec := json.NewDecoder(strings.NewReader(jsonGroup))
	err := dec.Decode(&group)

	return group, err
}
