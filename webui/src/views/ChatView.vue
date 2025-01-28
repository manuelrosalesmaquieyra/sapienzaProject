<script setup>
import { ref, onMounted, computed, nextTick, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import MainLayout from '../layouts/MainLayout.vue'
import { api } from '../services/api'

const route = useRoute()
const router = useRouter()
const messages = ref([])
const newMessage = ref('')
const loading = ref(false)
const error = ref('')
const conversationDetails = computed(() => route.params.conversation || { participants: [] })

// Get conversation ID from route
const conversationId = ref(route.params.conversation_id)

// Add this computed property
const currentUsername = computed(() => localStorage.getItem('username'))

// Add ref for messages container
const messagesContainer = ref(null)

// Add this computed property to get the other participant's name
const otherParticipant = computed(() => {
    const participants = conversationDetails.value?.participants || []
    return participants.find(p => p !== currentUsername.value) || 'Unknown User'
})

// Fetch messages
const fetchMessages = async () => {
    try {
        loading.value = true
        const data = await api.getConversationMessages(conversationId.value)
        messages.value = data.messages
        // Add this debug log
        console.log('Messages with reactions:', messages.value)
    } catch (err) {
        error.value = 'Failed to load messages'
        console.error('Error:', err)
    } finally {
        loading.value = false
    }
}

// Send message
const sendMessage = async () => {
    if (!newMessage.value.trim()) return
    
    try {
        const response = await api.sendMessage(conversationId.value, newMessage.value)
        // Add new message to the list instead of fetching all messages
        messages.value.push({
            id: response.message_id,
            conversation_id: conversationId.value,
            sender: currentUsername.value,
            content: newMessage.value,
            timestamp: new Date()
        })
        newMessage.value = ''
        
        // Scroll to bottom after new message
        await nextTick(() => {
            const container = document.querySelector('.messages-area')
            if (container) {
                container.scrollTop = container.scrollHeight
            }
        })
    } catch (err) {
        error.value = 'Failed to send message'
        console.error('Error:', err)
    }
}

const deleteMessage = async (messageId) => {
    if (!confirm('Are you sure you want to delete this message?')) return
    
    try {
        await api.deleteMessage(conversationId.value, messageId)
        // Remove message from local state
        messages.value = messages.value.filter(m => m.message_id !== messageId)
    } catch (err) {
        error.value = 'Failed to delete message'
        console.error('Error:', err)
    }
}

const showForwardModal = ref(false)
const messageToForward = ref(null)
const conversations = ref([])

// Get conversation list
const fetchConversations = async () => {
  try {
    const response = await api.getConversations(currentUsername.value)
    conversations.value = response
  } catch (err) {
    console.error('Error fetching conversations:', err)
  }
}

const forwardMessage = async (message) => {
  messageToForward.value = message
  await fetchConversations()
  showForwardModal.value = true
}

const getConversationName = (conv) => {
  return conv.participants
    .filter(p => p !== currentUsername.value)
    .join(', ')
}

const confirmForward = async (targetConversationId) => {
  try {
    await api.forwardMessage(targetConversationId, messageToForward.value.message_id)
    showForwardModal.value = false
    messageToForward.value = null
    
    // If forwarding to current conversation, refresh messages
    if (targetConversationId === conversationId.value) {
      await fetchMessages()  // Refresh the messages list
    }
  } catch (err) {
    console.error('Error forwarding message:', err)
  }
}

// Revert back to original reactions
const reactions = [':)', ':(', ':D', ':P', '<3']

// Add click away event listener
const closeReactionModal = (event) => {
  // Check if click is outside the reaction modal
  if (showReactionModal.value && 
      !event.target.closest('.reaction-modal') && 
      !event.target.closest('.react-btn')) {
    showReactionModal.value = false
    selectedMessage.value = null
  }
}

// Add method to delete reaction
const handleReactionDelete = (messageId, reaction) => {
  console.log('Button clicked!', {
    messageId,
    reaction,
    currentUsername: currentUsername.value
  })
  deleteReaction(messageId)
}

const deleteReaction = async (messageId) => {
  console.log('Attempting to delete reaction:', {
    messageId,
    conversationId: conversationId.value
  })
  try {
    await api.deleteReaction(messageId, conversationId.value)
    console.log('Reaction deleted successfully')
    await fetchMessages()
  } catch (err) {
    console.error('Error deleting reaction:', err)
  }
}

const showReactionModal = ref(false)
const selectedMessage = ref(null)

const showReactions = (msg) => {
  selectedMessage.value = msg.message_id
  showReactionModal.value = true
}

const addReaction = async (messageId, emoji) => {
  const container = document.querySelector('.messages-area')
  const scrollPosition = container ? container.scrollTop : 0
  
  try {
    await api.addReaction(messageId, emoji, conversationId.value)
    showReactionModal.value = false
    selectedMessage.value = null
    
    // Update messages but maintain scroll position
    await fetchMessages()
    await nextTick(() => {
      if (container) {
        container.scrollTop = scrollPosition
      }
    })
  } catch (err) {
    console.error('Error adding reaction:', err)
  }
}

const handleReactionClick = async (messageId, reaction) => {
  const container = document.querySelector('.messages-area')
  const scrollPosition = container ? container.scrollTop : 0
  
  try {
    await api.deleteReaction(messageId, conversationId.value)
    
    // Update messages but maintain scroll position
    await fetchMessages()
    await nextTick(() => {
      if (container) {
        container.scrollTop = scrollPosition
      }
    })
  } catch (err) {
    console.error('Error deleting reaction:', err)
  }
}

const getModalStyle = (msg) => {
  const messageEl = document.querySelector(`[data-message-id="${msg.message_id}"]`)
  if (!messageEl) return {}

  const rect = messageEl.getBoundingClientRect()
  const chatContainer = messageEl.closest('.chat-container')
  if (!chatContainer) return {}

  const containerRect = chatContainer.getBoundingClientRect()
  const topSpace = rect.top - containerRect.top
  const bottomSpace = containerRect.bottom - rect.bottom

  // If there's more space at the bottom, show below
  if (bottomSpace > topSpace) {
    return {
      top: '100%',
      bottom: 'auto',
      marginTop: '8px'
    }
  }
  // Otherwise show above
  return {
    bottom: '100%',
    top: 'auto',
    marginBottom: '8px'
  }
}

onMounted(() => {
    fetchMessages()
    document.addEventListener('click', closeReactionModal)
    // Add this debug log
    console.log('Current username:', currentUsername.value)
})

onUnmounted(() => {
    // Clean up listener
    document.removeEventListener('click', closeReactionModal)
})
</script>

<template>
  <MainLayout>
    <div class="chat-container">
      <!-- Chat Header -->
      <div class="chat-header">
        <button @click="router.push('/home')" class="back-btn">
          ← Back
        </button>
        <h2>Chat with {{ otherParticipant }}</h2>
      </div>

      <!-- Messages Area -->
      <div class="messages-area" ref="messagesContainer">
        <div v-if="loading" class="loading">
          Loading messages...
        </div>
        
        <div v-else-if="error" class="error-message">
          {{ error }}
        </div>
        
        <template v-else>
          <div v-for="msg in messages" 
               :key="msg.message_id" 
               :class="['message', msg.sender === currentUsername ? 'sent' : 'received']"
               style="position: relative;">
            <div class="message-sender">{{ msg.sender }}</div>
            <div class="message-content">
              {{ msg.content }}
              <!-- Show reactions -->
              <div v-if="msg.reactions && msg.reactions.length > 0" class="message-reaction">
                <span class="reaction-label">Reaction:</span>
                <template v-for="reaction in msg.reactions" :key="`${msg.message_id}-${reaction.user_id}`">
                  <div 
                    class="reaction-emoji"
                    @click="handleReactionClick(msg.message_id, reaction)"
                  >
                    {{ reaction.reaction }}
                  </div>
                </template>
              </div>
              <div class="message-actions">
                <button v-if="msg.sender === currentUsername" 
                        class="action-btn delete-btn"
                        @click="deleteMessage(msg.message_id)">
                  🗑️
                </button>
                <button class="action-btn forward-btn"
                        @click="forwardMessage(msg)">
                  ➡️
                </button>
                <button class="action-btn react-btn"
                        @click="showReactions(msg)">
                  😊
                </button>
              </div>
            </div>
            <div class="message-time">{{ new Date(msg.timestamp).toLocaleTimeString() }}</div>
            <!-- Reaction Modal -->
            <div v-if="showReactionModal && selectedMessage === msg.message_id" 
                 class="reaction-modal"
                 :class="{ 'sent-modal': msg.sender === currentUsername }"
                 :style="getModalStyle(msg)">
              <div class="reaction-list">
                <button v-for="emoji in reactions"
                        :key="emoji"
                        class="reaction-btn"
                        @click="addReaction(msg.message_id, emoji)">
                  {{ emoji }}
                </button>
              </div>
            </div>
          </div>
        </template>
      </div>

      <!-- Message Input -->
      <div class="message-input">
        <input 
          v-model="newMessage"
          @keyup.enter="sendMessage"
          type="text"
          placeholder="Type a message..."
        >
        <button @click="sendMessage" :disabled="!newMessage.trim()">
          Send
        </button>
      </div>

      <!-- Forward Modal -->
      <div v-if="showForwardModal" class="modal">
        <div class="modal-content">
          <h3>Forward Message</h3>
          <div class="conversation-list">
            <div v-for="conv in conversations" 
                 :key="conv.conversation_id"
                 class="conversation-item"
                 @click="confirmForward(conv.conversation_id)">
              {{ getConversationName(conv) }}
            </div>
          </div>
          <button class="cancel-btn" @click="showForwardModal = false">Cancel</button>
        </div>
      </div>
    </div>
  </MainLayout>
</template>

<style scoped>
.chat-container {
  display: flex;
  flex-direction: column;
  height: calc(100vh - 60px);
  max-width: 800px;
  margin: 0 auto;
}

.chat-header {
  padding: 1rem;
  background: white;
  border-bottom: 1px solid #eee;
  display: flex;
  align-items: center;
  gap: 1rem;
}

.back-btn {
  padding: 0.5rem 1rem;
  border: none;
  background: #f5f5f5;
  border-radius: 4px;
  cursor: pointer;
}

.messages-area {
  flex: 1;
  overflow-y: auto;
  padding: 1rem;
  background: #f5f5f5;
  display: flex;
  flex-direction: column;
}

.messages-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.message {
  max-width: 70%;
  padding: 0.5rem 1rem;
  border-radius: 1rem;
  margin: 0.5rem 0;
  position: relative;
}

.message.sent {
  align-self: flex-end;
  background: #0d6efd;
  color: white;
  border-bottom-right-radius: 0.25rem;
  margin-left: 30%;
}

.message.received {
  align-self: flex-start;
  background: white;
  color: black;
  border-bottom-left-radius: 0.25rem;
  margin-right: 30%;
}

.message.sent .message-sender {
  text-align: right;
}

.message.sent .message-time {
  text-align: right;
}

.message-time {
  font-size: 0.75rem;
  opacity: 0.7;
  margin-top: 0.25rem;
}

.message-input {
  padding: 1rem;
  background: white;
  border-top: 1px solid #eee;
  display: flex;
  gap: 1rem;
}

.message-input input {
  flex: 1;
  padding: 0.75rem;
  border: 1px solid #ddd;
  border-radius: 4px;
}

.message-input button {
  padding: 0.75rem 1.5rem;
  background: #0d6efd;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

.message-input button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.loading, .error-message {
  text-align: center;
  padding: 2rem;
  color: #666;
}

.error-message {
  color: #dc3545;
}

.message-sender {
    font-size: 0.8rem;
    opacity: 0.7;
    margin-bottom: 0.25rem;
}

.message-content {
    position: relative;
}

.message-actions {
    position: absolute;
    top: 50%;
    transform: translateY(-50%);
    display: flex;
    flex-direction: column;
    gap: 4px;
    opacity: 0;
    transition: opacity 0.2s;
}

/* Position buttons on right for received messages */
.received .message-actions {
    right: -35px;
}

/* Position buttons on left for sent messages */
.sent .message-actions {
    left: -35px;
}

.message:hover .message-actions {
    opacity: 1;
}

.action-btn {
    background: none;
    border: none;
    cursor: pointer;
    padding: 4px;
    font-size: 16px;
}

.action-btn:hover {
    opacity: 0.7;
}

.modal {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
}

.modal-content {
  background: white;
  padding: 20px;
  border-radius: 8px;
  min-width: 300px;
}

.conversation-list {
  max-height: 300px;
  overflow-y: auto;
}

.conversation-item {
  padding: 10px;
  cursor: pointer;
  border-bottom: 1px solid #eee;
}

.conversation-item:hover {
  background: #f5f5f5;
}

.cancel-btn {
  margin-top: 10px;
  padding: 8px 16px;
  background: #f44336;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

.cancel-btn:hover {
  background: #d32f2f;
}

.message-reaction {
  margin-top: 4px;
  font-size: 0.9em;
  color: #666;
  user-select: none;
}

.reaction-label {
  opacity: 0.7;
}

.reaction-emoji {
  display: inline-block;
  padding: 2px 6px;
  margin: 0 2px;
  font-family: monospace;
  border-radius: 4px;
}

.my-reaction {
  display: inline-block;
  padding: 2px 6px;
  margin: 0 2px;
  cursor: pointer;
  background-color: rgba(0,0,0,0.1);
  border-radius: 4px;
}

.sent .my-reaction {
  background-color: rgba(255,255,255,0.2);
}

.clickable {
  cursor: pointer;
}

.sent .message-reaction {
  color: rgba(255, 255, 255, 0.8);
}

.reaction-modal {
  position: absolute;
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.1);
  padding: 8px;
  z-index: 100;
  width: max-content;
  left: 0;
}

.sent-modal {
  left: auto;
  right: 0;
}

/* For modals near the bottom of the screen */
.message:last-child .reaction-modal {
  top: auto;
  bottom: 0;
}

.reaction-list {
  display: flex;
  gap: 8px;
}

.reaction-btn {
  background: none;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-family: monospace;
  font-size: 1.1em;
  cursor: pointer;
  padding: 4px 8px;
}

.reaction-btn:hover {
  background: #f0f0f0;
}

.action-btn {
  font-family: monospace;
}

.reaction-emoji.clickable {
  cursor: pointer;
}

.reaction-emoji.clickable:hover {
  opacity: 0.7;
}

.reaction-button {
  background: none;
  border: none;
  padding: 2px 6px;
  margin: 0 2px;
  cursor: pointer;
  color: inherit;
  font-family: inherit;
  font-size: inherit;
}

.reaction-button:hover {
  opacity: 0.7;
}
</style>