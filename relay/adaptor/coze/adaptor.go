package coze

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/relay/adaptor"
	"github.com/songquanpeng/one-api/relay/adaptor/coze/constant/version"
	"github.com/songquanpeng/one-api/relay/adaptor/openai"
	"github.com/songquanpeng/one-api/relay/meta"
	"github.com/songquanpeng/one-api/relay/model"
)

type Adaptor struct {
	meta *meta.Meta
}

func (a *Adaptor) Init(meta *meta.Meta) {
	a.meta = meta
}

func (a *Adaptor) GetRequestURL(meta *meta.Meta) (string, error) {
	apiVersion := getApiVersion()

	if apiVersion == version.V3 {
		// 如果指定了会话ID，则添加到URL参数中
		if conversationId := a.meta.Config.ConversationID; conversationId != "" {
			return fmt.Sprintf("%s/v3/chat?conversation_id=%s", meta.BaseURL, conversationId), nil
		}
		return fmt.Sprintf("%s/v3/chat", meta.BaseURL), nil
	}

	// 默认使用V2版本
	return fmt.Sprintf("%s/open_api/v2/chat", meta.BaseURL), nil
}

func (a *Adaptor) SetupRequestHeader(c *gin.Context, req *http.Request, meta *meta.Meta) error {
	adaptor.SetupCommonRequestHeader(c, req, meta)
	req.Header.Set("Authorization", "Bearer "+meta.APIKey)
	return nil
}

func (a *Adaptor) ConvertRequest(c *gin.Context, relayMode int, request *model.GeneralOpenAIRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}

	// 设置用户ID
	request.User = a.meta.Config.UserID

	// 根据环境变量决定使用的API版本
	apiVersion := getApiVersion()

	if apiVersion == version.V3 {
		return ConvertRequestV3(*request), nil
	}

	// 默认使用V2
	return ConvertRequest(*request), nil
}

func (a *Adaptor) ConvertImageRequest(request *model.ImageRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}
	return request, nil
}

func (a *Adaptor) DoRequest(c *gin.Context, meta *meta.Meta, requestBody io.Reader) (*http.Response, error) {
	return adaptor.DoRequestHelper(a, c, meta, requestBody)
}

func (a *Adaptor) DoResponse(c *gin.Context, resp *http.Response, meta *meta.Meta) (usage *model.Usage, err *model.ErrorWithStatusCode) {
	var responseText *string

	// 根据环境变量决定使用的API版本
	apiVersion := getApiVersion()

	if apiVersion == version.V3 {
		if meta.IsStream {
			err, responseText = StreamHandlerV3(c, resp)
		} else {
			err, responseText = HandlerV3(c, resp, meta.PromptTokens, meta.ActualModelName)
		}
	} else {
		// 默认使用V2
		if meta.IsStream {
			err, responseText = StreamHandler(c, resp)
		} else {
			err, responseText = Handler(c, resp, meta.PromptTokens, meta.ActualModelName)
		}
	}

	if responseText != nil {
		usage = openai.ResponseText2Usage(*responseText, meta.ActualModelName, meta.PromptTokens)
	} else {
		usage = &model.Usage{}
	}
	usage.PromptTokens = meta.PromptTokens
	usage.TotalTokens = usage.PromptTokens + usage.CompletionTokens
	return
}

func (a *Adaptor) GetModelList() []string {
	return ModelList
}

func (a *Adaptor) GetChannelName() string {
	return "coze"
}

// getApiVersion 根据环境变量决定使用的API版本
func getApiVersion() string {
	// 环境变量名称：COZE_API_VERSION
	apiVersion := os.Getenv("COZE_API_VERSION")
	if apiVersion == version.V2 {
		return version.V2
	}
	return version.V3
}
