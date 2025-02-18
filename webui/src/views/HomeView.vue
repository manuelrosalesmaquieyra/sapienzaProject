<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import MainLayout from '../layouts/MainLayout.vue'
import { api } from '../services/api'

const router = useRouter()
const conversations = ref([])
const groups = ref([])  // Add this for groups
const activeTab = ref('chats') // 'chats' or 'groups'
const loading = ref(false)
const error = ref('')
const showNewChatDialog = ref(false)
const newChatUsername = ref('')
const showNewGroupDialog = ref(false)
const newGroupName = ref('')
const newGroupMembers = ref('')  // Will be split into array

// Add this computed property
const currentUsername = computed(() => localStorage.getItem('username'))

// Fetch conversations
const fetchData = async () => {
    try {
        const username = localStorage.getItem('username')
        if (!username) {
            throw new Error('Username not found')
        }

        // Keep existing data if fetch fails
        const conversationsData = await api.getConversations(username)
        
        if (conversationsData) {
            // Only update lists if we got valid data
            conversations.value = conversationsData.filter(conv => !conv.is_group)
            groups.value = conversationsData.filter(conv => conv.is_group)
        }
        
        error.value = '' // Clear any existing errors
    } catch (err) {
        error.value = 'Failed to load conversations and groups'
        console.error('Error:', err)
        // Don't clear existing conversations/groups on error
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
    const targetUsername = newChatUsername.value.trim();
    
    try {
        if (!targetUsername) {
            error.value = 'Please enter a username';
            setTimeout(() => error.value = '', 2000); // Clear after 3 seconds
            showNewChatDialog.value = false;
            return;
        }

        if (targetUsername === currentUsername.value) {
            error.value = 'Cannot create chat with yourself';
            setTimeout(() => error.value = '', 2000); // Clear after 3 seconds
            showNewChatDialog.value = false;
            return;
        }

        const result = await api.createConversation(targetUsername);
        
        if (result) {
            error.value = '';
            await fetchData();
        }
    } catch (err) {
        error.value = err.message;
        setTimeout(() => error.value = '', 2000); // Clear after 3 seconds
    } finally {
        showNewChatDialog.value = false;
        newChatUsername.value = '';
    }
}

// Add this new function after your ref declarations
const validateUsers = async (usernames) => {
    try {
        const results = await Promise.all(
            usernames.map(async username => {
                const exists = await api.checkUserExists(username)
                return { username, exists }
            })
        )
        
        const invalidUsers = results
            .filter(result => !result.exists)
            .map(result => result.username)
            
        return invalidUsers
    } catch (err) {
        console.error('Error validating users:', err)
        throw err
    }
}

// Modify the handleNewGroup function
const handleNewGroup = async () => {
    try {
        // Convert comma-separated members string to array and trim whitespace
        const membersArray = newGroupMembers.value
            .split(',')
            .map(member => member.trim())
            .filter(member => member.length > 0)

        // Validate basic requirements
        if (newGroupName.value.length < 3 || newGroupName.value.length > 30) {
            error.value = 'Group name must be between 3 and 30 characters'
            showNewGroupDialog.value = false // Close dialog
            setTimeout(() => error.value = '', 2000) // Clear error after 2 seconds
            return
        }
        if (membersArray.length < 2) {
            error.value = 'Please add at least 2 members'
            showNewGroupDialog.value = false
            setTimeout(() => error.value = '', 2000)
            return
        }
        if (membersArray.length > 50) {
            error.value = 'Maximum 50 members allowed'
            showNewGroupDialog.value = false
            setTimeout(() => error.value = '', 2000)
            return
        }

        // Validate users exist
        const invalidUsers = await validateUsers(membersArray)
        if (invalidUsers.length > 0) {
            error.value = `Cannot create group: These users do not exist: ${invalidUsers.join(', ')}`
            showNewGroupDialog.value = false
            setTimeout(() => error.value = '', 2000)
            return
        }

        console.log('Creating group:', { name: newGroupName.value, members: membersArray })
        const result = await api.createGroup(newGroupName.value, membersArray)
        console.log('Group created:', result)
        
        // Reset form and close dialog
        showNewGroupDialog.value = false
        newGroupName.value = ''
        newGroupMembers.value = ''
        
        // Refresh conversations list
        await fetchData()
    } catch (err) {
        console.error('Error creating group:', err)
        error.value = 'Failed to create group: ' + err.message
        showNewGroupDialog.value = false // Close dialog on error
        setTimeout(() => error.value = '', 2000) // Clear error after 2 seconds
    }
}

const handleNewButton = () => {
    if (activeTab.value === 'chats') {
        showNewChatDialog.value = true
        showNewGroupDialog.value = false
    } else {
        showNewChatDialog.value = false
        showNewGroupDialog.value = true
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

      <div v-if="error" class="error-banner">
        {{ error }}
      </div>

      <!-- Modified Conversations/Groups List -->
      <div class="conversations-list">
        <div v-if="loading" class="loading">
          Loading...
        </div>
        <div v-else>
          <!-- Show empty state based on active tab -->
          <div v-if="(activeTab === 'chats' && conversations.length === 0) || 
                   (activeTab === 'groups' && groups.length === 0)" 
               class="empty-state">
            No {{ activeTab }} yet
          </div>
          
          <!-- Show conversations or groups based on active tab -->
          <div v-else>
            <div v-for="item in (activeTab === 'chats' ? conversations : groups)" 
                 :key="item.conversation_id"
                 class="conversation-item"
                 @click="router.push({
                    name: 'chat',
                    params: { 
                        conversation_id: item.conversation_id,
                        conversation: item
                    }
                 })"
            >
              <div class="conv-avatar">
                <img v-if="item.photo_url" 
                     :src="item.photo_url" 
                     alt="Profile photo"
                     class="avatar-img"
                >
                <div v-else class="avatar-placeholder">
                  {{ activeTab === 'chats' ? '👤' : '👥' }}
                </div>
              </div>
              <div class="conv-info">
                <h3>
                  {{ activeTab === 'chats' 
                      ? item.participants.find(p => p !== currentUsername) 
                      : item.name }}
                </h3>
                <p>
                  {{ item.last_message_is_reply ? '↩️ ' : '' }}
                  {{ item.last_message || 'No messages yet' }}
                </p>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- New Chat/Group Button -->
      <button 
        class="new-chat-btn"
        @click="handleNewButton"
      >
        {{ activeTab === 'chats' ? 'New Chat' : 'New Group' }}
      </button>

      <!-- Existing chat dialog -->
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

      <!-- Group dialog -->
      <div v-if="showNewGroupDialog" class="dialog-overlay">
        <div class="dialog">
            <h3>Create New Group</h3>
            <div class="form-group">
                <label for="groupName">Group Name:</label>
                <input 
                    id="groupName"
                    v-model="newGroupName"
                    placeholder="Enter group name (3-30 characters)"
                    type="text"
                    pattern="[a-zA-Z0-9_-]+"
                    minlength="3"
                    maxlength="30"
                >
            </div>
            <div class="form-group">
                <label for="groupMembers">Members:</label>
                <input 
                    id="groupMembers"
                    v-model="newGroupMembers"
                    placeholder="Enter usernames (comma-separated)"
                    type="text"
                >
                <small class="help-text">Add at least 2 members, separated by commas</small>
            </div>
            <div class="dialog-buttons">
                <button @click="showNewGroupDialog = false">Cancel</button>
                <button 
                    @click="handleNewGroup"
                    :disabled="!newGroupName || !newGroupMembers"
                >
                    Create Group
                </button>
            </div>
        </div>
      </div>
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
  overflow: hidden;
  background: #f0f0f0;
  display: flex;
  align-items: center;
  justify-content: center;
}

.avatar-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.avatar-placeholder {
  font-size: 1.5rem;
  color: #666;
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

.form-group {
    margin-bottom: 1rem;
}

.form-group label {
    display: block;
    margin-bottom: 0.5rem;
    font-weight: 500;
}

.help-text {
    display: block;
    font-size: 0.8rem;
    color: #666;
    margin-top: 0.25rem;
}

.error-banner {
    position: fixed;
    top: 60px;
    left: 50%;
    transform: translateX(-50%);
    background: #f8d7da;
    color: #721c24;
    padding: 10px 20px;
    border-radius: 4px;
    z-index: 1000;
    box-shadow: 0 2px 4px rgba(0,0,0,0.1);
    transition: opacity 0.5s ease-in-out;
    opacity: 1;
}

.error-banner.fade-out {
    opacity: 0;
}
</style>