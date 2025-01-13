package api

import (
	"net/http"
)

// Handler returns an instance of httprouter.Router that handle APIs registered here
func (rt *_router) Handler() http.Handler {
	// User routes
	rt.router.PUT("/users/{username}", rt.setMyUserName)
	rt.router.POST("/users/{username}/photo", rt.setMyPhoto)
	rt.router.GET("/users/{username}/conversations", rt.getUserConversations)
	// Conversation routes
	rt.router.GET("/conversations/:conversationId/messages", rt.getConversationMessages)
	rt.router.POST("/conversations/:conversationId/messages", rt.sendMessage)

	// Reaction routes
	rt.router.POST("/conversations/:conversationId/messages/:messageId/reactions", rt.addReaction)
	rt.router.DELETE("/conversations/:conversationId/messages/:messageId/reactions", rt.removeReaction)

	// Register routes
	rt.router.GET("/", rt.getHelloWorld)
	rt.router.GET("/context", rt.wrap(rt.getContextReply))

	// Special routes
	rt.router.GET("/liveness", rt.liveness)

	return rt.router
}
