package openai

import (
	adaptercommon "chat/adapter/common"
	"chat/globals"
	"chat/utils"
	"fmt"
	"net/url"
	"strings"
	"time"
)

type VideoRequest struct {
	Prompt         string  `json:"prompt"`
	Model          string  `json:"model,omitempty"`
	Seconds        *string `json:"seconds,omitempty"`
	Size           *string `json:"size,omitempty"`
	InputReference *string `json:"input_reference,omitempty"`
}

type VideoJob struct {
	CompletedAt        *int64  `json:"completed_at,omitempty"`
	CreatedAt          int64   `json:"created_at"`
	ExpiresAt          *int64  `json:"expires_at,omitempty"`
	Id                 string  `json:"id"`
	Model              string  `json:"model"`
	Object             string  `json:"object"`
	Progress           *int    `json:"progress,omitempty"`
	Prompt             string  `json:"prompt"`
	RemixedFromVideoId *string `json:"remixed_from_video_id,omitempty"`
	Seconds            string  `json:"seconds"`
	Size               string  `json:"size"`
	Status             string  `json:"status"`
	Url                string  `json:"url,omitempty"`
	Error              *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type haoduomiGenerateResponse struct {
	TaskID string `json:"task_id"`
	Data   struct {
		TaskID        string   `json:"task_id"`
		LegacyTaskID  string   `json:"任务id"`
		LegacyTaskIDs []string `json:"任务ids"`
	} `json:"data"`
}

type haoduomiTaskStatus struct {
	TaskID    string `json:"task_id"`
	State     string `json:"state"`
	IsFinal   bool   `json:"is_final"`
	Progress  *int   `json:"progress"`
	ResultURL string `json:"result_url"`
	Message   string `json:"message"`
	Error     string `json:"error"`
}

type haoduomiTaskStatusResponse struct {
	TaskID    string             `json:"task_id"`
	State     string             `json:"state"`
	IsFinal   bool               `json:"is_final"`
	Progress  *int               `json:"progress"`
	ResultURL string             `json:"result_url"`
	Message   string             `json:"message"`
	Error     string             `json:"error"`
	Data      haoduomiTaskStatus `json:"data"`
}

func (c *ChatInstance) isHaoduomiVideoEndpoint() bool {
	endpoint := strings.ToLower(c.GetEndpoint())
	return strings.Contains(endpoint, "lk888.ai") || strings.Contains(endpoint, "haoduomi.ai")
}

func (c *ChatInstance) getHaoduomiAPIBase() string {
	endpoint := strings.TrimRight(c.GetEndpoint(), "/")
	endpoint = strings.TrimSuffix(endpoint, "/api")
	endpoint = strings.TrimSuffix(endpoint, "/v1")
	return endpoint + "/api"
}

func (c *ChatInstance) createHaoduomiVideoRequest(props *adaptercommon.VideoProps, hook globals.Hook) error {
	if props.InputReference == nil || !strings.HasPrefix(strings.ToLower(strings.TrimSpace(*props.InputReference)), "http") {
		return fmt.Errorf("haoduomi video requires one public first-frame image URL")
	}

	duration, aspectRatio, resolution := "5", "16:9", "720p"
	if props.Seconds != nil {
		duration = *props.Seconds
	}
	if props.AspectRatio != nil {
		aspectRatio = *props.AspectRatio
	}
	if props.Resolution != nil {
		resolution = *props.Resolution
	}
	body := map[string]interface{}{
		"model":  props.Model,
		"prompt": props.Prompt,
		"params": map[string]interface{}{
			"images":       []string{*props.InputReference},
			"aspect_ratio": aspectRatio,
			"resolution":   resolution,
			"duration":     duration,
		},
	}

	res, err := utils.Post(c.getHaoduomiAPIBase()+"/v1/media/generate", c.GetHeader(), body, props.Proxy)
	if err != nil || res == nil {
		if err != nil {
			return fmt.Errorf("haoduomi video error: %s", err.Error())
		}
		return fmt.Errorf("haoduomi video error: empty response")
	}
	created := utils.MapToStruct[haoduomiGenerateResponse](res)
	if created == nil {
		return fmt.Errorf("haoduomi video error: cannot parse response")
	}
	taskID := created.TaskID
	if taskID == "" {
		taskID = created.Data.TaskID
	}
	if taskID == "" {
		taskID = created.Data.LegacyTaskID
	}
	if taskID == "" && len(created.Data.LegacyTaskIDs) > 0 {
		taskID = created.Data.LegacyTaskIDs[0]
	}
	if taskID == "" {
		return fmt.Errorf("haoduomi video error: response has no task_id")
	}

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	deadline := time.After(2 * time.Hour)
	begin, lastProgress := false, -1
	for {
		select {
		case <-ticker.C:
			endpoint := fmt.Sprintf("%s/v1/skills/task-status?task_id=%s", c.getHaoduomiAPIBase(), url.QueryEscape(taskID))
			data, getErr := utils.Get(endpoint, c.GetHeader(), props.Proxy)
			if getErr != nil || data == nil {
				continue
			}
			response := utils.MapToStruct[haoduomiTaskStatusResponse](data)
			if response == nil {
				continue
			}
			status := haoduomiTaskStatus{TaskID: response.TaskID, State: response.State, IsFinal: response.IsFinal, Progress: response.Progress, ResultURL: response.ResultURL, Message: response.Message, Error: response.Error}
			if response.Data.TaskID != "" || response.Data.State != "" {
				status = response.Data
			}
			if status.Progress != nil {
				if !begin {
					begin = true
					if err := hook(&globals.Chunk{Content: "```progress\n"}); err != nil {
						return err
					}
				}
				if *status.Progress != lastProgress {
					if err := hook(&globals.Chunk{Content: fmt.Sprintf("%d\n", *status.Progress)}); err != nil {
						return err
					}
					lastProgress = *status.Progress
				}
			}
			if !status.IsFinal {
				continue
			}
			if begin {
				if err := hook(&globals.Chunk{Content: "```\n"}); err != nil {
					return err
				}
			}
			if strings.EqualFold(status.State, "success") && status.ResultURL != "" {
				now := time.Now().Unix()
				job := VideoJob{Id: taskID, Model: props.Model, Object: "video", Prompt: props.Prompt, Seconds: duration, Size: resolution, Status: "completed", CompletedAt: &now, CreatedAt: now, Url: status.ResultURL}
				return hook(&globals.Chunk{Content: utils.Marshal(job)})
			}
			message := status.Message
			if message == "" {
				message = status.Error
			}
			if message == "" {
				message = "upstream task failed"
			}
			return fmt.Errorf("haoduomi video job failed: %s", message)
		case <-deadline:
			if begin {
				_ = hook(&globals.Chunk{Content: "```\n"})
			}
			return fmt.Errorf("haoduomi video job timeout")
		}
	}
}

func (c *ChatInstance) getVideoCreateEndpoint() string {
	return fmt.Sprintf("%s/v1/videos", c.GetEndpoint())
}

func (c *ChatInstance) getVideoQueryEndpoint(id string) string {
	return fmt.Sprintf("%s/v1/videos/%s", c.GetEndpoint(), id)
}

func (c *ChatInstance) CreateVideoRequest(props *adaptercommon.VideoProps, hook globals.Hook) error {
	if c.isHaoduomiVideoEndpoint() {
		return c.createHaoduomiVideoRequest(props, hook)
	}
	body := VideoRequest{
		Prompt:         props.Prompt,
		Model:          props.Model,
		Seconds:        props.Seconds,
		Size:           props.Size,
		InputReference: props.InputReference,
	}

	res, err := utils.Post(c.getVideoCreateEndpoint(), c.GetHeader(), body, props.Proxy)
	if err != nil || res == nil {
		if err != nil {
			return fmt.Errorf("openai video error: %s", err.Error())
		}
		return fmt.Errorf("openai video error: empty response")
	}

	job := utils.MapToStruct[VideoJob](res)
	if job == nil {
		return fmt.Errorf("openai video error: cannot parse response")
	}
	if job.Error != nil && (job.Error.Message != "") {
		return fmt.Errorf("openai video error: %s", job.Error.Message)
	}

	const maxTimeout = 30 * time.Minute
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	deadline := time.After(maxTimeout)

	var begin bool
	var lastProgress int = -1

	for {
		if job.Status == "completed" {
			if begin {
				if err := hook(&globals.Chunk{Content: "```\n"}); err != nil {
					return err
				}
			}
			return hook(&globals.Chunk{Content: utils.Marshal(job)})
		}
		if job.Status == "failed" {
			if begin {
				if err := hook(&globals.Chunk{Content: "```\n"}); err != nil {
					return err
				}
			}
			if job.Error != nil && job.Error.Message != "" {
				return fmt.Errorf("openai video job failed: %s", job.Error.Message)
			}
			return fmt.Errorf("openai video job failed")
		}

		select {
		case <-ticker.C:
			if job.Id == "" {
				return hook(&globals.Chunk{Content: utils.Marshal(job)})
			}
			data, gErr := utils.Get(c.getVideoQueryEndpoint(job.Id), c.GetHeader(), props.Proxy)
			if gErr != nil || data == nil {
				continue
			}
			if j := utils.MapToStruct[VideoJob](data); j != nil {
				job = j
			}

			progress := 0
			if job.Progress != nil {
				progress = *job.Progress
			}

			if !begin {
				begin = true
				if err := hook(&globals.Chunk{Content: "```progress\n"}); err != nil {
					return err
				}
			}

			if progress != lastProgress {
				if err := hook(&globals.Chunk{Content: fmt.Sprintf("%d\n", progress)}); err != nil {
					return err
				}
				lastProgress = progress
			}
		case <-deadline:
			if begin {
				if err := hook(&globals.Chunk{Content: "```\n"}); err != nil {
					return err
				}
			}
			return fmt.Errorf("openai video job timeout")
		}
	}
}
