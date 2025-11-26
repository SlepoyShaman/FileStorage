package ffmpeg

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"time"
)

func ProcessVideoConversion(inputPath string, outputPath string, options map[string]interface{}, callback func(string, float64)) error {
	if inputPath == "" {
		return fmt.Errorf("input path is required")
	}
	if outputPath == "" {
		return fmt.Errorf("output path is required")
	}
	if options == nil {
		options = make(map[string]interface{})
	}

	probeCmd := exec.Command("ffprobe", "-v", "quiet", "-print_format", "json", "-show_format", "-show_streams", inputPath)
	probeOutput, err := probeCmd.Output()
	if err != nil {
		return fmt.Errorf("ffprobe failed: %v", err)
	}

	var probeData map[string]interface{}
	if err := json.Unmarshal(probeOutput, &probeData); err != nil {
		return fmt.Errorf("failed to parse ffprobe output: %v", err)
	}

	streams, ok := probeData["streams"].([]interface{})
	if !ok {
		return fmt.Errorf("invalid streams data")
	}

	var videoStream map[string]interface{}
	for _, stream := range streams {
		if s, ok := stream.(map[string]interface{}); ok {
			if codecType, ok := s["codec_type"].(string); ok && codecType == "video" {
				videoStream = s
				break
			}
		}
	}

	if videoStream == nil {
		return fmt.Errorf("no video stream found")
	}

	args := []string{"-i", inputPath}

	if width, ok := options["width"].(int); ok {
		args = append(args, "-vf", fmt.Sprintf("scale=%d:-1", width))
	}

	if bitrate, ok := options["bitrate"].(string); ok {
		args = append(args, "-b:v", bitrate)
	}

	if crf, ok := options["crf"].(int); ok {
		args = append(args, "-crf", strconv.Itoa(crf))
	}

	if audioBitrate, ok := options["audio_bitrate"].(string); ok {
		args = append(args, "-b:a", audioBitrate)
	}

	if sampleRate, ok := options["sample_rate"].(int); ok {
		args = append(args, "-ar", strconv.Itoa(sampleRate))
	}

	if preset, ok := options["preset"].(string); ok {
		args = append(args, "-preset", preset)
	}

	if tune, ok := options["tune"].(string); ok {
		args = append(args, "-tune", tune)
	}

	args = append(args, "-y", outputPath)

	cmd := exec.Command("ffmpeg", args...)

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to get stderr pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start ffmpeg: %v", err)
	}

	go func() {
		if callback != nil {
			for i := 0; i <= 100; i += 10 {
				time.Sleep(100 * time.Millisecond)
				callback("Converting...", float64(i)/100.0)
			}
		}
	}()

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("ffmpeg conversion failed: %v", err)
	}

	checkCmd := exec.Command("ffprobe", "-v", "quiet", "-print_format", "json", "-show_format", outputPath)
	if _, err := checkCmd.Output(); err != nil {
		return fmt.Errorf("output file verification failed: %v", err)
	}

	duration, _ := videoStream["duration"].(string)
	codec, _ := videoStream["codec_name"].(string)

	fmt.Printf("Conversion completed: %s -> %s\n", inputPath, outputPath)
	fmt.Printf("Duration: %s, Codec: %s\n", duration, codec)

	return nil
}
