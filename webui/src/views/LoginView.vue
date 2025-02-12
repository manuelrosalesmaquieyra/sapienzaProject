<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import LoginLayout from '../layouts/LoginLayout.vue'

// Get base URL from environment or window location
const API_URL = import.meta.env.VITE_API_BASE_URL || `${window.location.protocol}//${window.location.hostname}:3000`

// Input field for username
const username = ref('')
// Error message state
const error = ref('')
const router = useRouter()

// Handle login form submission
const handleLogin = async () => {
  try {
    // Basic validation
    if (username.value.length < 3 || username.value.length > 16) {
      error.value = 'Username must be between 3 and 16 characters'
      return
    }

    // Call login API with dynamic URL
    const response = await fetch(`${API_URL}/session`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ name: username.value }),
    })

    if (!response.ok) {
      throw new Error('Login failed')
    }

    // Get session identifier
    const data = await response.json()
    
    // Store session identifier and username in localStorage
    localStorage.setItem('sessionId', data.session_id)
    localStorage.setItem('username', data.username)
    
    // Redirect to home page
    router.push('/home')
  } catch (err) {
    error.value = 'Login failed. Please try again.'
    console.error('Login error:', err)
  }
}
</script>

<template>
  <LoginLayout>
    <form @submit.prevent="handleLogin" class="login-form">
      <!-- Login title -->
      <h2>Login to WASAText</h2>
      
      <!-- Username input -->
      <div class="form-group">
        <label for="username">Username:</label>
        <input
          id="username"
          v-model="username"
          type="text"
          placeholder="Enter your username"
          required
          pattern="[a-zA-Z0-9_-]+"
          minlength="3"
          maxlength="16"
        >
      </div>

      <!-- Error message display -->
      <div v-if="error" class="error-message">
        {{ error }}
      </div>

      <!-- Submit button -->
      <button type="submit" class="login-button">
        Login
      </button>
    </form>
  </LoginLayout>
</template>

<style scoped>
.login-form {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.form-group label {
  font-weight: 500;
}

.form-group input {
  padding: 0.75rem;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 1rem;
}

.error-message {
  color: #dc3545;
  font-size: 0.875rem;
}

.login-button {
  padding: 0.75rem;
  background: #0d6efd;
  color: white;
  border: none;
  border-radius: 4px;
  font-size: 1rem;
  cursor: pointer;
  transition: background-color 0.2s;
}

.login-button:hover {
  background: #0b5ed7;
}
</style>