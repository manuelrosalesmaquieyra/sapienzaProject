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
const isFirstLoad = ref(true)

const otherParticipant = computed(() => {
    if (!conversation.value?.participants) return ''
    return conversation.value.participants.find(p => p !== currentUsername.value) || ''
})

const scrollToBottom = () => {
    nextTick(() => {
        const container = document.querySelector('.messages-area')
        if (container) {
            container.scrollTo({
                top: container.scrollHeight,
                behavior: isFirstLoad.value ? 'auto' : 'smooth'
            })
            isFirstLoad.value = false
        }
    })
}

const fetchMessages = async () => {
    try {
        loading.value = true
        const data = await api.getConversationMessages(conversationId.value)
        messages.value = data.messages
        
        data.messages.forEach(msg => {
            if (msg.reactions && msg.reactions.length > 0) {
                console.log('Message with reactions:', {
                    message_id: msg.message_id,
                    content: msg.content,
                    reactions: msg.reactions
                })
            }
        })
        
    } catch (err) {
        error.value = 'Failed to load messages'
        console.error('Error:', err)
    } finally {
        loading.value = false
    }
}

const replyingTo = ref(null)
const imageInput = ref(null)

const handleReply = (message) => {
    replyingTo.value = message;
    document.querySelector('.message-input input').focus();
};

const cancelReply = () => {
    replyingTo.value = null;
};

const handleImageUpload = async (event) => {
    const file = event.target.files[0];
    if (!file) return;
    
    try {
        const response = await api.sendImageMessage(conversationId.value, file);
        messages.value.push({
            id: response.message_id,
            conversation_id: conversationId.value,
            sender: currentUsername.value,
            image_url: response.image_url,
            timestamp: new Date()
        });
        
        event.target.value = '';
        
        await nextTick(() => {
            const container = document.querySelector('.messages-area');
            if (container) {
                container.scrollTop = container.scrollHeight;
            }
        });
    } catch (err) {
        error.value = 'Failed to send image';
        console.error('Error:', err);
    }
};

const sendMessage = async () => {
    if (!newMessage.value.trim()) return;
    
    try {
        let response;
        if (replyingTo.value) {
            response = await api.replyToMessage(
                conversationId.value,
                replyingTo.value.message_id,
                newMessage.value
            );
        } else {
            response = await api.sendMessage(conversationId.value, newMessage.value);
        }
        
        messages.value.push({
            id: response.message_id,
            conversation_id: conversationId.value,
            sender: currentUsername.value,
            content: newMessage.value,
            reply_to_id: replyingTo.value?.message_id,
            timestamp: new Date()
        });
        
        newMessage.value = '';
        replyingTo.value = null;
        
        await nextTick(() => {
            const container = document.querySelector('.messages-area');
            if (container) {
                container.scrollTop = container.scrollHeight;
            }
        });
    } catch (err) {
        error.value = 'Failed to send message';
        console.error('Error:', err);
    }
};

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
const availableUsers = ref([])
const forwardTabActive = ref('chats')

const fetchForwardData = async () => {
    try {
        const [convsResponse, usersResponse] = await Promise.all([
            api.getConversations(currentUsername.value),
            api.getAllUsers()
        ])
        
        conversations.value = convsResponse
        availableUsers.value = usersResponse.filter(username => 
            username !== currentUsername.value
        )
        
        console.log('Available users:', availableUsers.value)
    } catch (err) {
        console.error('Error fetching forward data:', err)
        error.value = 'Failed to load users and conversations'
    }
}

const forwardMessage = async (message) => {
    messageToForward.value = message
    await fetchForwardData()
    showForwardModal.value = true
}

const forwardToUser = async (username) => {
    try {
        console.log('Starting forward to user:', username);
        
        // First create a new conversation
        const conversationResponse = await api.createConversation(username);
        console.log('Created conversation:', conversationResponse);
        
        if (!conversationResponse || !conversationResponse.conversation_id) {
            throw new Error('Failed to create conversation');
        }
        
        // Then forward the message
        const response = await api.forwardMessage(
            conversationResponse.conversation_id,
            messageToForward.value.message_id || messageToForward.value.id
        );
        console.log('Forward response:', response);
        
        showForwardModal.value = false;
        messageToForward.value = null;
        
    } catch (err) {
        console.error('Error forwarding to user:', err);
        error.value = err.message || 'Failed to forward message';
    }
};

const confirmForward = async (targetConversationId) => {
    try {
        console.log('Confirming forward:', {
            targetConversationId,
            messageToForward: messageToForward.value
        })
        
        const response = await api.forwardMessage(
            targetConversationId, 
            messageToForward.value.message_id || messageToForward.value.id
        )
        console.log('Forward response:', response)
        
        showForwardModal.value = false
        messageToForward.value = null
        
        if (targetConversationId === conversationId.value) {
            await fetchMessages()
        }
    } catch (err) {
        console.error('Error forwarding message:', err)
        error.value = 'Failed to forward message'
    }
}

const getConversationName = (conv) => {
  return conv.participants
    .filter(p => p !== currentUsername.value)
    .join(', ')
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

const showGroupMenu = ref(false)
const newGroupName = ref('')
const newGroupPhoto = ref('')
const groupPhotoInput = ref(null)

const handleUpdateGroupName = async () => {
    if (!newGroupName.value.trim()) {
        alert('Please enter a group name')
        return
    }
    
    try {
        await api.updateGroupName(route.params.conversation_id, newGroupName.value)
        await fetchConversationDetails()
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
        await fetchConversationDetails()
        newGroupPhoto.value = ''
        showGroupMenu.value = false
    } catch (error) {
        console.error('Error updating group photo:', error)
        alert('Failed to update group photo: ' + error.message)
    }
}

const handleGroupPhotoUpload = async (event) => {
    const file = event.target.files[0];
    if (!file) return;
    
    if (!file.type.startsWith('image/')) {
        alert('Please select an image file');
        return;
    }
    
    try {
        await api.updateGroupPhoto(conversationId.value, file);
        await fetchConversationDetails();
        showGroupMenu.value = false;
    } catch (error) {
        console.error('Error updating group photo:', error);
        alert('Failed to update group photo: ' + error.message);
    }
};

const handleLeaveGroup = async () => {
    if (!confirm('Are you sure you want to leave this group?')) {
        return
    }
    
    try {
        await api.leaveGroup(route.params.conversation_id)
        router.push('/home')
    } catch (error) {
        console.error('Error leaving group:', error)
        alert('Failed to leave group: ' + error.message)
    }
}

const handleImageError = (e) => {
    e.target.style.display = 'none'
}

watch(messages, () => {
    if (messages.value.length > 0) {
        scrollToBottom()
    }
}, { immediate: true })

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
          <div class="menu-container">
            <button 
                v-if="conversation?.is_group" 
                @click="showGroupMenu = !showGroupMenu"
                class="group-menu-btn"
            >
                ⚙️
            </button>
            
            <!-- Group Menu Dropdown -->
            <div v-if="showGroupMenu" class="group-menu">
                <div class="menu-item">
                    <input 
                        v-model="newGroupName" 
                        placeholder="New group name"
                        @keyup.enter="handleUpdateGroupName"
                    >
                    <button @click="handleUpdateGroupName">Update Name</button>
                </div>

                <div class="menu-item">
                    <input 
                        type="file"
                        accept="image/*"
                        @change="handleGroupPhotoUpload"
                        ref="groupPhotoInput"
                        style="display: none"
                    >
                    <button @click="$refs.groupPhotoInput.click()" class="photo-btn">
                        Change Photo
                    </button>
                </div>

                <div class="menu-item">
                    <button @click="handleLeaveGroup" class="leave-btn">
                        Leave Group
                    </button>
                </div>
            </div>
          </div>
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
                 :data-message-id="msg.message_id"
                 style="position: relative;">
              <div class="message-content">
                <div v-if="conversation?.is_group && msg.sender !== currentUsername" 
                     class="message-sender">
                  {{ msg.sender }}
                </div>
                
                <div v-if="msg.reply_to_id" class="reply-info">
                  <div class="reply-sender">
                    {{ messages.find(m => m.message_id === msg.reply_to_id)?.sender || 'Unknown' }}
                  </div>
                  <div class="replied-content">
                    {{ messages.find(m => m.message_id === msg.reply_to_id)?.content || 
                       (messages.find(m => m.message_id === msg.reply_to_id)?.image_url ? '📷 Photo' : 'Message not found') }}
                  </div>
                </div>
                <div v-if="msg.image_url" class="message-image">
                  <img :src="msg.image_url" alt="Sent image" @error="handleImageError">
                </div>
                <div v-else class="message-text">{{ msg.content }}</div>
                <div v-if="msg.reactions && msg.reactions.length > 0" class="message-reaction">
                  <template v-for="reaction in msg.reactions" :key="`${msg.message_id}-${reaction.user_id}`">
                    <div 
                      class="reaction-emoji"
                      @click="handleReactionClick(msg.message_id, reaction)"
                    >
                      {{ reaction.reaction }} by {{ reaction.user_id }}
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
                  <button class="action-btn reply-btn"
                          @click="handleReply(msg)">
                    ↩️
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
          <div v-if="replyingTo" class="reply-preview">
            <div class="reply-content">
                Replying to: {{ replyingTo.content }}
            </div>
            <button class="cancel-reply" @click="cancelReply">×</button>
          </div>
          <button class="attach-btn" @click="$refs.imageInput.click()">
            📎
          </button>
          <input 
            v-model="newMessage"
            @keyup.enter="sendMessage"
            type="text"
            :placeholder="replyingTo ? 'Type your reply...' : 'Type a message...'"
          >
          <button @click="sendMessage" :disabled="!newMessage.trim()">
            Send
          </button>
        </div>

        <div v-if="showForwardModal" class="modal">
          <div class="modal-content">
            <h3>Forward Message</h3>
            
            <!-- Tabs -->
            <div class="forward-tabs">
                <button 
                    :class="['tab-btn', { active: forwardTabActive === 'chats' }]"
                    @click="forwardTabActive = 'chats'"
                >
                    Existing Chats
                </button>
                <button 
                    :class="['tab-btn', { active: forwardTabActive === 'users' }]"
                    @click="forwardTabActive = 'users'"
                >
                    All Users
                </button>
            </div>

            <!-- Content -->
            <div v-if="forwardTabActive === 'chats'" class="conversation-list">
                <div v-for="conv in conversations" 
                     :key="conv.conversation_id"
                     class="conversation-item"
                     @click="confirmForward(conv.conversation_id)">
                    {{ getConversationName(conv) }}
                </div>
                <div v-if="conversations.length === 0" class="empty-state">
                    No existing chats
                </div>
            </div>

            <div v-else class="users-list">
                <div v-for="username in availableUsers" 
                     :key="username"
                     class="user-item"
                     @click="forwardToUser(username)">
                    {{ username }}
                </div>
                <div v-if="availableUsers.length === 0" class="empty-state">
                    No other users found
                </div>
            </div>

            <button class="cancel-btn" @click="showForwardModal = false">Cancel</button>
          </div>
        </div>

        <input
            type="file"
            ref="imageInput"
            accept="image/*"
            style="display: none"
            @change="handleImageUpload"
        >
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
    background: rgba(255, 255, 255, 0.9);
    padding: 4px;
    border-radius: 20px;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.received .message-actions {
    right: -65px;
}

.sent .message-actions {
    left: -65px;
}

.message:hover .message-actions {
    opacity: 1;
}

.action-btn {
    background: none;
    border: none;
    cursor: pointer;
    padding: 6px;
    font-size: 16px;
    border-radius: 50%;
    transition: background-color 0.2s;
}

.action-btn:hover {
    background: rgba(0, 0, 0, 0.1);
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
  padding: 4px 8px;
  margin: 0 2px;
  font-family: monospace;
  border-radius: 4px;
  background-color: rgba(0, 0, 0, 0.05);
}

.sent .reaction-emoji {
  background-color: rgba(255, 255, 255, 0.1);
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
    width: 100%;
    padding: 8px;
    background: #fff;
    color: #dc3545;
    border: 1px solid #dc3545;
    border-radius: 4px;
    cursor: pointer;
    font-weight: 500;
}

.leave-btn:hover {
    background: #dc3545;
    color: white;
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

.message-sender {
    font-size: 0.8rem;
    font-weight: 600;
    color: #0d6efd;
    margin-bottom: 4px;
}

.sent .message-sender {
    color: rgba(255, 255, 255, 0.8);
}

.reply-preview {
    display: flex;
    align-items: center;
    padding: 8px 12px;
    background: #f8f9fa;
    border-left: 4px solid #0d6efd;
    margin-bottom: 8px;
}

.reply-content {
    flex: 1;
}

.reply-content-sender {
    font-weight: 600;
    color: #0d6efd;
    margin-bottom: 2px;
}

.reply-content-text {
    color: #666;
    font-size: 0.9em;
}

.cancel-reply {
    background: none;
    border: none;
    color: #666;
    font-size: 1.2em;
    cursor: pointer;
    padding: 0 8px;
}

.cancel-reply:hover {
    color: #dc3545;
}

.reply-info {
    background: rgba(0, 0, 0, 0.04);
    border-left: 4px solid #0d6efd;
    padding: 4px 8px;
    margin-bottom: 8px;
    border-radius: 4px;
    font-size: 0.9em;
}

.sent .reply-info {
    background: rgba(255, 255, 255, 0.1);
    border-left: 4px solid rgba(255, 255, 255, 0.5);
}

.reply-sender {
    font-weight: 600;
    color: #0d6efd;
    margin-bottom: 2px;
}

.sent .reply-sender {
    color: rgba(255, 255, 255, 0.9);
}

.replied-content {
    color: #666;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.sent .replied-content {
    color: rgba(255, 255, 255, 0.8);
}

.message-image img {
    max-width: 100%;
    max-height: 300px;
    border-radius: 4px;
}

.attach-btn {
    background: none;
    border: none;
    font-size: 1.2em;
    cursor: pointer;
    padding: 8px;
}

.menu-container {
    position: relative;
}

.group-menu {
    position: absolute;
    right: 0;
    top: 100%;
    background: white;
    border: 1px solid #ddd;
    border-radius: 4px;
    padding: 10px;
    box-shadow: 0 2px 8px rgba(0,0,0,0.1);
    z-index: 1000;
    min-width: 200px;
}

.menu-item {
    display: flex;
    gap: 8px;
    margin: 8px 0;
}

.menu-item input {
    flex: 1;
    padding: 4px 8px;
    border: 1px solid #ddd;
    border-radius: 4px;
}

.menu-item button {
    padding: 4px 8px;
    background: #007bff;
    color: white;
    border: none;
    border-radius: 4px;
    cursor: pointer;
}

.menu-item button:hover {
    background: #0056b3;
}

.photo-btn {
    width: 100%;
    padding: 8px;
    background: #28a745;
    color: white;
    border: none;
    border-radius: 4px;
    cursor: pointer;
}

.photo-btn:hover {
    background: #218838;
}

.leave-btn {
    width: 100%;
    padding: 8px;
    background: #fff;
    color: #dc3545;
    border: 1px solid #dc3545;
    border-radius: 4px;
    cursor: pointer;
    font-weight: 500;
}

.leave-btn:hover {
    background: #dc3545;
    color: white;
}

.forward-tabs {
    display: flex;
    gap: 1rem;
    margin-bottom: 1rem;
    border-bottom: 1px solid #eee;
    padding-bottom: 0.5rem;
}

.forward-tabs .tab-btn {
    padding: 0.5rem 1rem;
    border: none;
    background: #f5f5f5;
    border-radius: 4px;
    cursor: pointer;
}

.forward-tabs .tab-btn.active {
    background: #0d6efd;
    color: white;
}

.users-list {
    max-height: 300px;
    overflow-y: auto;
}

.user-item {
    padding: 10px;
    cursor: pointer;
    border-bottom: 1px solid #eee;
}

.user-item:hover {
    background: #f5f5f5;
}

.empty-state {
    text-align: center;
    padding: 1rem;
    color: #666;
}
</style>