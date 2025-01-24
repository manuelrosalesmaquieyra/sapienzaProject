<script setup>
import { ref, onMounted, computed, nextTick } from 'vue'
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
        // Store current scroll position
        const scrollPos = messagesContainer.value?.scrollTop
        const isScrolledToBottom = messagesContainer.value?.scrollHeight - messagesContainer.value?.scrollTop === messagesContainer.value?.clientHeight
        
        messages.value = data.messages
        
        // Wait for DOM update
        await nextTick(() => {
            if (isScrolledToBottom) {
                // If was at bottom, scroll to new bottom
                messagesContainer.value?.scrollTo(0, messagesContainer.value.scrollHeight)
            } else if (scrollPos) {
                // Otherwise restore previous position
                messagesContainer.value?.scrollTo(0, scrollPos)
            }
        })
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

const reactions = ['👍', '❤️', '😂', '😮', '😢', '😡']
const showReactionModal = ref(false)
const selectedMessage = ref(null)

const showReactions = (message) => {
  selectedMessage.value = message.message_id
  showReactionModal.value = true
}

const addReaction = async (messageId, emoji) => {
  try {
    // Verify emoji length
    if (emoji.length < 1 || emoji.length > 5) {
      console.error('Emoji must be between 1 and 5 characters')
      return
    }
    
    await api.addReaction(messageId, emoji, conversationId.value)
    showReactionModal.value = false
    selectedMessage.value = null
    await fetchMessages()
  } catch (err) {
    console.error('Error adding reaction:', err)
  }
}

onMounted(() => {
    fetchMessages()
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
               :class="['message', msg.sender === currentUsername ? 'sent' : 'received']">
            <div class="message-sender">{{ msg.sender }}</div>
            <div class="message-content">
              {{ msg.content }}
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
                  😀
                </button>
              </div>
              <!-- Reaction Modal -->
              <div v-if="showReactionModal && selectedMessage === msg.message_id" 
                   class="reaction-modal">
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
            <div class="message-time">{{ new Date(msg.timestamp).toLocaleTimeString() }}</div>
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

.reaction-modal {
  position: absolute;
  bottom: 100%;
  left: 50%;
  transform: translateX(-50%);
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.1);
  padding: 8px;
  margin-bottom: 8px;
  z-index: 100;
}

.reaction-list {
  display: flex;
  gap: 4px;
}

.reaction-btn {
  background: none;
  border: none;
  cursor: pointer;
  padding: 4px;
  font-size: 20px;
  transition: transform 0.2s;
}

.reaction-btn:hover {
  transform: scale(1.2);
}
</style>