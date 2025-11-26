package ffmpeg

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"time"
)

type VideoConverter struct {
	inputPath  string
	outputPath string
	options    map[string]interface{}
	callback   func(string, float64)

	probeData   map[string]interface{}
	videoStream map[string]interface{}
	ffmpegArgs  []string
}

func NewVideoConverter(inputPath, outputPath string, options map[string]interface{}, callback func(string, float64)) *VideoConverter {
	if options == nil {
		options = make(map[string]interface{})
	}

	return &VideoConverter{
		inputPath:  inputPath,
		outputPath: outputPath,
		options:    options,
		callback:   callback,
		ffmpegArgs: []string{"-i", inputPath},
	}
}

func ProcessVideoConversion(inputPath string, outputPath string, options map[string]interface{}, callback func(string, float64)) error {
	converter := NewVideoConverter(inputPath, outputPath, options, callback)
	return converter.Execute()
}

func (vc *VideoConverter) GetVideoInfo() (map[string]interface{}, error) {
	if err := vc.analyzeSourceVideo(); err != nil {
		return nil, err
	}
	return vc.videoStream, nil
}

func (vc *VideoConverter) BuildFFmpegCommand() (*exec.Cmd, error) {
	if err := vc.validateInputs(); err != nil {
		return nil, err
	}

	if err := vc.analyzeSourceVideo(); err != nil {
		return nil, err
	}

	vc.prepareVideoOptions()
	vc.prepareAudioOptions()
	vc.prepareAdditionalOptions()
	vc.finalizeArguments()

	return exec.Command("ffmpeg", vc.ffmpegArgs...), nil
}

func (vc *VideoConverter) Execute() error {
	if err := vc.validateInputs(); err != nil {
		return err
	}

	if err := vc.analyzeSourceVideo(); err != nil {
		return err
	}

	vc.prepareVideoOptions()
	vc.prepareAudioOptions()
	vc.prepareAdditionalOptions()
	vc.finalizeArguments()

	if err := vc.runConversion(); err != nil {
		return err
	}

	if err := vc.verifyOutput(); err != nil {
		return err
	}

	vc.logResults()

	return nil
}

func (vc *VideoConverter) validateInputs() error {
	if vc.inputPath == "" {
		return fmt.Errorf("input path is required")
	}
	if vc.outputPath == "" {
		return fmt.Errorf("output path is required")
	}
	return nil
}

func (vc *VideoConverter) analyzeSourceVideo() error {
	probeCmd := exec.Command("ffprobe", "-v", "quiet", "-print_format", "json", "-show_format", "-show_streams", vc.inputPath)
	probeOutput, err := probeCmd.Output()
	if err != nil {
		return fmt.Errorf("ffprobe failed: %v", err)
	}

	if err := json.Unmarshal(probeOutput, &vc.probeData); err != nil {
		return fmt.Errorf("failed to parse ffprobe output: %v", err)
	}

	return vc.findVideoStream()
}

func (vc *VideoConverter) findVideoStream() error {
	streams, ok := vc.probeData["streams"].([]interface{})
	if !ok {
		return fmt.Errorf("invalid streams data")
	}

	for _, stream := range streams {
		if s, ok := stream.(map[string]interface{}); ok {
			if codecType, ok := s["codec_type"].(string); ok && codecType == "video" {
				vc.videoStream = s
				return nil
			}
		}
	}

	return fmt.Errorf("no video stream found")
}

func (vc *VideoConverter) prepareVideoOptions() {
	if width, ok := vc.options["width"].(int); ok {
		vc.ffmpegArgs = append(vc.ffmpegArgs, "-vf", fmt.Sprintf("scale=%d:-1", width))
	}

	if bitrate, ok := vc.options["bitrate"].(string); ok {
		vc.ffmpegArgs = append(vc.ffmpegArgs, "-b:v", bitrate)
	}

	if crf, ok := vc.options["crf"].(int); ok {
		vc.ffmpegArgs = append(vc.ffmpegArgs, "-crf", strconv.Itoa(crf))
	}
}

func (vc *VideoConverter) prepareAudioOptions() {
	if audioBitrate, ok := vc.options["audio_bitrate"].(string); ok {
		vc.ffmpegArgs = append(vc.ffmpegArgs, "-b:a", audioBitrate)
	}

	if sampleRate, ok := vc.options["sample_rate"].(int); ok {
		vc.ffmpegArgs = append(vc.ffmpegArgs, "-ar", strconv.Itoa(sampleRate))
	}
}

func (vc *VideoConverter) prepareAdditionalOptions() {
	if preset, ok := vc.options["preset"].(string); ok {
		vc.ffmpegArgs = append(vc.ffmpegArgs, "-preset", preset)
	}

	if tune, ok := vc.options["tune"].(string); ok {
		vc.ffmpegArgs = append(vc.ffmpegArgs, "-tune", tune)
	}
}

func (vc *VideoConverter) finalizeArguments() {
	vc.ffmpegArgs = append(vc.ffmpegArgs, "-y", vc.outputPath)
}

func (vc *VideoConverter) runConversion() error {
	cmd := exec.Command("ffmpeg", vc.ffmpegArgs...)

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to get stderr pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start ffmpeg: %v", err)
	}

	vc.monitorProgress(stderr)

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("ffmpeg conversion failed: %v", err)
	}

	return nil
}

func (vc *VideoConverter) monitorProgress(stderr interface{}) {
	if vc.callback != nil {
		go func() {
			for i := 0; i <= 100; i += 10 {
				time.Sleep(100 * time.Millisecond)
				vc.callback("Converting...", float64(i)/100.0)
			}
		}()
	}
}

func (vc *VideoConverter) verifyOutput() error {
	checkCmd := exec.Command("ffprobe", "-v", "quiet", "-print_format", "json", "-show_format", vc.outputPath)
	if _, err := checkCmd.Output(); err != nil {
		return fmt.Errorf("output file verification failed: %v", err)
	}
	return nil
}

func (vc *VideoConverter) logResults() {
	duration, _ := vc.videoStream["duration"].(string)
	codec, _ := vc.videoStream["codec_name"].(string)

	fmt.Printf("Conversion completed: %s -> %s\n", vc.inputPath, vc.outputPath)
	fmt.Printf("Duration: %s, Codec: %s\n", duration, codec)
}
