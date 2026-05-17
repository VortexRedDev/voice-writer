package file

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

// DecodeWavToS16 decodes a standard PCM WAV file into int16 samples.
// It supports only 16-bit PCM for now.
func DecodeWavToS16(path string) ([]int16, int, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, 0, err
	}
	defer file.Close()

	// Read RIFF header
	var header [44]byte
	if _, err := io.ReadFull(file, header[:]); err != nil {
		return nil, 0, fmt.Errorf("failed to read wav header: %v", err)
	}

	if string(header[0:4]) != "RIFF" || string(header[8:12]) != "WAVE" {
		return nil, 0, fmt.Errorf("not a valid wav file")
	}

	// Format chunk
	channels := int(binary.LittleEndian.Uint16(header[22:24]))
	sampleRate := int(binary.LittleEndian.Uint32(header[24:28]))
	bitsPerSample := int(binary.LittleEndian.Uint16(header[34:36]))

	if bitsPerSample != 16 {
		return nil, 0, fmt.Errorf("only 16-bit PCM wav is supported, got %d bits", bitsPerSample)
	}

	// Data chunk
	dataSize := int(binary.LittleEndian.Uint32(header[40:44]))
	
	// Some WAV files have extra chunks before 'data', so we might need to search for 'data'
	if string(header[36:40]) != "data" {
		// Seek to find "data" chunk
		_, err = file.Seek(36, 0)
		if err != nil {
			return nil, 0, err
		}
		
		var chunkID [4]byte
		for {
			_, err = io.ReadFull(file, chunkID[:])
			if err != nil {
				return nil, 0, fmt.Errorf("could not find data chunk: %v", err)
			}
			var chunkSize uint32
			err = binary.Read(file, binary.LittleEndian, &chunkSize)
			if err != nil {
				return nil, 0, err
			}
			
			if string(chunkID[:]) == "data" {
				dataSize = int(chunkSize)
				break
			}
			
			// Skip this chunk
			_, err = file.Seek(int64(chunkSize), 1)
			if err != nil {
				return nil, 0, err
			}
		}
	}

	data := make([]byte, dataSize)
	_, err = io.ReadFull(file, data)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to read wav data: %v", err)
	}

	samples := make([]int16, dataSize/2)
	for i := 0; i < len(samples); i++ {
		samples[i] = int16(binary.LittleEndian.Uint16(data[i*2 : i*2+2]))
	}

	// If stereo, convert to mono by averaging (optional, or just take one channel)
	if channels == 2 {
		mono := make([]int16, len(samples)/2)
		for i := 0; i < len(mono); i++ {
			mono[i] = int16((int32(samples[i*2]) + int32(samples[i*2+1])) / 2)
		}
		return mono, sampleRate, nil
	}

	return samples, sampleRate, nil
}
