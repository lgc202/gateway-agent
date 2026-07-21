// Package chat 提供 Chat 和 Message HTTP 接口
package chat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/gin-gonic/gin"

	"github.com/lgc202/gateway-agent/internal/apiserver/handler/response"
	chatservice "github.com/lgc202/gateway-agent/internal/apiserver/service/chat"
	"github.com/lgc202/gateway-agent/internal/pkg/errorsx"
)

const (
	defaultMessageLimit = 50
	maxRequestBodyBytes = 384 * 1024
)

type appendMessageRequest struct {
	Content string `json:"content"`
}

type chatResponse struct {
	ID        uint64    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type messageResponse struct {
	ID        uint64                  `json:"id"`
	ChatID    uint64                  `json:"chat_id"`
	Role      chatservice.MessageRole `json:"role"`
	Content   string                  `json:"content"`
	CreatedAt time.Time               `json:"created_at"`
}

type messagePageResponse struct {
	Items       []messageResponse `json:"items"`
	NextAfterID uint64            `json:"next_after_id"`
}

// Handler 处理 Chat 和 Message HTTP 请求
type Handler struct {
	service *chatservice.Service
}

// New 创建 Chat Handler
func New(service *chatservice.Service) *Handler {
	return &Handler{service: service}
}

// Register 注册 Chat 和 Message 路由
func (h *Handler) Register(router *gin.RouterGroup) {
	router.POST("/chats", h.createChat)
	router.GET("/chats/:chat_id", h.getChat)
	router.POST("/chats/:chat_id/messages", h.appendMessage)
	router.GET("/chats/:chat_id/messages", h.listMessages)
}

func (h *Handler) createChat(ctx *gin.Context) {
	chat, err := h.service.CreateChat(ctx.Request.Context())
	if err != nil {
		response.WriteError(ctx, err)
		return
	}

	response.WriteSuccess(ctx, http.StatusCreated, toChatResponse(chat))
}

func (h *Handler) getChat(ctx *gin.Context) {
	chatID, err := parsePositiveUint64(ctx.Param("chat_id"), "chat_id")
	if err != nil {
		response.WriteError(ctx, err)
		return
	}

	chat, err := h.service.GetChat(ctx.Request.Context(), chatID)
	if err != nil {
		response.WriteError(ctx, err)
		return
	}

	response.WriteSuccess(ctx, http.StatusOK, toChatResponse(chat))
}

func (h *Handler) appendMessage(ctx *gin.Context) {
	chatID, err := parsePositiveUint64(ctx.Param("chat_id"), "chat_id")
	if err != nil {
		response.WriteError(ctx, err)
		return
	}
	if !hasJSONContentType(ctx.Request.Header.Values("Content-Type")) {
		response.WriteError(ctx, errorsx.NewUser(errorsx.CodeInvalidRequest, "Content-Type 必须是 application/json"))
		return
	}

	ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, maxRequestBodyBytes)
	req, err := decodeAppendMessageRequest(ctx.Request.Body)
	if err != nil {
		response.WriteError(ctx, errorsx.NewUser(errorsx.CodeInvalidRequest, "请求体格式错误"))
		return
	}

	message, err := h.service.AppendUserMessage(ctx.Request.Context(), chatID, req.Content)
	if err != nil {
		response.WriteError(ctx, err)
		return
	}

	response.WriteSuccess(ctx, http.StatusCreated, toMessageResponse(message))
}

func (h *Handler) listMessages(ctx *gin.Context) {
	chatID, err := parsePositiveUint64(ctx.Param("chat_id"), "chat_id")
	if err != nil {
		response.WriteError(ctx, err)
		return
	}

	afterID, err := parseOptionalUint64(ctx.Query("after_id"), "after_id", 0)
	if err != nil {
		response.WriteError(ctx, err)
		return
	}
	limit, err := parseLimit(ctx.Query("limit"))
	if err != nil {
		response.WriteError(ctx, err)
		return
	}

	messages, err := h.service.ListMessages(ctx.Request.Context(), chatID, afterID, limit)
	if err != nil {
		response.WriteError(ctx, err)
		return
	}

	items := make([]messageResponse, 0, len(messages))
	nextAfterID := afterID
	for _, message := range messages {
		items = append(items, toMessageResponse(message))
		nextAfterID = message.ID
	}

	response.WriteSuccess(ctx, http.StatusOK, messagePageResponse{
		Items:       items,
		NextAfterID: nextAfterID,
	})
}

func hasJSONContentType(values []string) bool {
	if len(values) != 1 {
		return false
	}

	mediaType, parameters, err := mime.ParseMediaType(values[0])
	if err != nil || mediaType != "application/json" {
		return false
	}
	for name, value := range parameters {
		if !strings.EqualFold(name, "charset") || !strings.EqualFold(value, "utf-8") {
			return false
		}
	}
	return true
}

func decodeAppendMessageRequest(body io.Reader) (appendMessageRequest, error) {
	data, err := io.ReadAll(body)
	if err != nil {
		return appendMessageRequest{}, err
	}
	if !utf8.Valid(data) {
		return appendMessageRequest{}, fmt.Errorf("请求体不是合法 UTF-8")
	}

	decoder := json.NewDecoder(bytes.NewReader(data))
	start, err := decoder.Token()
	if err != nil || start != json.Delim('{') {
		return appendMessageRequest{}, fmt.Errorf("请求体顶层必须是 JSON 对象")
	}

	var req appendMessageRequest
	contentSeen := false
	for decoder.More() {
		keyToken, err := decoder.Token()
		if err != nil {
			return appendMessageRequest{}, err
		}
		key, ok := keyToken.(string)
		if !ok || key != "content" || contentSeen {
			return appendMessageRequest{}, fmt.Errorf("请求体只允许唯一的 content 字段")
		}
		contentSeen = true
		var rawContent json.RawMessage
		if err := decoder.Decode(&rawContent); err != nil {
			return appendMessageRequest{}, fmt.Errorf("content 必须是 JSON 字符串")
		}
		if !validJSONStringUnicode(rawContent) {
			return appendMessageRequest{}, fmt.Errorf("content 包含无效的 Unicode 代理项")
		}
		if err := json.Unmarshal(rawContent, &req.Content); err != nil {
			return appendMessageRequest{}, fmt.Errorf("content 必须是 JSON 字符串")
		}
	}

	end, err := decoder.Token()
	if err != nil || end != json.Delim('}') || !contentSeen {
		return appendMessageRequest{}, fmt.Errorf("请求体必须包含 content 字符串字段")
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return appendMessageRequest{}, fmt.Errorf("请求体只能包含一个 JSON 对象")
	}

	return req, nil
}

func validJSONStringUnicode(value []byte) bool {
	if len(value) < 2 || value[0] != '"' || value[len(value)-1] != '"' {
		return false
	}

	for i := 1; i < len(value)-1; i++ {
		if value[i] != '\\' {
			continue
		}
		i++
		if i >= len(value)-1 || value[i] != 'u' {
			continue
		}
		if i+4 >= len(value) {
			return false
		}

		first, err := strconv.ParseUint(string(value[i+1:i+5]), 16, 16)
		if err != nil {
			return false
		}
		i += 4
		firstRune := rune(first)
		if !utf16.IsSurrogate(firstRune) {
			continue
		}
		if firstRune < 0xD800 || firstRune > 0xDBFF {
			return false
		}
		if i+6 >= len(value) || value[i+1] != '\\' || value[i+2] != 'u' {
			return false
		}

		second, err := strconv.ParseUint(string(value[i+3:i+7]), 16, 16)
		if err != nil || utf16.DecodeRune(firstRune, rune(second)) == utf8.RuneError {
			return false
		}
		i += 6
	}

	return true
}

func parsePositiveUint64(value, field string) (uint64, error) {
	parsed, err := strconv.ParseUint(value, 10, 64)
	if err != nil || parsed == 0 {
		return 0, errorsx.NewUser(errorsx.CodeInvalidRequest, fmt.Sprintf("%s 必须是大于 0 的整数", field))
	}
	return parsed, nil
}

func parseOptionalUint64(value, field string, defaultValue uint64) (uint64, error) {
	if value == "" {
		return defaultValue, nil
	}
	parsed, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0, errorsx.NewUser(errorsx.CodeInvalidRequest, fmt.Sprintf("%s 必须是大于等于 0 的整数", field))
	}
	return parsed, nil
}

func parseLimit(value string) (int32, error) {
	if value == "" {
		return defaultMessageLimit, nil
	}
	parsed, err := strconv.ParseInt(value, 10, 32)
	if err != nil || parsed < 1 || parsed > 200 {
		return 0, errorsx.NewUser(errorsx.CodeInvalidRequest, "limit 必须在 1 到 200 之间")
	}
	return int32(parsed), nil
}

func toChatResponse(chat chatservice.Chat) chatResponse {
	return chatResponse{
		ID:        chat.ID,
		CreatedAt: chat.CreatedAt,
		UpdatedAt: chat.UpdatedAt,
	}
}

func toMessageResponse(message chatservice.Message) messageResponse {
	return messageResponse{
		ID:        message.ID,
		ChatID:    message.ChatID,
		Role:      message.Role,
		Content:   message.Content,
		CreatedAt: message.CreatedAt,
	}
}
