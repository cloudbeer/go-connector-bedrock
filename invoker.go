package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	openai "github.com/sashabaranov/go-openai"
)


const defaultRegion = "us-east-1"

/**
 Model list: https://docs.aws.amazon.com/bedrock/latest/userguide/model-ids.html
**/
var supportedModels = [...]string{
	"anthropic.claude-3-sonnet-20240229-v1:0",
	"anthropic.claude-3-haiku-20240307-v1:0",
	"anthropic.claude-3-5-sonnet-20240620-v1:0",
	"anthropic.claude-3-opus-20240229-v1:0",
}

var brc *bedrockruntime.Client

func init() {
	fmt.Println("init...")
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = defaultRegion
	}

	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(region))
	if err != nil {
		log.Fatal(err)
	}

	brc = bedrockruntime.NewFromConfig(cfg)
}

func chooseModel(model string) *string {
	for _, m := range supportedModels {
		if m == model {
			return &model
		}
	}
	return &supportedModels[0]
}

// transfer ChatCompletionRequest to ConverseInput
func formatStreamInput(chatReq openai.ChatCompletionRequest) bedrockruntime.ConverseStreamInput {
	// var messages []types.Message
	chatMessages := chatReq.Messages
	var systemMessages []openai.ChatCompletionMessage
	var qaMessages []openai.ChatCompletionMessage
	var converseMessages []types.Message
	var converseSystem []types.SystemContentBlock

	for _, msg := range chatMessages {
		if msg.Role == "system" {
			systemMessages = append(systemMessages, msg)
		} else {
			qaMessages = append(qaMessages, msg)
		}
	}

	for _, msg := range qaMessages {
		// TODO: 这里要补全信息
		converseMessages = append(converseMessages, types.Message{
			Role: types.ConversationRole(msg.Role),
			Content: []types.ContentBlock{
				&types.ContentBlockMemberText{
					Value: msg.Content,
				},
			},
		})
	}
	for _, msg := range systemMessages {
		converseSystem = append(converseSystem, types.SystemContentBlock(&types.SystemContentBlockMemberText{
			Value: msg.Content,
		}))
	}

	return bedrockruntime.ConverseStreamInput{
		ModelId:  chooseModel(chatReq.Model),
		Messages: converseMessages,
		System:   converseSystem,
	}
}

// transfer ChatCompletionRequest to ConverseInput
func format(chatReq openai.ChatCompletionRequest) bedrockruntime.ConverseInput {
	// var messages []types.Message
	chatMessages := chatReq.Messages
	var systemMessages []openai.ChatCompletionMessage
	var qaMessages []openai.ChatCompletionMessage
	var converseMessages []types.Message
	var converseSystem []types.SystemContentBlock

	for _, msg := range chatMessages {
		if msg.Role == "system" {
			systemMessages = append(systemMessages, msg)
		} else {
			qaMessages = append(qaMessages, msg)
		}
	}

	for _, msg := range qaMessages {
		// TODO: 这里要补全信息
		converseMessages = append(converseMessages, types.Message{
			Role: types.ConversationRole(msg.Role),
			Content: []types.ContentBlock{
				&types.ContentBlockMemberText{
					Value: msg.Content,
				},
			},
		})
	}
	for _, msg := range systemMessages {
		converseSystem = append(converseSystem, types.SystemContentBlock(&types.SystemContentBlockMemberText{
			Value: msg.Content,
		}))
	}

	return bedrockruntime.ConverseInput{
		ModelId:  chooseModel(chatReq.Model),
		Messages: converseMessages,
		System:   converseSystem,
	}
}


func converse(brc *bedrockruntime.Client, converseInput bedrockruntime.ConverseInput) openai.ChatCompletionResponse {
	output, err := brc.Converse(context.Background(), &converseInput)
	// fmt.Println(output.GetStream().Events())
	if err != nil {
		fmt.Println(err)
	}

	usage := output.Usage

	// tmpJ, _ := json.Marshal(output)
	// fmt.Print(string(tmpJ), usage)

	response, _ := output.Output.(*types.ConverseOutputMemberMessage)

	responseContentBlock := response.Value.Content[0]
	text, _ := responseContentBlock.(*types.ContentBlockMemberText)

	return openai.ChatCompletionResponse{
		Choices: []openai.ChatCompletionChoice{
			{
				Message: openai.ChatCompletionMessage{
					Role:    "assistant",
					Content: text.Value,
				},
				FinishReason: openai.FinishReason(output.StopReason),
			},
		},
		Usage: openai.Usage{
			PromptTokens:     int(*usage.InputTokens),
			CompletionTokens: int(*usage.OutputTokens),
			TotalTokens:      int(*usage.TotalTokens),
		},
	}

}

// type StreamingOutputHandler func(ctx context.Context, part string) error

func convereStream(
	brc *bedrockruntime.Client,
	w http.ResponseWriter,
	converseInput bedrockruntime.ConverseStreamInput) {
	output, err := brc.ConverseStream(context.Background(), &converseInput)
	if err != nil {
		fmt.Println(err)
	}
	// msg := openai.ChatCompletionStreamChoice{}
	for event := range output.GetStream().Events() {

		// xxx, _ := json.Marshal(event)
		// fmt.Println(string(xxx))

		switch v := event.(type) {
		case *types.ConverseStreamOutputMemberMessageStart:
			// fmt.Println("ConverseStreamOutputMemberMessageStart")

		case *types.ConverseStreamOutputMemberContentBlockDelta:
			textResponse := v.Value.Delta.(*types.ContentBlockDeltaMemberText)
			response := openai.ChatCompletionStreamResponse{
				ID:      "",
				Object:  "",
				Created: 0,
				Model:   *converseInput.ModelId,
				Choices: []openai.ChatCompletionStreamChoice{
					{
						Delta: openai.ChatCompletionStreamChoiceDelta{
							Role:    "assistant",
							Content: textResponse.Value,
						},
						// FinishReason: openai.FinishReason(v.StopReason),
					},
				},
			}
			jsonResponse, err := json.Marshal(response)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.Write([]byte("data:"))
			w.Write(jsonResponse)
			w.Write([]byte("\n\n"))
			w.(http.Flusher).Flush()
		case *types.ConverseStreamOutputMemberMessageStop:

			w.Write([]byte("data:"))
			w.Write([]byte("[DONE]"))
			w.Write([]byte("\n\n"))
			w.(http.Flusher).Flush()

		case *types.UnknownUnionMember:
			fmt.Println("unknown tag:", v.Tag)
		}
	}

}
