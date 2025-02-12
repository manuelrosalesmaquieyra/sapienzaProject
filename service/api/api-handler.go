package api

import (
	"net/http"

	"github.com/gorilla/handlers"
)

// Handler returns an instance of httprouter.Router that handle APIs registered here
func (rt *_router) Handler() http.Handler {

	// Login route
	rt.router.POST("/session", rt.doLogin)

	// User routes
	rt.router.PUT("/users/:username", rt.setMyUserName)
	rt.router.POST("/users/:username/photo", rt.setMyPhoto)
	rt.router.GET("/users/:username", rt.getUser)

	// Group routes
	rt.router.POST("/groups", rt.createGroup)
	rt.router.POST("/groups/:group_id", rt.updateGroupName)
	rt.router.POST("/groups/:group_id/photo", rt.updateGroupPhoto)
	rt.router.POST("/groups/:group_id/leave", rt.leaveGroup)

	// Conversation routes
	rt.router.GET("/conversations/:conversationId/messages", rt.getConversationMessages)
	rt.router.GET("/users/:username/conversations", rt.getUserConversations)
	rt.router.POST("/conversations", rt.createConversation)
	rt.router.GET("/conversations/:conversationId", rt.getConversation)
	rt.router.GET("/conversations/:conversationId/details", rt.getConversationDetails)

	// Reaction routes
	rt.router.POST("/conversations/:conversationId/messages/:messageId/reactions", rt.addReaction)
	rt.router.DELETE("/conversations/:conversationId/messages/:messageId/reactions", rt.removeReaction)

	// Message routes
	rt.router.POST("/conversations/:conversationId/messages", rt.sendMessage)
	rt.router.DELETE("/conversations/:conversationId/messages/:messageId", rt.deleteMessage)
	rt.router.POST("/conversations/:conversationId/messages/:messageId", rt.forwardMessage)

	// Register routes
	// rt.router.GET("/", rt.getHelloWorld)
	// rt.router.GET("/context", rt.wrap(rt.getContextReply))

	// Special routes
	// rt.router.GET("/liveness", rt.liveness)

	// Create CORS handler
	corsMiddleware := handlers.CORS(
		handlers.AllowedOrigins([]string{"http://localhost:5173"}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
		handlers.AllowCredentials(),
	)

	// Apply CORS middleware to router
	return corsMiddleware(rt.router)
}
