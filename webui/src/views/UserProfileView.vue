<template>
  <div class="user-profile">
    <div class="profile-header">
      <button class="back-btn" @click="router.push('/home')">
        <span class="arrow">←</span> Back to Home
      </button>
      <h2>Profile Settings</h2>
    </div>
    
    <!-- Username Update Section -->
    <div class="profile-section">
      <h3>Update Username</h3>
      <div class="input-group">
        <input 
          v-model="newUsername" 
          type="text" 
          placeholder="New username"
          pattern="^[a-zA-Z0-9_-]+$"
          minlength="3"
          maxlength="16"
        >
        <button @click="handleUsernameUpdate" :disabled="!isValidUsername">
          Update Username
        </button>
      </div>
      <span v-if="usernameError" class="error">{{ usernameError }}</span>
    </div>

    <!-- Profile Photo Section -->
    <div class="profile-section">
      <h3>Profile Photo</h3>
      <div class="photo-section">
        <div class="photo-container">
          <div v-if="!currentPhotoUrl" class="empty-photo">
            <i class="photo-icon">📷</i>
          </div>
          <img 
            v-else
            :src="currentPhotoUrl" 
            alt="Profile photo"
            class="profile-photo"
          >
        </div>
        <div class="photo-actions">
          <input 
            type="file"
            ref="fileInput"
            @change="handleFileSelect"
            accept="image/*"
            class="file-input"
            style="display: none"
          >
          <button @click="$refs.fileInput.click()" class="upload-btn">
            Upload Photo
          </button>
          <button @click="handleRemovePhoto" class="remove-photo-btn">
            Remove Photo
          </button>
        </div>
        <span v-if="photoError" class="error">{{ photoError }}</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '@/services/api'

const router = useRouter()
const currentUsername = ref(localStorage.getItem('username'))
const newUsername = ref('')
const photoUrl = ref('')
const currentPhotoUrl = ref('')
const usernameError = ref('')
const photoError = ref('')

const isValidUsername = computed(() => {
  const regex = /^[a-zA-Z0-9_-]+$/
  return newUsername.value.length >= 3 && 
         newUsername.value.length <= 16 && 
         regex.test(newUsername.value)
})

const isValidUrl = computed(() => {
  try {
    new URL(photoUrl.value)
    return true
  } catch {
    return false
  }
})

const fetchUserData = async () => {
  try {
    const username = localStorage.getItem('username')
    console.log('Fetching data for username:', username)
    if (username) {
      const userData = await api.getUserProfile(username)
      console.log('Received user data:', userData)
      currentPhotoUrl.value = userData.photo_url
      console.log('Set currentPhotoUrl to:', currentPhotoUrl.value)
    }
  } catch (err) {
    console.error('Error fetching user data:', err)
  }
}

onMounted(() => {
  fetchUserData()
})

const handleUsernameUpdate = async () => {
  if (!isValidUsername.value) {
    usernameError.value = 'Username must be 3-16 characters and contain only letters, numbers, underscores, and hyphens'
    return
  }

  try {
    usernameError.value = ''
    const response = await api.updateUsername(currentUsername.value, newUsername.value)
    
    // Update local storage and state
    localStorage.setItem('username', response.username)
    currentUsername.value = response.username
    newUsername.value = ''
    
    // Show success message
    alert('Username updated successfully!')
    
    // Optional: redirect to home
    router.push('/home')
  } catch (err) {
    // Handle specific error cases
    if (err.message.includes('already taken')) {
      usernameError.value = 'This username is already taken'
    } else if (err.message.includes('invalid')) {
      usernameError.value = 'Invalid username format'
    } else if (err.message.includes('same as current')) {
      usernameError.value = 'This is already your current username'
    } else {
      usernameError.value = err.message || 'Failed to update username. Please try again.'
    }
    console.error('Error updating username:', err)
  }
}

const handleFileSelect = async (event) => {
  const file = event.target.files[0]
  if (!file) return

  // Validate file type
  if (!file.type.startsWith('image/')) {
    photoError.value = 'Please select an image file'
    return
  }

  // Validate file size (e.g., 5MB limit)
  if (file.size > 5 * 1024 * 1024) {
    photoError.value = 'Image size should be less than 5MB'
    return
  }

  try {
    photoError.value = ''
    const response = await api.uploadProfilePhoto(currentUsername.value, file)
    currentPhotoUrl.value = response.photo_url
    await fetchUserData()
  } catch (err) {
    photoError.value = 'Failed to upload profile photo'
    console.error('Error uploading photo:', err)
  }
}

const handleRemovePhoto = async () => {
  try {
    await api.updateProfilePhoto(currentUsername.value, "")  // Send empty string
    currentPhotoUrl.value = ""
  } catch (err) {
    photoError.value = 'Failed to remove photo'
    console.error('Error removing photo:', err)
  }
}
</script>

<style scoped>
.user-profile {
  max-width: 600px;
  margin: 0 auto;
  padding: 20px;
}

.profile-header {
  display: flex;
  align-items: center;
  margin-bottom: 2rem;
  gap: 2rem;
  padding-left: 1rem;
}

.back-btn {
  padding: 0.7rem 1.2rem;
  background: #0d6efd;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 1.1rem;
  font-weight: bold;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  transition: background-color 0.2s;
}

.back-btn:hover {
  background: #0b5ed7;
}

.arrow {
  font-size: 1.3rem;
}

.profile-section {
  background: white;
  padding: 1.5rem;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  margin-bottom: 2rem;
}

.profile-section h3 {
  margin: 0 0 1rem 0;
  color: #333;
}

.photo-section {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  align-items: center;
}

.photo-container {
  width: 150px;
  height: 150px;
  border-radius: 50%;
  overflow: hidden;
  border: 3px solid #f0f0f0;
  background: #f8f9fa;  /* Light background for empty state */
  display: flex;
  align-items: center;
  justify-content: center;
}

.empty-photo {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
  background: #f8f9fa;
}

.photo-icon {
  font-size: 3rem;
  color: #adb5bd;
  opacity: 0.5;
}

.profile-photo {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.input-group {
  display: flex;
  gap: 10px;
  width: 100%;
  margin: 0.5rem 0;
}

input {
  flex: 1;
  padding: 0.8rem;
  border: 1px solid #ccc;
  border-radius: 4px;
  font-size: 1rem;
}

button {
  padding: 0.8rem 1.2rem;
  background: #0d6efd;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-weight: bold;
}

button:disabled {
  background: #cccccc;
  cursor: not-allowed;
}

.error {
  color: #dc3545;
  font-size: 0.9rem;
  margin-top: 0.5rem;
}

.remove-photo-btn {
  background: #dc3545;  /* Red color for delete action */
  margin-top: 1rem;
}

.photo-actions {
  display: flex;
  gap: 1rem;
  margin-top: 1rem;
}

.upload-btn {
  background: #0d6efd;
  color: white;
  padding: 0.8rem 1.2rem;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-weight: bold;
}

.upload-btn:hover {
  background: #0b5ed7;
}
</style> 