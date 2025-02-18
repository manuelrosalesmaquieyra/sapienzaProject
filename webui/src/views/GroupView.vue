<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import MainLayout from '../layouts/MainLayout.vue'
import { api } from '../services/api'

const router = useRouter()
const route = useRoute()
const groups = ref([])
const loading = ref(false)
const error = ref('')
const photoError = ref('')
const showCreateGroupModal = ref(false)
const showGroupSettingsModal = ref(false)
const editingGroup = ref(null)

// Form data for new group
const newGroupName = ref('')
const selectedMembers = ref([])
const availableUsers = ref([])

const currentUsername = computed(() => localStorage.getItem('username'))
const groupId = computed(() => route.params.groupId)
const selectedGroup = ref(null)

// Fetch groups
const fetchGroups = async () => {
    try {
        loading.value = true
        const response = await api.getConversations(currentUsername.value)
        groups.value = response.filter(conv => conv.is_group)
    } catch (err) {
        error.value = 'Failed to load groups'
        console.error('Error:', err)
    } finally {
        loading.value = false
    }
}

// Fetch single group details
const fetchGroupDetails = async () => {
    try {
        loading.value = true
        const response = await api.getConversation(groupId.value)
        selectedGroup.value = response
    } catch (err) {
        error.value = 'Failed to load group details'
        console.error('Error:', err)
    } finally {
        loading.value = false
    }
}

// Handle photo upload
const handlePhotoUpload = async (event) => {
    const file = event.target.files[0]
    if (!file) return

    if (!file.type.startsWith('image/')) {
        photoError.value = 'Please select an image file'
        return
    }

    if (file.size > 5 * 1024 * 1024) {
        photoError.value = 'Image size should be less than 5MB'
        return
    }

    try {
        photoError.value = ''
        // Use editingGroup's conversation_id when in settings modal
        const groupIdToUpdate = editingGroup.value ? editingGroup.value.conversation_id : groupId.value
        const response = await api.updateGroupPhoto(groupIdToUpdate, file)
        
        if (editingGroup.value) {
            editingGroup.value.photo_url = response.photo_url
            await fetchGroups() // Refresh groups list
        } else {
            selectedGroup.value.photo_url = response.photo_url
        }
    } catch (err) {
        photoError.value = 'Failed to upload group photo'
        console.error('Error uploading photo:', err)
    }
}

// Handle leave group
const handleLeaveGroup = async () => {
    if (!confirm('Are you sure you want to leave this group?')) return

    try {
        // Use editingGroup's conversation_id when in settings modal
        const groupIdToLeave = editingGroup.value ? editingGroup.value.conversation_id : groupId.value
        await api.leaveGroup(groupIdToLeave)
        
        if (editingGroup.value) {
            showGroupSettingsModal.value = false
            await fetchGroups() // Refresh groups list
        } else {
            router.push('/groups')
        }
    } catch (err) {
        error.value = 'Failed to leave group'
        console.error('Error:', err)
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

// Modify the createGroup function
const createGroup = async () => {
    if (!newGroupName.value.trim() || selectedMembers.value.length < 1) {
        error.value = 'Please enter a group name and select members'
        return
    }

    try {
        // Validate users before creating group
        const invalidUsers = await validateUsers(selectedMembers.value)
        
        if (invalidUsers.length > 0) {
            error.value = `Cannot create group: These users do not exist: ${invalidUsers.join(', ')}`
            return
        }

        await api.createGroup(
            newGroupName.value,
            [...selectedMembers.value, currentUsername.value]
        )
        
        showCreateGroupModal.value = false
        newGroupName.value = ''
        selectedMembers.value = []
        error.value = '' // Clear any previous errors
        
        await fetchGroups()
    } catch (err) {
        error.value = 'Failed to create group: ' + err.message
        console.error('Error creating group:', err)
    }
}

// Open group settings
const openGroupSettings = (event, group) => {
    event.stopPropagation() // Prevent navigation to group detail
    editingGroup.value = {
        ...group,
        newName: group.name
    }
    showGroupSettingsModal.value = true
}

// Update group name
const updateGroupName = async () => {
    if (!editingGroup.value.newName.trim()) return

    try {
        await api.updateGroupName(
            editingGroup.value.conversation_id, 
            editingGroup.value.newName
        )
        await fetchGroups() // Refresh the groups list
        showGroupSettingsModal.value = false
    } catch (err) {
        error.value = 'Failed to update group name'
        console.error('Error:', err)
    }
}

onMounted(() => {
    if (groupId.value) {
        fetchGroupDetails()
    } else {
        fetchGroups()
    }
})
</script>

<template>
    <MainLayout>
        <!-- Group List View -->
        <div v-if="!groupId" class="groups-container">
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
                         class="group-item">
                        <div class="group-content" @click="router.push(`/groups/${group.conversation_id}`)">
                            <div class="group-name">{{ group.name }}</div>
                            <div class="group-info">
                                {{ group.participants.length }} members
                            </div>
                        </div>
                        <button class="config-btn" @click="(e) => openGroupSettings(e, group)">
                            ⚙️
                        </button>
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

            <!-- Group Settings Modal -->
            <div v-if="showGroupSettingsModal" class="modal">
                <div class="modal-content">
                    <h3>Group Settings</h3>
                    
                    <!-- Group Name Update -->
                    <div class="form-group">
                        <label>Group Name:</label>
                        <input 
                            v-model="editingGroup.newName" 
                            type="text" 
                            :placeholder="editingGroup.name"
                            maxlength="30"
                            pattern="^[a-zA-Z0-9_-]+$"
                        >
                        <button @click="updateGroupName" class="update-btn">
                            Update Name
                        </button>
                    </div>

                    <!-- Group Photo Update -->
                    <div class="form-group">
                        <label>Group Photo:</label>
                        <input 
                            type="file"
                            ref="fileInput"
                            @change="handlePhotoUpload"
                            accept="image/*"
                            style="display: none"
                        >
                        <button @click="$refs.fileInput.click()" class="update-btn">
                            Update Photo
                        </button>
                        <span v-if="photoError" class="error">{{ photoError }}</span>
                    </div>

                    <!-- Leave Group -->
                    <div class="form-group">
                        <button @click="handleLeaveGroup" class="leave-btn">
                            Leave Group
                        </button>
                    </div>

                    <div class="modal-actions">
                        <button @click="showGroupSettingsModal = false" class="cancel-btn">
                            Close
                        </button>
                    </div>
                </div>
            </div>
        </div>

        <!-- Group Detail View -->
        <div v-else class="group-detail-container">
            <div class="group-header">
                <button class="back-btn" @click="router.push('/groups')">
                    <span class="arrow">←</span> Back to Groups
                </button>
                <h2>{{ selectedGroup?.name || 'Loading...' }}</h2>
            </div>

            <div v-if="loading" class="loading">
                Loading group details...
            </div>

            <div v-else-if="error" class="error-message">
                {{ error }}
            </div>

            <div v-else class="group-content">
                <!-- Group Photo Section -->
                <div class="photo-section">
                    <div class="photo-container">
                        <div v-if="!selectedGroup.photo_url" class="empty-photo">
                            <i class="photo-icon">👥</i>
                        </div>
                        <img 
                            v-else
                            :src="selectedGroup.photo_url" 
                            alt="Group photo"
                            class="group-photo"
                        >
                    </div>
                    <div class="photo-actions">
                        <input 
                            type="file"
                            ref="fileInput"
                            @change="handlePhotoUpload"
                            accept="image/*"
                            style="display: none"
                        >
                        <button @click="$refs.fileInput.click()" class="upload-btn">
                            Update Photo
                        </button>
                        <span v-if="photoError" class="error">{{ photoError }}</span>
                    </div>
                </div>

                <!-- Group Info -->
                <div class="group-info">
                    <h3>Members ({{ selectedGroup.participants?.length || 0 }})</h3>
                    <div class="members-list">
                        <div v-for="member in selectedGroup.participants" 
                             :key="member"
                             class="member-item">
                            {{ member }}
                        </div>
                    </div>
                </div>

                <!-- Group Actions -->
                <div class="group-actions">
                    <button @click="handleLeaveGroup" class="leave-btn">
                        Leave Group
                    </button>
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
    display: flex;
    justify-content: space-between;
    align-items: center;
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

.group-detail-container {
    max-width: 800px;
    margin: 0 auto;
    padding: 1rem;
}

.group-header {
    display: flex;
    align-items: center;
    gap: 2rem;
    margin-bottom: 2rem;
}

.back-btn {
    padding: 0.7rem 1.2rem;
    background: #0d6efd;
    color: white;
    border: none;
    border-radius: 4px;
    cursor: pointer;
    display: flex;
    align-items: center;
    gap: 0.5rem;
}

.arrow {
    font-size: 1.3rem;
}

.photo-section {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 1rem;
    margin-bottom: 2rem;
}

.photo-container {
    width: 150px;
    height: 150px;
    border-radius: 50%;
    overflow: hidden;
    border: 3px solid #f0f0f0;
    background: #f8f9fa;
}

.empty-photo {
    width: 100%;
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
}

.photo-icon {
    font-size: 3rem;
    color: #adb5bd;
}

.group-photo {
    width: 100%;
    height: 100%;
    object-fit: cover;
}

.photo-actions {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.5rem;
}

.upload-btn {
    padding: 0.8rem 1.2rem;
    background: #0d6efd;
    color: white;
    border: none;
    border-radius: 4px;
    cursor: pointer;
    font-weight: bold;
}

.group-actions {
    display: flex;
    justify-content: center;
    margin-top: 2rem;
}

.leave-btn {
    padding: 0.8rem 1.2rem;
    background: #dc3545;
    color: white;
    border: none;
    border-radius: 4px;
    cursor: pointer;
    font-weight: bold;
}

.config-btn {
    padding: 0.5rem;
    background: none;
    border: none;
    cursor: pointer;
    font-size: 1.2rem;
    z-index: 1;
}

.config-btn:hover {
    opacity: 0.7;
}

.update-btn {
    padding: 0.5rem 1rem;
    background: #0d6efd;
    color: white;
    border: none;
    border-radius: 4px;
    cursor: pointer;
    margin-top: 0.5rem;
}

.group-menu {
    position: absolute;
    top: 100%;
    right: 0;
    background: white;
    border: 1px solid #ddd;
    border-radius: 4px;
    padding: 1rem;
    box-shadow: 0 2px 8px rgba(0,0,0,0.1);
    z-index: 1000;
    min-width: 250px;
}

.menu-item {
    margin: 1rem 0;
}

.menu-item input {
    width: 100%;
    padding: 0.5rem;
    border: 1px solid #ddd;
    border-radius: 4px;
    margin-bottom: 0.5rem;
}

.menu-item button {
    width: 100%;
    padding: 0.5rem;
    background: #0d6efd;
    color: white;
    border: none;
    border-radius: 4px;
    cursor: pointer;
}

.menu-item button.leave-btn {
    background: #dc3545;
}

.menu-item button:hover {
    opacity: 0.9;
}
</style> 