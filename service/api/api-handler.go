package api

import (
	"net/http"
)

// Handler returns an instance of httprouter.Router that handle APIs registered here
func (rt *_router) Handler() http.Handler {

	// Login route
	rt.router.POST("/session", rt.doLogin)

	// User routes
	rt.router.PUT("/users/{username}", rt.setMyUserName)
	rt.router.POST("/users/{username}/photo", rt.setMyPhoto)

	// Conversation routes
	rt.router.GET("/conversations/:conversationId/messages", rt.getConversationMessages)
	rt.router.GET("/users/{username}/conversations", rt.getUserConversations)

	// Reaction routes
	rt.router.POST("/conversations/:conversationId/messages/:messageId/reactions", rt.addReaction)
	rt.router.DELETE("/conversations/:conversationId/messages/:messageId/reactions", rt.removeReaction)

	// Message routes
	rt.router.POST("/conversations/:conversationId/messages", rt.sendMessage)
	rt.router.DELETE("/conversations/:conversationId/messages/:messageId", rt.deleteMessage)
	rt.router.POST("/conversations/:conversationId/messages/:messageId", rt.forwardMessage)

	// Group routes
	rt.router.POST("/conversations/groups", rt.createGroup)
	rt.router.POST("/conversations/groups/:group_id", rt.updateGroupName)
	rt.router.POST("/conversations/groups/:group_id/photo", rt.updateGroupPhoto)
	rt.router.POST("/conversations/groups/:group_id/leave", rt.leaveGroup)

	// Register routes
	// rt.router.GET("/", rt.getHelloWorld)
	// rt.router.GET("/context", rt.wrap(rt.getContextReply))

	// Special routes
	//rt.router.GET("/liveness", rt.liveness)

	return rt.router
}
