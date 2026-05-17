package asr

import (
	sherpa "github.com/k2-fsa/sherpa-onnx-go/sherpa_onnx"
)

// postProcessPunctuation adds commas and a period based on timestamps.
// This is used by Paraformer model which doesn't have built-in punctuation.
func postProcessPunctuation(result *sherpa.OfflineRecognizerResult) string {
	if len(result.Tokens) == 0 {
		return result.Text
	}

	var punctuatedText string
	tokens := result.Tokens
	timestamps := result.Timestamps

	// If no timestamps, just add a period at the end
	if len(timestamps) < len(tokens) {
		return result.Text + "。"
	}

	for i := 0; i < len(tokens); i++ {
		punctuatedText += tokens[i]

		// Check for pause after this token
		if i < len(tokens)-1 {
			pauseDuration := timestamps[i+1] - timestamps[i]
			// If pause is more than 0.8 seconds, add a comma
			if pauseDuration > 0.8 {
				punctuatedText += "，"
			}
		}
	}

	return punctuatedText + "。"
}
