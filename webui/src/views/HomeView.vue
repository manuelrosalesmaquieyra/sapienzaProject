<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import MainLayout from '../layouts/MainLayout.vue'
import { api } from '../services/api'

const router = useRouter()
const conversations = ref([])
const activeTab = ref('chats') // 'chats' or 'groups'
const loading = ref(false)
const error = ref('')
const showNewChatDialog = ref(false)
const newChatUsername = ref('')

// Add this computed property
const currentUsername = computed(() => localStorage.getItem('username'))

// Fetch conversations
const fetchData = async () => {
    loading.value = true
    error.value = ''
    try {
        // Get username from localStorage (we need to store this during login)
        const username = localStorage.getItem('username')
        if (!username) {
            throw new Error('Username not found')
        }

        const data = await api.getConversations(username)
        conversations.value = data
        console.log('Conversations data:', data)
        console.log('Current username:', currentUsername.value)
    } catch (err) {
        error.value = 'Failed to load conversations'
        console.error('Error:', err)
    } finally {
        loading.value = false
    }
}

// Initial load
onMounted(() => {
    const sessionId = localStorage.getItem('sessionId')
    if (!sessionId) {
        router.push('/login')
        return
    }
    fetchData()
})

const handleNewChat = async () => {
    try {
        console.log('Creating chat with:', newChatUsername.value)
        const result = await api.createConversation(newChatUsername.value)
        console.log('Chat created:', result)
        showNewChatDialog.value = false
        newChatUsername.value = ''
        await fetchData()
    } catch (err) {
        console.error('Error creating chat:', err)
        error.value = 'Failed to create conversation: ' + err.message
    }
}
</script>

<template>
  <MainLayout>
    <div class="home-container">
      <!-- Tabs -->
      <div class="tabs">
        <button 
          :class="['tab-btn', { active: activeTab === 'chats' }]"
          @click="activeTab = 'chats'"
        >
          Chats
        </button>
        <button 
          :class="['tab-btn', { active: activeTab === 'groups' }]"
          @click="activeTab = 'groups'"
        >
          Groups
        </button>
      </div>

      <!-- Loading and Error states -->
      <div v-if="loading" class="loading">
        Loading...
      </div>

      <div v-if="error" class="error-message">
        {{ error }}
      </div>

      <!-- Conversations List -->
      <div class="conversations-list">
        <div v-if="conversations.length === 0" class="empty-state">
          No {{ activeTab }} yet
        </div>
        <div 
          v-else
          v-for="conv in conversations" 
          :key="conv.conversation_id"
          class="conversation-item"
          @click="router.push({
            name: 'chat',
            params: { 
              conversation_id: conv.conversation_id,
              conversation: conv  // Pass the full conversation object
            }
          })"
        >
          <!-- Placeholder for conversation -->
          <div class="conv-avatar"></div>
          <div class="conv-info">
            <h3>
              {{ conv.participants.find(p => p !== currentUsername) || 'Unknown User' }}
            </h3>
            <p>{{ conv.last_message || 'No messages yet' }}</p>
          </div>
        </div>
      </div>

      <!-- New Chat/Group Button -->
      <div v-if="showNewChatDialog" class="dialog-overlay">
        <div class="dialog">
          <h3>Start New Chat</h3>
          <input 
            v-model="newChatUsername"
            placeholder="Enter username"
            type="text"
          >
          <div class="dialog-buttons">
            <button @click="showNewChatDialog = false">Cancel</button>
            <button @click="handleNewChat">Start Chat</button>
          </div>
        </div>
      </div>

      <button 
        class="new-chat-btn"
        @click="showNewChatDialog = true"
      >
        New {{ activeTab === 'chats' ? 'Chat' : 'Group' }}
      </button>
    </div>
  </MainLayout>
</template>

<style scoped>
.home-container {
  max-width: 800px;
  margin: 0 auto;
  padding: 1rem;
}

.tabs {
  display: flex;
  gap: 1rem;
  margin-bottom: 1rem;
}

.tab-btn {
  padding: 0.5rem 1rem;
  border: none;
  background: #f5f5f5;
  border-radius: 4px;
  cursor: pointer;
}

.tab-btn.active {
  background: #0d6efd;
  color: white;
}

.conversations-list {
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  min-height: 400px;
}

.empty-state {
  padding: 2rem;
  text-align: center;
  color: #666;
}

.conversation-item {
  display: flex;
  padding: 1rem;
  border-bottom: 1px solid #eee;
  cursor: pointer;
}

.conversation-item:hover {
  background: #f5f5f5;
}

.conv-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: #ddd;
  margin-right: 1rem;
}

.conv-info h3 {
  margin: 0;
  font-size: 1rem;
}

.conv-info p {
  margin: 0.25rem 0 0;
  font-size: 0.875rem;
  color: #666;
}

.new-chat-btn {
  position: fixed;
  bottom: 2rem;
  right: 2rem;
  padding: 1rem 2rem;
  background: #0d6efd;
  color: white;
  border: none;
  border-radius: 50px;
  cursor: pointer;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
}

.new-chat-btn:hover {
  background: #0b5ed7;
}

.loading {
    text-align: center;
    padding: 2rem;
    color: #666;
}

.error-message {
    color: #dc3545;
    text-align: center;
    padding: 1rem;
    margin-bottom: 1rem;
}

.dialog-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.dialog {
  background: white;
  padding: 2rem;
  border-radius: 8px;
  min-width: 300px;
}

.dialog input {
  width: 100%;
  padding: 0.5rem;
  margin: 1rem 0;
}

.dialog-buttons {
  display: flex;
  gap: 1rem;
  justify-content: flex-end;
}

.dialog-buttons button {
  padding: 0.5rem 1rem;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

.dialog-buttons button:first-child {
  background: #f5f5f5;
}

.dialog-buttons button:last-child {
  background: #0d6efd;
  color: white;
}
</style>