<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import MainLayout from '../layouts/MainLayout.vue'
import { api } from '../services/api'

const router = useRouter()
const groups = ref([])
const loading = ref(false)
const error = ref('')
const showCreateGroupModal = ref(false)

// Form data for new group
const newGroupName = ref('')
const selectedMembers = ref([])
const availableUsers = ref([]) // We'll need to add an API endpoint to get users

const currentUsername = computed(() => localStorage.getItem('username'))

// Fetch groups
const fetchGroups = async () => {
    try {
        loading.value = true
        // We'll need to implement this API endpoint
        const response = await api.getConversations(currentUsername.value)
        // Filter for group conversations
        groups.value = response.filter(conv => conv.is_group)
    } catch (err) {
        error.value = 'Failed to load groups'
        console.error('Error:', err)
    } finally {
        loading.value = false
    }
}

// Create new group
const createGroup = async () => {
    if (!newGroupName.value.trim() || selectedMembers.value.length < 1) {
        error.value = 'Please enter a group name and select members'
        return
    }

    try {
        const response = await api.createGroup(
            newGroupName.value,
            [...selectedMembers.value, currentUsername.value]
        )
        
        showCreateGroupModal.value = false
        newGroupName.value = ''
        selectedMembers.value = []
        
        // Refresh groups list
        await fetchGroups()
    } catch (err) {
        error.value = 'Failed to create group'
        console.error('Error:', err)
    }
}

onMounted(() => {
    fetchGroups()
})
</script>

<template>
    <MainLayout>
        <div class="groups-container">
            <!-- Groups Header -->
            <div class="groups-header">
                <h2>Groups</h2>
                <button @click="showCreateGroupModal = true" class="create-group-btn">
                    Create New Group
                </button>
            </div>

            <!-- Groups List -->
            <div class="groups-list">
                <div v-if="loading" class="loading">
                    Loading groups...
                </div>
                
                <div v-else-if="error" class="error-message">
                    {{ error }}
                </div>
                
                <template v-else>
                    <div v-for="group in groups" 
                         :key="group.conversation_id"
                         class="group-item"
                         @click="router.push(`/groups/${group.conversation_id}`)">
                        <div class="group-name">{{ group.name }}</div>
                        <div class="group-info">
                            {{ group.participants.length }} members
                        </div>
                    </div>
                </template>
            </div>

            <!-- Create Group Modal -->
            <div v-if="showCreateGroupModal" class="modal">
                <div class="modal-content">
                    <h3>Create New Group</h3>
                    <div class="form-group">
                        <label>Group Name:</label>
                        <input v-model="newGroupName" 
                               type="text" 
                               placeholder="Enter group name"
                               maxlength="30"
                               pattern="^[a-zA-Z0-9_-]+$">
                    </div>
                    
                    <div class="form-group">
                        <label>Select Members:</label>
                        <div class="members-list">
                            <div v-for="user in availableUsers" 
                                 :key="user"
                                 class="member-item">
                                <input type="checkbox" 
                                       :value="user"
                                       v-model="selectedMembers">
                                <span>{{ user }}</span>
                            </div>
                        </div>
                    </div>

                    <div class="modal-actions">
                        <button @click="createGroup" 
                                :disabled="!newGroupName.trim() || selectedMembers.length < 1"
                                class="create-btn">
                            Create
                        </button>
                        <button @click="showCreateGroupModal = false" 
                                class="cancel-btn">
                            Cancel
                        </button>
                    </div>
                </div>
            </div>
        </div>
    </MainLayout>
</template>

<style scoped>
.groups-container {
    display: flex;
    flex-direction: column;
    height: calc(100vh - 60px);
    max-width: 800px;
    margin: 0 auto;
    padding: 1rem;
}

.groups-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1rem;
}

.create-group-btn {
    padding: 0.5rem 1rem;
    background: #0d6efd;
    color: white;
    border: none;
    border-radius: 4px;
    cursor: pointer;
}

.groups-list {
    flex: 1;
    overflow-y: auto;
}

.group-item {
    padding: 1rem;
    background: white;
    border-radius: 8px;
    margin-bottom: 0.5rem;
    cursor: pointer;
    transition: background-color 0.2s;
}

.group-item:hover {
    background: #f8f9fa;
}

.group-name {
    font-weight: bold;
    margin-bottom: 0.25rem;
}

.group-info {
    font-size: 0.875rem;
    color: #6c757d;
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
    padding: 2rem;
    border-radius: 8px;
    width: 90%;
    max-width: 500px;
}

.form-group {
    margin-bottom: 1rem;
}

.form-group label {
    display: block;
    margin-bottom: 0.5rem;
}

.form-group input[type="text"] {
    width: 100%;
    padding: 0.5rem;
    border: 1px solid #ddd;
    border-radius: 4px;
}

.members-list {
    max-height: 200px;
    overflow-y: auto;
    border: 1px solid #ddd;
    border-radius: 4px;
    padding: 0.5rem;
}

.member-item {
    padding: 0.5rem;
    display: flex;
    align-items: center;
    gap: 0.5rem;
}

.modal-actions {
    display: flex;
    justify-content: flex-end;
    gap: 1rem;
    margin-top: 1rem;
}

.create-btn {
    padding: 0.5rem 1rem;
    background: #0d6efd;
    color: white;
    border: none;
    border-radius: 4px;
    cursor: pointer;
}

.create-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
}

.cancel-btn {
    padding: 0.5rem 1rem;
    background: #dc3545;
    color: white;
    border: none;
    border-radius: 4px;
    cursor: pointer;
}

.loading, .error-message {
    text-align: center;
    padding: 2rem;
    color: #666;
}

.error-message {
    color: #dc3545;
}
</style> 