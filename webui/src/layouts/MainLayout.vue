<script setup>
import { useRouter } from 'vue-router'

const router = useRouter()

const handleLogout = () => {
  localStorage.removeItem('sessionId')
  router.push('/login')
}
</script>

<template>
  <div class="app-container">
    <header class="app-header">
      <h1>WASAText</h1>
      <div class="header-controls">
        <router-link to="/profile" class="profile-btn">Profile Settings</router-link>
        <button @click="handleLogout" class="logout-btn">Logout</button>
      </div>
    </header>
    
    <div class="main-content">
      <slot></slot>
    </div>
  </div>
</template>

<style scoped>
.app-container {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

.app-header {
  background: #333;
  color: white;
  padding: 1rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-controls {
  display: flex;
  gap: 1rem;
  align-items: center;
}

.profile-btn {
  padding: 0.5rem 1rem;
  background: #4CAF50;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  text-decoration: none;
}

.profile-btn:hover {
  background: #45a049;
}

.logout-btn {
  padding: 0.5rem 1rem;
  background: #dc3545;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

.logout-btn:hover {
  background: #c82333;
}

.main-content {
  flex: 1;
  padding: 1rem;
}
</style>