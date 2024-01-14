package webrtc

import (
	"bytes"
	"encoding/binary"
	"strconv"

	"github.com/pion/webrtc/v3"

	"encoding/json"
)

const (
	OP_MOVE     = 0x01
	OP_SCROLL   = 0x02
	OP_KEY_DOWN = 0x03
	OP_KEY_UP   = 0x04
	OP_KEY_CLK  = 0x05
)

type PayloadHeader struct {
	Event  uint8
	Length uint16
}

type PayloadMove struct {
	PayloadHeader
	X uint16
	Y uint16
}

type PayloadScroll struct {
	PayloadHeader
	X int16
	Y int16
}

type PayloadKey struct {
	PayloadHeader
	Key uint64 // TODO: uint32
}

// Your data structure in Go
type PayloadMzelo struct {
	Event string `json:"event"`
	Value string `json:"value"`
}

// Decode JSON string to Go object
func jsonToStruct(jsonString string) (*PayloadMzelo, error) {
	var obj PayloadMzelo
	err := json.Unmarshal([]byte(jsonString), &obj)
	if err != nil {
		return nil, err
	}
	return &obj, nil
}

// var mu = sync.Mutex{}

func (manager *WebRTCManager) handle(id string, msg webrtc.DataChannelMessage, connection *webrtc.PeerConnection, videoTrack *webrtc.TrackLocalStaticSample, videosdTrack *webrtc.TrackLocalStaticSample) error {
	if (!manager.config.ImplicitControl && !manager.sessions.IsHost(id)) || (manager.config.ImplicitControl && !manager.sessions.CanControl(id)) {
		return nil
	}

	// manager.logger.Info().Msgf("data message isString %s", msg.IsString)

	if (msg.IsString) {
		// Decode JSON string to Go object
		obj, err := jsonToStruct(string(msg.Data))
		if err != nil {
			manager.logger.Info().Msgf("Error decoding JSON: %s", err)
			return nil
		}

		// Use the decoded object
		manager.logger.Info().Msgf("Decoded Object: %s", obj)

		if obj.Event=="quality435345" {

			// connection.mu.Lock()
			// defer connection.mu.Unlock()

			// Find the transceiver with the existing video track
			var videoTransceiver *webrtc.RTPTransceiver
			for _, transceiver := range connection.GetTransceivers() {
				if transceiver.Sender() != nil && transceiver.Sender().Track() != nil && transceiver.Sender().Track().Kind() == webrtc.RTPCodecTypeVideo {
					videoTransceiver = transceiver
					break
				}
			}

			if videoTransceiver == nil {
				manager.logger.Info().Msgf("existing video track not found")
				return nil
			}

			// Check the transceiver state before switching
			if videoTransceiver.Direction() != webrtc.RTPTransceiverDirectionSendrecv {
				manager.logger.Info().Msgf("video transceiver not in sendrecv state")
				return nil
			}
		
			// Log the current transceiver state before switching
			manager.logger.Info().Msgf("Current video transceiver state: %s", videoTransceiver.Direction().String())

			// // Remove the existing video track
			// if err := videoTransceiver.Sender().RemoveTrack(); err != nil {
			// 	manager.logger.Info().Msgf("failed to remove video track: %w", err)
			// 	return nil
			// }

			// Remove the existing video track
			// if err := connection.RemoveTrack(videoTransceiver.Sender()); err != nil {
			// 	manager.logger.Info().Msgf("failed to remove video track: %w", err)
			// 	return nil
			// }
	

			// Remove the existing video track by replacing it with an empty track
			// emptyTrack, err := webrtc.NewTrackLocalStaticRTP(manager.capture.Video().Codec().Capability, "empty", "empty")
			// if err != nil {
			// 	manager.logger.Info().Msgf("failed to create empty track: %w", err)
			// 	return nil
			// }
			
			// if err := videoTransceiver.Sender().ReplaceTrack(emptyTrack); err != nil {
			// 	manager.logger.Info().Msgf("failed to switch video track: %w", err)
			// 	return nil
			// }
	
			// if obj.Value == "HD" {
			// 	// Replace the empty track with the HD video track
			// 	if err := videoTransceiver.Sender().ReplaceTrack(videoTrack); err != nil {
			// 		manager.logger.Info().Msgf("failed to switch video track: %w", err)
			// 	}
			// } else if obj.Value == "SD" {
			// 	// Replace the empty track with the SD video track
			// 	if err := videoTransceiver.Sender().ReplaceTrack(videosdTrack); err != nil {
			// 		manager.logger.Info().Msgf("failed to switch video track: %w", err)
			// 	}
			// } else {
			// 	// Handle other cases if needed
			// }
	
			// if obj.Value == "HD" {
			// 	// Add the HD video track
			// 	if _, err := connection.AddTrack(videoTrack); err != nil {
			// 		manager.logger.Info().Msgf("failed to add HD video track: %w", err)
			// 		return nil
			// 	}
			// } else if obj.Value == "SD" {
			// 	// Add the SD video track
			// 	if _, err := connection.AddTrack(videosdTrack); err != nil {
			// 		manager.logger.Info().Msgf("failed to add SD video track: %w", err)
			// 		return nil
			// 	}
			// } else {
			// 	// Handle other cases if needed
			// }
	
			// manager.logger.Info().Msgf("Video track switched successfully.")
			
			// if obj.Value == "HD" {
			// 	// Replace the existing video track with the new one
			// 	if err := videoTransceiver.Sender().ReplaceTrack(videoTrack); err != nil {
			// 		manager.logger.Info().Msgf("failed to switch video track: %w", err)
			// 	}
			// } else if (obj.Value == "SD") {
			// 	// Replace the existing video track with the new one
			// 	if err := videoTransceiver.Sender().ReplaceTrack(videosdTrack); err != nil {
			// 		manager.logger.Info().Msgf("failed to switch video track: %w", err)
			// 	}
			// } else {
				
			// }

			manager.logger.Info().Msgf("Video track switched successfully. New transceiver state: %s", videoTransceiver.Direction().String())


		}
		return nil
	}


	buffer := bytes.NewBuffer(msg.Data)
	header := &PayloadHeader{}
	hbytes := make([]byte, 3)

	if _, err := buffer.Read(hbytes); err != nil {
		return err
	}

	if err := binary.Read(bytes.NewBuffer(hbytes), binary.LittleEndian, header); err != nil {
		return err
	}

	buffer = bytes.NewBuffer(msg.Data)

	switch header.Event {
	case OP_MOVE:
		payload := &PayloadMove{}
		if err := binary.Read(buffer, binary.LittleEndian, payload); err != nil {
			return err
		}

		manager.desktop.Move(int(payload.X), int(payload.Y))
	case OP_SCROLL:
		payload := &PayloadScroll{}
		if err := binary.Read(buffer, binary.LittleEndian, payload); err != nil {
			return err
		}

		manager.logger.
			Debug().
			Str("x", strconv.Itoa(int(payload.X))).
			Str("y", strconv.Itoa(int(payload.Y))).
			Msg("scroll")

		manager.desktop.Scroll(int(payload.X), int(payload.Y))
	case OP_KEY_DOWN:
		payload := &PayloadKey{}
		if err := binary.Read(buffer, binary.LittleEndian, payload); err != nil {
			return err
		}

		if payload.Key < 8 {
			err := manager.desktop.ButtonDown(uint32(payload.Key))
			if err != nil {
				manager.logger.Warn().Err(err).Msg("button down failed")
				return nil
			}

			manager.logger.Debug().Msgf("button down %d", payload.Key)
		} else {
			err := manager.desktop.KeyDown(uint32(payload.Key))
			if err != nil {
				manager.logger.Warn().Err(err).Msg("key down failed")
				return nil
			}

			manager.logger.Debug().Msgf("key down %d", payload.Key)
		}
	case OP_KEY_UP:
		payload := &PayloadKey{}
		err := binary.Read(buffer, binary.LittleEndian, payload)
		if err != nil {
			return err
		}

		if payload.Key < 8 {
			err := manager.desktop.ButtonUp(uint32(payload.Key))
			if err != nil {
				manager.logger.Warn().Err(err).Msg("button up failed")
				return nil
			}

			manager.logger.Debug().Msgf("button up %d", payload.Key)
		} else {
			err := manager.desktop.KeyUp(uint32(payload.Key))
			if err != nil {
				manager.logger.Warn().Err(err).Msg("key up failed")
				return nil
			}

			manager.logger.Debug().Msgf("key up %d", payload.Key)
		}
	case OP_KEY_CLK:
		// unused
		break
	}

	return nil
}
