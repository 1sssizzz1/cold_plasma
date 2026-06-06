package handler

import (
	"net/http"
	"strings"

	"cold-plasma-server/internal/ai"
	"cold-plasma-server/internal/service"
	"cold-plasma-server/internal/transport/middleware"

	"github.com/gin-gonic/gin"
)

type ChatHandler struct {
	chat *service.ChatService
}

func NewChatHandler(chat *service.ChatService) *ChatHandler {
	return &ChatHandler{chat: chat}
}

type chatReq struct {
	SessionID string           `json:"session_id"`
	Message   string           `json:"message"`
	UserName  string           `json:"user_name"`
	UserEmail string           `json:"user_email"`
	UserPhone string           `json:"user_phone"`
	History   []chatHistoryMsg `json:"history"`
}

type chatHistoryMsg struct {
	Role string `json:"role"`
	Text string `json:"text"`
}

type chatSessionReq struct {
	SessionID string `json:"session_id"`
	UserName  string `json:"user_name"`
	UserEmail string `json:"user_email"`
	UserPhone string `json:"user_phone"`
}

func (h *ChatHandler) OpenSession(c *gin.Context) {
	var req chatSessionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "РќРµРєРѕСЂСЂРµРєС‚РЅС‹Р№ JSON")
		return
	}
	if err := h.chat.OpenSession(c.Request.Context(), chatActorFromRequest(c, req.UserName, req.UserEmail, req.UserPhone), req.SessionID); err != nil {
		handleServiceErr(c, err)
		return
	}
	ok(c, gin.H{"session_id": strings.TrimSpace(req.SessionID)})
}

func (h *ChatHandler) CloseSession(c *gin.Context) {
	var req chatSessionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "РќРµРєРѕСЂСЂРµРєС‚РЅС‹Р№ JSON")
		return
	}
	if err := h.chat.CloseSession(c.Request.Context(), req.SessionID); err != nil {
		handleServiceErr(c, err)
		return
	}
	ok(c, gin.H{"session_id": strings.TrimSpace(req.SessionID)})
}

func (h *ChatHandler) Ask(c *gin.Context) {
	var req chatReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "Некорректный JSON")
		return
	}

	answer, err := h.chat.Ask(c.Request.Context(), chatActorFromRequest(c, req.UserName, req.UserEmail, req.UserPhone), req.SessionID, req.Message, toAIHistory(req.History))
	if err != nil {
		handleServiceErr(c, err)
		return
	}
	ok(c, gin.H{"text": answer.Text, "model": answer.Model, "intent": answer.Intent})
}

func chatActorFromRequest(c *gin.Context, name, email, phone string) service.ChatActor {
	actor := service.ChatActor{
		Name:  strings.TrimSpace(name),
		Email: strings.TrimSpace(email),
		Phone: strings.TrimSpace(phone),
	}
	if u, ok := middleware.GetUser(c); ok {
		actor.UserID = &u.UserID
		if actor.Name == "" {
			actor.Name = u.Name
		}
		actor.Email = u.Email
	}
	return actor
}

func toAIHistory(items []chatHistoryMsg) []ai.ChatMessage {
	out := make([]ai.ChatMessage, 0, len(items))
	for _, item := range items {
		out = append(out, ai.ChatMessage{Role: item.Role, Text: item.Text})
	}
	return out
}
