<script setup>
import { ref, onMounted, computed, nextTick, onUnmounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import MainLayout from '../layouts/MainLayout.vue'
import { api } from '../services/api'

const route = useRoute()
const router = useRouter()
const conversation = ref(null)
const messages = ref([])
const newMessage = ref('')
const loading = ref(false)
const error = ref('')
const conversationDetails = ref({ participants: [] })
const conversationId = ref(route.params.conversation_id)
const currentUsername = computed(() => localStorage.getItem('username'))
const messagesContainer = ref(null)

const otherParticipant = computed(() => {
    if (!conversation.value?.participants) return ''
    return conversation.value.participants.find(p => p !== currentUsername.value) || ''
})

const fetchMessages = async () => {
    try {
        loading.value = true
        const data = await api.getConversationMessages(conversationId.value)
        messages.value = data.messages
        console.log('Messages with reactions:', messages.value)
    } catch (err) {
        error.value = 'Failed to load messages'
        console.error('Error:', err)
    } finally {
        loading.value = false
    }
}

const sendMessage = async () => {
    if (!newMessage.value.trim()) return
    
    try {
        const response = await api.sendMessage(conversationId.value, newMessage.value)
        messages.value.push({
            id: response.message_id,
            conversation_id: conversationId.value,
            sender: currentUsername.value,
            content: newMessage.value,
            timestamp: new Date()
        })
        newMessage.value = ''
        
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
        messages.value = messages.value.filter(m => m.message_id !== messageId)
    } catch (err) {
        error.value = 'Failed to delete message'
        console.error('Error:', err)
    }
}

const showForwardModal = ref(false)
const messageToForward = ref(null)
const conversations = ref([])

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
    
    if (targetConversationId === conversationId.value) {
      await fetchMessages()
    }
  } catch (err) {
    console.error('Error forwarding message:', err)
  }
}

const reactions = [':)', ':(', ':D', ':P', '<3']

const closeReactionModal = (event) => {
  if (showReactionModal.value && 
      !event.target.closest('.reaction-modal') && 
      !event.target.closest('.react-btn')) {
    showReactionModal.value = false
    selectedMessage.value = null
  }
}

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

  if (bottomSpace > topSpace) {
    return {
      top: '100%',
      bottom: 'auto',
      marginTop: '8px'
    }
  }
  return {
    bottom: '100%',
    top: 'auto',
    marginBottom: '8px'
  }
}

const fetchConversationDetails = async () => {
    try {
        const details = await api.getConversationDetails(route.params.conversation_id)
        conversation.value = details
        console.log('Conversation loaded:', details)
    } catch (error) {
        console.error('Error:', error)
    }
}

// Add these refs for group management
const showGroupMenu = ref(false)
const newGroupName = ref('')
const newGroupPhoto = ref('')

// Add these methods for group management
const handleUpdateGroupName = async () => {
    if (!newGroupName.value.trim()) {
        alert('Please enter a group name')
        return
    }
    
    try {
        await api.updateGroupName(route.params.conversation_id, newGroupName.value)
        // Refresh conversation details to show new name
        await fetchConversationDetails()
        // Clear input and close menu
        newGroupName.value = ''
        showGroupMenu.value = false
    } catch (error) {
        console.error('Error updating group name:', error)
        alert('Failed to update group name: ' + error.message)
    }
}

const handleUpdateGroupPhoto = async () => {
    if (!newGroupPhoto.value.trim()) {
        alert('Please enter a photo URL')
        return
    }
    
    try {
        await api.updateGroupPhoto(route.params.conversation_id, newGroupPhoto.value)
        // Refresh conversation details to show new photo
        await fetchConversationDetails()
        // Clear input and close menu
        newGroupPhoto.value = ''
        showGroupMenu.value = false
    } catch (error) {
        console.error('Error updating group photo:', error)
        alert('Failed to update group photo: ' + error.message)
    }
}

const handleLeaveGroup = async () => {
    if (!confirm('Are you sure you want to leave this group?')) {
        return
    }
    
    try {
        await api.leaveGroup(route.params.conversation_id)
        // Redirect to home after leaving
        router.push('/home')
    } catch (error) {
        console.error('Error leaving group:', error)
        alert('Failed to leave group: ' + error.message)
    }
}

// Add this function to handle image loading errors
const handleImageError = (e) => {
    e.target.style.display = 'none'
}

onMounted(() => {
    fetchConversationDetails()
    fetchMessages()
    document.addEventListener('click', closeReactionModal)
    console.log('Current username:', currentUsername.value)
})

onUnmounted(() => {
    document.removeEventListener('click', closeReactionModal)
})
</script>

<template>
  <MainLayout>
    <div class="chat-container">
      <div class="chat-sidebar">
        <button @click="router.push('/home')" class="back-btn" title="Back to Chats">
          <span class="arrow">←</span>
          <span class="text">Back</span>
        </button>
      </div>
      
      <div class="chat-main">
        <div class="chat-header">
          <div class="header-info">
            <div class="header-content">
              <img v-if="conversation?.photo_url" 
                   :src="conversation.photo_url" 
                   class="group-photo"
                   @error="handleImageError"
                   alt="Profile photo">
              <div class="text-content">
                <h2 v-if="conversation?.is_group">
                  {{ conversation.name || 'Loading...' }}
                  <p class="participants">{{ conversation?.participants?.join(', ') }}</p>
                </h2>
                <h2 v-else>
                  {{ otherParticipant }}
                </h2>
              </div>
            </div>
          </div>
          <button 
            v-if="conversation?.is_group" 
            @click="showGroupMenu = !showGroupMenu"
            class="group-menu-btn"
          >
            ⚙️
          </button>
        </div>

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
              <div class="message-content">
                {{ msg.content }}
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

        <!-- Group management menu -->
        <div v-if="showGroupMenu" class="group-menu">
            <div class="menu-item">
                <input v-model="newGroupName" 
                       placeholder="New group name"
                       @keyup.enter="handleUpdateGroupName">
                <button @click="handleUpdateGroupName">Update Name</button>
            </div>

            <div class="menu-item">
                <input v-model="newGroupPhoto" 
                       placeholder="New photo URL"
                       @keyup.enter="handleUpdateGroupPhoto">
                <button @click="handleUpdateGroupPhoto">Update Photo</button>
            </div>

            <div class="menu-item">
                <button @click="handleLeaveGroup" class="leave-btn">
                    Leave Group
                </button>
            </div>
        </div>
      </div>
    </div>
  </MainLayout>
</template>

<style scoped>
.chat-container {
  display: flex;
  min-height: calc(100vh - 50px);
  max-width: 1400px;
  margin: 0 auto;
  gap: 1rem;
  padding: 0.5rem;
  position: fixed;
  top: 50px;
  left: 0;
  right: 0;
  bottom: 0;
}

.chat-sidebar {
  width: 150px;
  padding-right: 0.5rem;
  border-right: 1px solid #eee;
}

.chat-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
  margin-right: 0.5rem;
  overflow: hidden;
}

.back-btn {
  width: auto;
  min-width: 80px;
  height: 40px;
  padding: 0.5rem 1rem;
  border: none;
  background: #f5f5f5;
  border-radius: 8px;
  cursor: pointer;
  font-size: 0.9rem;
  display: flex;
  align-items: center;
  gap: 4px;
  margin-left: 0.5rem;
}

.back-btn .arrow {
  font-size: 1.2rem;
}

.back-btn:hover {
  background: #e9ecef;
}

.chat-header {
  padding: 1rem;
  background: #f5f5f5;
  border-bottom: 1px solid #ddd;
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
}

.header-info {
  flex: 1;
}

.header-info h2 {
  margin: 0;
  padding: 0;
}

.participants {
  margin: 5px 0 0 0;
  font-size: 0.9rem;
  color: #666;
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
  padding: 0.8rem 1rem;
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

.received .message-actions {
    right: -35px;
}

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

.group-menu {
    position: absolute;
    top: 60px;
    right: 10px;
    background: white;
    border: 1px solid #ddd;
    border-radius: 4px;
    padding: 10px;
    box-shadow: 0 2px 8px rgba(0,0,0,0.1);
    z-index: 100;
}

.menu-item {
    margin: 10px 0;
}

.menu-item input {
    margin-right: 10px;
    padding: 4px 8px;
}

.leave-btn {
    color: red;
    border-color: red;
}

.group-menu-btn {
    background: none;
    border: none;
    font-size: 1.2rem;
    cursor: pointer;
    padding: 5px;
}

.group-menu-btn:hover {
    background: #eee;
    border-radius: 50%;
}

.header-content {
    display: flex;
    align-items: center;
    gap: 12px;
}

.group-photo {
    width: 40px;
    height: 40px;
    border-radius: 50%;
    object-fit: cover;
    border: 1px solid #ddd;
}

.text-content {
    flex: 1;
}
</style>