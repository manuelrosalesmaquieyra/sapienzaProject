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
  padding: 0.5rem 1rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
  height: 50px;
}

.app-header h1 {
  font-size: 1.5rem;
  margin: 0;
}

.header-controls {
  display: flex;
  gap: 0.5rem;
  align-items: center;
}

.profile-btn, .logout-btn {
  padding: 0.3rem 0.8rem;
  font-size: 0.9rem;
  border-radius: 4px;
  cursor: pointer;
}

.profile-btn {
  background: #4CAF50;
  color: white;
  border: none;
  text-decoration: none;
}

.profile-btn:hover {
  background: #45a049;
}

.logout-btn {
  background: #dc3545;
  color: white;
  border: none;
}

.logout-btn:hover {
  background: #c82333;
}

.main-content {
  flex: 1;
  padding: 1rem;
}
</style>