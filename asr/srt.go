package asr

import (
	"fmt"
	"strings"
	"time"
)

// ToSRT converts TimestampResult to SRT format.
func ToSRT(result *TimestampResult) string {
	if result == nil || len(result.Tokens) == 0 {
		return ""
	}

	var srt strings.Builder
	tokens := result.Tokens
	timestamps := result.Timestamps

	// If no timestamps, we can't create a valid SRT.
	if len(timestamps) < len(tokens) {
		return ""
	}

	segmentIndex := 1
	var segmentTokens []string
	var segmentStartTime float32
	
	// Max segment duration in seconds
	const maxSegmentDuration = 3.0
	// Max characters per segment
	const maxCharsPerSegment = 20
	// Pause threshold to split segment
	const pauseThreshold = 0.5

	segmentStartTime = timestamps[0]
	
	for i := 0; i < len(tokens); i++ {
		segmentTokens = append(segmentTokens, tokens[i])
		
		currentTime := timestamps[i]
		
		// Decide whether to end current segment
		shouldEnd := false
		
		// 1. Check for pause after this token
		if i < len(tokens)-1 {
			pauseDuration := timestamps[i+1] - timestamps[i]
			if pauseDuration > pauseThreshold {
				shouldEnd = true
			}
		} else {
			// End of all tokens
			shouldEnd = true
		}
		
		// 2. Check for segment duration
		if currentTime - segmentStartTime > maxSegmentDuration {
			shouldEnd = true
		}
		
		// 3. Check for segment length
		if len(strings.Join(segmentTokens, "")) > maxCharsPerSegment {
			shouldEnd = true
		}

		if shouldEnd {
			endTime := currentTime
			if i < len(tokens)-1 {
				// End time is slightly before next token
				endTime = timestamps[i+1] - 0.01
			} else {
				// For the last token, add a small buffer
				endTime = currentTime + 0.5
			}
			
			writeSRTSegment(&srt, segmentIndex, segmentStartTime, endTime, strings.Join(segmentTokens, ""))
			segmentIndex++
			
			// Reset for next segment
			segmentTokens = nil
			if i < len(tokens)-1 {
				segmentStartTime = timestamps[i+1]
			}
		}
	}

	return srt.String()
}

func writeSRTSegment(w *strings.Builder, index int, start, end float32, text string) {
	w.WriteString(fmt.Sprintf("%d\n", index))
	w.WriteString(fmt.Sprintf("%s --> %s\n", formatSRTTime(start), formatSRTTime(end)))
	w.WriteString(fmt.Sprintf("%s\n\n", strings.TrimSpace(text)))
}

func formatSRTTime(seconds float32) string {
	t := time.Unix(0, int64(seconds*float32(time.Second))).UTC()
	return fmt.Sprintf("%02d:%02d:%02d,%03d",
		t.Hour(), t.Minute(), t.Second(), t.Nanosecond()/1e6)
}
